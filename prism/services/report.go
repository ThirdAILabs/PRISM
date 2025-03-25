package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/licensing"
	"prism/prism/reports"
	"prism/prism/schema"
	"prism/prism/services/auth"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReportService struct {
	manager        *reports.ReportManager
	licensing      *licensing.LicenseVerifier
	resourceFolder string
}

type ReportRequest struct {
	Id                    string                     `json:"Id"`
	LastAccessedAt        string                     `json:"LastAccessedAt"`
	AuthorId              string                     `json:"AuthorId"`
	AuthorName            string                     `json:"AuthorName"`
	Source                string                     `json:"Source"`
	Status                string                     `json:"Status"`
	Content               map[string]json.RawMessage `json:"Content"`
	ContainsReportContent bool                       `json:"ContainsReportContent"`
	ContainsDisclosure    bool                       `json:"ContainsDisclosure"`
	TimeRange             string                     `json:"TimeRange"`
}

func (request *ReportRequest) Validate() error {

	if !request.ContainsReportContent {
		return nil
	}

	if request.Id == "" {
		return fmt.Errorf("report id is required")
	}

	if request.LastAccessedAt == "" {
		return fmt.Errorf("last accessed at is required")
	}

	if request.AuthorId == "" {
		return fmt.Errorf("author id is required")
	}

	if request.AuthorName == "" {
		return fmt.Errorf("author name is required")
	}

	if request.Source == "" {
		return fmt.Errorf("source is required")
	}

	if request.Status == "" {
		return fmt.Errorf("status is required")
	}

	if request.Content == nil {
		return fmt.Errorf("content is required")
	}
	return nil
}

func convertReportRequestToReport(request *ReportRequest) (api.Report, string, error) {

	id, err := uuid.Parse(request.Id)
	if err != nil {
		return api.Report{}, "", fmt.Errorf("error parsing report id: %w", err)
	}

	lastAccessedAt, err := time.Parse(time.RFC3339, request.LastAccessedAt)
	if err != nil {
		return api.Report{}, "", fmt.Errorf("error parsing report last accessed at: %w", err)
	}

	report := api.Report{
		Id:             id,
		LastAccessedAt: lastAccessedAt,
		AuthorId:       request.AuthorId,
		AuthorName:     request.AuthorName,
		Source:         request.Source,
		Status:         request.Status,
		Content:        make(map[string][]api.Flag),
	}

	for flagType, flagsData := range request.Content {
		var flagsArray []json.RawMessage
		if err := json.Unmarshal(flagsData, &flagsArray); err != nil {
			return api.Report{}, "", fmt.Errorf("error deserializing flags: %w", err)
		}

		flags := make([]api.Flag, 0, len(flagsArray))
		for _, rawFlag := range flagsArray {
			flag, err := api.ParseFlag(flagType, rawFlag)
			if err != nil {
				return api.Report{}, "", fmt.Errorf("error creating empty flag of the type %s: %w", flagType, err)
			}

			if err := json.Unmarshal(rawFlag, flag); err != nil {
				return api.Report{}, "", fmt.Errorf("error deserializing flag of the type %s: %w", flagType, err)
			}
			flags = append(flags, flag)
		}
		report.Content[flagType] = flags
	}

	return report, request.TimeRange, nil
}

func (s *ReportService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/author", func(r chi.Router) {
		r.Get("/list", WrapRestHandler(s.List))
		r.Post("/create", WrapRestHandler(s.CreateReport))
		r.Get("/{report_id}", WrapRestHandler(s.GetReport))
		r.Delete("/{report_id}", WrapRestHandler(s.DeleteAuthorReport))
		r.Post("/{report_id}/check-disclosure", WrapRestHandler(s.CheckDisclosure))
		r.Post("/{report_id}/download", s.DownloadReport)
		r.Post("/{report_id}/feedback", WrapRestHandler(s.LogFlagFeedback))
	})

	r.Route("/university", func(r chi.Router) {
		r.Get("/list", WrapRestHandler(s.ListUniversityReports))
		r.Post("/create", WrapRestHandler(s.CreateUniversityReport))
		r.Get("/{report_id}", WrapRestHandler(s.GetUniversityReport))
		r.Delete("/{report_id}", WrapRestHandler(s.DeleteUniversityReport))
	})

	return r
}

func (s *ReportService) List(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reports, err := s.manager.ListAuthorReports(userId)
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

	params, err := ParseRequestBody[api.CreateAuthorReportRequest](r)
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

	if err := s.licensing.VerifyLicense(); err != nil {
		slog.Error("cannot create new report, unable to verify license", "error", err)
		return nil, CodedError(err, licensingErrorStatus(err))
	}

	id, err := s.manager.CreateAuthorReport(userId, params.AuthorId, params.AuthorName, params.Source)
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

	id, err := URLParamUUID(r, "report_id")
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	report, err := s.manager.GetAuthorReport(userId, id)
	if err != nil {
		return nil, CodedError(err, reportErrorStatus(err))
	}

	return report, nil
}

func (s *ReportService) DeleteAuthorReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	id, err := URLParamUUID(r, "report_id")
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if err := s.manager.DeleteAuthorReport(userId, id); err != nil {
		return nil, CodedError(err, reportErrorStatus(err))
	}

	return nil, nil
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

	report, err := s.manager.GetAuthorReport(userId, reportId)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	if report.Status != schema.ReportCompleted {
		return nil, CodedError(errors.New("cannot process disclosures for report unless report status is complete"), http.StatusUnprocessableEntity)
	}

	for _, flags := range report.Content {
		for _, flag := range flags {
			updateFlagDisclosure(flag, allFileTexts)
		}
	}
	return report, nil
}

func (s *ReportService) DownloadReport(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reportId, err := URLParamUUID(r, "report_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var requestBody ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if err := requestBody.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	containsReportContent := requestBody.ContainsReportContent

	var report api.Report
	var timeRange string
	if containsReportContent {
		report, timeRange, err = convertReportRequestToReport(&requestBody)
	} else {
		report, err = s.manager.GetAuthorReport(userId, reportId)
	}
	if err != nil {
		http.Error(w, err.Error(), reportErrorStatus(err))
		return
	}

	if report.Status != schema.ReportCompleted {
		http.Error(w, "cannot download report unless report status is complete", http.StatusUnprocessableEntity)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	var fileBytes []byte
	var contentType, filename string
	switch format {
	case "csv":
		fileBytes, err = generateCSV(report)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		contentType = "text/csv"
		filename = fmt.Sprintf("%s Report.csv", report.AuthorName)
	case "pdf":
		fileBytes, err = generatePDF(report, s.resourceFolder, requestBody.ContainsDisclosure, timeRange)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		contentType = "application/pdf"
		filename = fmt.Sprintf("%s Report.pdf", report.AuthorName)
	case "excel", "xlsx":
		fileBytes, err = generateExcel(report)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = fmt.Sprintf("%s Report.xlsx", report.AuthorName)
	default:
		http.Error(w, "unsupported format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(fileBytes); err != nil {
		slog.Error("error writing file bytes", "error", err)
		http.Error(w, "error writing file", http.StatusInternalServerError)
	}
}

func (s *ReportService) ListUniversityReports(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	reports, err := s.manager.ListUniversityReports(userId)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return reports, nil
}

func (s *ReportService) CreateUniversityReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	params, err := ParseRequestBody[api.CreateUniversityReportRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if params.UniversityId == "" {
		return nil, CodedError(errors.New("UniversityId must be specified"), http.StatusUnprocessableEntity)
	}

	if params.UniversityName == "" {
		return nil, CodedError(errors.New("UniversityName must be specified"), http.StatusUnprocessableEntity)
	}

	if err := s.licensing.VerifyLicense(); err != nil {
		slog.Error("cannot create new report, unable to verify license", "error", err)
		return nil, CodedError(err, licensingErrorStatus(err))
	}

	id, err := s.manager.CreateUniversityReport(userId, params.UniversityId, params.UniversityName)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return api.CreateReportResponse{Id: id}, nil
}

func (s *ReportService) GetUniversityReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	id, err := URLParamUUID(r, "report_id")
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	report, err := s.manager.GetUniversityReport(userId, id)
	if err != nil {
		return nil, CodedError(err, reportErrorStatus(err))
	}

	return report, nil
}

func (s *ReportService) DeleteUniversityReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	id, err := URLParamUUID(r, "report_id")
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if err := s.manager.DeleteUniversityReport(userId, id); err != nil {
		return nil, CodedError(err, reportErrorStatus(err))
	}

	return nil, nil
}

func (s *ReportService) LogFlagFeedback(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, err
	}

	reportId, err := URLParamUUID(r, "report_id")
	if err != nil {
		return nil, err
	}

	params, err := ParseRequestBody[api.FlagFeedbackRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	if err := s.manager.SaveFlagFeedback(reportId, userId, params.FlagHash, params.Feedback); err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return nil, nil
}
