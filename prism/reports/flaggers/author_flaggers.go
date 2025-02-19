package flaggers

import (
	"fmt"
	"log/slog"
	"prism/prism/api"
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

func (flagger *AuthorIsFacultyAtEOCFlagger) Name() string {
	return "PotentialFacultyAtEOC"
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Flag(logger *slog.Logger, authorName string) ([]api.Flag, error) {
	logger.Info("checking if author is faculty at EOC", "author_name", authorName)

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

			logger.Info("found author in listing for EOC university", "author_name", authorName, "university", university)
		}
	}

	logger.Info("finished checking for faculty at EOC", "n_flags", len(flags))

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

func (flagger *AuthorIsAssociatedWithEOCFlagger) findFirstSecondHopEntities(logger *slog.Logger, authorName string, works []openalex.Work) ([]api.Flag, error) {
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

		for _, result := range results {
			if !matcher.matches(result.Text) {
				continue
			}

			url, _ := result.Metadata["url"].(string)
			if seen[url] {
				continue
			}

			seen[url] = true

			title, _ := result.Metadata["title"].(string)
			entities := result.Metadata["entities"].(string)

			if primaryMatcher.matches(author.author) {
				logger.Info("found primary connection", "doc", title, "url", url)
				flags = append(flags, &api.MiscHighRiskAssociationFlag{
					Message:         "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:        title,
					DocUrl:          url,
					DocEntities:     strings.Split(entities, ";"),
					EntityMentioned: author.author,
				})
				logger.Info("author is assoiciated with EOC", "author", author.author, "doc", title, "entities", entities)
			} else {
				logger.Info("found coauthor connection", "doc", title, "url", url)
				flags = append(flags, &api.MiscHighRiskAssociationFlag{
					Message:          "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:         title,
					DocUrl:           url,
					DocEntities:      strings.Split(entities, ";"),
					EntityMentioned:  author.author,
					Connections:      []api.Connection{{DocTitle: author.author + " (frequent coauthor)", DocUrl: ""}},
					FrequentCoauthor: &author.author,
				})
				logger.Info("frequent coauthor is assoiciated with EOC", "coauthor", author.author, "doc", title, "entities", entities)
			}
		}
	}

	logger.Info("first/second level flags", "n_flags", len(flags))

	return flags, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) findSecondThirdHopEntities(logger *slog.Logger, authorName string) ([]api.Flag, error) {
	seen := make(map[string]bool)

	primaryMatcher, validName := newNameMatcher(authorName)
	if !validName {
		slog.Error("author name is empty")
		return nil, nil
	}

	queryToConn := make(map[string][]api.Connection)

	results, err := flagger.auxNDB.Query(authorName, 5, nil)
	if err != nil {
		return nil, fmt.Errorf("error querying ndb: %w", err)
	}

	for _, result := range results {
		if !primaryMatcher.matches(result.Text) {
			continue
		}

		url, _ := result.Metadata["url"].(string)
		if seen[url] {
			continue
		}
		seen[url] = true

		entities, _ := result.Metadata["entities"].(string)
		for _, entity := range strings.Split(entities, ";") {
			if _, ok := queryToConn[entity]; !ok {
				title, _ := result.Metadata["title"].(string)
				logger.Info("found first hop match", "doc", title, "url", url)
				queryToConn[entity] = []api.Connection{{DocTitle: title, DocUrl: url}}
			}
		}
	}

	// Second map to avoid mutating the original while iterating over it
	level2Entities := make(map[string][]api.Connection)

	for query, level1Entity := range queryToConn {
		results, err := flagger.auxNDB.Query(query, 5, nil)
		if err != nil {
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
					logger.Info("found second hop match", "doc", title, "url", url)
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
			return nil, fmt.Errorf("error querying ndb: %w", err)
		}

		for _, result := range results {
			if !strings.Contains(result.Text, query) {
				continue
			}

			title, _ := result.Metadata["title"].(string)
			url, _ := result.Metadata["url"].(string)
			entities, _ := result.Metadata["entities"].(string)

			logger.Info("found complex connection", "doc", title, "url", url, "steps", conns)

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

	logger.Info("second/third level flags", "n_flags", len(flags))

	return flags, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) Flag(logger *slog.Logger, authorName string, works []openalex.Work) ([]api.Flag, error) {
	logger.Info("checking if author is associated with EOC", "author_name", authorName)

	firstSecondLevelFlags, err := flagger.findFirstSecondHopEntities(logger, authorName, works)
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

	logger.Info("finished checking if author is associated with EOC", "n_flags", len(flags))

	return flags, nil
}
