package gscholar

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"prism/prism/api"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
)

const ApiKey = "24e6015d3c452c7678abe92dd7e585b2cbcf2886b5b8ce7ed685d26e98d0930d"

var client = resty.New().
	SetBaseURL("https://serpapi.com").
	SetQueryParam("api_key", ApiKey).
	AddRetryCondition(func(response *resty.Response, err error) bool {
		if err != nil {
			return true // The err can be non nil for some network errors.
		}
		// There's no reason to retry other 400 requests since the outcome should not change
		return response != nil && (response.StatusCode() > 499 || response.StatusCode() == http.StatusTooManyRequests)
	}).
	SetRetryCount(2)

var (
	ErrGoogleScholarSearchFailed = errors.New("google scholar search failed")
	ErrInvalidCursor             = errors.New("invalid cursor")
	ErrCursorCreationFailed      = errors.New("cursor creation failed")
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
		AuthorName:   profile.Name,
		Institutions: strings.Split(profile.Affilliations, ", "),
		Source:       api.GoogleScholarSource,
	}
}

func nextGScholarPageV1(query string, nextPageToken *string) ([]api.Author, *string, error) {
	type gscholarResults struct {
		Profiles   []gscholarProfile `json:"profiles"`
		Pagination struct {
			NextPageToken *string `json:"next_page_token"`
		} `json:"pagination"`
	}

	params := map[string]string{
		"engine":   "google_scholar_profiles",
		"mauthors": query,
	}
	if nextPageToken != nil {
		params["after_author"] = *nextPageToken
	}

	res, err := client.R().
		SetResult(&gscholarResults{}).
		SetQueryParams(params).
		Get("/search.json")

	if err != nil {
		slog.Error("google scholar profile search: search failed", "query", query, "error", err)
		return nil, nil, ErrGoogleScholarSearchFailed
	}

	if !res.IsSuccess() {
		slog.Error("google scholar profile search returned error", "status_code", res.StatusCode(), "body", res.String())
		return nil, nil, ErrGoogleScholarSearchFailed
	}

	results := res.Result().(*gscholarResults)

	authors := make([]api.Author, 0, len(results.Profiles))
	for _, profile := range results.Profiles {
		authors = append(authors, profileToAuthor(profile))
	}

	return authors, results.Pagination.NextPageToken, nil
}

func getAuthorDetails(authorId string) (api.Author, error) {
	type authorDetailsResult struct {
		Author gscholarProfile `json:"author"`
	}

	res, err := client.R().
		SetResult(&authorDetailsResult{}).
		SetQueryParam("engine", "google_scholar_author").
		SetQueryParam("author_id", authorId).
		SetQueryParam("num", "0").
		SetQueryParam("start", "0").
		SetQueryParam("sort", "pubdate").
		Get("/search")

	if err != nil {
		slog.Error("google scholar author search failed", "author_id", authorId, "error", err)
		return api.Author{}, ErrGoogleScholarSearchFailed
	}

	if !res.IsSuccess() {
		slog.Error("google scholar author search returned error", "status_code", res.StatusCode(), "body", res.String())
		return api.Author{}, ErrGoogleScholarSearchFailed
	}

	result := res.Result().(*authorDetailsResult)
	result.Author.AuthorId = authorId // For some reason this isn't part of the response from the endpoint

	return profileToAuthor(result.Author), nil
}

func getAuthorId(link string) string {
	start := strings.Index(link, "user=")
	distToEnd := strings.Index(link[start:], "&")
	if start >= 0 && distToEnd >= 0 {
		return link[start+5 : start+distToEnd]
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

func nextGScholarPageV2(query string, nextIdx *int, seen map[string]bool) ([]api.Author, *int, error) {
	type gscholarResults struct {
		OrganicResults []struct {
			PublicationInfo struct {
				Authors []struct {
					Name string `json:"name"`
					Link string `json:"link"`
				} `json:"authors"`
			} `json:"publication_info"`
		} `json:"organic_results"`
	}
	const nResultsPerPage = 20

	queryTokens := strings.Split(strings.ToLower(query), " ")

	authorIds := make([]string, 0)

	startIdx := 0
	if nextIdx != nil {
		startIdx = *nextIdx
	}

	seenInbatch := make(map[string]bool)

	for i := 0; i < 5 && len(authorIds) == 0; i++ {
		res, err := client.R().
			SetResult(&gscholarResults{}).
			SetQueryParam("engine", "google_scholar").
			SetQueryParam("q", query).
			SetQueryParam("start", strconv.Itoa(startIdx)).
			SetQueryParam("num", strconv.Itoa(nResultsPerPage)).
			Get("/search.json")

		if err != nil {
			slog.Error("google scholar search: search failed", "query", query, "error", err)
			return nil, nil, ErrGoogleScholarSearchFailed
		}

		startIdx += nResultsPerPage

		if !res.IsSuccess() {
			slog.Error("google scholar search returned error", "status_code", res.StatusCode(), "body", res.String())
			return nil, nil, ErrGoogleScholarSearchFailed
		}

		results := res.Result().(*gscholarResults)

		for _, result := range results.OrganicResults {
			for _, author := range result.PublicationInfo.Authors {
				if authorNameInQuery(author.Name, queryTokens) {
					if authorId := getAuthorId(author.Link); authorId != "" && !seen[authorId] && !seenInbatch[authorId] {
						seenInbatch[authorId] = true
						authorIds = append(authorIds, authorId)
					}
				}
			}
		}

		if len(authorIds) > 0 {
			break
		}
	}

	if len(authorIds) == 0 {
		return nil, nil, nil
	}

	authorsCh := make(chan api.Author, len(authorIds))
	wg := sync.WaitGroup{}
	wg.Add(len(authorIds))

	for _, id := range authorIds {
		go func(authorId string) {
			defer wg.Done()
			details, err := getAuthorDetails(authorId)
			if err != nil {
				slog.Error("google scholar search: error getting author details", "author_id", authorId, "error", err)
			} else {
				authorsCh <- details
			}
		}(id)
	}

	wg.Wait()
	close(authorsCh)

	authors := make([]api.Author, 0, len(authorsCh))
	for author := range authorsCh {
		authors = append(authors, author)
	}

	if len(authors) != len(authorIds) {
		slog.Error("google scholar search: could not get author details for all author ids", "expected_cnt", len(authorIds), "actual_cnt", len(authors))
		if len(authors) == 0 {
			// If some of the calls succeed we long the errors and return success. If they all fail we return an error.
			return nil, nil, ErrGoogleScholarSearchFailed
		}
	}

	return authors, &startIdx, nil
}

type cursorPayload struct {
	V1Cursor *string
	V2Cursor *int
	Seen     []string
}

func parseCursor(token string) (cursorPayload, error) {
	if token == "" {
		return cursorPayload{}, nil
	}

	bytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		slog.Error("unable to decode cursor token", "error", err)
		return cursorPayload{}, ErrInvalidCursor
	}

	var payload cursorPayload
	if err := json.Unmarshal(bytes, &payload); err != nil {
		slog.Error("unable to parse cursor token", "error", err)
		return cursorPayload{}, ErrInvalidCursor
	}

	return payload, nil
}

func encodeCursor(payload cursorPayload) (string, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("error encoding cursor to token", "error", err)
		return "", ErrCursorCreationFailed
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func NextGScholarPage(query, cursorToken string) ([]api.Author, string, error) {
	cursor, err := parseCursor(cursorToken)
	if err != nil {
		return nil, "", err
	}

	seen := make(map[string]bool)
	for _, x := range cursor.Seen {
		seen[x] = true
	}

	var v1Results, v2Results []api.Author
	var v1Cursor *string
	var v2Cursor *int
	var v1Err, v2Err error

	wg := sync.WaitGroup{}

	if len(seen) == 0 || cursor.V1Cursor != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v1Results, v1Cursor, v1Err = nextGScholarPageV1(query, cursor.V1Cursor)
		}()
	}

	if len(seen) == 0 || cursor.V2Cursor != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v2Results, v2Cursor, v2Err = nextGScholarPageV2(query, cursor.V2Cursor, seen)
		}()
	}

	wg.Wait()

	if v1Err != nil {
		slog.Error("v1 gscholar search failed", "error", v1Err)
	}

	if v2Err != nil {
		slog.Error("v2 gscholar search failed", "error", v2Err)
	}

	if v1Err != nil && v2Err != nil {
		err := errors.Join(v1Err, v2Err)
		slog.Error("both v1 and v2 next page failed", "error", err)
		return nil, "", err
	}

	newCursor := cursorPayload{V1Cursor: v1Cursor, V2Cursor: v2Cursor, Seen: cursor.Seen}

	results := make([]api.Author, 0, len(v1Results)+len(v2Results))
	for _, res := range slices.Concat(v1Results, v2Results) {
		if !seen[res.AuthorId] {
			results = append(results, res)
			seen[res.AuthorId] = true
			newCursor.Seen = append(newCursor.Seen, res.AuthorId)
		}
	}

	newCursorToken, err := encodeCursor(newCursor)
	if err != nil {
		return nil, "", err
	}

	return results, newCursorToken, nil
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

	type gscholarPapers struct {
		Articles []gscholarPaper `json:"articles"`
	}

	const batchSize = 100

	res, err := client.R().
		SetResult(&gscholarPapers{}).
		SetQueryParam("engine", "google_scholar_author").
		SetQueryParam("author_id", iter.authorId).
		SetQueryParam("num", strconv.Itoa(batchSize)).
		SetQueryParam("start", strconv.Itoa(iter.start)).
		SetQueryParam("sort", "pubdate").
		Get("/search")

	if err != nil {
		slog.Error("google scholar paper search failed", "author_id", iter.authorId, "error", err)
		return nil, fmt.Errorf("google scholar error: %w", err)
	}

	if !res.IsSuccess() {
		slog.Error("google scholar paper search returned error", "status_code", res.StatusCode(), "body", res.String())
		return nil, ErrGoogleScholarSearchFailed
	}

	results := res.Result().(*gscholarPapers)

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
