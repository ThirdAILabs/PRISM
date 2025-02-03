package openalex

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"prism/api"
	"slices"
)

var ErrSearchFailed = errors.New("error performing openalex search")

type RemoteOpenAlex struct{}

func autocompleteHelper(component, query string, dest interface{}) error {
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

func (oa *RemoteOpenAlex) AutocompleteAuthor(query string) ([]api.Author, error) {
	var suggestions struct {
		Results []struct {
			Id          string `json:"id"`
			DisplayName string `json:"display_name"`
			Hint        string `json:"hint"`
		} `json:"results"`
	}

	if err := autocompleteHelper("authors", query, &suggestions); err != nil {
		return nil, err
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

func (oa *RemoteOpenAlex) AutocompleteInstitution(query string) ([]api.Institution, error) {
	var suggestions struct {
		Results []struct {
			DisplayName string `json:"display_name"`
		} `json:"results"`
	}

	if err := autocompleteHelper("institutions", query, &suggestions); err != nil {
		return nil, err
	}

	institutions := make([]api.Institution, 0, len(suggestions.Results))
	for _, institution := range suggestions.Results {
		institutions = append(institutions, api.Institution{
			DisplayName: institution.DisplayName,
		})
	}

	return institutions, nil
}

func (oa *RemoteOpenAlex) FindAuthors(author, institution string) ([]api.Author, error) {
	url := fmt.Sprintf(
		"https://api.openalex.org/authors?filter=display_name.search:%s,affiliations.institution.id:%s&mailto=kartik@thirdai.com",
		url.QueryEscape(author), url.QueryEscape(institution),
	)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("open alex search failed", "author", author, "institution", institution, "error", err)
		return nil, ErrSearchFailed
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
		slog.Error("open alex search failed: error parsing reponse from author search", "author", author, "institution", institution, "error", err)
		return nil, ErrSearchFailed
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

	return authors, nil
}
