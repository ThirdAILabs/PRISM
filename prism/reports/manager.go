package reports

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"prism/prism/api"
	"prism/prism/monitoring"
	"prism/prism/schema"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AuthorReportTimeout            time.Duration = time.Minute * 20
	UniversityReportTimeout        time.Duration = time.Minute * 45
	UniversityReportUpdateInterval time.Duration = time.Hour * 24 * 30
	AuthorReportUpdateInterval     time.Duration = time.Hour * 24 * 14
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
	db                             *gorm.DB
	authorReportUpdateInterval     time.Duration
	authorReportTimeout            time.Duration
	universityReportUpdateInterval time.Duration
	universityReportTimeout        time.Duration
	stopReportUpdate               chan struct{}
}

func NewManager(db *gorm.DB) *ReportManager {
	return &ReportManager{
		db:                             db,
		authorReportUpdateInterval:     AuthorReportUpdateInterval,
		authorReportTimeout:            AuthorReportTimeout,
		universityReportUpdateInterval: UniversityReportUpdateInterval,
		universityReportTimeout:        UniversityReportTimeout,
	}
}

func (r *ReportManager) StartReportUpdateCheck() {
	r.stopReportUpdate = make(chan struct{})
	go func() {
		ticker := time.Tick(10 * time.Minute)

		for {
			select {
			case <-ticker:
				if err := r.CheckForStaleAuthorReports(); err != nil {
					slog.Error("error checking for stale author reports", "error", err)
				}
				if err := r.CheckForStaleUniversityReports(); err != nil {
					slog.Error("error checking for stale university reports", "error", err)
				}
			case <-r.stopReportUpdate:
				return
			}
		}
	}()
}

func (r *ReportManager) StopReportUpdateCheck() {
	if r.stopReportUpdate != nil {
		close(r.stopReportUpdate)
	}
}

func (r *ReportManager) SetAuthorReportUpdateInterval(interval time.Duration) *ReportManager {
	r.authorReportUpdateInterval = interval
	return r
}

func (r *ReportManager) SetAuthorReportTimeout(timeout time.Duration) *ReportManager {
	r.authorReportTimeout = timeout
	return r
}

func (r *ReportManager) SetUniversityReportUpdateInterval(interval time.Duration) *ReportManager {
	r.universityReportUpdateInterval = interval
	return r
}

func (r *ReportManager) SetUniversityReportTimeout(timeout time.Duration) *ReportManager {
	r.universityReportTimeout = timeout
	return r
}

func (r *ReportManager) ListAuthorReports(userId uuid.UUID) ([]api.Report, error) {
	var reports []schema.UserAuthorReport

	if err := r.db.Preload("Report").Order("last_accessed_at DESC").Find(&reports, "user_id = ?", userId).Error; err != nil {
		slog.Error("error finding list of reports ")
		return nil, ErrReportAccessFailed
	}

	results := make([]api.Report, 0, len(reports))
	for _, report := range reports {
		res, err := ConvertReport(report)
		if err != nil {
			return nil, ErrReportAccessFailed
		}
		results = append(results, res)
	}

	return results, nil
}

func createOrGetAuthorReport(txn *gorm.DB, authorId, authorName, source, affiliations, researchInterests string, forUniversityReport bool) (schema.AuthorReport, error) {
	var report schema.AuthorReport
	result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
		Limit(1).
		Find(&report, "author_id = ? AND source = ? AND for_university_report = ?", authorId, source, forUniversityReport)
	if result.Error != nil {
		slog.Error("error checking for existing author report", "error", result.Error)
		return schema.AuthorReport{}, ErrReportCreationFailed
	}

	if result.RowsAffected == 0 {
		report = schema.AuthorReport{
			Id:                  uuid.New(),
			LastUpdatedAt:       EarliestReportDate,
			AuthorId:            authorId,
			AuthorName:          authorName,
			Source:              source,
			Affiliations:        affiliations,
			ResearchInterests:   researchInterests,
			Status:              schema.ReportQueued,
			StatusUpdatedAt:     time.Now().UTC(),
			ForUniversityReport: forUniversityReport,
		}

		if err := txn.Create(&report).Error; err != nil {
			slog.Error("error creating new author report", "error", err)
			return schema.AuthorReport{}, ErrReportCreationFailed
		}
	} else {
		if forUniversityReport {
			monitoring.UniAuthorReportsFoundInCache.WithLabelValues("TODO ORG").Inc()
		} else {
			monitoring.AuthorReportsFoundInCache.WithLabelValues("TODO ORG").Inc()
		}
	}

	if forUniversityReport {
		monitoring.UniAuthorReportsCreated.WithLabelValues("TODO ORG").Inc()
	} else {
		monitoring.AuthorReportsCreated.WithLabelValues("TODO ORG").Inc()
	}

	return report, nil
}

func (r *ReportManager) CreateAuthorReport(userId uuid.UUID, authorId, authorName, source, affiliations, researchInterests string) (uuid.UUID, error) {
	var userReport schema.UserAuthorReport
	var userReportId uuid.UUID
	now := time.Now().UTC()

	err := r.db.Transaction(func(txn *gorm.DB) error {
		report, err := createOrGetAuthorReport(txn, authorId, authorName, source, affiliations, researchInterests, false /*forUniversityReport*/)
		if err != nil {
			return err
		}

		result := txn.Where("user_id = ? AND report_id = ?", userId, report.Id).Limit(1).Find(&userReport)
		if result.Error != nil {
			slog.Error("error finding existing user author report", "error", result.Error)
			return ErrReportCreationFailed
		}

		if result.RowsAffected == 0 {
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

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return userReportId, nil
}

func (r *ReportManager) CheckForStaleAuthorReports() error {
	// Check for author reports that are stale relative to the update frequency specified in a user author report.
	if err := r.db.Model(&schema.AuthorReport{}).
		Where("status != ? AND status != ? AND for_university_report = ?", schema.ReportInProgress, schema.ReportQueued, false).
		Where("last_updated_at < ?", time.Now().UTC().Add(-r.authorReportUpdateInterval)).
		Updates(map[string]any{"status": schema.ReportQueued, "status_updated_at": time.Now().UTC()}).Error; err != nil {
		slog.Error("error checking for stale author reports", "error", err)
		return ErrReportAccessFailed
	}

	if err := r.db.Model(&schema.AuthorReport{}).
		Where("status != ? AND status != ? AND for_university_report = ?", schema.ReportInProgress, schema.ReportQueued, false).
		Where("EXISTS (?)", r.db.Model(&schema.AuthorReportHook{}).
			Select("1").
			Joins("JOIN user_author_reports ON user_author_reports.id = author_report_hooks.user_report_id").
			Where("user_author_reports.report_id = author_reports.id").
			Where("author_report_hooks.last_ran_at < NOW() - (author_report_hooks.interval || ' seconds')::interval"),
		).
		Updates(map[string]any{"status": schema.ReportQueued, "status_updated_at": time.Now().UTC()}).Error; err != nil {
		slog.Error("error checking report updates required by hooks", "error", err)
		return ErrReportAccessFailed
	}

	// Check for author reports that are stale relative to a university report.
	if err := r.db.Model(&schema.AuthorReport{}).
		Where("status != ? AND status != ? AND for_university_report = ?", schema.ReportInProgress, schema.ReportQueued, true).
		Where("last_updated_at < ?", time.Now().UTC().Add(-r.universityReportUpdateInterval)).
		Updates(map[string]any{"status": schema.ReportQueued, "status_updated_at": time.Now().UTC()}).Error; err != nil {
		slog.Error("error checking for stale author reports for university reports", "error", err)
		return ErrReportAccessFailed
	}

	// Check for author reports that have been in progress for too long.
	if err := r.db.Model(&schema.AuthorReport{}).
		Where("status = ?", schema.ReportInProgress).
		Where("status_updated_at < ?", time.Now().UTC().Add(-r.authorReportTimeout)).
		Updates(map[string]any{"status": schema.ReportQueued, "status_updated_at": time.Now().UTC()}).Error; err != nil {
		slog.Error("error checking for long running author reports", "error", err)
		return ErrReportAccessFailed
	}

	return nil
}

func (r *ReportManager) GetAuthorReport(userId, reportId uuid.UUID) (api.Report, error) {
	var report schema.UserAuthorReport

	if err := r.db.Preload("Report").Preload("Report.Flags").
		First(&report, "id = ?", reportId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return api.Report{}, ErrReportNotFound
		}
		slog.Error("error getting user author report", "author_report_id", reportId, "error", err)
		return api.Report{}, ErrReportAccessFailed
	}

	if report.UserId != userId {
		return api.Report{}, ErrUserCannotAccessReport
	}

	return ConvertReport(report)
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
	Id                  uuid.UUID
	AuthorId            string
	AuthorName          string
	Source              string
	StartDate           time.Time
	EndDate             time.Time
	ForUniversityReport bool
	Affiliations        string
}

func (r *ReportManager) findNextAuthorReport(txn *gorm.DB) (*schema.AuthorReport, error) {
	var report schema.AuthorReport
	for _, for_university_report := range [2]bool{false, true} {
		result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Limit(1).Order("status_updated_at ASC").
			Where("for_university_report = ?", for_university_report).
			Where("status = ?", schema.ReportQueued).
			Find(&report)
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
		report, err = r.findNextAuthorReport(txn)
		if err != nil {
			return err
		}

		if report != nil {
			updates := map[string]any{"status": schema.ReportInProgress, "status_updated_at": time.Now().UTC()}
			if err := txn.Model(report).Updates(updates).Error; err != nil {
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
			Id:                  report.Id,
			AuthorId:            report.AuthorId,
			AuthorName:          report.AuthorName,
			Source:              report.Source,
			StartDate:           report.LastUpdatedAt,
			EndDate:             time.Now().UTC(),
			ForUniversityReport: report.ForUniversityReport,
			Affiliations:        report.Affiliations,
		}, nil
	}

	return nil, nil
}

func (r *ReportManager) UpdateAuthorReport(id uuid.UUID, status string, updateTime time.Time, updateFlags []api.Flag) error {
	return r.db.Transaction(func(txn *gorm.DB) error {
		updates := map[string]any{"status": status, "status_updated_at": updateTime}
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
}

func (r *ReportManager) CreateAuthorReportHook(userId, reportId uuid.UUID, action string, data []byte, interval int) error {
	return r.db.Transaction(func(txn *gorm.DB) error {
		var userReport schema.UserAuthorReport
		if err := txn.First(&userReport, "id = ?", reportId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReportNotFound
			}
			slog.Error("error retrieving user author report", "error", err)
			return ErrReportCreationFailed
		}

		if userReport.UserId != userId {
			return ErrUserCannotAccessReport
		}

		hook := schema.AuthorReportHook{
			Id:           uuid.New(),
			UserReportId: reportId,
			Action:       action,
			Data:         data,
			LastRanAt:    EarliestReportDate,
			Interval:     interval,
		}

		if err := txn.Create(&hook).Error; err != nil {
			slog.Error("error creating author report hook", "error", err)
			return ErrReportCreationFailed
		}

		return nil
	})
}

func flagsToReportContent(flags []schema.AuthorFlag) (map[string][]api.Flag, error) {
	content := make(map[string][]api.Flag)

	for _, flag := range flags {
		output, err := api.ParseFlag(flag.FlagType, flag.Data)
		if err != nil {
			return nil, err
		}
		content[output.Type()] = append(content[output.Type()], output)
	}

	return content, nil
}

func ConvertReport(report schema.UserAuthorReport) (api.Report, error) {
	content, err := flagsToReportContent(report.Report.Flags)
	if err != nil {
		return api.Report{}, err
	}

	return api.Report{
		Id:                report.Id,
		LastAccessedAt:    report.LastAccessedAt,
		AuthorId:          report.Report.AuthorId,
		AuthorName:        report.Report.AuthorName,
		Source:            report.Report.Source,
		Affiliations:      report.Report.Affiliations,
		ResearchInterests: report.Report.ResearchInterests,
		Status:            report.Report.Status,
		Content:           content,
	}, nil
}

func (r *ReportManager) ListUniversityReports(userId uuid.UUID) ([]api.UniversityReport, error) {
	var reports []schema.UserUniversityReport

	if err := r.db.Preload("Report").Preload("Report.Authors").Order("last_accessed_at DESC").Find(&reports, "user_id = ?", userId).Error; err != nil {
		slog.Error("error finding list of reports ")
		return nil, ErrReportAccessFailed
	}

	results := make([]api.UniversityReport, 0, len(reports))
	for _, report := range reports {
		content := api.UniversityReportContent{}
		for _, author := range report.Report.Authors {
			switch author.Status {
			case schema.ReportCompleted, schema.ReportFailed:
				content.AuthorsReviewed++
			}
			content.TotalAuthors++
		}

		results = append(results, convertUniversityReport(report, content))
	}

	return results, nil
}

func (r *ReportManager) CreateUniversityReport(userId uuid.UUID, universityId, universityName, universityLocation string) (uuid.UUID, error) {
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
				Id:                 reportId,
				LastUpdatedAt:      EarliestReportDate,
				UniversityId:       universityId,
				UniversityName:     universityName,
				UniversityLocation: universityLocation,
				StatusUpdatedAt:    time.Now().UTC(),
				Status:             schema.ReportQueued,
			}

			if err := txn.Create(&report).Error; err != nil {
				slog.Error("error creating new university report", "error", err)
				return ErrReportCreationFailed
			}
		} else {
			monitoring.UniAuthorReportsFoundInCache.WithLabelValues("TODO ORG").Inc()
		}

		monitoring.UniAuthorReportsCreated.WithLabelValues("TODO ORG").Inc()

		result = txn.Where("user_id = ? AND report_id = ?", userId, report.Id).
			Limit(1).Find(&userReport)
		if result.Error != nil {
			slog.Error("error finding existing user university report", "error", result.Error)
			return ErrReportCreationFailed
		}

		if result.RowsAffected == 0 {
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

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return userReportId, nil
}

func (r *ReportManager) CheckForStaleUniversityReports() error {
	// Check for university reports that are stale relative to the university report update frequency.
	if err := r.db.Model(&schema.UniversityReport{}).
		Where("status != ? AND status != ?", schema.ReportInProgress, schema.ReportQueued).
		Where("last_updated_at < ?", time.Now().UTC().Add(-r.universityReportUpdateInterval)).
		Updates(map[string]any{"status": schema.ReportQueued, "status_updated_at": time.Now().UTC()}).Error; err != nil {
		slog.Error("error checking for stale university reports", "error", err)
		return ErrReportAccessFailed
	}

	// Check for university reports that have been in progress for too long.
	if err := r.db.Model(&schema.UniversityReport{}).
		Where("status = ?", schema.ReportInProgress).
		Where("status_updated_at < ?", time.Now().UTC().Add(-r.universityReportTimeout)).
		Updates(map[string]any{"status": schema.ReportQueued, "status_updated_at": time.Now().UTC()}).Error; err != nil {
		slog.Error("error checking for long running university reports", "error", err)
		return ErrReportAccessFailed
	}

	return nil
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
		if err := txn.Preload("Report").First(&report, "id = ?", reportId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReportNotFound
			}
			slog.Error("error getting user university report", "university_report_id", reportId, "error", err)
			return ErrReportAccessFailed
		}

		if report.UserId != userId {
			return ErrUserCannotAccessReport
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
			Where("author_flags.date IS NULL OR author_flags.date > ?", time.Now().UTC().AddDate(-yearsInUniversityReport, 0, 0)).
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

	for _, flagList := range content.Flags {
		sort.Slice(flagList, func(i, j int) bool {
			if flagList[i].FlagCount == flagList[j].FlagCount {
				return flagList[i].AuthorName < flagList[j].AuthorName
			}
			return flagList[i].FlagCount > flagList[j].FlagCount
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
	Id                 uuid.UUID
	UniversityId       string
	UniversityName     string
	UniversityLocation string
	UpdateDate         time.Time
}

func (r *ReportManager) GetNextUniversityReport() (*UniversityReportUpdateTask, error) {
	retryTimestamp := time.Now().UTC().Add(-r.universityReportTimeout)

	found := false
	var report schema.UniversityReport

	err := r.db.Transaction(func(txn *gorm.DB) error {
		result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Limit(1).Order("status_updated_at ASC").
			Where("status = ? OR (status = ? AND status_updated_at < ?)", schema.ReportQueued, schema.ReportInProgress, retryTimestamp).
			Find(&report)
		if result.Error != nil {
			slog.Error("error getting next university report from queue", "error", result.Error)
			return ErrReportAccessFailed
		}

		if result.RowsAffected != 1 {
			return nil
		}

		updates := map[string]any{"status": schema.ReportInProgress, "status_updated_at": time.Now().UTC()}
		if err := txn.Model(&report).Updates(updates).Error; err != nil {
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
			Id:                 report.Id,
			UniversityId:       report.UniversityId,
			UniversityName:     report.UniversityName,
			UniversityLocation: report.UniversityLocation,
			UpdateDate:         time.Now().UTC(),
		}, nil
	}

	return nil, nil
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

		updates := map[string]any{"status": status, "status_updated_at": time.Now().UTC()}

		if status == schema.ReportCompleted {
			updates["last_updated_at"] = updateTime

			var authorReports []schema.AuthorReport

			for _, author := range authors {
				report, err := createOrGetAuthorReport(txn, author.AuthorId, author.AuthorName, author.Source, "", "", true /*forUniversityReport*/)
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
		Id:                 report.Id,
		LastAccessedAt:     report.LastAccessedAt,
		UniversityId:       report.Report.UniversityId,
		UniversityName:     report.Report.UniversityName,
		UniversityLocation: report.Report.UniversityLocation,
		Status:             report.Report.Status,
		Content:            content,
	}
}
