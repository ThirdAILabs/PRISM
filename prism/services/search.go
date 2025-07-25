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
	"prism/prism/reports"
	"prism/prism/reports/utils"
	"prism/prism/search"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const institutionNameSimilarityThreshold = 0.8

type SearchService struct {
	openalex openalex.KnowledgeBase

	// entitySearch EntitySearch
	entitySearch *search.ManyToOneIndex[api.MatchedEntity]
}

func NewSearchService(oa openalex.KnowledgeBase, entities []api.MatchedEntity) SearchService {
	return SearchService{
		openalex:     oa,
		entitySearch: NewEntitySearch(entities),
	}
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
		if similarity := utils.JaroWinklerSimilarity(originalInstitution, inst); similarity >= similarityThreshold {
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

	r.Get("/authors", WrapRestHandler(s.SearchOpenAlex))
	r.Get("/authors-advanced", WrapRestHandler(s.SearchGoogleScholar))
	r.Get("/match-entities", WrapRestHandler(s.MatchEntities))

	return r
}

var ErrSearchFailed = errors.New("error performing search")

func (s *SearchService) SearchOpenAlex(r *http.Request) (any, error) {
	query := r.URL.Query()

	if query.Get("author_name") != "" {
		author, institution, institutionName := query.Get("author_name"), query.Get("institution_id"), query.Get("institution_name")
		if institution == "" || institutionName == "" {
			return nil, CodedError(errors.New("institution_id and institution_name must be specified when searching by author name"), http.StatusBadRequest)
		}
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
	} else if query.Get("orcid") != "" {
		author, err := s.openalex.FindAuthorByOrcidId(query.Get("orcid"))
		if err != nil {
			if errors.Is(err, openalex.ErrAuthorNotFound) {
				return []api.Author{}, nil
			}
			return nil, CodedError(err, http.StatusInternalServerError)
		}
		return []api.Author{{
			AuthorId:     author.AuthorId,
			AuthorName:   author.DisplayName,
			Institutions: author.InstitutionNames(),
			Source:       api.OpenAlexSource,
			Interests:    author.Concepts,
		}}, nil
	} else if query.Get("paper_title") != "" {
		papers, err := s.openalex.FindWorksByTitle([]string{query.Get("paper_title")}, reports.EarliestReportDate, time.Now())
		if err != nil {
			return nil, CodedError(err, http.StatusInternalServerError)
		}

		if len(papers) < 1 {
			return nil, CodedError(errors.New("no papers found for title"), http.StatusNotFound)
		}

		authors := make([]api.Author, 0, len(papers[0].Authors))
		for _, author := range papers[0].Authors {
			authors = append(authors, api.Author{
				AuthorId:     author.AuthorId,
				AuthorName:   author.DisplayName,
				Institutions: author.InstitutionNames(),
				Source:       api.OpenAlexSource,
				Interests:    author.Concepts,
			})
		}

		return authors, nil
	} else {
		return nil, CodedError(errors.New("one of 'author_name', 'orcid', or 'paper_title' query parameters must be specified"), http.StatusBadRequest)
	}

}

func filterAuthorsBySimilarity(authors []api.Author, queryName string) []api.Author {
	const minSimilarity = 0.5

	type pair struct {
		author *api.Author
		sim    float64
	}
	authorSims := make([]pair, 0, len(authors))
	for _, author := range authors {
		sim := utils.IndelSimilarity(author.AuthorName, queryName)
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

func NewEntitySearch(entities []api.MatchedEntity) *search.ManyToOneIndex[api.MatchedEntity] {
	entityNames := make([][]string, 0, len(entities))
	for _, entity := range entities {
		entityNames = append(entityNames, strings.Split(entity.Names, "\n"))
	}

	return search.NewManyToOneIndex(entityNames, entities)
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

	results := s.entitySearch.Query(query, 15)

	idToEntity := make(map[int]api.MatchedEntity)

	seen := make(map[string]bool)
	candidates := make([]string, 0, len(results))
	for id, entity := range results {
		if seen[entity.Metadata.Names] {
			continue
		}
		seen[entity.Metadata.Names] = true
		idToEntity[id] = entity.Metadata
		candidates = append(candidates, formatForPrompt(entity.Metadata, id))
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
