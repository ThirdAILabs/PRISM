package services

import (
	"net/http"
	"prism/api"
	"prism/openalex"

	"github.com/go-chi/chi/v5"
)

type AutocompleteService struct {
	openalex openalex.KnowledgeBase
}

func (s *AutocompleteService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/author", WrapRestHandler(s.AutocompleteAuthor))
	r.Get("/institution", WrapRestHandler(s.AutocompleteInstitution))

	return r
}

func (s *AutocompleteService) AutocompleteAuthor(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	authors, err := s.openalex.AutocompleteAuthor(query)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	results := make([]api.Author, 0, len(authors))
	for _, author := range authors {
		results = append(results, api.Author{
			AuthorId:     author.AuthorId,
			AuthorName:   author.DisplayName,
			Institutions: author.InstitutionNames(),
			Source:       api.OpenAlexSource,
		})
	}

	return results, nil
}

func (s *AutocompleteService) AutocompleteInstitution(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	institutions, err := s.openalex.AutocompleteInstitution(query)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	results := make([]api.Institution, 0, len(institutions))
	for _, inst := range institutions {
		results = append(results, api.Institution{
			InstitutionName: inst.InstitutionName,
		})
	}

	return results, nil
}
