package flaggers

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/search"
	"sync"
)

type ReportProcessor struct {
	openalex                openalex.KnowledgeBase
	workFlaggers            []WorkFlagger
	authorFacultyAtEOC      *AuthorIsFacultyAtEOCFlagger
	authorAssociatedWithEOC *AuthorIsAssociatedWithEOCFlagger
}

type ReportProcessorOptions struct {
	UniversityNDB search.NeuralDB
	DocNDB        search.NeuralDB
	AuxNDB        search.NeuralDB

	EntityLookup *EntityStore

	ConcerningEntities     eoc.EocSet
	ConcerningInstitutions eoc.EocSet
	ConcerningFunders      eoc.EocSet
	ConcerningPublishers   eoc.EocSet

	SussyBakas []string

	GrobidEndpoint string

	WorkDir string
}

// TODO(Nicholas): How to do cleanup for this, or just let it get cleaned up at the end of the process?
func NewReportProcessor(opts ReportProcessorOptions) (*ReportProcessor, error) {
	authorCache, err := NewCache[openalex.Author]("authors", filepath.Join(opts.WorkDir, "authors.cache"))
	if err != nil {
		return nil, fmt.Errorf("error loading author cache: %w", err)
	}
	ackCache, err := NewCache[Acknowledgements]("acks", filepath.Join(opts.WorkDir, "acks.cache"))
	if err != nil {
		return nil, fmt.Errorf("error loading ack cache: %w", err)
	}

	return &ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		workFlaggers: []WorkFlagger{
			&OpenAlexMultipleAffiliationsFlagger{},
			&OpenAlexFunderIsEOC{
				concerningFunders:  opts.ConcerningFunders,
				concerningEntities: opts.ConcerningEntities,
			},
			&OpenAlexPublisherIsEOC{
				concerningPublishers: opts.ConcerningPublishers,
			},
			&OpenAlexCoauthorIsEOC{
				concerningEntities: opts.ConcerningEntities,
			},
			&OpenAlexAuthorAffiliationIsEOC{
				concerningEntities:     opts.ConcerningEntities,
				concerningInstitutions: opts.ConcerningInstitutions,
			},
			&OpenAlexCoauthorAffiliationIsEOC{
				concerningEntities:     opts.ConcerningEntities,
				concerningInstitutions: opts.ConcerningInstitutions,
			},
			&OpenAlexAcknowledgementIsEOC{
				openalex:     openalex.NewRemoteKnowledgeBase(),
				entityLookup: opts.EntityLookup,
				authorCache:  authorCache,
				extractor:    NewGrobidExtractor(ackCache, opts.GrobidEndpoint, opts.WorkDir),
				sussyBakas:   opts.SussyBakas,
			},
		},
		authorFacultyAtEOC: &AuthorIsFacultyAtEOCFlagger{
			universityNDB: opts.UniversityNDB,
		},
		authorAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlagger{
			docNDB: opts.DocNDB,
			auxNDB: opts.AuxNDB,
		},
	}, nil
}

func (processor *ReportProcessor) getWorkStream(report reports.ReportUpdateTask) (chan openalex.WorkBatch, error) {
	switch report.Source {
	case api.OpenAlexSource:
		return streamOpenAlexWorks(processor.openalex, report.AuthorId, report.StartDate, report.EndDate), nil
	case api.GoogleScholarSource:
		return streamGScholarWorks(processor.openalex, report.AuthorName, report.AuthorId, report.StartDate, report.EndDate), nil
	// case api.UnstructuredSource:
	// 	return streamUnstructuredWorks(processor.openalex, report.AuthorName, "what should the text be", report.StartYear, report.EndYear), nil
	// case api.ScopusSource:
	// 	return streamScopusWorks()
	default:
		return nil, fmt.Errorf("invalid report source '%s'", report.Source)
	}
}

func (processor *ReportProcessor) processWorks(logger *slog.Logger, authorName string, workStream chan openalex.WorkBatch, flagsCh chan []api.Flag) {
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

			go func(flagger WorkFlagger, works []openalex.Work, authorIds []string) {
				defer wg.Done()

				logger := logger.With("flagger", flagger.Name(), "batch", batch)
				logger.Info("starting batch with flagger")

				flags, err := flagger.Flag(logger, works, authorIds)
				if err != nil {
					logger.Error("flagger error", "error", err)
				} else {
					flagsCh <- flags
					logger.Info("batch complete")
				}
			}(flagger, works.Works, works.TargetAuthorIds)
		}

		wg.Add(1)
		go func(batch int, works []openalex.Work) {
			defer wg.Done()

			flagger := processor.authorAssociatedWithEOC
			if flagger == nil {
				return
			}

			logger := logger.With("flagger", flagger.Name(), "batch", batch)
			logger.Info("starting batch with flagger")

			flags, err := flagger.Flag(logger, authorName, works)
			if err != nil {
				logger.Error("flagger error", "error", err)
			} else {
				flagsCh <- flags
				logger.Info("batch complete")
			}

		}(batch, works.Works)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		flagger := processor.authorFacultyAtEOC
		if flagger == nil {
			return
		}

		logger := logger.With("flagger", flagger.Name())
		logger.Info("starting author faculty at eoc with flagger")

		flags, err := flagger.Flag(logger, authorName)
		if err != nil {
			logger.Error("flagger error", "error", err)
		} else {
			flagsCh <- flags
			logger.Info("batch complete")
		}
	}()

	wg.Wait()
	close(flagsCh)
}

func (processor *ReportProcessor) ProcessReport(report reports.ReportUpdateTask) (api.ReportContent, error) {
	logger := slog.With("report_id", report.Id)

	logger.Info("starting report processing")

	workStream, err := processor.getWorkStream(report)
	if err != nil {
		logger.Error("unable to get work stream", "error", err)
		return api.ReportContent{}, fmt.Errorf("unable to get works: %w", err)
	}

	flagsSeen := make(map[string]bool)
	flagsCh := make(chan []api.Flag, 100)

	go processor.processWorks(logger, report.AuthorName, workStream, flagsCh)

	content := make(api.ReportContent)
	for flags := range flagsCh {
		for _, flag := range flags {
			if key := flag.Key(); !flagsSeen[key] {
				flagsSeen[key] = true

				content[string(flag.Type())] = append(content[string(flag.Type())], flag)
			}
		}
	}

	logger.Info("report complete", "n_flags", len(flagsSeen))

	return content, nil
}
