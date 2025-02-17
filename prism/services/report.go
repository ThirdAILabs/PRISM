package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/reports"
	"prism/prism/services/auth"
	"prism/prism/services/licensing"
	"strings"
	"time"

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

	if params.StartYear == 0 {
		params.StartYear = time.Now().Year() - 4
	}

	if params.EndYear == 0 {
		params.EndYear = time.Now().Year()
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

	id, err := s.manager.CreateReport(licenseId, userId, params.AuthorId, params.AuthorName, params.Source, params.StartYear, params.EndYear)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return api.CreateReportResponse{Id: id}, nil
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

	report, err := s.manager.GetReport(userId, id, false)
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

	params, err := ParseRequestBody[api.AddLicenseUserRequest](r)
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

func (s *ReportService) CheckDisclosure(r *http.Request) (any, error) {
	// Authenticate the user.
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	// Get the report ID from the URL.
	reportIdParam := chi.URLParam(r, "report_id")
	reportId, err := uuid.Parse(reportIdParam)
	if err != nil {
		return nil, CodedError(fmt.Errorf("invalid report id '%v': %w", reportIdParam, err), http.StatusBadRequest)
	}

	// Parse the multipart form (limit 32 MB).
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	// Retrieve files under the field "files".
	fileHeaders := r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		return nil, CodedError(errors.New("no files uploaded"), http.StatusBadRequest)
	}

	// Aggregate text extracted from all uploaded files.
	var allFileTexts []string
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			slog.Error("error opening uploaded file", "filename", fileHeader.Filename, "error", err)
			continue // or return error
		}

		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			slog.Error("error reading file bytes", "filename", fileHeader.Filename, "error", err)
			continue
		}

		// Determine file extension.
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		text, err := parseFileContent(ext, fileBytes)
		if err != nil {
			slog.Error("error parsing file content", "filename", fileHeader.Filename, "error", err)
			continue
		}

		allFileTexts = append(allFileTexts, text)
	}

	// Retrieve the stored report.
	report, err := s.manager.GetReport(userId, reportId, true)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	// Since convertReport stored the content as a ReportContent, we assert that:
	rc, ok := report.Content.(reports.ReportContent)
	if !ok {
		return nil, CodedError(fmt.Errorf("unexpected content type"), http.StatusInternalServerError)
	}

	// Now you can work directly with rc.Connections.
	// For each connection group, update the Disclosed slice as needed:
	for i, connField := range rc.Connections {
		disclosedSlice := make([]bool, len(connField.Connections))
		for j, connection := range connField.Connections {
			matchFound := false
			titleLower := strings.ToLower(connection.Title)
			for _, txt := range allFileTexts { // assuming you already built this slice from uploaded files
				if strings.Contains(strings.ToLower(txt), titleLower) {
					matchFound = true
					break
				}
			}

			// If no match in the title, then check the corresponding detail if available.
			if !matchFound && j < len(connField.Details) {
				tokens := extractTokens(connField.Details[j])
				for _, token := range tokens {
					tokenLower := strings.ToLower(token)
					for _, txt := range allFileTexts {
						if strings.Contains(strings.ToLower(txt), tokenLower) {
							matchFound = true
							break
						}
					}
					if matchFound {
						break
					}
				}
			}
			disclosedSlice[j] = matchFound
		}
		rc.Connections[i].Disclosed = disclosedSlice
	}

	// Update the report with the modified ReportContent.
	updatedContentBytes, err := json.Marshal(rc)
	if err != nil {
		slog.Error("error serializing updated report content", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	if err := s.manager.UpdateReport(report.Id, "complete", updatedContentBytes); err != nil {
		slog.Error("error updating report with disclosure information", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	for i := range rc.Connections {
		rc.Connections[i].Details = nil
	}

	// Optionally, update the report's Content field with our modified content.
	report.Content = rc

	return report, nil

}

// parseFileContent is a stub to extract text from a file based on its extension.
// Implement actual parsing logic for different file types as needed.
func parseFileContent(ext string, fileBytes []byte) (string, error) {
	switch ext {
	case ".txt":
		return string(fileBytes), nil
	case ".pdf":
		return extractTextFromPDF(fileBytes)
	case ".docx":
		// TODO: Implement DOCX text extraction.
		return "", errors.New("DOCX parsing not implemented")
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

func extractTokens(detail interface{}) []string {
	var tokens []string
	switch v := detail.(type) {
	case string:
		// Split the string into words.
		tokens = append(tokens, strings.Fields(v)...)
	case []interface{}:
		for _, item := range v {
			tokens = append(tokens, extractTokens(item)...)
		}
	case map[string]interface{}:
		for _, val := range v {
			tokens = append(tokens, extractTokens(val)...)
		}
	default:
		// Fallback: try to marshal to JSON and split (not ideal, but something)
		if b, err := json.Marshal(v); err == nil {
			tokens = append(tokens, strings.Fields(string(b))...)
		}
	}
	return tokens
}
