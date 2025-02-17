package flaggers

import (
	"fmt"
	"log/slog"
	"prism/api"
	"prism/openalex"
	"prism/search"
	"sync"
)

type ReportProcessor struct {
	openalex                openalex.KnowledgeBase
	workFlaggers            []WorkFlagger
	authorFacultyAtEOC      AuthorIsFacultyAtEOCFlagger
	authorAssociatedWithEOC AuthorIsAssociatedWithEOCFlagger
}

type ReportProcessorOptions struct {
	UniversityNDB search.NeuralDB
	DocNDB        search.NeuralDB
	AuxNDB        search.NeuralDB

	EntityLookup *EntityStore

	ConcerningEntities     []string
	ConcerningInstitutions []string
	ConcerningFunders      []string
	ConcerningPublishers   []string

	SussyBakas []string

	GrobidEndpoint string

	WorkDir string
}

func convertToSet(list []string) eocSet {
	set := make(eocSet)
	for _, item := range list {
		set[item] = struct{}{}
	}
	return set
}

// TODO(Nicholas): How to do cleanup for this, or just let it get cleaned up at the end of the process?
func NewReportProcessor(opts ReportProcessorOptions) (*ReportProcessor, error) {
	// ackFlagCache, err := NewCache[cachedAckFlag]("ack_flags", filepath.Join(opts.WorkDir, "ack_flags.cache"))
	// if err != nil {
	// 	return nil, fmt.Errorf("error loading ack flag cache: %w", err)
	// }
	// authorCache, err := NewCache[openalex.Author]("authors", filepath.Join(opts.WorkDir, "authors.cache"))
	// if err != nil {
	// 	return nil, fmt.Errorf("error loading author cache: %w", err)
	// }
	// ackCache, err := NewCache[Acknowledgements]("acks", filepath.Join(opts.WorkDir, "acks.cache"))
	// if err != nil {
	// 	return nil, fmt.Errorf("error loading ack cache: %w", err)
	// }

	concerningEntities := convertToSet(opts.ConcerningEntities)
	concerningInstitutions := convertToSet(opts.ConcerningInstitutions)
	concerningFunders := convertToSet(opts.ConcerningFunders)
	concerningPublishers := convertToSet(opts.ConcerningPublishers)

	return &ReportProcessor{
		openalex: openalex.NewRemoteKnowledgeBase(),
		workFlaggers: []WorkFlagger{
			&OpenAlexMultipleAffiliationsFlagger{},
			&OpenAlexFunderIsEOC{
				concerningFunders:  concerningFunders,
				concerningEntities: concerningEntities,
			},
			&OpenAlexPublisherIsEOC{
				concerningPublishers: concerningPublishers,
			},
			&OpenAlexCoauthorIsEOC{
				concerningEntities: concerningEntities,
			},
			&OpenAlexAuthorAffiliationIsEOC{
				concerningEntities:     concerningEntities,
				concerningInstitutions: concerningInstitutions,
			},
			&OpenAlexCoauthorAffiliationIsEOC{
				concerningEntities:     concerningEntities,
				concerningInstitutions: concerningInstitutions,
			},
			// &OpenAlexAcknowledgementIsEOC{
			// 	openalex:     openalex.NewRemoteKnowledgeBase(),
			// 	entityLookup: opts.EntityLookup,
			// 	flagCache:    ackFlagCache,
			// 	authorCache:  authorCache,
			// 	extractor: &GrobidAcknowledgementsExtractor{
			// 		cache:          ackCache,
			// 		maxWorkers:     10,
			// 		grobidEndpoint: opts.GrobidEndpoint,
			// 		downloadDir:    opts.WorkDir,
			// 	},
			// 	sussyBakas: opts.SussyBakas,
			// },
		},
		authorFacultyAtEOC: AuthorIsFacultyAtEOCFlagger{
			universityNDB: opts.UniversityNDB,
		},
		authorAssociatedWithEOC: AuthorIsAssociatedWithEOCFlagger{
			docNDB: opts.DocNDB,
			auxNDB: opts.AuxNDB,
		},
	}, nil
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

func (processor *ReportProcessor) ProcessReport(report api.Report) ([]Flag, error) {
	logger := slog.With("report_id", report.Id)

	logger.Info("starting report processing")

	workStream, err := processor.getWorkStream(report)
	if err != nil {
		logger.Error("unable to get work stream", "error", err)
		return nil, fmt.Errorf("unable to get works: %w", err)
	}

	flagsSeen := make(map[string]bool)
	flagsCh := make(chan []Flag, 100)

	go processor.processWorks(logger, report.AuthorName, workStream, flagsCh)

	allFlags := make([]Flag, 0)
	for flags := range flagsCh {
		for _, flag := range flags {
			if key := flag.Key(); !flagsSeen[key] {
				flagsSeen[key] = true
				allFlags = append(allFlags, flag)
			}
		}
	}

	logger.Info("report complete", "n_flags", len(allFlags))

	return allFlags, nil
}
