package flaggers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"prism/prism/api"
	"prism/prism/llms"
	"prism/prism/openalex"
	"prism/prism/search"
	"regexp"
	"slices"
	"strings"
)

type AuthorIsFacultyAtEOCFlagger struct {
	universityNDB search.NeuralDB
}

type nameMatcher struct {
	re *regexp.Regexp
}

func newNameMatcher(name string) (nameMatcher, bool) {
	fields := strings.Fields(strings.ToLower(name))
	if len(fields) == 0 {
		return nameMatcher{}, false
	}

	if len(fields) == 1 {
		return nameMatcher{regexp.MustCompile(fields[0])}, true
	}

	firstname, lastname := fields[0], fields[len(fields)-1]

	maxChars := max(len(name)-(len(firstname)+len(lastname)), 10)

	re := regexp.MustCompile(fmt.Sprintf(`(\b%s[\w\s\.\-\,]{0,%d}%s\b)|(\b%s[\w\s\.\-\,]{0,%d}%s\b)`, firstname, maxChars, lastname, lastname, maxChars, firstname))

	return nameMatcher{re: re}, true
}

func (n *nameMatcher) matches(candidate string) bool {
	return n.re.MatchString(strings.ToLower(candidate))
}

func (n *nameMatcher) findMatches(candidate string) []string {
	return n.re.FindAllString(strings.ToLower(candidate), -1)
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Name() string {
	return "PotentialFacultyAtEOC"
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Flag(logger *slog.Logger, authorName string) ([]api.Flag, error) {
	results, err := flagger.universityNDB.Query(authorName, 5, nil)
	if err != nil {
		logger.Error("error querying ndb", "error", err)
		return nil, fmt.Errorf("error querying ndb: %w", err)
	}

	matcher, validName := newNameMatcher(authorName)
	if !validName {
		slog.Error("author name is empty")
		return nil, nil
	}

	flags := make([]api.Flag, 0)

	for _, result := range results {
		if matcher.matches(result.Text) {
			university, _ := result.Metadata["university"].(string)
			if university == "" {
				logger.Error("missing university metadata", "result", result)
				continue
			}

			url, _ := result.Metadata["url"].(string)

			flags = append(flags, &api.PotentialAuthorAffiliationFlag{
				Message:       fmt.Sprintf("The author %s may be associated with this concerning entity: %s\n", authorName, university),
				University:    university,
				UniversityUrl: url,
			})
		}
	}

	return flags, nil
}

type AuthorIsAssociatedWithEOCFlagger struct {
	docNDB search.NeuralDB
	auxNDB search.NeuralDB
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) Name() string {
	return "MiscAssociationWithEOC"
}

type authorCnt struct {
	author string
	cnt    int
}

func topCoauthors(works []openalex.Work) []authorCnt {
	authors := make(map[string]int)
	for _, work := range works {
		for _, author := range work.Authors {
			authors[author.DisplayName]++
		}
	}

	topAuthors := make([]authorCnt, 0)
	for author, cnt := range authors {
		topAuthors = append(topAuthors, authorCnt{author: author, cnt: cnt})
	}

	slices.SortFunc(topAuthors, func(a, b authorCnt) int {
		if a.cnt > b.cnt {
			return -1
		}
		if a.cnt < b.cnt {
			return 1
		}
		return 0
	})

	return topAuthors[:min(len(topAuthors), 4)]
}

const llmMatchValidationPromptTemplate = `Return a python list of True or False indicating whether the entity matches against any of the entities in the list.

Syntax:
Inputs:
- Name (String)
- Possible aliases (List of List of Strings)

Output : [True, False, ...]

Example :
Input : 
Name : Marie C. Smith
Possible aliases : [["Marie Smith", "Marie C. Smith"], ["Smith, Marge", "M. Phillipe Smith", "Marie Smiths"], ["Smith, Marie", "M.W. Smith"]]
Output : [True, False, True]

Explanation :
- Marie C. Smith matches against ["Marie Smith", "Marie C. Smith"]
- Marie C. Smith does not match against ["Smith, Marge", "M. Phillipe Smith", "Marie Smiths"]
- Marie C. Smith matches against ["Smith, Marie"] but not against ["M.W. Smith"] 

Use general knowledge to determine if the entity matches against any of the entities in the list. Answer with a python list of True or False, and nothing else. Do not provide any explanation or extra words/characters. Ensure that the python syntax is absolutely correct and runnable. Ensure that the length of the output list is the same as the length of the possible aliases list.

Input : 
- Name: %s
- Possible aliases: ["%s"]

Output:
`

func runLLMVerification(name string, possibleAliases [][]string) ([]bool, error) {
	llm := llms.New()

	// Convert the nested slice into a properly formatted string representation
	aliasesJSON, err := json.Marshal(possibleAliases)
	if err != nil {
		return nil, fmt.Errorf("error marshalling aliases: %w", err)
	}

	prompt := fmt.Sprintf(llmMatchValidationPromptTemplate, name, string(aliasesJSON))
	slog.Info("prompt", "prompt", prompt)
	res, err := llm.Generate(prompt, &llms.Options{
		Model:        llms.GPT4oMini,
		ZeroTemp:     true,
		SystemPrompt: "You are a helpful python assistant who responds in python lists only.",
	})
	if err != nil {
		slog.Error("error running llm", "error", err)
		return nil, fmt.Errorf("error running llm: %w", err)
	}

	// Clean the response to handle potential formatting issues
	res = strings.TrimSpace(res)
	res = strings.Trim(res, "[]")
	res = strings.ReplaceAll(res, " ", "")
	flags := strings.Split(res, ",")

	if len(flags) != len(possibleAliases) {
		slog.Error("llm returned incorrect number of flags", "expected", len(possibleAliases), "got", len(flags))
		slog.Error("llm response", "response", res)
		slog.Error("possible aliases", "aliases", possibleAliases)
		return nil, fmt.Errorf("llm returned incorrect number of flags: %d", len(flags))
	}

	results := make([]bool, len(flags))
	for i, flag := range flags {
		results[i] = strings.TrimSpace(flag) == "True"
	}

	return results, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) findFirstSecondHopEntities(authorName string, works []openalex.Work) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	seen := make(map[string]bool)

	primaryMatcher, validName := newNameMatcher(authorName)
	if !validName {
		slog.Error("author name is empty")
		return nil, nil
	}

	frequentAuthors := topCoauthors(works)
	for _, author := range frequentAuthors {

		matcher, validName := newNameMatcher(author.author)
		if !validName {
			slog.Error("co-author name is empty")
			continue
		}

		// TODO(question): do we need to use the name combinations, since the tokenizer will split on whitespace and lowercase?
		results, err := flagger.docNDB.Query(author.author, 5, nil)
		if err != nil {
			return nil, fmt.Errorf("error querying ndb: %w", err)
		}

		temporaryFlags := make([]api.Flag, 0)
		matches := make([][]string, 0)

		for _, result := range results {
			if !matcher.matches(result.Text) {
				continue
			}

			matches = append(matches, matcher.findMatches(result.Text))

			url, _ := result.Metadata["url"].(string)
			if seen[url] {
				continue
			}

			seen[url] = true

			title, _ := result.Metadata["title"].(string)
			entities := result.Metadata["entities"].(string)

			if primaryMatcher.matches(author.author) {
				slog.Info("primary match", "author", author.author, "title", title, "url", url, "entities", entities)
				temporaryFlags = append(temporaryFlags, &api.MiscHighRiskAssociationFlag{
					Message:         "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:        title,
					DocUrl:          url,
					DocEntities:     strings.Split(entities, ";"),
					EntityMentioned: author.author,
				})
			} else {
				slog.Info("secondary match", "author", author.author, "title", title, "url", url, "entities", entities)
				temporaryFlags = append(temporaryFlags, &api.MiscHighRiskAssociationFlag{
					Message:          "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:         title,
					DocUrl:           url,
					DocEntities:      strings.Split(entities, ";"),
					EntityMentioned:  author.author,
					Connections:      []api.Connection{{DocTitle: author.author + " (frequent coauthor)", DocUrl: ""}},
					FrequentCoauthor: &author.author,
				})
			}
		}
		// if there are no matches, we don't need to run the LLM
		// if len(matches) == 0 {
		// 	continue
		// }
		// llmResults, err := runLLMVerification(author.author, matches)
		// if err != nil {
		// 	return nil, fmt.Errorf("error running llm: %w", err)
		// }

		// for index, llmResult := range llmResults {
		// 	if llmResult {
		// 		flags = append(flags, temporaryFlags[index])
		// 		slog.Info("flag", "author", author.author, "matches", matches[index], "llmResult", llmResult)
		// 	}
		// }

		flags = append(flags, temporaryFlags...)

	}

	return flags, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) findSecondThirdHopEntities(logger *slog.Logger, authorName string) ([]api.Flag, error) {
	seen := make(map[string]bool)

	primaryMatcher, validName := newNameMatcher(authorName)
	if !validName {
		logger.Error("author name is empty")
		return nil, nil
	}

	queryToConn := make(map[string][]api.Connection)

	results, err := flagger.auxNDB.Query(authorName, 5, nil)
	if err != nil {
		logger.Error("error querying aux ndb", "error", err)
		return nil, fmt.Errorf("error querying ndb: %w", err)
	}

	for _, result := range results {
		if !primaryMatcher.matches(result.Text) {
			continue
		}

		matches := primaryMatcher.findMatches(result.Text)
		slog.Info("matches inside third hop", "matches", matches)

		url, _ := result.Metadata["url"].(string)
		if seen[url] {
			continue
		}
		seen[url] = true

		entities, _ := result.Metadata["entities"].(string)
		for _, entity := range strings.Split(entities, ";") {
			if _, ok := queryToConn[entity]; !ok {
				title, _ := result.Metadata["title"].(string)
				queryToConn[entity] = []api.Connection{{DocTitle: title, DocUrl: url}}
			}
		}
	}

	// Second map to avoid mutating the original while iterating over it
	level2Entities := make(map[string][]api.Connection)

	for query, level1Entity := range queryToConn {
		results, err := flagger.auxNDB.Query(query, 5, nil)
		if err != nil {
			logger.Error("error querying aux ndb", "error", err)
			return nil, fmt.Errorf("error querying ndb: %w", err)
		}

		for _, result := range results {
			if !strings.Contains(result.Text, query) {
				continue
			}

			url, _ := result.Metadata["url"].(string)
			if seen[url] {
				continue
			}
			seen[url] = true

			entities, _ := result.Metadata["entities"].(string)
			for _, entity := range strings.Split(entities, ";") {
				if _, ok := level2Entities[entity]; !ok {
					title, _ := result.Metadata["title"].(string)
					level2Entities[entity] = append(level1Entity, api.Connection{DocTitle: title, DocUrl: url})
				}
			}
		}
	}

	for k, v := range level2Entities {
		queryToConn[k] = v
	}

	flags := make([]api.Flag, 0)

	for query, conns := range queryToConn {
		results, err := flagger.docNDB.Query(query, 5, nil)
		if err != nil {
			slog.Error("error querying doc ndb", "error", err)
			return nil, fmt.Errorf("error querying ndb: %w", err)
		}

		for _, result := range results {
			if !strings.Contains(result.Text, query) {
				continue
			}

			title, _ := result.Metadata["title"].(string)
			url, _ := result.Metadata["url"].(string)
			entities, _ := result.Metadata["entities"].(string)

			slog.Info("third level match", "query", query, "title", title, "url", url, "entities", entities)
			flags = append(flags, &api.MiscHighRiskAssociationFlag{
				Message:         "The author may be associated be an entity who/which may be mentioned in a press release.\n",
				DocTitle:        title,
				DocUrl:          url,
				DocEntities:     strings.Split(entities, ";"),
				EntityMentioned: query,
				Connections:     conns,
			})
		}
	}

	return flags, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) Flag(logger *slog.Logger, authorName string, works []openalex.Work) ([]api.Flag, error) {
	firstSecondLevelFlags, err := flagger.findFirstSecondHopEntities(authorName, works)
	if err != nil {
		logger.Error("error checking first/second level flags", "error", err)
		return nil, err
	}

	slog.Info("first/second level flags", "flags", firstSecondLevelFlags)

	secondThirdLevelFlags, err := flagger.findSecondThirdHopEntities(logger, authorName)
	if err != nil {
		logger.Error("error checking second/third level flags", "error", err)
		return nil, err
	}

	slog.Info("second/third level flags", "flags", secondThirdLevelFlags)

	flags := slices.Concat(firstSecondLevelFlags, secondThirdLevelFlags)

	return flags, nil
}
