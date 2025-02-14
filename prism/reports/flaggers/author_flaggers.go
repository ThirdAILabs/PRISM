package flaggers

import (
	"fmt"
	"log/slog"
	"prism/openalex"
	"prism/search"
	"slices"
	"strings"
)

type AuthorIsFacultyAtEOCFlagger struct {
	universityNDB search.NeuralDB
}

type nameMatcher struct {
	nameParts []string
}

func newNameMatcher(name string) nameMatcher {
	return nameMatcher{
		nameParts: strings.Fields(strings.ToLower(name)),
	}
}

func (n *nameMatcher) matches(candidate string) bool {
	candidate = strings.ToLower(candidate)
	for _, part := range n.nameParts {
		if !strings.Contains(candidate, part) {
			return false
		}
	}
	return true
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Name() flagType {
	return AuthorIsFacultyAtEOC
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Flag(logger *slog.Logger, authorName string) ([]Flag, error) {
	logger.Info("checking if author is faculty at EOC", "author_name", authorName)

	results, err := flagger.universityNDB.Query(authorName, 5, nil)
	if err != nil {
		logger.Error("error querying ndb", "error", err)
		return nil, fmt.Errorf("error querying ndb: %w", err)
	}

	matcher := newNameMatcher(authorName)

	flags := make([]Flag, 0)

	for _, result := range results {
		if matcher.matches(result.Text) {
			university, _ := result.Metadata["university"].(string)
			if university == "" {
				logger.Error("missing university metadata", "result", result)
				continue
			}

			url, _ := result.Metadata["url"].(string)

			flags = append(flags, &AuthorIsFacultyAtEOCFlag{
				FlagTitle:     "Person may be affiliated with this university",
				FlagMessage:   fmt.Sprintf("The author %s may be associated with this concerning entity: %s\n", authorName, university),
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

func (flagger *AuthorIsAssociatedWithEOCFlagger) Name() flagType {
	return AuthorIsAssociatedWithEOC
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

func (flagger *AuthorIsAssociatedWithEOCFlagger) findFirstSecondHopEntities(logger *slog.Logger, authorName string, works []openalex.Work) ([]Flag, error) {
	flags := make([]Flag, 0)

	seen := make(map[string]bool)

	primaryMatcher := newNameMatcher(authorName)

	frequentAuthors := topCoauthors(works)
	for _, author := range frequentAuthors {

		matcher := newNameMatcher(author.author)

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
				flags = append(flags, &AuthorIsAssociatedWithEOCFlag{
					FlagTitle:       "Person may be affiliated with someone mentioned in a press release.",
					FlagMessage:     "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:        title,
					DocUrl:          url,
					DocEntities:     strings.Split(entities, ";"),
					EntityMentioned: strings.ToTitle(author.author),
					ConnectionLevel: "primary",
				})
				logger.Info("author is assoiciated with EOC", "author", author.author, "doc", title, "entities", entities)
			} else {
				coauthor := strings.ToTitle(author.author)
				flags = append(flags, &AuthorIsAssociatedWithEOCFlag{
					FlagTitle:        "The author's frequent coauthor may be mentioned in a press release.",
					FlagMessage:      "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:         title,
					DocUrl:           url,
					DocEntities:      strings.Split(entities, ";"),
					EntityMentioned:  coauthor,
					ConnectionLevel:  "secondary",
					Nodes:            []Node{{DocTitle: coauthor + " (frequent coauthor)", DocUrl: ""}},
					FrequentCoauthor: &coauthor,
				})
				logger.Info("frequent coauthor is assoiciated with EOC", "coauthor", author.author, "doc", title, "entities", entities)
			}
		}
	}

	logger.Info("first/second level flags", "n_flags", len(flags))

	return flags, nil
}

type entityMetadata struct {
	connection string
	nodes      []Node
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) findSecondThirdHopEntities(logger *slog.Logger, authorName string) ([]Flag, error) {
	seen := make(map[string]bool)

	primaryMatcher := newNameMatcher(authorName)

	queryToEntities := make(map[string]entityMetadata)

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
			if _, ok := queryToEntities[entity]; !ok {
				title, _ := result.Metadata["title"].(string)
				queryToEntities[entity] = entityMetadata{
					connection: "secondary",
					nodes:      []Node{{DocTitle: title, DocUrl: url}},
				}
			}
		}
	}

	// Second map to avoid mutating the original while iterating over it
	level2Entities := make(map[string]entityMetadata)

	for query, level1Entity := range queryToEntities {
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

					level2Entities[entity] = entityMetadata{
						connection: "tertiary",
						nodes:      append(level1Entity.nodes, Node{DocTitle: title, DocUrl: url}),
					}
				}
			}
		}
	}

	for k, v := range level2Entities {
		queryToEntities[k] = v
	}

	flags := make([]Flag, 0)

	for query, entity := range queryToEntities {
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

			flags = append(flags, &AuthorIsAssociatedWithEOCFlag{
				FlagTitle:       "Author may be affiliated with an entity whose associate may be mentioned in a press release.",
				FlagMessage:     "The author may be associated be an entity who/which may be mentioned in a press release.\n",
				DocTitle:        title,
				DocUrl:          url,
				DocEntities:     strings.Split(entities, ";"),
				EntityMentioned: query,
				ConnectionLevel: entity.connection,
				Nodes:           entity.nodes,
			})
		}
	}

	logger.Info("second/third level flags", "n_flags", len(flags))

	return flags, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) Flag(logger *slog.Logger, authorName string, works []openalex.Work) ([]Flag, error) {
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
