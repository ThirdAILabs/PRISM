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
	"prism/ndb"
	"prism/services/utils"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

type SearchService struct {
	entityNdb ndb.NeuralDB
}

func (s *SearchService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/regular", WrapRestHandler(s.SearchOpenAlex))
	r.Get("/advanced", s.SearchGoogleScholar)
	r.Get("/formal-relations", WrapRestHandler(s.FormalRelations))
	r.Get("/match-entities", WrapRestHandler(s.MatchEntities))

	return r
}

var ErrSearchFailed = errors.New("error performing search")

func (s *SearchService) SearchOpenAlex(r *http.Request) (any, error) {
	query := r.URL.Query()
	author, institution := query.Get("author"), query.Get("institution")

	url := fmt.Sprintf(
		"https://api.openalex.org/authors?filter=display_name.search:%s,affiliations.institution.id:%s&mailto=kartik@thirdai.com",
		url.QueryEscape(author), url.QueryEscape(institution),
	)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("open alex search failed", "author", author, "institution", institution, "error", err)
		return nil, CodedError(ErrSearchFailed, http.StatusInternalServerError)
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
		return nil, CodedError(ErrSearchFailed, http.StatusInternalServerError)
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

	googleResults, err := utils.GoogleSearch(fmt.Sprintf(`"%s" "%s"`, author, institution))
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

	answer, err := llm.Generate(prompt)
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

	verification, err := llm.Generate(verificationPrompt)
	if err != nil {
		slog.Error("formal relations: verification llm generaton failed", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	hasRelation := !strings.Contains(strings.ToLower(verification), "no")
	return api.FormalRelationResponse{HasFormalRelation: hasRelation}, nil
}

func cleanEntry(id uint64, text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, "[ADDRESS]") {
			// TODO(question): this used split before, is Cut ok?
			_, addr, _ := strings.Cut(line, "[ADDRESS]")
			addr, _, _ = strings.Cut(addr, ";")
			lines[i] = "[ADDRESS] " + addr
		}
	}
	return fmt.Sprintf("[ID] %d\n", id) + strings.Join(lines, "\n")
}

const (
	matchEntitiesPrompt = `I will give you a list of entities, each formatted as a '[ID] <id value> [ENTITY START] ... [ENTITY END]' block. Each block contains known aliases of the entity, the address, and some other information.
    If "%s" in the list of entities, return the value of the [ID] tag of all blocks that contain the entity, formatting it in a CSV list like this:

    <id1>,<id2>,<id3>

    And so on. Otherwise return nothing.
    It may not be a perfect string match, but you should not return entities that can be reasoned to be a mismatch.`
)

func (s *SearchService) MatchEntities(r *http.Request) (any, error) {
	query := r.URL.Query().Get("query")

	results, err := s.entityNdb.Query(query, 15, nil)
	if err != nil {
		slog.Error("match entities: ndb search failed", "query", query, "error", err)
		return nil, CodedError(errors.New("entity search failed"), http.StatusInternalServerError)
	}

	idToText := make(map[uint64]string)

	candidates := make([]string, 0, len(results))
	for _, result := range results {
		idToText[result.Id] = result.Text
		candidates = append(candidates, cleanEntry(result.Id, result.Text))
	}

	// TODO(question): does the prompt make sense with the entities in front?
	prompt := strings.Join(candidates, "\n") + "\n\n" + fmt.Sprintf(matchEntitiesPrompt, query)

	response, err := llms.New().Generate(prompt)
	if err != nil {
		slog.Error("match entities: llm generaton failed", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	entities := make([]string, 0)
	for _, id := range strings.Split(strings.Trim(response, "`"), ",") {
		parsed, err := strconv.ParseUint(id, 10, 64)
		if err == nil {
			if entity, ok := idToText[parsed]; ok {
				entities = append(entities, entity)
			} else {
				slog.Error("match entities: invalid id returned from llm")
			}
		} else {
			slog.Error("match entities: malformed id returned from llm", "id", id, "error", err)
		}
	}

	return api.MatchEntitiesResponse{Entities: entities}, nil
}
