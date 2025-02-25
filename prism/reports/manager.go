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

func (r *ReportManager) ListReports(userId uuid.UUID) ([]api.Report, error) {
	var reports []schema.UserAuthorReport

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

func (r *ReportManager) queueReportUpdateIfNeeded(txn *gorm.DB, report *schema.AuthorReport) error {
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
		var report schema.AuthorReport
		result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&report, "author_id = ? AND source = ?", authorId, source)
		if result.Error != nil {
			slog.Error("error checking for existing report", "error", result.Error)
			return ErrReportCreationFailed
		}

		if result.RowsAffected == 0 {
			reportId := uuid.New()
			report = schema.AuthorReport{
				Id:            reportId,
				LastUpdatedAt: EarliestReportDate,
				AuthorId:      authorId,
				AuthorName:    authorName,
				Source:        source,
				Status:        schema.ReportQueued,
			}

			if err := txn.Create(&report).Error; err != nil {
				slog.Error("error creating new report", "error", err)
				return ErrReportCreationFailed
			}
		}

		if err := r.queueReportUpdateIfNeeded(txn, &report); err != nil {
			return err
		}

		userReport := schema.UserAuthorReport{
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
			LicenseId:          licenseId,
			UserAuthorReportId: userReport.Id,
			UserId:             userId,
			Timestamp:          time.Now().UTC(),
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
	var report schema.UserAuthorReport

	if err := r.db.Transaction(func(txn *gorm.DB) error {
		if err := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Report").Preload("Report.Flags").
			First(&report, "id = ?", reportId).Error; err != nil {
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
	var report schema.AuthorReport

	err := r.db.Transaction(func(txn *gorm.DB) error {
		result := txn.Clauses(clause.Locking{Strength: "UPDATE"}).
			Limit(1).Order("queued_at ASC").
			Find(&report, "status = ?", schema.ReportQueued)
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

func flagsToReportContent(flags []schema.AuthorFlag) (api.ReportContent, error) {
	content := make(api.ReportContent)

	for _, flag := range flags {
		output, err := api.EmptyFlag(flag.FlagType)
		if err != nil {
			return nil, nil
		}
		if err := json.Unmarshal(flag.Data, output); err != nil {
			return nil, fmt.Errorf("error deserializing flag: %w", err)
		}
		content[output.Type()] = append(content[output.Type()], output)
	}

	return content, nil
}

func (r *ReportManager) UpdateReport(id uuid.UUID, status string, updateTime time.Time, updateContent api.ReportContent) error {
	return r.db.Transaction(func(txn *gorm.DB) error {
		var report schema.AuthorReport

		if err := txn.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Flags").First(&report, "id = ?", id).Error; err != nil {
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

			seen := make(map[string]bool)
			for _, flag := range report.Flags {
				seen[flag.FlagKey] = true
			}

			newFlags := make([]schema.AuthorFlag, 0)
			for _, flags := range updateContent {
				for _, flag := range flags {
					if key := flag.Key(); !seen[key] {
						seen[key] = true

						data, err := json.Marshal(flag)
						if err != nil {
							return fmt.Errorf("error serializing flag: %w", err)
						}

						newFlags = append(newFlags, schema.AuthorFlag{
							Id:       uuid.New(),
							ReportId: report.Id,
							FlagType: flag.Type(),
							FlagKey:  key,
							Data:     data,
						})
					}
				}
			}

			if len(newFlags) > 0 {
				if err := txn.Save(newFlags).Error; err != nil {
					slog.Error("error adding new flags to report", "report_id", id, "error", err)
					return ErrReportAccessFailed
				}
			}
		}

		if err := txn.Model(&report).Updates(updates).Error; err != nil {
			slog.Error("error updating report status", "report_id", id, "error", err)
			return ErrReportAccessFailed
		}
		return nil
	})
}

func convertReport(report schema.UserAuthorReport) (api.Report, error) {
	content, err := flagsToReportContent(report.Report.Flags)
	if err != nil {
		return api.Report{}, err
	}

	return api.Report{
		Id:         report.Id,
		CreatedAt:  report.CreatedAt,
		AuthorId:   report.Report.AuthorId,
		AuthorName: report.Report.AuthorName,
		Source:     report.Report.Source,
		Status:     report.Report.Status,
		Content:    content,
	}, nil
}
