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
	entityDB search.NeuralDB
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

func (flagger *AuthorIsFacultyAtEOCFlagger) Name() string {
	return AuthorIsFacultyAtEOC
}

func (flagger *AuthorIsFacultyAtEOCFlagger) Flag(logger *slog.Logger, authorName string) ([]Flag, error) {
	logger.Info("checking if author is faculty at EOC", "author_name", authorName)

	results, err := flagger.entityDB.Query(authorName, 5, nil)
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

			flags = append(flags, &AuthorFlag{
				FlaggerType: AuthorIsFacultyAtEOC,
				Title:       "Person may be affiliated with this university",
				Message:     fmt.Sprintf("The author %s may be associated with this concerning entity: %s\n", authorName, university),
				AuthorIsFacultyAtEOC: &AuthorIsFacultyAtEOCFlag{
					University:    university,
					UniversityUrl: url,
				},
			})

			logger.Info("found author in listing for EOC university", "author_name", authorName, "university", university)
		}
	}

	logger.Info("finished checking for faculty at EOC", "n_flags", len(flags))

	return flags, nil
}

type AuthorIsAssociatedWithEOCFlagger struct {
	prDB  search.NeuralDB
	auxDB search.NeuralDB
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) Name() string {
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
		results, err := flagger.prDB.Query(author.author, 5, nil)
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
				flags = append(flags, &AuthorFlag{
					FlaggerType: AuthorIsAssociatedWithEOC,
					Title:       "Person may be affiliated with someone mentioned in a press release.",
					Message:     "The author or a frequent associate may be mentioned in a press release.",
					AuthorIsAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlag{
						DocTitle:          title,
						DocUrl:            url,
						DocEntities:       strings.Split(entities, ";"),
						EntitiesMentioned: []string{strings.ToTitle(author.author)},
						Connection:        "primary",
					},
				})
				logger.Info("author is assoiciated with EOC", "author", author.author, "doc", title, "entities", entities)
			} else {
				coauthor := strings.ToTitle(author.author)
				flags = append(flags, &AuthorFlag{
					FlaggerType: AuthorIsAssociatedWithEOC,
					Title:       "The author's frequent coauthor may be mentioned in a press release.",
					Message:     "The author or a frequent associate may be mentioned in a press release.",
					AuthorIsAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlag{
						DocTitle:          title,
						DocUrl:            url,
						DocEntities:       strings.Split(entities, ";"),
						EntitiesMentioned: []string{coauthor},
						Connection:        "secondary",
						Nodes:             []Node{{DocTitle: coauthor + " (frequent coauthor)", DocUrl: ""}},
						FrequentCoauthor:  &coauthor,
					},
				})
				logger.Info("frequent coauthor is assoiciated with EOC", "coauthor", author.author, "doc", title, "entities", entities)
			}
		}
	}

	logger.Info("first/second level flags", "n_flags", len(flags))

	return flags, nil
}

type entityMetadata struct {
	level      int
	node1Title string
	node1Url   string
	node2Title string
	node2Url   string
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) findSecondThirdHopEntities(logger *slog.Logger, authorName string) ([]Flag, error) {
	seen := make(map[string]bool)

	primaryMatcher := newNameMatcher(authorName)

	queryToEntities := make(map[string]entityMetadata)

	results, err := flagger.auxDB.Query(authorName, 5, nil)
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

		entities, _ := result.Metadata["entities"].([]string)
		for _, entity := range entities {
			if _, ok := queryToEntities[entity]; !ok {
				title, _ := result.Metadata["title"].(string)
				queryToEntities[entity] = entityMetadata{
					level:      1,
					node1Title: title,
					node1Url:   url,
				}
			}
		}
	}

	for query, level1Entity := range queryToEntities {
		results, err := flagger.auxDB.Query(query, 5, nil)
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

			entities, _ := result.Metadata["entities"].([]string)
			for _, entity := range entities {
				if _, ok := queryToEntities[entity]; !ok {
					title, _ := result.Metadata["title"].(string)

					queryToEntities[entity] = entityMetadata{
						level:      2,
						node1Title: level1Entity.node1Title,
						node1Url:   level1Entity.node1Url,
						node2Title: title,
						node2Url:   url,
					}
				}
			}
		}
	}

	logger.Info("queries at first/second level", "n_queries", len(queryToEntities))

	flags := make([]Flag, 0)

	for query, entity := range queryToEntities {
		results, err := flagger.prDB.Query(query, 5, nil)
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

			flag := &AuthorFlag{
				FlaggerType: AuthorIsAssociatedWithEOC,
				Title:       "Author may be affiliated with an entity whose associate may be mentioned in a press release.",
				Message:     "The author may be associated be an entity who/which may be mentioned in a press release.\n",
				AuthorIsAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlag{
					DocTitle:          title,
					DocUrl:            url,
					DocEntities:       strings.Split(entities, ";"),
					EntitiesMentioned: []string{query},
					Nodes: []Node{
						{DocTitle: entity.node1Title, DocUrl: entity.node1Url},
					},
				},
			}
			if entity.level == 1 {
				flag.AuthorIsAssociatedWithEOC.Connection = "secondary"
			} else {
				flag.AuthorIsAssociatedWithEOC.Connection = "tertiary"
				flag.AuthorIsAssociatedWithEOC.Nodes = append(flag.AuthorIsAssociatedWithEOC.Nodes, Node{
					DocTitle: entity.node2Title,
					DocUrl:   entity.node2Url,
				})
			}

			flags = append(flags, flag)
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
