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
	"prism/llms"
	"prism/services/utils"
	"slices"
	"strings"
	"sync"

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

var ErrSearchFailed = errors.New("error performing search")

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
		http.Error(w, ErrSearchFailed.Error(), http.StatusInternalServerError)
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
		slog.Error("open alex search failed: error parsing reponse from", "query", query, "institution", institution, "error", err)
		http.Error(w, ErrSearchFailed.Error(), http.StatusInternalServerError)
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
	query := strings.ReplaceAll(strings.ToLower(r.URL.Query().Get("query")), "@", " ")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "http response does not support chunking", http.StatusInternalServerError)
		return
	}

	seen := make(map[string]bool)

	resultsCh := make(chan []api.Author, 10)

	v1Crawler := utils.NewProfilePageCrawler(query, resultsCh)

	v2Crawler := utils.NewGScholarCrawler(query, resultsCh)

	stop := make(chan bool)

	go func() {
		for {
			select {
			case authors := <-resultsCh:
				unseenAuthors := make([]api.Author, 0)
				for _, author := range authors {
					if !seen[author.AuthorId] {
						seen[author.AuthorId] = true
						unseenAuthors = append(unseenAuthors, author)
					}
				}

				if len(unseenAuthors) > 0 {
					if err := json.NewEncoder(w).Encode(unseenAuthors); err != nil {
						slog.Error("error sending authors chunk", "error", err)
						http.Error(w, "error sending response chunk", http.StatusInternalServerError)
						return
					}
					flusher.Flush()
				}
			case <-stop:
				break
			}
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		v1Crawler.Run()
	}()
	go func() {
		defer wg.Done()
		v2Crawler.Run()
	}()

	wg.Done()
	close(stop)
}

func findLinkInResponse(response string) (string, error) {
	start := max(strings.Index(response, "https://"), strings.Index(response, "http://"))
	if start < 0 {
		return "", fmt.Errorf("missing link in response")
	}
	return response[start:], nil // TODO(reliability): what happens if there's text after the link?
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

	text := make([]byte, maxBytes) // TODO(reliability): this seems like a lot?
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

func (s *SearchService) FormalRelations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	author, institution := query.Get("author"), query.Get("institution")

	googleResults, err := utils.GoogleSearch(fmt.Sprintf(`"%s" "%s"`, author, institution))
	if err != nil {
		slog.Error("formal relations: google search error", "error", err)
		http.Error(w, ErrSearchFailed.Error(), http.StatusInternalServerError)
		return
	}

	resultsJson, err := json.MarshalIndent(googleResults, "", "    ")
	if err != nil {
		slog.Error("formal relations: error serializing google search results", "error", err)
		http.Error(w, ErrSearchFailed.Error(), http.StatusInternalServerError)
		return
	}

	prompt := fmt.Sprintf(formalRelationsInitalResultsPromptTemplate, string(resultsJson), author, institution, author, institution)

	llm := llms.New()

	answer, err := llm.Generate(prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(strings.ToLower(answer), "i cannot answer") {
		WriteJsonResponse(w, api.FormalRelationResponse{HasFormalRelation: false})
		return
	}

	link, err := findLinkInResponse(answer)
	if err != nil {
		slog.Error("formal relations: unable to find link in generated response", "error", err)
		http.Error(w, "invalid response from llm", http.StatusInternalServerError)
		return
	}

	content, err := fetchLink(link)
	if err != nil {
		slog.Error("formal relations: unable to fetch content from link", "link", link, "error", err)
		http.Error(w, "error loading content from link", http.StatusInternalServerError)
		return
	}

	verificationPrompt := fmt.Sprintf(formalRelationsVerficationPromptTemplate, author, institution, link, content, answer)

	verification, err := llm.Generate(verificationPrompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasRelation := !strings.Contains(strings.ToLower(verification), "no")
	WriteJsonResponse(w, api.FormalRelationResponse{HasFormalRelation: hasRelation})
}

func (s *SearchService) MatchEntities(w http.ResponseWriter, r *http.Request) {

}
