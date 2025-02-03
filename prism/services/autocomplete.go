package services

import (
	"net/http"
	"prism/openalex"

	"github.com/go-chi/chi/v5"
)

type AutocompleteService struct {
	openalex openalex.KnowledgeBase
}

func (s *AutocompleteService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/author", WrapRestHandler(s.AutocompleteAuthor))
	r.Get("/institution", WrapRestHandler(s.AutocompleteAuthor))

	return r
}

func (s *AutocompleteService) AutocompleteAuthor(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	authors, err := s.openalex.AutocompleteAuthor(query)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return authors, nil
}

func (s *AutocompleteService) AutocompleteInstitution(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	institutions, err := s.openalex.AutocompleteInstitution(query)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return institutions, nil
}
