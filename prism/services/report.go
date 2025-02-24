package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/reports"
	"prism/prism/schema"
	"prism/prism/services/auth"
	"prism/prism/services/licensing"
	"strings"
	"time"

	"github.com/bbalet/stopwords"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
	"gorm.io/gorm"
)

type ReportService struct {
	manager *reports.ReportManager
	db      *gorm.DB
}

func (s *ReportService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/list", WrapRestHandler(s.List))
	r.Post("/create", WrapRestHandler(s.CreateReport))
	r.Get("/{report_id}", WrapRestHandler(s.GetReport))
	r.Post("/{report_id}/check-disclosure", WrapRestHandler(s.CheckDisclosure))

	r.Post("/activate-license", WrapRestHandler(s.UseLicense))

	return r
}

func (s *ReportService) List(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reports, err := s.manager.ListReports(userId)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return reports, nil
}

func (s *ReportService) CreateReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	params, err := ParseRequestBody[api.CreateReportRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if params.AuthorId == "" {
		return nil, CodedError(errors.New("AuthorId must be specified"), http.StatusUnprocessableEntity)
	}

	if params.AuthorId == "" {
		return nil, CodedError(errors.New("AuthorName must be specified"), http.StatusUnprocessableEntity)
	}

	switch params.Source {
	case api.OpenAlexSource, api.GoogleScholarSource, api.ScopusSource, api.UnstructuredSource:
		// ok
	default:
		return nil, CodedError(errors.New("invalid Source"), http.StatusUnprocessableEntity)
	}

	licenseId, err := licensing.VerifyLicenseForReport(s.db, userId)
	if err != nil {
		slog.Error("cannot create new report, unable to verify license", "error", err)
		switch {
		case errors.Is(err, licensing.ErrMissingLicense):
			return nil, CodedError(err, http.StatusUnprocessableEntity)
		case errors.Is(err, licensing.ErrExpiredLicense):
			return nil, CodedError(err, http.StatusForbidden)
		case errors.Is(err, licensing.ErrDeactivatedLicense):
			return nil, CodedError(err, http.StatusForbidden)
		default:
			return nil, CodedError(err, http.StatusInternalServerError)
		}
	}

	id, err := s.manager.CreateReport(licenseId, userId, params.AuthorId, params.AuthorName, params.Source)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return api.CreateReportResponse{Id: id}, nil
}

func parseDateParam(param string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, param)
	if err != nil {
		return time.Time{}, CodedError(fmt.Errorf("error parsing date constraint: %w", err), http.StatusBadRequest)
	}
	return date, nil
}

func (s *ReportService) GetReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	param := chi.URLParam(r, "report_id")
	id, err := uuid.Parse(param)
	if err != nil {
		return nil, CodedError(fmt.Errorf("invalid uuid '%v' provided: %w", param, err), http.StatusBadRequest)
	}

	report, err := s.manager.GetReport(userId, id)
	if err != nil {
		switch {
		case errors.Is(err, reports.ErrReportNotFound):
			return nil, CodedError(err, http.StatusNotFound)
		case errors.Is(err, reports.ErrUserCannotAccessReport):
			return nil, CodedError(err, http.StatusForbidden)
		default:
			return nil, CodedError(err, http.StatusInternalServerError)
		}
	}

	return report, nil
}

func (s *ReportService) UseLicense(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	params, err := ParseRequestBody[api.ActivateLicenseRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if err := s.db.Transaction(func(txn *gorm.DB) error {
		return licensing.AddLicenseUser(txn, params.License, userId)
	}); err != nil {
		switch {
		case errors.Is(err, licensing.ErrLicenseNotFound):
			return nil, CodedError(err, http.StatusNotFound)
		case errors.Is(err, licensing.ErrInvalidLicense):
			return nil, CodedError(err, http.StatusUnprocessableEntity)
		case errors.Is(err, licensing.ErrExpiredLicense):
			return nil, CodedError(err, http.StatusForbidden)
		case errors.Is(err, licensing.ErrDeactivatedLicense):
			return nil, CodedError(err, http.StatusForbidden)
		default:
			return nil, CodedError(err, http.StatusInternalServerError)
		}
	}

	return nil, nil
}

func updateFlagDisclosure(flag api.Flag, allFileTexts []string) {
	entities := filterTokens(flag.GetEntities())

DisclosureCheck:
	for _, entity := range entities {
		for _, txt := range allFileTexts {
			if strings.Contains(strings.ToLower(txt), entity) {
				flag.MarkDisclosed()
				break DisclosureCheck
			}
		}
	}
}

func updateDisclosures[T api.Flag](flags []T, allTexts []string) {
	for _, flag := range flags {
		updateFlagDisclosure(flag, allTexts)
	}
}

func (s *ReportService) CheckDisclosure(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reportIdParam := chi.URLParam(r, "report_id")
	reportId, err := uuid.Parse(reportIdParam)
	if err != nil {
		return nil, CodedError(fmt.Errorf("invalid report id '%v': %w", reportIdParam, err), http.StatusBadRequest)
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	fileHeaders := r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		return nil, CodedError(errors.New("no files uploaded"), http.StatusBadRequest)
	}

	var allFileTexts []string
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			slog.Error("error opening uploaded file", "filename", fileHeader.Filename, "error", err)
			continue
		}

		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			slog.Error("error reading file bytes", "filename", fileHeader.Filename, "error", err)
			continue
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		text, err := parseFileContent(ext, fileBytes)
		if err != nil {
			slog.Error("error parsing file content", "filename", fileHeader.Filename, "error", err)
			continue
		}

		allFileTexts = append(allFileTexts, text)
	}

	report, err := s.manager.GetReport(userId, reportId)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	if report.Status != schema.ReportCompleted {
		return nil, CodedError(errors.New("cannot process disclosures for report unless report status is complete"), http.StatusUnprocessableEntity)
	}

	updateDisclosures(report.Content.TalentContracts, allFileTexts)
	updateDisclosures(report.Content.AssociationsWithDeniedEntities, allFileTexts)
	updateDisclosures(report.Content.HighRiskFunders, allFileTexts)
	updateDisclosures(report.Content.AuthorAffiliations, allFileTexts)
	updateDisclosures(report.Content.PotentialAuthorAffiliations, allFileTexts)
	updateDisclosures(report.Content.MiscHighRiskAssociations, allFileTexts)
	updateDisclosures(report.Content.CoauthorAffiliations, allFileTexts)

	return report, nil
}

func parseFileContent(ext string, fileBytes []byte) (string, error) {
	switch ext {
	case ".txt":
		return string(fileBytes), nil
	case ".pdf":
		return extractTextFromPDF(fileBytes)
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func extractTextFromPDF(fileBytes []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder
	numPages := reader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		textBuilder.WriteString(pageText)
	}

	return textBuilder.String(), nil
}

func removeStopwords(text string) string {
	return stopwords.CleanString(text, "en", false)
}

func filterTokens(tokens []string) []string {
	var filtered []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		// If the token contains spaces, assume it’s a multi-word phrase and keep it as-is.
		if strings.Contains(token, " ") {
			filtered = append(filtered, strings.ToLower(token))
		} else {
			cleaned := removeStopwords(token)
			if cleaned != "" {
				filtered = append(filtered, cleaned)
			}
		}
	}
	return filtered
}
