package services

import (
	"net/http"
	"prism/prism/api"
	"prism/prism/entity_search"
	"prism/prism/licensing"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/services/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type BackendService struct {
	report       ReportService
	search       SearchService
	autocomplete AutocompleteService

	userAuth auth.TokenVerifier
}

func NewBackend(db *gorm.DB, oa openalex.KnowledgeBase, entitySearch *entity_search.EntityIndex[api.MatchedEntity], userAuth auth.TokenVerifier, licensing *licensing.LicenseVerifier, resourceFolder string) *BackendService {
	return &BackendService{
		report: ReportService{
			manager:        reports.NewManager(db),
			db:             db,
			licensing:      licensing,
			resourceFolder: resourceFolder,
		},
		search: SearchService{
			openalex:     oa,
			entitySearch: entitySearch,
		},
		autocomplete: AutocompleteService{
			openalex: oa,
		},
		userAuth: userAuth,
	}
}

func (s *BackendService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.With(auth.Middleware(s.userAuth)).Mount("/report", s.report.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/search", s.search.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/autocomplete", s.autocomplete.Routes())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
