package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"prism/api"

	"github.com/go-chi/chi/v5"
)

type AutocompleteService struct{}

func (s *AutocompleteService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/author", WrapRestHandler(s.AutocompleteAuthor))
	r.Get("/institution", WrapRestHandler(s.AutocompleteAuthor))

	return r
}

func openAlexAutocompletion(component, query string, dest interface{}) error {
	url := fmt.Sprintf("https://api.openalex.org/autocomplete/%s?q=%s&mailto=kartik@thirdai.com", component, url.QueryEscape(query))

	res, err := http.Get(url)
	if err != nil {
		slog.Error("autocompletion failed: open alex returned error", "query", query, "component", component, "error", err)
		return fmt.Errorf("unable to get autocomplete suggestions")
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(dest); err != nil {
		slog.Error("autocompletion failed: error parsing reponse from open alex", "query", query, "component", component, "error", err)
		return fmt.Errorf("unable to get autocomplete suggestions")
	}

	return nil
}

func (s *AutocompleteService) AutocompleteAuthor(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	var suggestions struct {
		Results []struct {
			Id          string `json:"id"`
			DisplayName string `json:"display_name"`
			Hint        string `json:"hint"`
		} `json:"results"`
	}

	if err := openAlexAutocompletion("authors", query, &suggestions); err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	authors := make([]api.Author, 0, len(suggestions.Results))
	for _, author := range suggestions.Results {
		authors = append(authors, api.Author{
			AuthorId:     author.Id,
			DisplayName:  author.DisplayName,
			Institutions: []string{author.Hint},
			Source:       api.OpenAlexSource,
		})
	}

	return authors, nil
}

func (s *AutocompleteService) AutocompleteInstitution(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	var suggestions struct {
		Results []struct {
			DisplayName string `json:"display_name"`
		} `json:"results"`
	}

	if err := openAlexAutocompletion("institutions", query, &suggestions); err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	institutions := make([]api.Institution, 0, len(suggestions.Results))
	for _, institution := range suggestions.Results {
		institutions = append(institutions, api.Institution{
			DisplayName: institution.DisplayName,
		})
	}

	return institutions, nil
}
