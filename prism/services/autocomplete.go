package services

import (
	"net/http"
	"prism/prism/api"
	"prism/prism/openalex"

	"github.com/go-chi/chi/v5"
)

type AutocompleteService struct {
	openalex openalex.KnowledgeBase
}

func (s *AutocompleteService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/author", WrapRestHandler(s.AutocompleteAuthor))
	r.Get("/institution", WrapRestHandler(s.AutocompleteInstitution))
	r.Get("/paper", WrapRestHandler(s.AutocompletePaper))

	return r
}

func (s *AutocompleteService) AutocompleteAuthor(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")
	institutionId := r.URL.Query().Get("institution_id")

	authors, err := s.openalex.AutocompleteAuthor(query, institutionId)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	seen := make(map[string]bool)
	results := make([]api.Autocompletion, 0, len(authors))
	for _, author := range authors {
		if !seen[author.Name] {
			results = append(results, author)
			seen[author.Name] = true
		}
	}

	return results, nil
}

func (s *AutocompleteService) AutocompleteInstitution(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	institutions, err := s.openalex.AutocompleteInstitution(query)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	seen := make(map[string]bool)
	results := make([]api.Autocompletion, 0, len(institutions))
	for _, inst := range institutions {
		key := inst.Name + ";;" + inst.Hint
		if !seen[key] {
			results = append(results, inst)
			seen[key] = true
		}
	}

	return results, nil
}

func (s *AutocompleteService) AutocompletePaper(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	institutions, err := s.openalex.AutocompletePaper(query)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return institutions, nil
}
