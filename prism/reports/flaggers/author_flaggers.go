package flaggers

import (
	"fmt"
	"log/slog"
	"prism/api"
	"prism/ndb"
	"prism/openalex"
	"slices"
	"strings"
)

type AuthorIsFacultyAtEOCFlagger struct {
	entityDB ndb.NeuralDB
}

type nameMatcher struct {
	nameParts []string
}

func newNameMatcher(name string) nameMatcher {
	return nameMatcher{
		nameParts: strings.Split(strings.ToLower(name), " "),
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

func (flagger *AuthorIsFacultyAtEOCFlagger) Flag(author api.Author) ([]Flag, error) {
	results, err := flagger.entityDB.Query(author.DisplayName, 5, nil)
	if err != nil {
		return nil, fmt.Errorf("error querying ndb: %w", err)
	}

	matcher := newNameMatcher(author.DisplayName)

	flags := make([]Flag, 0)

	for _, result := range results {
		if matcher.matches(result.Text) {
			university, _ := result.Metadata["university"].(string)
			if university == "" {
				slog.Error("missing university metadata", "result", result)
				continue
			}

			url, _ := result.Metadata["url"].(string)

			flags = append(flags, Flag{
				FlaggerType:   AuthorIsFacultyAtEOC,
				Title:         "Person may be affiliated with this university",
				Message:       fmt.Sprintf("The author %s may be associated with this concerning entity: %s\n", author.DisplayName, university),
				UniversityUrl: url,
				Affiliations:  []string{university},
				Metadata: map[string]any{
					"university": university,
					"url":        url,
					"entities":   []string{university},
				},
			})
		}
	}

	return flags, nil
}

type AuthorIsAssociatedWithEOCFlagger struct {
	prDB  ndb.NeuralDB
	auxDB ndb.NeuralDB
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

func (flagger *AuthorIsAssociatedWithEOCFlagger) findFirstSecondHopEntities(author api.Author, works []openalex.Work) ([]Flag, error) {
	flags := make([]Flag, 0)

	seen := make(map[string]bool)

	primaryMatcher := newNameMatcher(author.DisplayName)

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

			university, _ := result.Metadata["university"].(string)

			if primaryMatcher.matches(author.author) {
				flags = append(flags, Flag{
					FlaggerType:   AuthorIsAssociatedWithEOC,
					Title:         "Person may be affiliated with someone mentioned in a press release.",
					Message:       "The author or a frequent associate may be mentioned in a press release.",
					UniversityUrl: url,
					Affiliations:  []string{university},
					Metadata: map[string]any{
						"title":              result.Metadata["title"],
						"connection":         "primary",
						"url":                url,
						"entities":           result.Metadata["entities"],
						"entities_mentioned": []string{strings.ToTitle(author.author)},
					},
				})
			} else {
				flags = append(flags, Flag{
					FlaggerType:   AuthorIsAssociatedWithEOC,
					Title:         "The author's frequent coauthor may be mentioned in a press release.",
					Message:       "The author or a frequent associate may be mentioned in a press release.",
					UniversityUrl: "",
					Affiliations:  []string{},
					Metadata: map[string]any{
						"title":              result.Metadata["title"],
						"connection":         "secondary",
						"url":                url,
						"doj_url":            url,
						"doj_title":          result.Metadata["title"],
						"entities":           result.Metadata["entities"],
						"entities_mentioned": []string{strings.ToTitle(author.author)},
						"node1_title":        strings.ToTitle(author.author) + " (frequent coauthor)",
						"node1_url":          "",
						"frequent_coauthor":  strings.ToTitle(author.author),
					},
				})
			}
		}
	}

	return flags, nil
}

type entityMetadata struct {
	level      int
	node1Title string
	node1Url   string
	node2Title string
	node2Url   string
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) findSecondThirdHopEntities(author api.Author) ([]Flag, error) {
	seen := make(map[string]bool)

	primaryMatcher := newNameMatcher(author.DisplayName)

	queryToEntities := make(map[string]entityMetadata)

	results, err := flagger.auxDB.Query(author.DisplayName, 5, nil)
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
			entities, _ := result.Metadata["entities"].([]string)

			flag := Flag{
				FlaggerType:   AuthorIsAssociatedWithEOC,
				Title:         "Author may be affiliated with an entity whose associate may be mentioned in a press release.",
				Message:       "The author may be associated be an entity who/which may be mentioned in a press release.\n",
				UniversityUrl: "",
				Affiliations:  []string{},
				Metadata: map[string]any{
					"title":              title,
					"url":                url,
					"doj_url":            url,
					"doj_title":          title,
					"entities":           entities,
					"entities_mentioned": []string{query},
					"node1_title":        entity.node1Title,
					"node1_url":          entity.node1Url,
				},
			}
			if entity.level == 1 {
				flag.Metadata["connection"] = "secondary"
			} else {
				flag.Metadata["connection"] = "tertiary"
				flag.Metadata["node2_title"] = entity.node2Title
				flag.Metadata["node2_url"] = entity.node2Url
			}

			flags = append(flags, flag)
		}
	}

	return flags, nil
}

func (flagger *AuthorIsAssociatedWithEOCFlagger) Flag(author api.Author, works []openalex.Work) ([]Flag, error) {
	firstSecondLevelFlags, err := flagger.findFirstSecondHopEntities(author, works)
	if err != nil {
		return nil, err
	}

	secondThirdLevelFlags, err := flagger.findSecondThirdHopEntities(author)
	if err != nil {
		return nil, err
	}

	return slices.Concat(firstSecondLevelFlags, secondThirdLevelFlags), nil
}
