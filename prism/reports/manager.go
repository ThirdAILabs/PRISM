package reports

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"prism/prism/api"
	"prism/prism/schema"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

func (r *ReportManager) ListReports(userId uuid.UUID) ([]api.Report, error) {
	var reports []schema.UserReport

	if err := r.db.Preload("Report").Order("created_at ASC").Find(&reports, "user_id = ?", userId).Error; err != nil {
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

func (r *ReportManager) queueReportUpdateIfNeeded(txn *gorm.DB, report *schema.Report) error {
	if time.Now().UTC().Sub(report.LastUpdatedAt) > r.staleReportThreshold &&
		report.Status != schema.ReportInProgress && report.Status != schema.ReportQueued {
		updates := map[string]any{"status": schema.ReportQueued, "queued_at": time.Now().UTC()}
		if err := txn.Model(&report).Updates(updates).Error; err != nil {
			slog.Error("error queueing stale report for update", "report_id", report.Id, "error", err)
			return ErrReportAccessFailed
		}
		report.Status = schema.ReportQueued
	}
	return nil
}

func (r *ReportManager) CreateReport(licenseId, userId uuid.UUID, authorId, authorName, source string) (uuid.UUID, error) {
	userReportId := uuid.New()

	err := r.db.Transaction(func(txn *gorm.DB) error {
		var report schema.Report
		result := txn.Limit(1).Find(&report, "author_id = ? AND source = ?", authorId, source)
		if result.Error != nil {
			slog.Error("error checking for existing report", "error", result.Error)
			return ErrReportCreationFailed
		}

		if result.RowsAffected == 0 {
			reportId := uuid.New()
			report = schema.Report{
				Id:            reportId,
				LastUpdatedAt: EarliestReportDate,
				AuthorId:      authorId,
				AuthorName:    authorName,
				Source:        source,
				Status:        schema.ReportQueued,
				Content: &schema.ReportContent{
					ReportId: reportId,
					Content:  []byte(`{}`),
				},
			}

			if err := txn.Create(&report).Error; err != nil {
				slog.Error("error creating new report", "error", err)
				return ErrReportCreationFailed
			}
		}

		if err := r.queueReportUpdateIfNeeded(txn, &report); err != nil {
			return err
		}

		userReport := schema.UserReport{
			Id:        userReportId,
			UserId:    userId,
			CreatedAt: time.Now().UTC(),
			ReportId:  report.Id,
		}

		if err := txn.Create(&userReport).Error; err != nil {
			slog.Error("error creating new user report", "error", err)
			return ErrReportCreationFailed
		}

		usage := schema.LicenseUsage{
			LicenseId:    licenseId,
			UserReportId: userReport.Id,
			UserId:       userId,
			Timestamp:    time.Now().UTC(),
		}
		if err := txn.Create(&usage).Error; err != nil {
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

func (r *ReportManager) GetReport(userId, reportId uuid.UUID) (api.Report, error) {
	var report schema.UserReport

	if err := r.db.Transaction(func(txn *gorm.DB) error {
		if err := txn.Preload("Report").Preload("Report.Content").First(&report, "id = ?", reportId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReportNotFound
			}
			slog.Error("error getting user report", "report_id", reportId, "error", err)
			return ErrReportAccessFailed
		}

		if report.UserId != userId {
			return ErrUserCannotAccessReport
		}

		if err := r.queueReportUpdateIfNeeded(txn, report.Report); err != nil {
			return err
		}
		return nil
	}); err != nil {
		slog.Error("error getting user report", "report_id", reportId, "error", err)
		return api.Report{}, err
	}

	return convertReport(report)
}

type ReportUpdateTask struct {
	Id         uuid.UUID
	AuthorId   string
	AuthorName string
	Source     string
	StartDate  time.Time
	EndDate    time.Time
}

func (r *ReportManager) GetNextReport() (*ReportUpdateTask, error) {
	found := false
	var report schema.Report

	err := r.db.Transaction(func(txn *gorm.DB) error {
		result := txn.Limit(1).Order("queued_at ASC").Find(&report, "status = ?", schema.ReportQueued)
		if result.Error != nil {
			slog.Error("error getting next report from queue", "error", result.Error)
			return ErrReportAccessFailed
		}

		if result.RowsAffected != 1 {
			return nil
		}

		if err := txn.Model(&report).Update("status", schema.ReportInProgress).Error; err != nil {
			slog.Error("error updating report status to in progress", "error", err)
			return ErrReportAccessFailed
		}

		found = true
		return nil
	})

	if err != nil {
		return nil, err
	}

	if found {
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

func mergeFlags[T api.Flag](old, new []T) []T {
	seen := make(map[string]bool)
	output := make([]T, 0, len(old)+len(new))
	for _, flag := range old {
		if key := flag.Key(); !seen[key] {
			output = append(output, flag)
			seen[key] = true
		}
	}
	for _, flag := range new {
		if key := flag.Key(); !seen[key] {
			output = append(output, flag)
			seen[key] = true
		}
	}
	return output
}

func deserializeReportContent(data []byte) (api.ReportContent, error) {
	var content api.ReportContent
	if err := json.Unmarshal(data, &content); err != nil {
		slog.Error("error deserializing report content", "error", err)
		return api.ReportContent{}, fmt.Errorf("error deserializing report content: %w", err)
	}
	return content, nil
}

func serializeReportContent(content api.ReportContent) ([]byte, error) {
	data, err := json.Marshal(content)
	if err != nil {
		slog.Error("error serializing report content", "error", err)
		return nil, fmt.Errorf("error serializing report content: %w", err)
	}
	return data, nil
}

func mergeContents(old, new api.ReportContent) api.ReportContent {
	return api.ReportContent{
		TalentContracts:                mergeFlags(old.TalentContracts, new.TalentContracts),
		AssociationsWithDeniedEntities: mergeFlags(old.AssociationsWithDeniedEntities, new.AssociationsWithDeniedEntities),
		HighRiskFunders:                mergeFlags(old.HighRiskFunders, new.HighRiskFunders),
		AuthorAffiliations:             mergeFlags(old.AuthorAffiliations, new.AuthorAffiliations),
		PotentialAuthorAffiliations:    mergeFlags(old.PotentialAuthorAffiliations, new.PotentialAuthorAffiliations),
		MiscHighRiskAssociations:       mergeFlags(old.MiscHighRiskAssociations, new.MiscHighRiskAssociations),
		CoauthorAffiliations:           mergeFlags(old.CoauthorAffiliations, new.CoauthorAffiliations),
	}
}

func (r *ReportManager) UpdateReport(id uuid.UUID, status string, updateTime time.Time, updateContent api.ReportContent) error {
	return r.db.Transaction(func(txn *gorm.DB) error {
		var report schema.Report

		if err := txn.Preload("Content").First(&report, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("cannot update status of report, report not found", "report_id", id, "status", status)
				return ErrReportNotFound
			}
			slog.Error("error getting report to update status", "report_id", id, "status", status, "error", err)
			return ErrReportAccessFailed
		}

		updates := map[string]any{"status": status}

		if status == schema.ReportCompleted {
			updates["last_updated_at"] = updateTime

			oldContent, err := deserializeReportContent(report.Content.Content)
			if err != nil {
				return err
			}

			newContent, err := serializeReportContent(mergeContents(oldContent, updateContent))
			if err != nil {
				return err
			}

			if err := txn.Save(&schema.ReportContent{ReportId: id, Content: newContent}).Error; err != nil {
				slog.Error("error updating report content", "report_id", id, "error", err)
				return ErrReportAccessFailed
			}
		}

		if err := txn.Model(&report).Updates(updates).Error; err != nil {
			slog.Error("error updating report status", "report_id", id, "error", err)
			return ErrReportAccessFailed
		}
		return nil
	})
}

func convertReport(report schema.UserReport) (api.Report, error) {
	result := api.Report{
		Id:         report.Id,
		CreatedAt:  report.CreatedAt,
		AuthorId:   report.Report.AuthorId,
		AuthorName: report.Report.AuthorName,
		Source:     report.Report.Source,
		Status:     report.Report.Status,
	}

	if report.Report.Content != nil {
		content, err := deserializeReportContent(report.Report.Content.Content)
		if err != nil {
			return api.Report{}, ErrReportAccessFailed
		}
		result.Content = content
	}

	return result, nil
}
