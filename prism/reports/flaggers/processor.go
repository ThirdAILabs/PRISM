package flaggers

import (
	"fmt"
	"log/slog"
	"prism/api"
	"prism/openalex"
	"sync"
)

type ReportProcessor struct {
	openalex                openalex.KnowledgeBase
	workFlaggers            []WorkFlagger
	authorFacultyAtEOC      AuthorIsFacultyAtEOCFlagger
	authorAssociatedWithEOC AuthorIsAssociatedWithEOCFlagger
}

func (processor *ReportProcessor) getWorkStream(report api.Report) (chan openalex.WorkBatch, error) {
	switch report.Source {
	case api.OpenAlexSource:
		return streamOpenAlexWorks(processor.openalex, report.AuthorId, report.StartYear, report.EndYear), nil
	case api.GoogleScholarSource:
		return streamGScholarWorks(processor.openalex, report.AuthorName, report.AuthorId, report.StartYear, report.EndYear), nil
	// case api.UnstructuredSource:
	// 	return streamUnstructuredWorks(processor.openalex, report.AuthorName, "what should the text be", report.StartYear, report.EndYear), nil
	// case api.ScopusSource:
	// 	return streamScopusWorks()
	default:
		return nil, fmt.Errorf("invalid report source '%s'", report.Source)
	}
}

func (processor *ReportProcessor) processWorks(logger *slog.Logger, authorName string, workStream chan openalex.WorkBatch, flagsCh chan []Flag) {
	wg := sync.WaitGroup{}

	batch := -1
	for works := range workStream {
		batch++
		if works.Error != nil {
			logger.Error("error getting next batch of author works", "batch", batch, "error", works.Error)
			continue
		}
		logger.Info("got next batch of works", "batch", batch, "n_works", len(works.Works))
		for _, flagger := range processor.workFlaggers {
			wg.Add(1)

			go func(batch int, works []openalex.Work, authorIds []string) {
				defer wg.Done()

				logger.Info("starting batch with flagger", "flagger", flagger.Name(), "batch", batch)
				flags, err := flagger.Flag(works, authorIds)
				if err != nil {
					logger.Error("flagger error", "flagger", flagger.Name(), "batch", batch, "error", err)
				} else {
					flagsCh <- flags
					logger.Info("batch complete", "flagger", flagger.Name(), "batch", batch)
				}
			}(batch, works.Works, works.TargetAuthorIds)
		}

		wg.Add(1)
		go func(batch int, works []openalex.Work) {
			defer wg.Done()

			flagger := processor.authorAssociatedWithEOC

			logger.Info("starting batch with flagger", "flagger", flagger.Name(), "batch", batch)

			flags, err := flagger.Flag(authorName, works)
			if err != nil {
				logger.Error("flagger error", "flagger", flagger.Name(), "batch", batch, "error", err)
			} else {
				flagsCh <- flags
				logger.Info("batch complete", "flagger", flagger.Name(), "batch", batch)
			}

		}(batch, works.Works)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		flagger := processor.authorFacultyAtEOC

		logger.Info("starting batch with flagger", "flagger", flagger.Name(), "batch", batch)

		flags, err := flagger.Flag(authorName)
		if err != nil {
			logger.Error("flagger error", "flagger", flagger.Name(), "batch", batch, "error", err)
		} else {
			flagsCh <- flags
			logger.Info("batch complete", "flagger", flagger.Name(), "batch", batch)
		}
	}()

	wg.Wait()
	close(flagsCh)
}

func (processor *ReportProcessor) ProcessReport(report api.Report) (any, error) {
	logger := slog.With("report_id", report.Id)

	logger.Info("starting report processing")

	workStream, err := processor.getWorkStream(report)
	if err != nil {
		logger.Error("unable to get work stream", "error", err)
		return nil, fmt.Errorf("unable to get works: %w", err)
	}

	flagsCh := make(chan []Flag, 100)

	go processor.processWorks(logger, report.AuthorName, workStream, flagsCh)

	allFlags := make([]Flag, 0)
	for flags := range flagsCh {
		allFlags = append(allFlags, flags...)
	}

	logger.Info("report complete", "n_flags", len(allFlags))

	return allFlags, nil
}
