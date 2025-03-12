package flaggers

import (
	"fmt"
	"log/slog"
	"prism/prism/gscholar"
	"prism/prism/llms"
	"prism/prism/openalex"
	"prism/prism/reports"
	"regexp"
	"strings"
	"time"
)

func streamOpenAlexWorks(openalex openalex.KnowledgeBase, authorId string, startDate, endDate time.Time) chan openalex.WorkBatch {
	defer reports.LogTiming("streamOpenAlexWorks")()
	return openalex.StreamWorks(authorId, startDate, endDate)
}

func findOAAuthorId(work openalex.Work, targetAuthorName string) string {
	defer reports.LogTiming("findOAAuthorId")()
	authorId := ""
	maxSim := 0.0

	for _, author := range work.Authors {
		name := author.DisplayName
		if len(name) == 0 && author.RawAuthorName != nil {
			name = *author.RawAuthorName
		}

		if sim := IndelSimilarity(name, targetAuthorName); sim > maxSim {
			maxSim = sim
			authorId = author.AuthorId
		}
	}

	return authorId
}

func findTargetAuthorIds(works []openalex.Work, targetAuthorName string) []string {
	defer reports.LogTiming("findTargetAuthorIds")()
	targetAuthorIds := make([]string, 0)
	for _, work := range works {
		if oaId := findOAAuthorId(work, targetAuthorName); len(oaId) > 0 {
			targetAuthorIds = append(targetAuthorIds, oaId)
		}
	}
	return targetAuthorIds
}

func streamGScholarWorks(oa openalex.KnowledgeBase, authorName, gScholarAuthorId string, startDate, endDate time.Time) chan openalex.WorkBatch {
	defer reports.LogTiming("streamGScholarWorks")()
	outputCh := make(chan openalex.WorkBatch, 10)

	go func() {
		defer reports.LogTiming("streamGScholarWorks goroutine")()
		defer close(outputCh)

		workTitleIterator := gscholar.NewAuthorPaperIterator(gScholarAuthorId)
		for {
			batch, err := workTitleIterator.Next()
			if err != nil {
				slog.Error("error iterating over work titles in google scholar", "error", err)
				outputCh <- openalex.WorkBatch{Works: nil, TargetAuthorIds: nil, Error: err}
				break
			}
			if batch == nil {
				break
			}

			works, err := oa.FindWorksByTitle(batch, startDate, endDate)
			if err != nil {
				slog.Error("error getting works from openalex", "error", err)
				outputCh <- openalex.WorkBatch{Works: nil, TargetAuthorIds: nil, Error: err}
				break
			}

			outputCh <- openalex.WorkBatch{Works: works, TargetAuthorIds: findTargetAuthorIds(works, authorName)}
		}
	}()

	return outputCh
}

const extractTitlesPromptTemplate = `Extract all academic paper titles from this snippet. Return each title in a block like this:

[TITLE START] <title 1> [TITLE END]
[TITLE START] <title 2> [TITLE END]

And so on. Here comes the snippet:

%s
`

//lint:ignore U1000 streamUnstructuredWorks
func streamUnstructuredWorks(oa openalex.KnowledgeBase, authorName, text string, startDate, endDate time.Time) chan openalex.WorkBatch {
	defer reports.LogTiming("streamUnstructuredWorks")()
	outputCh := make(chan openalex.WorkBatch, 10)

	go func() {
		defer reports.LogTiming("streamUnstructuredWorks goroutine")()
		defer close(outputCh)

		llm := llms.New()

		answer, err := llm.Generate(fmt.Sprintf(extractTitlesPromptTemplate, text), &llms.Options{Model: llms.GPT4oMini})
		if err != nil {
			slog.Error("error getting title extraction response", "error", err)
			outputCh <- openalex.WorkBatch{Works: nil, TargetAuthorIds: nil, Error: fmt.Errorf("error extracting titles: %w", err)}
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
			works, err := oa.FindWorksByTitle(titles[i:min(len(titles), i+batchSize)], startDate, endDate)
			if err != nil {
				slog.Error("error finding works for titles", "error", err)
				outputCh <- openalex.WorkBatch{Works: nil, TargetAuthorIds: nil, Error: fmt.Errorf("error finding works: %w", err)}
				break
			}

			outputCh <- openalex.WorkBatch{Works: works, TargetAuthorIds: findTargetAuthorIds(works, authorName), Error: nil}
		}
	}()

	return outputCh
}

//lint:ignore U1000 streamScopusWorks
func streamScopusWorks(oa openalex.KnowledgeBase, authorName string, titles []string, startDate, endDate time.Time) chan openalex.WorkBatch {
	defer reports.LogTiming("streamScopusWorks")()
	outputCh := make(chan openalex.WorkBatch, 10)

	go func() {
		defer reports.LogTiming("streamScopusWorks goroutine")()
		defer close(outputCh)

		const batchSize = 20
		for i := 0; i < len(titles); i += batchSize {
			works, err := oa.FindWorksByTitle(titles[i:min(len(titles), i+batchSize)], startDate, endDate)
			if err != nil {
				slog.Error("error finding works for titles", "error", err)
				outputCh <- openalex.WorkBatch{Works: nil, TargetAuthorIds: nil, Error: fmt.Errorf("error finding works: %w", err)}
				break
			}

			outputCh <- openalex.WorkBatch{Works: works, TargetAuthorIds: findTargetAuthorIds(works, authorName), Error: nil}
		}
	}()

	return outputCh
}
