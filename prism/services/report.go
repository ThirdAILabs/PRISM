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

	r.Get("/list", WrapRestHandler(s.List))
	r.Post("/new", WrapRestHandler(s.NewReport))
	r.Get("/{report_id}", WrapRestHandler(s.GetReport))

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

func (s *ReportService) NewReport(r *http.Request) (any, error) {
	userId, err := auth.GetUserId(r)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	params, err := ParseRequestBody[api.CreateReportRequest](r)
	if err != nil {
		return nil, CodedError(err, http.StatusBadRequest)
	}

	id, err := s.manager.CreateReport(userId, params.AuthorId, params.DisplayName, params.Source, params.StartYear, params.EndYear)
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
