package flaggers

import (
	"fmt"
	"log/slog"
	"prism/api"
	"prism/gscholar"
	"prism/llms"
	"prism/openalex"
	"regexp"
	"strings"
)

func streamOpenAlexWorks(openalex openalex.KnowledgeBase, authorId string, startYear, endYear int) (chan []api.Work, chan error) {
	return openalex.FindWorks(authorId, startYear, endYear)
}

func streamGScholarWorks(openalex openalex.KnowledgeBase, authorId string, startYear, endYear int) (chan []api.Work, chan error) {
	workCh := make(chan []api.Work, 10)
	errorCh := make(chan error, 10)

	go func() {
		workTitleIterator := gscholar.NewAuthorPaperIterator(authorId)
		for {
			batch, err := workTitleIterator.Next()
			if err != nil {
				slog.Error("error iterating over work titles in google scholar", "erorr", err)
				errorCh <- err
				break
			}
			if batch == nil {
				break
			}

			works, err := openalex.FindWorksByTitle(batch, startYear, endYear)
			if err != nil {
				slog.Error("error getting works from openalex", "error", err)
				errorCh <- err
				break
			}

			workCh <- works
		}
	}()

	close(workCh)
	close(errorCh)

	return workCh, errorCh
}

const extractTitlesPromptTemplate = `Extract all academic paper titles from this snippet. Return each title in a block like this:

[TITLE START] <title 1> [TITLE END]
[TITLE START] <title 2> [TITLE END]

And so on. Here comes the snippet:

%s
`

func streamUnstructuredWorks(openalex openalex.KnowledgeBase, text string, startYear, endYear int) (chan []api.Work, chan error) {
	workCh := make(chan []api.Work, 10)
	errorCh := make(chan error, 10)

	go func() {
		llm := llms.New() // TODO: should this be using gpt-o1-mini

		answer, err := llm.Generate(fmt.Sprintf(extractTitlesPromptTemplate, text))
		if err != nil {
			slog.Error("error getting title extraction response", "error", err)
			errorCh <- fmt.Errorf("error extracting titles: %w", err)
			close(workCh)
			close(errorCh)
			return
		}

		re := regexp.MustCompile(`\[TITLE START\](.+)\[TITLE END\]`)
		matches := re.FindAllStringSubmatch(answer, -1)

		titles := make([]string, 0, len(matches))
		for _, match := range matches {
			match := strings.TrimSpace(match[1])
			if len(match) > 0 {
				titles = append(titles, match)
			}
		}

		const batchSize = 20
		for i := 0; i < len(titles); i += batchSize {
			works, err := openalex.FindWorksByTitle(titles[i:min(len(titles), i+batchSize)], startYear, endYear)
			if err != nil {
				slog.Error("error finding works for titles", "error", err)
				errorCh <- fmt.Errorf("error finding works: %w", err)
				break
			}

			workCh <- works
		}

		close(workCh)
		close(errorCh)
	}()

	return workCh, errorCh
}

func streamScopusWorks(openalex openalex.KnowledgeBase, titles []string, startYear, endYear int) (chan []api.Work, chan error) {
	workCh := make(chan []api.Work, 10)
	errorCh := make(chan error, 10)

	go func() {
		const batchSize = 20
		for i := 0; i < len(titles); i += batchSize {
			works, err := openalex.FindWorksByTitle(titles[i:min(len(titles), i+batchSize)], startYear, endYear)
			if err != nil {
				slog.Error("error finding works for titles", "error", err)
				errorCh <- fmt.Errorf("error finding works: %w", err)
				break
			}

			workCh <- works
		}

		close(workCh)
		close(errorCh)
	}()

	return workCh, errorCh
}
