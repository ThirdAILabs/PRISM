package services

import (
	"prism/services/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type BackendService struct {
	report       ReportService
	search       SearchService
	autocomplete AutocompleteService
	licensing    LicenseService

	userAuth  auth.TokenVerifier
	adminAuth auth.TokenVerifier
}

func (s *BackendService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.With(auth.Middleware(s.userAuth)).Mount("/report", s.report.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/search", s.search.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/autocomplete", s.autocomplete.Routes())

	r.With(auth.Middleware(s.adminAuth)).Mount("/license", s.licensing.Routes())

	return r
}
