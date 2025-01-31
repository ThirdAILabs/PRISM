package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"prism/api"
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

const apiKey = "24e6015d3c452c7678abe92dd7e585b2cbcf2886b5b8ce7ed685d26e98d0930d"

type gscholarProfile struct {
	AuthorId      string  `json:"author_id"`
	Name          string  `json:"name"`
	Email         *string `json:"email"`
	Affilliations string  `json:"affiliations"`
}

func profileToAuthor(profile gscholarProfile) api.Author {
	return api.Author{
		AuthorId:     profile.AuthorId,
		DisplayName:  profile.Name,
		Institutions: strings.Split(profile.Affilliations, ", "),
		// TODO: should we use email?
		Source: api.GoogleScholarSource,
	}
}

type profilePageCrawler struct {
	query         string
	nextPageToken *string
	results       chan []api.Author
	successCnt    int
	errorCnt      int
}

func (crawler *profilePageCrawler) nextPageFilter() string {
	if crawler.nextPageToken != nil {
		return fmt.Sprintf("&after_author=%s", *crawler.nextPageToken) // Should this be escaped?
	}
	return ""
}

const (
	profilesUrlTemplate = `https://serpapi.com/search.json?engine=google_scholar_profiles&mauthors=%s&api_key=%s%s`
)

func (crawler *profilePageCrawler) nextPage() ([]api.Author, bool) {
	url := fmt.Sprintf(profilesUrlTemplate, url.QueryEscape(crawler.query), apiKey, crawler.nextPageFilter())

	res, err := http.Get(url)
	if err != nil {
		slog.Error("google scholar profile crawler: search failed", "query", crawler.query, "error", err)
		crawler.errorCnt++
		return nil, true
	}

	var results struct {
		Profiles   []gscholarProfile `json:"profiles"`
		Pagination struct {
			NextPageToken *string `json:"next_page_token"`
		} `json:"pagination"`
	}

	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		slog.Error("google scholar profile crawler: error parsing reponse", "query", crawler.query, "error", err)
		crawler.errorCnt++
		return nil, true
	}

	authors := make([]api.Author, 0, len(results.Profiles))
	for _, profile := range results.Profiles {
		authors = append(authors, profileToAuthor(profile))
	}

	crawler.nextPageToken = results.Pagination.NextPageToken
	crawler.successCnt++

	return authors, crawler.nextPageToken == nil
}

func (crawler *profilePageCrawler) run() {
	for {
		authors, done := crawler.nextPage()
		if len(authors) > 0 {
			crawler.results <- authors
		}
		if done {
			slog.Info("google scholar profile crawler: complete", "errors", crawler.errorCnt, "successes", crawler.successCnt)
			return
		}
	}
}

func getAuthorDetails(authorId string) (api.Author, error) {
	url := fmt.Sprintf("https://serpapi.com/search?engine=google_scholar_author&author_id=%s&num=0&start=0&sort=pubdate&api_key=%s", authorId, apiKey)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("google scholar author search failed", "author_id", authorId, "error", err)
		return api.Author{}, ErrSearchFailed
	}

	var result struct {
		Author gscholarProfile `json:"author"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		slog.Error("google scholar author search failed: error parsing reponse", "author_id", authorId, "error", err)
		return api.Author{}, ErrSearchFailed
	}

	return profileToAuthor(result.Author), nil
}

type gscholarCrawler struct {
	query      string
	nextIdx    int
	results    chan []api.Author
	successCnt int
	errorCnt   int
	seen       map[string]bool
}

const (
	nResultsPerPage     = 20
	gscholarUrlTemplate = `https://serpapi.com/search.json?engine=google_scholar&q=%s&api_key=%s&start=%d&num=%d`
)

func getAuthorId(link string) string {
	start := strings.Index(link, "user=")
	end := strings.Index(link[start:], "&")
	if start >= 0 && end >= 0 {
		return link[start+5 : end]
	}
	return ""
}

func authorNameInQuery(name string, queryTokens []string) bool {
	for _, part := range strings.Split(strings.ToLower(name), " ") {
		if slices.Contains(queryTokens, part) {
			return true
		}
	}
	return false
}

func (crawler *gscholarCrawler) nextPage() ([]api.Author, bool) {
	queryTokens := strings.Split(strings.ToLower(crawler.query), " ")

	authorIds := make([]string, 0)

	for i := 0; i < 5 && len(authorIds) == 0; i++ {
		url := fmt.Sprintf(gscholarUrlTemplate, url.QueryEscape(crawler.query), apiKey, crawler.nextIdx, nResultsPerPage)
		crawler.nextIdx += nResultsPerPage

		res, err := http.Get(url)
		if err != nil {
			slog.Error("google scholar crawler: search failed", "query", crawler.query, "error", err)
			crawler.errorCnt++
			return nil, true
		}

		var results struct {
			OrganicResults []struct {
				PublicationInfo struct {
					Authors []struct {
						Name string `json:"name"`
						Link string `json:"link"`
					} `json:"authors"`
				} `json:"publication_info"`
			} `json:"organic_results"`
		}

		if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
			slog.Error("google scholar crawler: error parsing reponse", "query", crawler.query, "error", err)
			crawler.errorCnt++
			return nil, true
		}

		for _, result := range results.OrganicResults {
			for _, author := range result.PublicationInfo.Authors {
				if authorNameInQuery(author.Name, queryTokens) {
					if authorId := getAuthorId(author.Link); authorId != "" && !crawler.seen[authorId] {
						crawler.seen[authorId] = true
						authorIds = append(authorIds, authorId)
					}
				}
			}
		}
	}

	if len(authorIds) == 0 {
		return nil, true
	}

	authorsCh := make(chan api.Author, len(authorIds))
	wg := sync.WaitGroup{}
	wg.Add(len(authorIds))

	for _, id := range authorIds {
		go func(authorId string) {
			defer wg.Done()
			details, err := getAuthorDetails(authorId)
			if err != nil {
				slog.Error("google scholar crawler: error getting author details", "author_id", authorId, "error", err)
			} else {
				authorsCh <- details
			}
		}(id)
	}

	wg.Wait()

	authors := make([]api.Author, 0, len(authorsCh))
	for author := range authorsCh {
		authors = append(authors, author)
	}

	if len(authors) == len(authorIds) {
		crawler.successCnt++
	} else {
		crawler.errorCnt++
		slog.Error("google scholar crawler: could not get author details for all author ids", "expected_cnt", len(authorIds), "actual_cnt", len(authors))
	}

	return authors, false
}

func (crawler *gscholarCrawler) run() {
	for {
		authors, done := crawler.nextPage()
		if len(authors) > 0 {
			crawler.results <- authors
		}
		if done {
			slog.Info("google scholar crawler: complete", "errors", crawler.errorCnt, "successes", crawler.successCnt)
			return
		}
	}
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

	v1Crawler := profilePageCrawler{
		query:         query,
		nextPageToken: nil,
		results:       resultsCh,
		successCnt:    0,
		errorCnt:      0,
	}

	v2Crawler := gscholarCrawler{
		query:      query,
		nextIdx:    0,
		results:    resultsCh,
		successCnt: 0,
		errorCnt:   0,
		seen:       make(map[string]bool),
	}

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
		v1Crawler.run()
	}()
	go func() {
		defer wg.Done()
		v2Crawler.run()
	}()

	wg.Done()
	close(stop)
}

func (s *SearchService) FormalRelations(w http.ResponseWriter, r *http.Request) {

}

func (s *SearchService) MatchEntities(w http.ResponseWriter, r *http.Request) {

}
