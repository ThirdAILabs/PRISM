package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"prism/api"
	"slices"

	"github.com/go-chi/chi/v5"
)

type SearchService struct{}

func (s *SearchService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/regular", s.SearchOpenAlex)
	r.Get("/advanced", s.SearchGoogleScholar)
	r.Get("/formal-relations", s.FormalRelations)
	r.Get("/match-entities", s.MatchEntities)

	return r
}

func (s *SearchService) SearchOpenAlex(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	author, institution := query.Get("author"), query.Get("institution")

	url := fmt.Sprintf(
		"https://api.openalex.org/authors?filter=display_name.search:%s,affiliations.institution.id:%s&mailto=kartik@thirdai.com",
		url.QueryEscape(author), url.QueryEscape(institution),
	)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("open alex search failed", "author", author, "institution", institution, "error", err)
		http.Error(w, "error performing search", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	var results struct {
		Results []struct {
			Id            string `json:"id"`
			DisplayName   string `json:"diplay_name"`
			WorksCount    int    `json:"works_count"`
			Affilliations []struct {
				Institution struct {
					DisplayName string `json:"display_name"`
					CountryCode string `json:"country_code"`
				} `json:"institution"`
			} `json:"affiliations"`
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		slog.Error("open alex search failed: error parsing reponse from open alex", "query", query, "institution", institution, "error", err)
		http.Error(w, "error performing search", http.StatusInternalServerError)
		return
	}

	authors := make([]api.Author, 0, len(results.Results))
	for _, result := range results.Results {
		if result.WorksCount > 0 {
			institutions := make([]string, 0)
			for i, inst := range result.Affilliations {
				if i < 3 || (inst.Institution.CountryCode == "US" && !slices.Contains(institutions, inst.Institution.DisplayName)) {
					institutions = append(institutions, inst.Institution.DisplayName)
				}
			}

			authors = append(authors, api.Author{
				AuthorId:     result.Id,
				DisplayName:  result.DisplayName,
				Source:       api.OpenAlexSource,
				Institutions: institutions,
				WorksCount:   result.WorksCount,
			})
		}
	}

	WriteJsonResponse(w, authors)
}

func (s *SearchService) SearchGoogleScholar(w http.ResponseWriter, r *http.Request) {

}

func (s *SearchService) FormalRelations(w http.ResponseWriter, r *http.Request) {

}

func (s *SearchService) MatchEntities(w http.ResponseWriter, r *http.Request) {

}
