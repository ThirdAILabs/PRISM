package services

import (
	"net/http"
	"prism/prism/monitoring"
	"prism/prism/services/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type BackendService struct {
	report       ReportService
	search       SearchService
	autocomplete AutocompleteService
	hooks        HookService

	userAuth auth.TokenVerifier
}

func NewBackend(report ReportService, search SearchService, autocomplete AutocompleteService, hooks HookService, userAuth auth.TokenVerifier) *BackendService {
	return &BackendService{
		report:       report,
		search:       search,
		autocomplete: autocomplete,
		hooks:        hooks,
		userAuth:     userAuth,
	}
}

func (s *BackendService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(monitoring.HandlerMetrics)
	r.Use(middleware.Recoverer)

	r.With(auth.Middleware(s.userAuth)).Mount("/report", s.report.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/search", s.search.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/autocomplete", s.autocomplete.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/hooks", s.hooks.Routes())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
