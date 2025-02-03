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
	licensing    LicenseService

	userAuth  *auth.KeycloakAuth
	adminAuth *auth.KeycloakAuth
}

func (s *BackendService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.With(s.userAuth.Middleware()).Mount("/report", s.report.Routes())
	r.With(s.userAuth.Middleware()).Mount("/search", s.search.Routes())
	r.With(s.userAuth.Middleware()).Mount("/autocomplete", s.autocomplete.Routes())

	r.With(s.adminAuth.Middleware()).Mount("/license", s.licensing.Routes())

	return r
}
