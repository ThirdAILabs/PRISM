package reports

import (
	"encoding/json"
	"errors"
	"log/slog"
	"prism/prism/api"
	"prism/prism/schema"
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

func NewManager(db *gorm.DB) *ReportManager {
	return &ReportManager{db: db}
}

func (r *ReportManager) ListReports(userId uuid.UUID) ([]api.Report, error) {
	var reports []schema.Report

	if err := r.db.Order("created_at ASC").Find(&reports, "user_id = ?", userId).Error; err != nil {
		slog.Error("error finding list of reports ")
		return nil, ErrReportAccessFailed
	}

	results := make([]api.Report, 0, len(reports))
	for _, report := range reports {
		res, err := convertReport(report, false)
		if err != nil {
			return nil, ErrReportAccessFailed
		}
		results = append(results, res)
	}

	return results, nil
}

func (r *ReportManager) CreateReport(licenseId, userId uuid.UUID, authorId, authorName, source string, startYear, endYear int) (uuid.UUID, error) {
	report := schema.Report{
		Id:         uuid.New(),
		UserId:     userId,
		CreatedAt:  time.Now().UTC(),
		AuthorId:   authorId,
		AuthorName: authorName,
		Source:     source,
		StartYear:  startYear,
		EndYear:    endYear,
		Status:     schema.ReportQueued,
	}

	err := r.db.Transaction(func(txn *gorm.DB) error {
		if err := txn.Create(&report).Error; err != nil {
			slog.Error("error creating new report", "error", err)
			return ErrReportAccessFailed
		}

		usage := schema.LicenseUsage{
			LicenseId: licenseId,
			ReportId:  report.Id,
			UserId:    userId,
			Timestamp: time.Now().UTC(),
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

	return report.Id, nil
}

func (r *ReportManager) GetReport(userId, id uuid.UUID, withDetails bool) (api.Report, error) {
	report, err := getReport(r.db, id, true)
	if err != nil {
		return api.Report{}, err
	}

	if report.UserId != userId {
		return api.Report{}, ErrUserCannotAccessReport
	}

	return convertReport(report, withDetails)
}

func (r *ReportManager) GetNextReport() (*api.Report, error) {
	found := false
	var report schema.Report

	err := r.db.Transaction(func(txn *gorm.DB) error {
		result := txn.Limit(1).Order("created_at ASC").Find(&report, "status = ?", schema.ReportQueued)
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
		result, err := convertReport(report, false)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	return nil, nil
}

func (r *ReportManager) UpdateReport(id uuid.UUID, status string, content []byte) error {
	return r.db.Transaction(func(txn *gorm.DB) error {
		report, err := getReport(txn, id, false)
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

func getReport(txn *gorm.DB, id uuid.UUID, withContent bool) (schema.Report, error) {
	var report schema.Report

	query := txn
	if withContent {
		query = query.Preload("Content")
	}

	if err := query.First(&report, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return schema.Report{}, ErrReportNotFound
		}
		slog.Error("error getting report", "report_id", id, "error", err)
		return schema.Report{}, ErrReportAccessFailed
	}

	return report, nil
}

func convertReport(report schema.Report, withDetails bool) (api.Report, error) {
	result := api.Report{
		Id:         report.Id,
		CreatedAt:  report.CreatedAt,
		AuthorId:   report.AuthorId,
		AuthorName: report.AuthorName,
		Source:     report.Source,
		StartYear:  report.StartYear,
		EndYear:    report.EndYear,
		Status:     report.Status,
	}

	if report.Content != nil {
		var content ReportContent
		err := json.Unmarshal(report.Content.Content, &content)
		if err != nil {
			slog.Error("error parsing report content", "error", err)
			return api.Report{}, ErrReportAccessFailed
		}
		if !withDetails {
			for i := range content.Connections {
				content.Connections[i].Details = nil
			}
		}

		result.Content = content
	}

	return result, nil
}
