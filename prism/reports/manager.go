package reports

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"prism/prism/api"
	"prism/prism/schema"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	StaleReportThreshold time.Duration = time.Hour * 24 * 14
)

var (
	EarliestReportDate time.Time = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
)

var (
	ErrReportAccessFailed     = errors.New("report access failed")
	ErrReportCreationFailed   = errors.New("report creation failed")
	ErrReportNotFound         = errors.New("report not found")
	ErrUserCannotAccessReport = errors.New("user cannot access report")
)

type ReportManager struct {
	db                   *gorm.DB
	staleReportThreshold time.Duration
}

func NewManager(db *gorm.DB, staleReportThreshold time.Duration) *ReportManager {
	return &ReportManager{db: db, staleReportThreshold: staleReportThreshold}
}

func (r *ReportManager) ListAuthorReports(userId uuid.UUID) ([]api.Report, error) {
	var reports []schema.UserAuthorReport

	if err := r.db.Preload("Report").Order("last_accessed_at DESC").Find(&reports, "user_id = ?", userId).Error; err != nil {
		slog.Error("error finding list of reports ")
		return nil, ErrReportAccessFailed
	}

	results := make([]api.Report, 0, len(reports))
	for _, report := range reports {
		res, err := convertReport(report)
		if err != nil {
			return nil, ErrReportAccessFailed
		}
		results = append(results, res)
	}

	return results, nil
}

// This function is only called from GetAuthorReport or CreateAuthorReport. The university reports
// use a different methods that can check all the author reports associated with the university report
// at once.
func (r *ReportManager) queueAuthorReportUpdateIfNeeded(txn *gorm.DB, report *schema.AuthorReport) error {
	if time.Now().UTC().Sub(report.LastUpdatedAt) > r.staleReportThreshold &&
		report.Status != schema.ReportInProgress && report.Status != schema.ReportQueued {
		updates := map[string]any{"status": schema.ReportQueued, "queued_at": time.Now().UTC(), "queued_by_user": true}
		if err := txn.Model(&report).Updates(updates).Error; err != nil {
			slog.Error("error queueing stale author report for update", "author_report_id", report.Id, "error", err)
			return ErrReportAccessFailed
		}
		report.Status = schema.ReportQueued
	}
	return nil
}

func createOrGetAuthorReport(txn *gorm.DB, authorId, authorName, source string, fromUserReq bool) (schema.AuthorReport, error) {
	var report schema.AuthorReport
	result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&report, "author_id = ? AND source = ?", authorId, source)
	if result.Error != nil {
		slog.Error("error checking for existing author report", "error", result.Error)
		return schema.AuthorReport{}, ErrReportCreationFailed
	}

	if result.RowsAffected == 0 {
		report = schema.AuthorReport{
			Id:            uuid.New(),
			LastUpdatedAt: EarliestReportDate,
			AuthorId:      authorId,
			AuthorName:    authorName,
			Source:        source,
			Status:        schema.ReportQueued,
			QueuedAt:      time.Now(),
			QueuedByUser:  fromUserReq,
		}

		if err := txn.Create(&report).Error; err != nil {
			slog.Error("error creating new author report", "error", err)
			return schema.AuthorReport{}, ErrReportCreationFailed
		}
	}

	return report, nil
}

func (r *ReportManager) CreateAuthorReport(licenseId, userId uuid.UUID, authorId, authorName, source string) (uuid.UUID, error) {
	var userReport schema.UserAuthorReport
	var userReportId uuid.UUID
	now := time.Now().UTC()

	err := r.db.Transaction(func(txn *gorm.DB) error {
		report, err := createOrGetAuthorReport(txn, authorId, authorName, source, true /*fromUserReq*/)
		if err != nil {
			return err
		}

		if err := r.queueAuthorReportUpdateIfNeeded(txn, &report); err != nil {
			return err
		}

		result := txn.Where("user_id = ? AND report_id = ?", userId, report.Id).First(&userReport)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Error("error finding existing user author report", "error", result.Error)
			return ErrReportCreationFailed
		}

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			userReportId = uuid.New()
			userReport = schema.UserAuthorReport{
				Id:             userReportId,
				UserId:         userId,
				LastAccessedAt: now,
				ReportId:       report.Id,
			}
			if err := txn.Create(&userReport).Error; err != nil {
				slog.Error("error creating new user author report", "error", err)
				return ErrReportCreationFailed
			}
		} else {
			userReport.LastAccessedAt = now
			if err := txn.Save(&userReport).Error; err != nil {
				slog.Error("error updating user author report", "error", err)
				return ErrReportCreationFailed
			}
			userReportId = userReport.Id
		}

		usage := schema.LicenseUsage{
			LicenseId:  licenseId,
			ReportId:   userReport.Id,
			ReportType: schema.AuthorReportType,
			UserId:     userId,
			Timestamp:  time.Now().UTC(),
		}
		if err := txn.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&usage).Error; err != nil {
			slog.Error("error logging license usage", "error", err)
			return errors.New("error updating license usage")
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return userReportId, nil
}

func (r *ReportManager) GetAuthorReport(userId, reportId uuid.UUID) (api.Report, error) {
	var report schema.UserAuthorReport

	if err := r.db.Transaction(func(txn *gorm.DB) error {
		if err := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Report").Preload("Report.Flags").
			First(&report, "id = ?", reportId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReportNotFound
			}
			slog.Error("error getting user author report", "author_report_id", reportId, "error", err)
			return ErrReportAccessFailed
		}

		if report.UserId != userId {
			return ErrUserCannotAccessReport
		}

		if err := r.queueAuthorReportUpdateIfNeeded(txn, report.Report); err != nil {
			return err
		}

		if err := txn.Model(&report).Update("last_accessed_at", time.Now().UTC()).Error; err != nil {
			slog.Error("error updating user author report last_accessed_at", "error", err)
			return ErrReportAccessFailed
		}

		return nil
	}); err != nil {
		return api.Report{}, err
	}

	slog.Info("GET AUTHOR REPORT", "report", fmt.Sprintf("%+v", report))

	return convertReport(report)
}

func (r *ReportManager) DeleteAuthorReport(userId, reportId uuid.UUID) error {
	result := r.db.Delete(&schema.UserAuthorReport{}, "id = ? AND user_id = ?", reportId, userId)
	if result.Error != nil {
		slog.Error("error deleting author report", "author_report_id", reportId, "error", result.Error)
		return ErrReportAccessFailed
	}
	if result.RowsAffected != 1 {
		return ErrReportNotFound
	}
	return nil
}

type ReportUpdateTask struct {
	Id         uuid.UUID
	AuthorId   string
	AuthorName string
	Source     string
	StartDate  time.Time
	EndDate    time.Time
}

func findNextAuthorReport(txn *gorm.DB) (*schema.AuthorReport, error) {
	var report schema.AuthorReport
	for _, queuedByUser := range [2]bool{true, false} {
		result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Limit(1).Order("queued_at ASC").
			Find(&report, "status = ? AND queued_by_user = ?", schema.ReportQueued, queuedByUser)
		if result.Error != nil {
			slog.Error("error getting next author report from queue", "error", result.Error)
			return nil, ErrReportAccessFailed
		}

		if result.RowsAffected == 1 {
			return &report, nil
		}
	}
	return nil, nil
}

func (r *ReportManager) GetNextAuthorReport() (*ReportUpdateTask, error) {
	var report *schema.AuthorReport

	err := r.db.Transaction(func(txn *gorm.DB) error {
		var err error
		report, err = findNextAuthorReport(txn)
		if err != nil {
			return err
		}

		if report != nil {
			if err := txn.Model(report).Update("status", schema.ReportInProgress).Error; err != nil {
				slog.Error("error updating author report status to in progress", "error", err)
				return ErrReportAccessFailed
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if report != nil {
		return &ReportUpdateTask{
			Id:         report.Id,
			AuthorId:   report.AuthorId,
			AuthorName: report.AuthorName,
			Source:     report.Source,
			StartDate:  report.LastUpdatedAt,
			EndDate:    time.Now().UTC(),
		}, nil
	}

	return nil, nil
}

func flagsToReportContent(flags []schema.AuthorFlag) (api.ReportContent, error) {
	content := make(api.ReportContent)

	for _, flag := range flags {
		output, err := api.EmptyFlag(flag.FlagType)
		if err != nil {
			slog.Error("error creating empty flag", "error", err)
			return nil, err
		}
		if err := json.Unmarshal(flag.Data, output); err != nil {
			return nil, fmt.Errorf("error deserializing flag: %w", err)
		}
		content[output.Type()] = append(content[output.Type()], output)
	}

	return content, nil
}

func (r *ReportManager) UpdateAuthorReport(id uuid.UUID, status string, updateTime time.Time, updateFlags []api.Flag) error {
	e := r.db.Transaction(func(txn *gorm.DB) error {
		updates := map[string]any{"status": status}
		if status == schema.ReportCompleted {
			updates["last_updated_at"] = updateTime
		}

		result := txn.Model(&schema.AuthorReport{Id: id}).Updates(updates)
		if result.Error != nil {
			slog.Error("error updating author report status", "author_report_id", id, "error", result.Error)
			return ErrReportAccessFailed
		}

		if result.RowsAffected != 1 {
			slog.Error("cannot update status of author report, report not found", "author_report_id", id, "status", status)
			return ErrReportNotFound
		}

		if len(updateFlags) == 0 {
			return nil
		}

		newFlags := make([]schema.AuthorFlag, 0)
		for _, flag := range updateFlags {
			data, err := json.Marshal(flag)
			if err != nil {
				return fmt.Errorf("error serializing flag: %w", err)
			}

			flagHash := flag.Hash()

			date, dateValid := flag.Date()

			newFlags = append(newFlags, schema.AuthorFlag{
				ReportId: id,
				FlagHash: hex.EncodeToString(flagHash[:]),
				FlagType: flag.Type(),
				Date:     sql.NullTime{Time: date, Valid: dateValid},
				Data:     data,
			})
		}

		if err := txn.Save(&newFlags).Error; err != nil {
			slog.Error("error adding new flags to author report", "author_report_id", id, "error", err)
			return ErrReportAccessFailed
		}

		return nil
	})

	slog.Info("REPORT UPDATE COMPLETED", "id", id, "status", status, "update_flags", updateFlags)

	return e
}

func convertReport(report schema.UserAuthorReport) (api.Report, error) {
	content, err := flagsToReportContent(report.Report.Flags)
	if err != nil {
		return api.Report{}, err
	}

	return api.Report{
		Id:             report.Id,
		LastAccessedAt: report.LastAccessedAt,
		AuthorId:       report.Report.AuthorId,
		AuthorName:     report.Report.AuthorName,
		Source:         report.Report.Source,
		Status:         report.Report.Status,
		Content:        content,
	}, nil
}

func (r *ReportManager) ListUniversityReports(userId uuid.UUID) ([]api.UniversityReport, error) {
	var reports []schema.UserUniversityReport

	if err := r.db.Preload("Report").Order("last_accessed_at DESC").Find(&reports, "user_id = ?", userId).Error; err != nil {
		slog.Error("error finding list of reports ")
		return nil, ErrReportAccessFailed
	}

	results := make([]api.UniversityReport, 0, len(reports))
	for _, report := range reports {
		results = append(results, convertUniversityReport(report, api.UniversityReportContent{}))
	}

	return results, nil
}

func (r *ReportManager) queueUniversityReportUpdateIfNeeded(txn *gorm.DB, report *schema.UniversityReport) error {
	if time.Now().UTC().Sub(report.LastUpdatedAt) > r.staleReportThreshold &&
		report.Status != schema.ReportInProgress && report.Status != schema.ReportQueued {
		updates := map[string]any{"status": schema.ReportQueued, "queued_at": time.Now().UTC()}
		if err := txn.Model(&report).Updates(updates).Error; err != nil {
			slog.Error("error queueing stale university report for update", "university_report_id", report.Id, "error", err)
			return ErrReportAccessFailed
		}
		report.Status = schema.ReportQueued
	}
	return nil
}

func (r *ReportManager) CreateUniversityReport(licenseId, userId uuid.UUID, universityId, universityName string) (uuid.UUID, error) {
	var userReport schema.UserUniversityReport
	var userReportId uuid.UUID
	now := time.Now().UTC()

	err := r.db.Transaction(func(txn *gorm.DB) error {
		var report schema.UniversityReport
		result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&report, "university_id = ?", universityId)
		if result.Error != nil {
			slog.Error("error checking for existing university report", "error", result.Error)
			return ErrReportCreationFailed
		}

		if result.RowsAffected == 0 {
			reportId := uuid.New()
			report = schema.UniversityReport{
				Id:             reportId,
				LastUpdatedAt:  EarliestReportDate,
				UniversityId:   universityId,
				UniversityName: universityName,
				Status:         schema.ReportQueued,
			}

			if err := txn.Create(&report).Error; err != nil {
				slog.Error("error creating new university report", "error", err)
				return ErrReportCreationFailed
			}
		}

		if err := r.queueUniversityReportUpdateIfNeeded(txn, &report); err != nil {
			return err
		}

		result = txn.Where("user_id = ? AND report_id = ?", userId, report.Id).
			First(&userReport)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Error("error finding existing user university report", "error", result.Error)
			return ErrReportCreationFailed
		}

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			userReportId = uuid.New()
			userReport = schema.UserUniversityReport{
				Id:             userReportId,
				UserId:         userId,
				ReportId:       report.Id,
				LastAccessedAt: now,
			}
			if err := txn.Create(&userReport).Error; err != nil {
				slog.Error("error creating new user university report", "error", err)
				return ErrReportCreationFailed
			}
		} else {
			if err := txn.Model(&userReport).Update("last_accessed_at", now).Error; err != nil {
				slog.Error("error updating user university report last_accessed_at", "error", err)
				return ErrReportCreationFailed
			}

			userReportId = userReport.Id
		}

		usage := schema.LicenseUsage{
			LicenseId:  licenseId,
			ReportId:   userReport.Id,
			ReportType: schema.UniversityReportType,
			UserId:     userId,
			Timestamp:  time.Now().UTC(),
		}
		if err := txn.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&usage).Error; err != nil {
			slog.Error("error logging license usage", "error", err)
			return errors.New("error updating license usage")
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return userReportId, nil
}

const (
	// The last N years of flags that will be added to the university report.
	yearsInUniversityReport = 5
)

func (r *ReportManager) GetUniversityReport(userId, reportId uuid.UUID) (api.UniversityReport, error) {
	type universityAuthorFlags struct {
		AuthorName string
		AuthorId   string
		Source     string
		FlagType   string
		Count      int
	}

	type authorReportStatus struct {
		Status string
		Count  int
	}

	var report schema.UserUniversityReport

	var statusCounts []authorReportStatus
	var flags []universityAuthorFlags

	if err := r.db.Transaction(func(txn *gorm.DB) error {
		if err := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Report").
			First(&report, "id = ?", reportId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReportNotFound
			}
			slog.Error("error getting user university report", "university_report_id", reportId, "error", err)
			return ErrReportAccessFailed
		}

		if report.UserId != userId {
			return ErrUserCannotAccessReport
		}

		if err := r.queueUniversityReportUpdateIfNeeded(txn, report.Report); err != nil {
			return err
		}

		report.LastAccessedAt = time.Now().UTC()
		if err := txn.Save(&report).Error; err != nil {
			slog.Error("error updating user university report last_accessed_at", "error", err)
			return ErrReportAccessFailed
		}

		if err := txn.Model(&schema.AuthorReport{}).
			Select("author_reports.status, count(*) as count").
			Joins("JOIN university_authors ON author_reports.id = university_authors.author_report_id AND university_authors.university_report_id = ?", report.ReportId).
			Group("author_reports.status").
			Find(&statusCounts).Error; err != nil {
			slog.Error("error checking status for author reports linked to university report", "university_report_id", reportId, "error", err)
			return ErrReportAccessFailed
		}

		// Note: it is ok to select author_reports.author_name/author_id/source even though
		// they are not not part of the group by clause or have an aggregator becuase
		// we are grouping by the primary key of the table they come from, so only a single value is possible.
		if err := txn.Model(&schema.AuthorFlag{}).
			Select("author_reports.author_name, author_reports.author_id, author_reports.source, author_flags.flag_type, count(*) as count").
			Joins("JOIN author_reports ON author_flags.report_id = author_reports.id").
			Joins("JOIN university_authors ON author_reports.id = university_authors.author_report_id AND university_authors.university_report_id = ?", report.ReportId).
			Where("author_flags.date IS NULL OR author_flags.date > ?", time.Now().AddDate(-yearsInUniversityReport, 0, 0)).
			Group("author_reports.id, author_flags.flag_type").
			Find(&flags).Error; err != nil {
			slog.Error("error querying flags for author reports linked to university report", "university_report_id", reportId, "error", err)
			return ErrReportAccessFailed
		}

		return nil
	}); err != nil {
		return api.UniversityReport{}, err
	}

	content := api.UniversityReportContent{
		Flags: make(map[string][]api.UniversityAuthorFlag),
	}

	for _, flag := range flags {
		content.Flags[flag.FlagType] = append(content.Flags[flag.FlagType], api.UniversityAuthorFlag{
			AuthorId:   flag.AuthorId,
			AuthorName: flag.AuthorName,
			Source:     flag.Source,
			FlagCount:  flag.Count,
		})
	}

	for _, status := range statusCounts {
		switch status.Status {
		case schema.ReportCompleted, schema.ReportFailed:
			content.AuthorsReviewed += status.Count
		}
		content.TotalAuthors += status.Count
	}

	return convertUniversityReport(report, content), nil
}

func (r *ReportManager) DeleteUniversityReport(userId, reportId uuid.UUID) error {
	result := r.db.Delete(&schema.UserUniversityReport{}, "id = ? AND user_id = ?", reportId, userId)
	if result.Error != nil {
		slog.Error("error deleting university report", "university_report_id", reportId, "error", result.Error)
		return ErrReportAccessFailed
	}
	if result.RowsAffected != 1 {
		return ErrReportNotFound
	}
	return nil
}

type UniversityReportUpdateTask struct {
	Id             uuid.UUID
	UniversityId   string
	UniversityName string
	UpdateDate     time.Time
}

func (r *ReportManager) GetNextUniversityReport() (*UniversityReportUpdateTask, error) {
	found := false
	var report schema.UniversityReport

	err := r.db.Transaction(func(txn *gorm.DB) error {
		result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Limit(1).Order("queued_at ASC").
			Find(&report, "status = ?", schema.ReportQueued)
		if result.Error != nil {
			slog.Error("error getting next university report from queue", "error", result.Error)
			return ErrReportAccessFailed
		}

		if result.RowsAffected != 1 {
			return nil
		}

		if err := txn.Model(&report).Update("status", schema.ReportInProgress).Error; err != nil {
			slog.Error("error updating university report status to in progress", "error", err)
			return ErrReportAccessFailed
		}

		found = true
		return nil
	})

	if err != nil {
		return nil, err
	}

	if found {
		return &UniversityReportUpdateTask{
			Id:             report.Id,
			UniversityId:   report.UniversityId,
			UniversityName: report.UniversityName,
			UpdateDate:     time.Now().UTC(),
		}, nil
	}

	return nil, nil
}

func (r *ReportManager) queueAuthorReportUpdatesForUniversityReport(txn *gorm.DB, universityReportId uuid.UUID) error {
	staleCutoff := time.Now().UTC().Add(-r.staleReportThreshold)

	result := txn.Model(&schema.AuthorReport{}).
		Where("EXISTS (?)", txn.Table("university_authors").Where("university_authors.author_report_id = author_reports.id AND university_authors.university_report_id = ?", universityReportId)).
		Where("author_reports.last_updated_at < ?", staleCutoff).
		Where("author_reports.status IN ?", []string{schema.ReportFailed, schema.ReportCompleted}).
		Updates(map[string]any{"status": schema.ReportQueued, "queued_at": time.Now()})

	if result.Error != nil {
		slog.Error("error queueing stale author reports for university report", "university_report_id", universityReportId, "error", result.Error)
		return ErrReportAccessFailed
	}

	slog.Info("queued updates for author reports for university report", "n_author_reports", result.RowsAffected, "university_report_id", universityReportId)
	return nil
}

type UniversityAuthorReport struct {
	AuthorId   string
	AuthorName string
	Source     string
}

func (r *ReportManager) UpdateUniversityReport(id uuid.UUID, status string, updateTime time.Time, authors []UniversityAuthorReport) error {
	return r.db.Transaction(func(txn *gorm.DB) error {
		var report schema.UniversityReport

		if err := txn.Clauses(clause.Locking{Strength: "UPDATE"}).First(&report, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("cannot update status of university report, report not found", "university_report_id", id, "status", status)
				return ErrReportNotFound
			}
			slog.Error("error getting university report to update status", "university_report_id", id, "status", status, "error", err)
			return ErrReportAccessFailed
		}

		updates := map[string]any{"status": status}

		if status == schema.ReportCompleted {
			updates["last_updated_at"] = updateTime

			var authorReports []schema.AuthorReport

			for _, author := range authors {
				report, err := createOrGetAuthorReport(txn, author.AuthorId, author.AuthorName, author.Source, false /*fromUserReq*/)
				if err != nil {
					slog.Error("error getting author report to add to university report", "author_id", author.AuthorId, "university_report_id", id, "error", err)
					return err
				}
				authorReports = append(authorReports, report)
			}

			if err := txn.Model(&schema.UniversityReport{Id: id}).Association("Authors").Replace(authorReports); err != nil {
				slog.Error("error replacing author univerisity associations", "university_report_id", id, "error", err)
				return ErrReportAccessFailed
			}

			if err := r.queueAuthorReportUpdatesForUniversityReport(txn, id); err != nil {
				slog.Error("error queueing author report updates for university report", "university_report_id", id, "error", err)
				return err
			}
		}

		if err := txn.Model(&report).Updates(updates).Error; err != nil {
			slog.Error("error updating university report status", "university_report_id", id, "error", err)
			return ErrReportAccessFailed
		}

		return nil
	})

}

func convertUniversityReport(report schema.UserUniversityReport, content api.UniversityReportContent) api.UniversityReport {
	return api.UniversityReport{
		Id:             report.Id,
		LastAccessedAt: report.LastAccessedAt,
		UniversityId:   report.Report.UniversityId,
		UniversityName: report.Report.UniversityName,
		Status:         report.Report.Status,
		Content:        content,
	}
}
