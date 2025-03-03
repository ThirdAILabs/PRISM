package services

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"prism/prism/api"
	"prism/prism/gscholar"
	"prism/prism/llms"
	"prism/prism/openalex"
	"prism/prism/reports/flaggers"
	"prism/prism/search"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

const institutionNameSimilarityThreshold = 0.8

type SearchService struct {
	openalex openalex.KnowledgeBase

	entitySearch EntitySearch
}

func hybridInstitutionNamesSort(originalInstitution string, institutionNames []string, similarityThreshold float64) []string {
	// put the institutes with name similar to the Original Institution at the front
	// sort the rest alphabetically

	type InstitutionSimilarity struct {
		Name       string
		Similarity float64
	}
	var similarInstitutesSimilarity []InstitutionSimilarity
	var differentInstitutes []string

	for _, inst := range institutionNames {
		if similarity := flaggers.JaroWinklerSimilarity(originalInstitution, inst); similarity >= similarityThreshold {
			similarInstitutesSimilarity = append(similarInstitutesSimilarity, InstitutionSimilarity{Name: inst, Similarity: similarity})
		} else {
			differentInstitutes = append(differentInstitutes, inst)
		}
	}

	// sort the similar institutes by similarity
	sort.Slice(similarInstitutesSimilarity, func(i, j int) bool {
		return similarInstitutesSimilarity[i].Similarity > similarInstitutesSimilarity[j].Similarity
	})
	similarInstitutes := make([]string, len(similarInstitutesSimilarity))
	for i, inst := range similarInstitutesSimilarity {
		similarInstitutes[i] = inst.Name
	}

	// sort the different institutes alphabetically
	sort.Strings(differentInstitutes)
	finalList := append(similarInstitutes, differentInstitutes...)
	return finalList
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
	author, institution, institutionName := query.Get("author_name"), query.Get("institution_id"), query.Get("institution_name")

	slog.Info("searching openalex", "author_name", author, "institution_id", institution)

	authors, err := s.openalex.FindAuthors(author, institution)
	if err != nil {
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	results := make([]api.Author, 0, len(authors))
	for _, author := range authors {
		sortedInstitutions := hybridInstitutionNamesSort(institutionName, author.InstitutionNames(), institutionNameSimilarityThreshold)
		results = append(results, api.Author{
			AuthorId:     author.AuthorId,
			AuthorName:   author.DisplayName,
			Institutions: sortedInstitutions,
			Source:       api.OpenAlexSource,
			Interests:    author.Concepts,
		})
	}

	return results, nil
}

func filterAuthorsBySimilarity(authors []api.Author, queryName string) []api.Author {
	const minSimilarity = 0.5

	type pair struct {
		author *api.Author
		sim    float64
	}
	authorSims := make([]pair, 0, len(authors))
	for _, author := range authors {
		sim := flaggers.IndelSimilarity(author.AuthorName, queryName)
		if sim > minSimilarity {
			authorSims = append(authorSims, pair{author: &author, sim: sim})
		}
	}
	sort.Slice(authorSims, func(i, j int) bool {
		return authorSims[i].sim > authorSims[j].sim
	})

	sortedAuthors := make([]api.Author, 0, len(authors))
	for _, pair := range authorSims {
		sortedAuthors = append(sortedAuthors, *pair.author)
	}
	return sortedAuthors
}

func (s *SearchService) SearchGoogleScholar(r *http.Request) (any, error) {
	query := r.URL.Query()
	author, institution, cursor := query.Get("author_name"), query.Get("institution_name"), r.URL.Query().Get("cursor")
	if author == "" || institution == "" {
		return nil, CodedError(errors.New("author_name and institution_name must be specified"), http.StatusBadRequest)
	}

	slog.Info("searching google scholar", "author_name", author, "institution_name", institution, "cursor", cursor)

	results, nextCursor, err := gscholar.NextGScholarPage(strings.ToLower(fmt.Sprintf("%s %s", author, institution)), cursor)
	if err != nil {
		if errors.Is(err, gscholar.ErrInvalidCursor) {
			return nil, CodedError(err, http.StatusBadRequest)
		}
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	return api.GScholarSearchResults{Authors: filterAuthorsBySimilarity(results, author), Cursor: nextCursor}, nil
}

func formatForPrompt(e api.MatchedEntity, id int) string {
	base := fmt.Sprintf("[ID] %d\n[ENTITY START]\n[NAMES START]\n%s\n[NAMES END]\n", id, e.Names)
	if e.Address != "" {
		addr, _, _ := strings.Cut(e.Address, ";")
		base += "[ADDRESS] " + addr + "\n"
	}
	if e.Country != "" {
		base += "[COUNTRY] " + e.Country + "\n"
	}
	if e.Type != "" {
		base += "[TYPE] " + e.Type + "\n"
	}
	if e.Resource != "" {
		base += "[RESOURCE] " + e.Resource + "\n"
	}

	return base + "[ENTITY END]"
}

type EntitySearch struct {
	ndb search.NeuralDB
}

func NewEntitySearch(path string) (EntitySearch, error) {
	ndb, err := search.NewNeuralDB(path)
	if err != nil {
		return EntitySearch{}, err
	}
	return EntitySearch{ndb: ndb}, nil
}

func (es *EntitySearch) Insert(entities []api.MatchedEntity) error {
	data := make([]string, 0, len(entities))
	metadata := make([]map[string]interface{}, 0, len(entities))
	for _, entity := range entities {
		data = append(data, entity.Names)

		metadata = append(metadata, map[string]interface{}{
			"address":  entity.Address,
			"country":  entity.Country,
			"type":     entity.Type,
			"resource": entity.Resource,
		})
	}

	return es.ndb.Insert("entities", "id", data, metadata, nil)
}

func (es *EntitySearch) Query(query string, topk int) ([]api.MatchedEntity, error) {
	results, err := es.ndb.Query(query, topk, nil)
	if err != nil {
		slog.Error("match entities: ndb search failed", "query", query, "error", err)
		return nil, errors.New("entity search failed")
	}

	entities := make([]api.MatchedEntity, 0, len(results))
	for i := range results {
		entities = append(entities, api.MatchedEntity{
			Names:    results[i].Text,
			Address:  results[i].Metadata["address"].(string),
			Country:  results[i].Metadata["country"].(string),
			Type:     results[i].Metadata["type"].(string),
			Resource: results[i].Metadata["resource"].(string),
		})
	}
	return entities, nil
}

func (es *EntitySearch) Free() {
	es.ndb.Free()
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

	results, err := s.entitySearch.Query(query, 15)
	if err != nil {
		slog.Error("match entities: ndb search failed", "query", query, "error", err)
		return nil, CodedError(errors.New("entity search failed"), http.StatusInternalServerError)
	}

	idToEntity := make(map[int]api.MatchedEntity)

	candidates := make([]string, 0, len(results))
	for id, entity := range results {
		idToEntity[id] = entity
		candidates = append(candidates, formatForPrompt(entity, id))
	}

	// TODO(question): does the prompt make sense with the entities in front?
	prompt := strings.Join(candidates, "\n") + "\n\n" + fmt.Sprintf(matchEntitiesPrompt, query)

	response, err := llms.New().Generate(prompt, nil)
	if err != nil {
		slog.Error("match entities: llm generaton failed", "error", err)
		return nil, CodedError(err, http.StatusInternalServerError)
	}

	entities := make([]api.MatchedEntity, 0)
	for _, id := range strings.Split(strings.Trim(response, "`"), ",") {
		parsed, err := strconv.Atoi(id)
		if err == nil {
			if entity, ok := idToEntity[parsed]; ok {
				entities = append(entities, entity)
			} else {
				slog.Error("match entities: invalid id returned from llm")
			}
		} else {
			slog.Error("match entities: malformed id returned from llm", "id", id, "error", err)
		}
	}

	return entities, nil
}
