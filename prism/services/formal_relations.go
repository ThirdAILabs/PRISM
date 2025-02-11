// NOTE: this is currently unused, just keeping the code in case we want it in the future
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"prism/api"
	"prism/gscholar"
	"prism/llms"
	"strings"
)

var (
	ErrGoogleSearchFailed = errors.New("google search failed")
)

func GoogleSearch(query string) (any, error) {
	url := fmt.Sprintf("https://serpapi.com/search.json?engine=google&q=%s&api_key=%s'", url.QueryEscape(query), gscholar.ApiKey)

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

func findLinkInResponse(response string) (string, error) {
	start := max(strings.Index(response, "https://"), strings.Index(response, "http://"))
	if start < 0 {
		return "", fmt.Errorf("missing link in response")
	}
	return response[start:], nil // TODO(question): what happens if there's text after the link?
}

func fetchLink(link string) (string, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}

	headers := map[string]string{
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Accept-Language":           "en-US,en;q=0.9",
		"Connection":                "keep-alive",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Referer":                   "https://www.google.com/",
		"Upgrade-Insecure-Requests": "1",
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	const maxBytes = 300000

	text := make([]byte, maxBytes) // TODO(question): this seems like a lot?
	n, err := io.ReadFull(res.Body, text)
	if err != nil && err != io.ErrUnexpectedEOF { // UnexpectedEOF is returned if < N bytes are read
		return "", err
	}

	return string(text[:n]), nil
}

const (
	formalRelationsInitalResultsPromptTemplate = "These are the results of a Google search, in JSON format:\n\n%s\n\n" +
		"Based on the results above, infer whether %s has a formal position at %s. If there is not, strictly reply " +
		"\"I cannot answer\" and nothing else. HOWEVER, if there is, use the results to draw a conclusion " +
		"about the formal relationship between %s and %s. Answer briefly and INCLUDE THE LINK " +
		"associated with the snippet that allowed you to make that conclusion IN A SEPARATE LINE."

	formalRelationsVerficationPromptTemplate = "You previously thought %s has a formal position at %s based on this link %s\n\n" +
		"This is an html snippet from the link:\n\n%s\n\n" +
		"If you still agree with \"%s\", strictly answer \"yes\" and nothing else. Otherwise, strictly answer \"no\" and nothing else."
)

func (s *SearchService) FormalRelations(r *http.Request) (any, error) {
	query := r.URL.Query()
	author, institution := query.Get("author"), query.Get("institution")

	googleResults, err := GoogleSearch(fmt.Sprintf(`"%s" "%s"`, author, institution))
	if err != nil {
		slog.Error("formal relations: google search error", "error", err)
		return nil, CodedError(ErrSearchFailed, http.StatusInternalServerError)

	}

	resultsJson, err := json.MarshalIndent(googleResults, "", "    ")
	if err != nil {
		slog.Error("formal relations: error serializing google search results", "error", err)
		return nil, CodedError(ErrSearchFailed, http.StatusInternalServerError)
	}

	prompt := fmt.Sprintf(formalRelationsInitalResultsPromptTemplate, string(resultsJson), author, institution, author, institution)

	llm := llms.New()

	answer, err := llm.Generate(prompt, nil)
	if err != nil {
		slog.Error("formal relations: initial llm generaton failed", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)

	}

	if strings.Contains(strings.ToLower(answer), "i cannot answer") {
		return api.FormalRelationResponse{HasFormalRelation: false}, nil
	}

	link, err := findLinkInResponse(answer)
	if err != nil {
		slog.Error("formal relations: unable to find link in generated response", "error", err)
		return nil, CodedError(errors.New("invalid response from llm"), http.StatusInternalServerError)
	}

	content, err := fetchLink(link)
	if err != nil {
		slog.Error("formal relations: unable to fetch content from link", "link", link, "error", err)
		return nil, CodedError(errors.New("error loading content from link"), http.StatusInternalServerError)
	}

	verificationPrompt := fmt.Sprintf(formalRelationsVerficationPromptTemplate, author, institution, link, content, answer)

	verification, err := llm.Generate(verificationPrompt, nil)
	if err != nil {
		slog.Error("formal relations: verification llm generaton failed", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	hasRelation := !strings.Contains(strings.ToLower(verification), "no")
	return api.FormalRelationResponse{HasFormalRelation: hasRelation}, nil
}
