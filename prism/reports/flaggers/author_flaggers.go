package flaggers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"prism/prism/api"
	"prism/prism/gscholar"
	"prism/prism/llms"
	"prism/prism/openalex"
	"prism/prism/search"
	"regexp"
	"slices"
	"strings"
	"time"
)

type AuthorFlagger interface {
	Flag(logger *slog.Logger, authorName string) ([]api.Flag, error)

	Name() string
}

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
				flags = append(flags, &api.MiscHighRiskAssociationFlag{
					Message:         "The author or a frequent associate may be mentioned in a press release.",
					DocTitle:        title,
					DocUrl:          url,
					DocEntities:     strings.Split(entities, ";"),
					EntityMentioned: author.author,
				})
			} else {
				flags = append(flags, &api.MiscHighRiskAssociationFlag{
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
	llm llms.LLM
}

func (flagger *AuthorNewsArticlesFlagger) Name() string {
	return "NewsArticles"
}

func fetchArticleWebpage(link string) (string, error) {
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

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request returned status: %d", res.StatusCode)
	}

	const maxBytes = 300000

	text := make([]byte, maxBytes) // TODO(question): this seems like a lot?
	n, err := io.ReadFull(res.Body, text)
	if err != nil && err != io.ErrUnexpectedEOF { // UnexpectedEOF is returned if < N bytes are read
		return "", err
	}

	return string(text[:n]), nil
}

const articleCheckTemplate = `I will give you an author name, the title of a news article, and part of the html webpage for the article. 
If the article is about the author, and indicates misconduct of some kind you must reply with just the word "flag".
If the article is not about the author, or does not indicate misconduct, you must reply with just the word "none".

Author Name: %s
Title: %s
Html Content: %s
`

type checkArticleTask struct {
	article    gscholar.NewsArticle
	authorName string
}

type checkArticleResult struct {
	article gscholar.NewsArticle
	flag    bool
}

func (flagger *AuthorNewsArticlesFlagger) checkArticle(task checkArticleTask) (checkArticleResult, error) {
	html, err := fetchArticleWebpage(task.article.Link)
	if err != nil {
		return checkArticleResult{article: task.article, flag: false}, err
	}

	response, err := flagger.llm.Generate(fmt.Sprintf(articleCheckTemplate, task.authorName, task.article.Title, html), nil)
	if err != nil {
		return checkArticleResult{article: task.article, flag: false}, err
	}

	return checkArticleResult{article: task.article, flag: strings.Contains(strings.ToLower(response), "flag")}, nil
}

const maxArticles = 20

func (flagger *AuthorNewsArticlesFlagger) Flag(logger *slog.Logger, authorName string) ([]api.Flag, error) {
	articles, err := gscholar.GetNewsArticles(authorName, "")
	if err != nil {
		logger.Error("error getting news articles for author", "author_name", authorName, "error", err)
		return nil, err
	}

	articles = articles[:min(maxArticles, len(articles))]

	queue := make(chan checkArticleTask, len(articles))

	for _, article := range articles {
		queue <- checkArticleTask{article: article, authorName: authorName}
	}
	close(queue)

	completed := make(chan CompletedTask[checkArticleResult], len(articles))

	RunInPool(flagger.checkArticle, queue, completed, 5)

	flags := make([]api.Flag, 0)
	for result := range completed {
		if result.Error != nil {
			logger.Error("error checking article", "error", err)
		} else {
			if result.Result.flag {
				flags = append(flags, &api.NewsArticleFlag{
					Message:     "The author may be mentioned in a news article that indicates misconduct.",
					Title:       result.Result.article.Title,
					Link:        result.Result.article.Link,
					ArticleDate: result.Result.article.Date,
				})
			}
		}
	}

	return flags, nil
}
