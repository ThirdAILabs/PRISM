package reports

import (
	"errors"
	"log/slog"
	"prism/api"
	"prism/schema"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrReportAccessFailed     = errors.New("report access failed")
	ErrReportCreationFailed   = errors.New("report creation failed")
	ErrReportNotFound         = errors.New("report not found")
	ErrUserCannotAccessReport = errors.New("user cannot access report")
)

type ReportManager struct {
	db *gorm.DB
}

func (r *ReportManager) ListReports(userId uuid.UUID) ([]api.Report, error) {
	var reports []schema.Report

	if err := r.db.Find(&reports, "user_id = ?", userId).Error; err != nil {
		slog.Error("error finding list of reports ")
		return nil, ErrReportAccessFailed
	}

	results := make([]api.Report, 0, len(reports))
	for _, report := range reports {
		results = append(results, convertReport(report))
	}

	return results, nil
}

func (r *ReportManager) CreateReport(userId uuid.UUID, authorId, displayName, source string, startYear, endYear int) (uuid.UUID, error) {
	report := schema.Report{
		Id:          uuid.New(),
		UserId:      userId,
		CreatedAt:   time.Now().UTC(),
		AuthorId:    authorId,
		DisplayName: displayName,
		Source:      source,
		StartYear:   startYear,
		EndYear:     endYear,
		Status:      schema.ReportQueued,
	}

	if err := r.db.Create(&report).Error; err != nil {
		slog.Error("error creating new report", "error", err)
		return uuid.Nil, ErrReportAccessFailed
	}

	return report.Id, nil
}

func (r *ReportManager) GetReport(userId, id uuid.UUID) (api.Report, error) {
	report, err := getReport(r.db, id)
	if err != nil {
		return api.Report{}, err
	}

	if report.UserId != userId {
		return api.Report{}, ErrUserCannotAccessReport
	}

	return convertReport(report), nil
}

func (r *ReportManager) GetNextReport() (*api.Report, error) {
	found := false
	var report schema.Report

	err := r.db.Transaction(func(txn *gorm.DB) error {
		if err := txn.First(&report, "status = ?", schema.ReportQueued).Order("created_at ASC").Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			slog.Error("error getting next report from queue", "error", err)
			return ErrReportAccessFailed
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
		result := convertReport(report)
		return &result, nil
	}

	return nil, nil
}

func (r *ReportManager) UpdateReport(id uuid.UUID, status string, content []byte) error {
	return r.db.Transaction(func(txn *gorm.DB) error {
		report, err := getReport(txn, id)
		if err != nil {
			return err
		}

		report.Status = status
		report.Content = &schema.ReportContent{
			ReportId: id, Content: content,
		}

		if err := txn.Save(&report).Error; err != nil {
			slog.Error("error updating report content and status", "report_id", id, "error", err)
			return ErrReportAccessFailed
		}
		return nil
	})
}

func getReport(txn *gorm.DB, id uuid.UUID) (schema.Report, error) {
	var report schema.Report

	if err := txn.First(&report, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return schema.Report{}, ErrReportNotFound
		}
		slog.Error("error getting report", "report_id", id, "error", err)
		return schema.Report{}, ErrReportAccessFailed
	}

	return report, nil
}

func convertReport(report schema.Report) api.Report {
	return api.Report{
		Id:          report.Id,
		CreatedAt:   report.CreatedAt,
		AuthorId:    report.AuthorId,
		DisplayName: report.DisplayName,
		Source:      report.Source,
		StartYear:   report.StartYear,
		EndYear:     report.EndYear,
		Status:      report.Status,
	}
}
