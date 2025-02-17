package services

import (
	"net/http"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/search"
	"prism/prism/services/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type BackendService struct {
	report       ReportService
	search       SearchService
	autocomplete AutocompleteService
	licensing    LicenseService

	userAuth  auth.TokenVerifier
	adminAuth auth.TokenVerifier
}

func NewBackend(db *gorm.DB, oa openalex.KnowledgeBase, entityNdb search.NeuralDB, userAuth, adminAuth auth.TokenVerifier) *BackendService {
	return &BackendService{
		report: ReportService{
			manager: reports.NewManager(db),
			db:      db,
		},
		search: SearchService{
			openalex:  oa,
			entityNdb: entityNdb,
		},
		autocomplete: AutocompleteService{
			openalex: oa,
		},
		licensing: LicenseService{
			db: db,
		},
		userAuth:  userAuth,
		adminAuth: adminAuth,
	}
}

func (s *BackendService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.With(auth.Middleware(s.userAuth)).Mount("/report", s.report.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/search", s.search.Routes())
	r.With(auth.Middleware(s.userAuth)).Mount("/autocomplete", s.autocomplete.Routes())

	r.With(auth.Middleware(s.adminAuth)).Mount("/license", s.licensing.Routes())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
