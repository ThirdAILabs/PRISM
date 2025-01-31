package services

import (
	"errors"
	"fmt"
	"net/http"
	"prism/api"
	"prism/reports"
	"prism/services/auth"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReportService struct {
	manager *reports.ReportManager
}

func (s *ReportService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/list", s.List)
	r.Post("/new", s.NewReport)
	r.Get("/{report_id}", s.GetReport)

	return r
}

func (s *ReportService) List(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reports, err := s.manager.ListReports(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, reports)
}

func (s *ReportService) NewReport(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params, err := ParseRequestBody[api.CreateReportRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.manager.CreateReport(userId, params.AuthorId, params.DisplayName, params.Source, params.StartYear, params.EndYear)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, api.CreateReportResponse{Id: id})
}

func (s *ReportService) GetReport(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	param := chi.URLParam(r, "report_id")
	id, err := uuid.Parse(param)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid uuid '%v' provided: %v", param, err), http.StatusBadRequest)
		return
	}

	report, err := s.manager.GetReport(userId, id)
	if err != nil {
		switch {
		case errors.Is(err, reports.ErrReportNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, reports.ErrUserCannotAccessReport):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	WriteJsonResponse(w, report)
}
