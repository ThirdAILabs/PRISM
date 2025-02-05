package gscholar

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

func GoogleSearch(query string) (any, error) {
	url := fmt.Sprintf("https://serpapi.com/search.json?engine=google&q=%s&api_key=%s'", url.QueryEscape(query), apiKey)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("google search failed", "query", query, "error", err)
		return nil, ErrGoogleSearchFailed
	}
	defer res.Body.Close()

	var results struct {
		OrganicResults any `json:"organic_results"`
	}

	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		slog.Error("google search failed: error parsing reponse", "query", query, "error", err)
		return nil, ErrGoogleSearchFailed
	}

	return results.OrganicResults, nil
}
