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

	return r
}

func (s *AutocompleteService) AutocompleteAuthor(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	authors, err := s.openalex.AutocompleteAuthor(query)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	seen := make(map[string]bool)
	results := make([]api.Author, 0, len(authors))
	for _, author := range authors {
		if !seen[author.DisplayName] {
			results = append(results, api.Author{
				AuthorId:     author.AuthorId,
				AuthorName:   author.DisplayName,
				Institutions: author.InstitutionNames(),
				Source:       api.OpenAlexSource,
			})
			seen[author.DisplayName] = true
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
	results := make([]api.Institution, 0, len(institutions))
	for _, inst := range institutions {
		key := inst.InstitutionName + ";;" + inst.Location
		if !seen[key] {
			results = append(results, api.Institution{
				InstitutionId:   inst.InstitutionId,
				InstitutionName: inst.InstitutionName,
				Location:        inst.Location,
			})
			seen[key] = true
		}
	}

	return results, nil
}
