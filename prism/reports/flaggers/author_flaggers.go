package flaggers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"prism/prism/api"
	"prism/prism/llms"
	"prism/prism/openalex"
	"prism/prism/search"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
)

type AuthorIsFacultyAtEOCFlagger struct {
	universityNDB search.NeuralDB
}

func NewAuthorIsFacultyAtEOCFlagger(universityNDB search.NeuralDB) *AuthorIsFacultyAtEOCFlagger {
	return &AuthorIsFacultyAtEOCFlagger{universityNDB: universityNDB}
}

type nameMatcher struct {
	text_regex   *regexp.Regexp
	entity_regex *regexp.Regexp
}

const (
	numUniversityDocumentsToRetrieve = 5
	numAuxillaryDocumentsToRetrieve  = 5
	numDOJDocumentsToRetrieve        = 5
	useLLMVerification               = true
)

func newNameMatcher(name string) (nameMatcher, bool) {
	fields := strings.Fields(strings.ToLower(name))
	if len(fields) == 0 {
		return nameMatcher{}, false
	}

	if len(fields) == 1 {
		return nameMatcher{text_regex: regexp.MustCompile(fields[0]), entity_regex: regexp.MustCompile(fields[0])}, true
	}

	firstname, lastname := regexp.QuoteMeta(fields[0]), regexp.QuoteMeta(fields[len(fields)-1])

	maxChars := max(len(name)-(len(firstname)+len(lastname)), 10)
	namepattern := regexp.QuoteMeta(strings.Join(fields, `\s+`))

	text_regex, err := regexp.Compile(fmt.Sprintf(`(\b%s[\w\s\.\-]{0,%d}%s\b)|(\b%s[\w\s\.\-\,]{0,%d}%s\b)|(\b%s\b)`, firstname, maxChars, lastname, lastname, maxChars, firstname, namepattern))
	if err != nil {
		return nameMatcher{}, false
	}

	// anchored regex to the start and end of the string
	entity_regex, err := regexp.Compile(fmt.Sprintf(`^((%s[\w\s\.\-]{0,%d}%s)|(%s[\w\s\.\-\,]{0,%d}%s)|(%s))$`,
		firstname, maxChars, lastname,
		lastname, maxChars, firstname,
		namepattern))
	if err != nil {
		return nameMatcher{}, false
	}

	return nameMatcher{text_regex: text_regex, entity_regex: entity_regex}, true
}

func (n *nameMatcher) matchesText(candidate string) bool {
	return n.text_regex.MatchString(strings.ToLower(candidate))
}

func (n *nameMatcher) matchesEntity(candidate string) bool {
	return n.entity_regex.MatchString(strings.ToLower(candidate))
}

func (n *nameMatcher) matchesAnyEntity(candidates []string) bool {
	for _, candidate := range candidates {
		if n.matchesEntity(strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

type MatchResult struct {
	Match   string
	Context string
}

func (n *nameMatcher) findMatchesInText(candidate string) []MatchResult {
	lowercaseCandidate := strings.ToLower(candidate)
	matchIndices := n.text_regex.FindAllStringIndex(lowercaseCandidate, -1)

	results := make([]MatchResult, 0, len(matchIndices))

	// Get positions of all words
	wordBounds := regexp.MustCompile(`\S+`).FindAllStringIndex(candidate, -1)

	for _, indices := range matchIndices {
		start, end := indices[0], indices[1]
		match := candidate[start:end]

		matchStartWord, matchEndWord := 0, len(wordBounds)-1

		// find the start and end word indices
		for i, bounds := range wordBounds {
			wordStart, wordEnd := bounds[0], bounds[1]

			if start >= wordStart && start < wordEnd {
				matchStartWord = i
			}

			if end > wordStart && end <= wordEnd {
				matchEndWord = i
				break
			}
		}

		contextStartWord := max(0, matchStartWord-2)
		contextEndWord := min(len(wordBounds)-1, matchEndWord+2)
		context := candidate[wordBounds[contextStartWord][0]:wordBounds[contextEndWord][1]]

		results = append(results, MatchResult{
			Match:   match,
			Context: context,
		})
	}

	return results
}

const llmMatchValidationPromptTemplate = `Return a python list of True or False indicating whether the entity matches against any of the entities in each group.

Syntax:
Inputs:
- Name (String)
- Possible matches grouped by page (List of List of Dict with 'match' and 'context' keys)

Output : [True, False, ...]

Example :
Input : 
Name : Marie C. Smith
Possible matches : [
  [
    {"match": "Marie Smith", "context": "Professor Marie Smith from Harvard"},
    {"match": "Marie C Smith", "context": "Professor Marie C Smith teaches biology"}
  ],
  [
    {"match": "Smith, Marge", "context": "Dr. Smith, Marge at Stanford"},
    {"match": "Smith, M", "context": "Smith, M is the lead author"}
  ]
]
Output : [True, False]

Explanation :
- The first group contains "Marie Smith" and "Marie C Smith" which are valid matches for "Marie C. Smith"
- The second group contains "Smith, Marge" and "Smith, M" which are not valid matches for "Marie C. Smith"

Example : 
Input : 
Name : J. Phillip
Possible matches : [
  [
    {"match": "J Phillip", "context": "Professor Donovan J Phillip"}
  ],
  [
    {"match": "J. Phillip", "context": "J. Phillip Smith"},
    {"match": "J Phillip", "context": "research by J Phillip in 2020"}
  ],
	[
	]
]
Output : [False, True, False]

Explanation :
- First group: "J Phillip" in "Professor Donovan J Phillip" is not a match because Donovan is part of the name
- Second group: Contains "J. Phillip Smith" which is a valid match for "J. Phillip"
- Third group: No matches found

Return True for a group if ANY match in that group correctly refers to the input name. Use the context to determine if a match is legitimate. Answer with a python list of True or False, and nothing else. Do not use markdown or any other formatting. Only return ["True", "False", ...].

Input : 
- Name: %s
- Possible matches: %s

Output:
`

func runLLMVerification(name string, texts []string) ([]bool, error) {
	// returns a list of boolean values indicating whether page i contains a match for the name

	matcher, validName := newNameMatcher(name)
	if !validName {
		return nil, fmt.Errorf("author name is empty")
	}

	possibleAliases := make([][]MatchResult, len(texts))
	for i, text := range texts {
		possibleAliases[i] = matcher.findMatchesInText(text)
	}

	llm := llms.New()

	// Convert the nested slice into a properly formatted string representation
	aliasesJSON, err := json.Marshal(possibleAliases)
	if err != nil {
		return nil, fmt.Errorf("error marshalling aliases: %w", err)
	}

	prompt := fmt.Sprintf(llmMatchValidationPromptTemplate, name, string(aliasesJSON))
	res, err := llm.Generate(prompt, &llms.Options{
		Model:        llms.GPT4o,
		ZeroTemp:     true,
		SystemPrompt: "You are a helpful python assistant who responds in python lists only.",
	})
	if err != nil {
		slog.Error("error running llm", "error", err)
		return nil, fmt.Errorf("error running llm: %w", err)
	}

	// Clean the response
	res = strings.TrimSpace(res)
	res = strings.Trim(res, "[]")
	res = strings.ReplaceAll(res, " ", "")
	flags := strings.Split(res, ",")

	if len(flags) != len(possibleAliases) {
		slog.Error("llm returned incorrect number of flags", "expected", len(possibleAliases), "got", len(flags))
		return nil, fmt.Errorf("llm returned incorrect number of flags: %d", len(flags))
	}

	results := make([]bool, len(flags))
	for i, flag := range flags {
		results[i] = strings.TrimSpace(flag) == "True"
	}

	return results, nil
}

func filterFlagsWithLLM(flags []api.Flag, texts []string, name string) ([]api.Flag, error) {
	if len(texts) == 0 {
		return flags, nil
	}

	if len(flags) != len(texts) {
		return nil, fmt.Errorf("flags and texts have different lengths")
	}

	llmResults, err := runLLMVerification(name, texts)
	if err != nil {
		return nil, fmt.Errorf("error running llm: %w", err)
	}

	filteredFlags := make([]api.Flag, 0)
	for i, flag := range flags {
		if llmResults[i] {
			filteredFlags = append(filteredFlags, flag)
		}
	}

	return filteredFlags, nil
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Name() string {
	return "PotentialFacultyAtEOC"
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Flag(logger *slog.Logger, authorName, affiliations string) ([]api.Flag, error) {
	results, err := flagger.universityNDB.Query(authorName, numUniversityDocumentsToRetrieve, nil)
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
		if matcher.matchesText(result.Text) {
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

type LinkMetadata struct {
	Title    string
	Url      string
	Entities []string
	Text     string
}

type AuthorIsAssociatedWithEOCFlagger struct {
	docIndex *search.ManyToOneIndex[LinkMetadata]
	auxIndex *search.ManyToOneIndex[LinkMetadata]
}

func NewAuthorIsAssociatedWithEOCFlagger(docIndex, auxIndex *search.ManyToOneIndex[LinkMetadata]) *AuthorIsAssociatedWithEOCFlagger {
	return &AuthorIsAssociatedWithEOCFlagger{docIndex: docIndex, auxIndex: auxIndex}
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) DisableForUniversityReport() bool {
	return false
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

	if len(topAuthors) <= 4 {
		return topAuthors
	}
	// Find the count of the 4th top author
	thresholdCount := topAuthors[3].cnt

	// Return all authors with count >= threshold
	var result []authorCnt
	for _, author := range topAuthors {
		if author.cnt >= thresholdCount {
			result = append(result, author)
		} else {
			break
		}
	}
	return result
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
		results := flagger.docIndex.Query(author.author, 5)

		temporaryFlags := make([]api.Flag, 0)
		texts := make([]string, 0)
		for _, result := range results {
			if !matcher.matchesText(result.Entity) {
				continue
			}

			texts = append(texts, result.Metadata.Text)

			if seen[result.Metadata.Url] {
				continue
			}

			seen[result.Metadata.Url] = true

			if primaryMatcher.matchesEntity(author.author) {
				temporaryFlags = append(temporaryFlags, &api.MiscHighRiskAssociationFlag{
					Message:         "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:        result.Metadata.Title,
					DocUrl:          result.Metadata.Url,
					DocEntities:     result.Metadata.Entities,
					EntityMentioned: author.author,
				})
			} else {
				temporaryFlags = append(temporaryFlags, &api.MiscHighRiskAssociationFlag{
					Message:          "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:         result.Metadata.Title,
					DocUrl:           result.Metadata.Url,
					DocEntities:      result.Metadata.Entities,
					EntityMentioned:  author.author,
					Connections:      []api.Connection{{DocTitle: author.author + " (frequent coauthor)", DocUrl: ""}},
					FrequentCoauthor: &author.author,
				})
			}
		}

		if len(temporaryFlags) == 0 {
			continue
		}

		if useLLMVerification {
			temporaryFlags, err := filterFlagsWithLLM(temporaryFlags, texts, author.author)
			if err != nil {
				return nil, fmt.Errorf("error filtering flags: %w", err)
			}
			flags = append(flags, temporaryFlags...)
		} else {
			flags = append(flags, temporaryFlags...)
		}
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

	results := flagger.auxIndex.Query(authorName, 5)

	for _, result := range results {
		// skip if already seen the URL
		if seen[result.Metadata.Url] {
			continue
		}
		seen[result.Metadata.Url] = true

		// iterate over entities and check if any of them match the primary matcher
		// skip if no entity matches the primary matcher
		// this is not always accurate as Thomas J. Smith will match with J. Smith
		if !primaryMatcher.matchesAnyEntity(result.Metadata.Entities) {
			continue
		}

		// add neighbouring entities to the queryToConn map
		for _, entity := range result.Metadata.Entities {
			if _, ok := queryToConn[entity]; !ok {
				queryToConn[entity] = []api.Connection{{DocTitle: result.Metadata.Title, DocUrl: result.Metadata.Url}}
			}
		}
	}

	// Second map to avoid mutating the original while iterating over it
	level2Entities := make(map[string][]api.Connection)

	for query, level1Entity := range queryToConn {
		results := flagger.auxIndex.Query(query, 5)

		secondaryMatcher, validName := newNameMatcher(query)
		if !validName {
			slog.Error("query name is empty")
			continue
		}

		for _, result := range results {
			if !strings.Contains(result.Entity, query) {
				continue
			}

			if seen[result.Metadata.Url] {
				continue
			}
			seen[result.Metadata.Url] = true

			// skip if no entity matches the secondary matcher
			if !secondaryMatcher.matchesAnyEntity(result.Metadata.Entities) {
				continue
			}

			for _, entity := range result.Metadata.Entities {
				if _, ok := level2Entities[entity]; !ok {
					level2Entities[entity] = append(level1Entity, api.Connection{DocTitle: result.Metadata.Title, DocUrl: result.Metadata.Url})
				}
			}
		}
	}

	for k, v := range level2Entities {
		queryToConn[k] = v
	}

	flags := make([]api.Flag, 0)
	seenFlags := make(map[[sha256.Size]byte]bool)
	for query, conns := range queryToConn {
		results := flagger.docIndex.Query(query, 5)

		tempFlags := make([]api.Flag, 0)

		// rather than using direct string comparison, we use llm to verify matches here
		texts := make([]string, 0)
		for _, result := range results {
			// searching for exact match
			// this increases the false negatives for the names
			if !strings.Contains(result.Entity, query) {
				continue
			}

			flag := &api.MiscHighRiskAssociationFlag{
				Message:         "The author may be associated be an entity who/which may be mentioned in a press release.\n",
				DocTitle:        result.Metadata.Title,
				DocUrl:          result.Metadata.Url,
				DocEntities:     result.Metadata.Entities,
				EntityMentioned: query,
				Connections:     conns,
			}

			hash := flag.Hash()

			if _, ok := seenFlags[hash]; ok {
				continue
			}

			seenFlags[hash] = true
			tempFlags = append(tempFlags, flag)
			texts = append(texts, result.Metadata.Text)
		}

		if useLLMVerification {
			filteredFlags, err := filterFlagsWithLLM(tempFlags, texts, query)
			if err != nil {
				return nil, fmt.Errorf("error filtering flags: %w", err)
			}
			flags = append(flags, filteredFlags...)
		} else {
			flags = append(flags, tempFlags...)
		}
	}
	return flags, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string, authorName string) ([]api.Flag, error) {
	firstSecondLevelFlags, err := flagger.findFirstSecondHopEntities(authorName, works)
	if err != nil {
		logger.Error("error checking first/second level flags", "error", err)
		return nil, err
	}

	secondThirdLevelFlags, err := flagger.findSecondThirdHopEntities(logger, authorName)
	if err != nil {
		logger.Error("error checking second/third level flags", "error", err)
		return nil, err
	}

	flags := slices.Concat(firstSecondLevelFlags, secondThirdLevelFlags)

	return flags, nil
}

type AuthorNewsArticlesFlagger struct {
	llm *llms.PerplexityAI
}

func NewAuthorNewsArticlesFlagger(ppxApiKey string) *AuthorNewsArticlesFlagger {
	return &AuthorNewsArticlesFlagger{llm: llms.NewPerplexityAI(ppxApiKey)}
}

func (flagger *AuthorNewsArticlesFlagger) Name() string {
	return "NewsArticles"
}

func (flagger *AuthorNewsArticlesFlagger) authorPrompts(authorName, affiliation string) (string, string) {
	systemPrompt := `You are a research assistant specializing in investigative analysis.
Your job is to assist with background checks on academic or professional authors by gathering and summarizing news articles
in a short sentence involving misconduct of some kind indicated by the article.
These indictments may or may not be connected to their institutional affiliation.`

	userPrompt := fmt.Sprintf("Search for news articles about the author %s", authorName)
	if affiliation != "" {
		userPrompt += fmt.Sprintf(" who is/was affiliated with %s.", affiliation)
	}
	userPrompt += fmt.Sprintf(`Focus specifically on:
1. Any articles that mention misconduct, ethical violations, or professional controversies
2. The specific role %s played in these incidents

For each relevant article:
- Provide a concise one-sentence summary of the alleged misconduct

Limit your response to the 5 most significant sources. If no articles mentioning misconduct are found, return an empty response format.`, authorName)

	return systemPrompt, userPrompt
}

type NewsFormat struct {
	Citation int    `json:"citation"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Date     string `json:"date"`
}

// ResponseFormat represents the overall response structure
type ResponseFormat struct {
	News []NewsFormat `json:"news"`
}

func (flagger *AuthorNewsArticlesFlagger) responseFormat() map[string]interface{} {
	jsonSchema := map[string]interface{}{
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"news": map[string]interface{}{
					"type":        "array",
					"description": "List of news items",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"citation": map[string]interface{}{
								"type":        "integer",
								"description": "Corresponding citation number of the source",
							},
							"title": map[string]interface{}{
								"type":        "string",
								"description": "Title of the source",
							},
							"summary": map[string]interface{}{
								"type":        "string",
								"description": "Summary of the source",
							},
							"date": map[string]interface{}{
								"type":        "string",
								"description": "Date of the source in DD-MM-YYYY format",
								"format":      "date",
							},
						},
					},
				},
			},
		},
	}

	return map[string]interface{}{
		"type":        "json_schema",
		"json_schema": jsonSchema,
	}
}
func (flagger *AuthorNewsArticlesFlagger) Flag(logger *slog.Logger, authorName, affiliations string) ([]api.Flag, error) {
	systemPrompt, userPrompt := flagger.authorPrompts(authorName, strings.Split(affiliations, ",")[0])

	responseFormatSchema := flagger.responseFormat()
	response, citations, err := flagger.llm.Generate(
		userPrompt, systemPrompt, &llms.PerplexityOptions{
			Model: "sonar-pro",
			WebSearchOptions: map[string]interface{}{
				"search_context_size": "high",
			},
			ResponseFormat: responseFormatSchema,
		},
	)

	if err != nil {
		logger.Error("error generating response", "error", err)
		return nil, err
	}

	// parsing response
	var parsedResponse ResponseFormat
	err = json.Unmarshal([]byte(response), &parsedResponse)
	if err != nil {
		logger.Error("error parsing response", "error", err)
		return nil, err
	}

	sort.Slice(parsedResponse.News, func(i, j int) bool {
		dateI, err := time.Parse("2006-01-02", parsedResponse.News[i].Date) // Assuming the date format is YYYY-MM-DD
		if err != nil {
			return false
		}
		dateJ, err := time.Parse("2006-01-02", parsedResponse.News[j].Date)
		if err != nil {
			return false
		}
		return dateI.After(dateJ)
	})

	parsedResponse.News = parsedResponse.News[:min(5, len(parsedResponse.News))]

	flags := make([]api.Flag, 0, len(parsedResponse.News))
	for _, result := range parsedResponse.News {
		docUrl := ""
		if len(citations) >= result.Citation {
			// citation is 1-indexed and result.Citation index is not out-of-bound
			docUrl = citations[result.Citation]
		}

		flags = append(flags, &api.MiscHighRiskAssociationFlag{
			Message:  result.Summary,
			DocTitle: result.Title,
			DocUrl:   docUrl,
			// TODO: add date support in this flag and use article.Date
			EntityMentioned: authorName,
		})
	}

	return flags, nil
}
