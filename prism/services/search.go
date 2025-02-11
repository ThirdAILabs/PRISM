package services

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"prism/api"
	"prism/gscholar"
	"prism/llms"
	"prism/openalex"
	"prism/search"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type SearchService struct {
	openalex openalex.KnowledgeBase

	entityNdb search.NeuralDB
}

func (s *SearchService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/regular", WrapRestHandler(s.SearchOpenAlex))
	r.Get("/advanced", WrapRestHandler(s.SearchGoogleScholar))
	r.Get("/match-entities", WrapRestHandler(s.MatchEntities))

	return r
}

var ErrSearchFailed = errors.New("error performing search")

func (s *SearchService) SearchOpenAlex(r *http.Request) (any, error) {
	query := r.URL.Query()
	author, institution := query.Get("author"), query.Get("institution")

	authors, err := s.openalex.FindAuthors(author, institution)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	results := make([]api.Author, 0, len(authors))
	for _, author := range authors {
		results = append(results, api.Author{
			AuthorId:     author.AuthorId,
			AuthorName:   author.DisplayName,
			Institutions: author.InstitutionNames(),
			Source:       api.OpenAlexSource,
		})
	}

	return results, nil
}

func (s *SearchService) SearchGoogleScholar(r *http.Request) (any, error) {
	query := strings.ReplaceAll(strings.ToLower(r.URL.Query().Get("query")), "@", " ")
	cursor := r.URL.Query().Get("cursor")

	results, nextCursor, err := gscholar.NextGScholarPage(query, cursor)
	if err != nil {
		if errors.Is(err, gscholar.ErrInvalidCursor) {
			return nil, CodedError(err, http.StatusBadRequest)
		}
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return api.GScholarSearchResults{Authors: results, Cursor: nextCursor}, nil
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

	response, err := llms.New().Generate(prompt, nil)
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
