package services

import (
	"prism/services/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type BackendService struct {
	report       ReportService
	search       SearchService
	autocomplete AutocomplenService

	auth *auth.KeycloakAuth
}

func (s *BackendService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(s.auth.Middleware())

	r.Mount("/report", s.report.Routes())
	r.Mount("/search", s.search.Routes())
	r.Mount("/autocomplete", s.autocomplete.Routes())

	return r
}
