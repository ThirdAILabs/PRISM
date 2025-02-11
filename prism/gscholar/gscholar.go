package gscholar

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
)

const ApiKey = "24e6015d3c452c7678abe92dd7e585b2cbcf2886b5b8ce7ed685d26e98d0930d"

var (
	ErrGoogleScholarSearchFailed = errors.New("google scholar search failed")
)

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

type ProfilePageCrawler struct {
	query         string
	nextPageToken *string
	results       chan []api.Author
	successCnt    int
	errorCnt      int
}

func NewProfilePageCrawler(query string, results chan []api.Author) *ProfilePageCrawler {
	return &ProfilePageCrawler{
		query:         query,
		nextPageToken: nil,
		results:       results,
		successCnt:    0,
		errorCnt:      0,
	}
}

func (crawler *ProfilePageCrawler) nextPageFilter() string {
	if crawler.nextPageToken != nil {
		return fmt.Sprintf("&after_author=%s", *crawler.nextPageToken) // Should this be escaped?
	}
	return ""
}

const (
	profilesUrlTemplate = `https://serpapi.com/search.json?engine=google_scholar_profiles&mauthors=%s&api_key=%s%s`
)

func (crawler *ProfilePageCrawler) nextPage() ([]api.Author, bool) {
	url := fmt.Sprintf(profilesUrlTemplate, url.QueryEscape(crawler.query), ApiKey, crawler.nextPageFilter())

	res, err := http.Get(url)
	if err != nil {
		slog.Error("google scholar profile crawler: search failed", "query", crawler.query, "error", err)
		crawler.errorCnt++
		return nil, true
	}
	defer res.Body.Close()

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

func (crawler *ProfilePageCrawler) Run() {
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
	url := fmt.Sprintf("https://serpapi.com/search?engine=google_scholar_author&author_id=%s&num=0&start=0&sort=pubdate&api_key=%s", authorId, ApiKey)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("google scholar author search failed", "author_id", authorId, "error", err)
		return api.Author{}, ErrGoogleScholarSearchFailed
	}
	defer res.Body.Close()

	var result struct {
		Author gscholarProfile `json:"author"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		slog.Error("google scholar author search failed: error parsing reponse", "author_id", authorId, "error", err)
		return api.Author{}, ErrGoogleScholarSearchFailed
	}

	return profileToAuthor(result.Author), nil
}

type GScholarCrawler struct {
	query      string
	nextIdx    int
	results    chan []api.Author
	successCnt int
	errorCnt   int
	seen       map[string]bool
}

func NewGScholarCrawler(query string, results chan []api.Author) *GScholarCrawler {
	return &GScholarCrawler{
		query:      query,
		nextIdx:    0,
		results:    results,
		successCnt: 0,
		errorCnt:   0,
		seen:       make(map[string]bool),
	}
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

func (crawler *GScholarCrawler) nextPage() ([]api.Author, bool) {
	queryTokens := strings.Split(strings.ToLower(crawler.query), " ")

	authorIds := make([]string, 0)

	for i := 0; i < 5 && len(authorIds) == 0; i++ {
		url := fmt.Sprintf(gscholarUrlTemplate, url.QueryEscape(crawler.query), ApiKey, crawler.nextIdx, nResultsPerPage)
		crawler.nextIdx += nResultsPerPage

		res, err := http.Get(url)
		if err != nil {
			slog.Error("google scholar crawler: search failed", "query", crawler.query, "error", err)
			crawler.errorCnt++
			return nil, true
		}
		defer res.Body.Close()

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

func (crawler *GScholarCrawler) Run() {
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

type gscholarPaper struct {
	Title string `json:"title"`
}

type AuthorPaperIterator struct {
	authorId string
	start    int
	stopped  bool
}

func NewAuthorPaperIterator(authorId string) *AuthorPaperIterator {
	return &AuthorPaperIterator{authorId: authorId, start: 0, stopped: false}
}

func (iter *AuthorPaperIterator) Next() ([]string, error) {
	if iter.stopped {
		return nil, nil
	}

	const batchSize = 100

	url := fmt.Sprintf("https://serpapi.com/search?engine=google_scholar_author&author_id=%s&num=%d&start=%d&sort=pubdate&api_key=%s", iter.authorId, batchSize, iter.start, ApiKey)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("google scholar: error getting author papers", "author_id", iter.authorId, "error", err)
		return nil, fmt.Errorf("google scholar error: %w", err)
	}

	var results struct {
		Articles []gscholarPaper `json:"articles"`
	}

	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		slog.Error("google scholar: error parsing papers reponse", "author_id", iter.authorId, "error", err)
		return nil, fmt.Errorf("error parsing response from google scholar: %w", err)
	}

	titles := make([]string, 0, len(results.Articles))
	for _, paper := range results.Articles {
		titles = append(titles, paper.Title)
	}

	if len(results.Articles) < batchSize {
		iter.stopped = true
	}
	iter.start += batchSize

	return titles, nil
}
