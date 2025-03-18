package flaggers

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/schema"
	"prism/prism/search"
	"prism/prism/triangulation"
	"sync"
	"time"
)

type ReportProcessor struct {
	openalex                openalex.KnowledgeBase
	workFlaggers            []WorkFlagger
	authorFacultyAtEOC      *AuthorIsFacultyAtEOCFlagger
	authorAssociatedWithEOC *AuthorIsAssociatedWithEOCFlagger
	manager                 *reports.ReportManager
}

type ReportProcessorOptions struct {
	UniversityNDB   search.NeuralDB
	DocNDB          search.NeuralDB
	AuxNDB          search.NeuralDB
	TriangulationDB *triangulation.TriangulationDB

	EntityLookup *EntityStore

	ConcerningEntities     eoc.EocSet
	ConcerningInstitutions eoc.EocSet
	ConcerningFunders      eoc.EocSet
	ConcerningPublishers   eoc.EocSet

	SussyBakas []string

	GrobidEndpoint string

	WorkDir string

	MaxDownloadThreads int
	MaxGrobidThreads   int
}

// TODO(Nicholas): How to do cleanup for this, or just let it get cleaned up at the end of the process?
func NewReportProcessor(manager *reports.ReportManager, opts ReportProcessorOptions) (*ReportProcessor, error) {
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
			&OpenAlexFunderIsEOC{
				concerningFunders:  opts.ConcerningFunders,
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
				openalex:        openalex.NewRemoteKnowledgeBase(),
				entityLookup:    opts.EntityLookup,
				authorCache:     authorCache,
				extractor:       NewGrobidExtractor(ackCache, opts.GrobidEndpoint, opts.WorkDir, opts.MaxDownloadThreads, opts.MaxGrobidThreads),
				sussyBakas:      opts.SussyBakas,
				triangulationDB: opts.TriangulationDB,
			},
		},
		authorFacultyAtEOC: &AuthorIsFacultyAtEOCFlagger{
			universityNDB: opts.UniversityNDB,
		},
		authorAssociatedWithEOC: &AuthorIsAssociatedWithEOCFlagger{
			docNDB: opts.DocNDB,
			auxNDB: opts.AuxNDB,
		},
		manager: manager,
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

func (processor *ReportProcessor) processWorks(logger *slog.Logger, authorName string, workStream chan openalex.WorkBatch, flagsCh chan []api.Flag, ForUniversityReport bool) {
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

			if ForUniversityReport && flagger.DisableForUniversityReport() {
				continue
			}

			wg.Add(1)

			go func(flagger WorkFlagger, works []openalex.Work, authorIds []string) {
				defer wg.Done()

				logger := logger.With("flagger", flagger.Name(), "batch", batch)

				flags, err := flagger.Flag(logger, works, authorIds)
				if err != nil {
					logger.Error("flagger error", "error", err)
				} else {
					flagsCh <- flags
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

			flags, err := flagger.Flag(logger, authorName, works)
			if err != nil {
				logger.Error("flagger error", "error", err)
			} else {
				flagsCh <- flags
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

		flags, err := flagger.Flag(logger, authorName)
		if err != nil {
			logger.Error("flagger error", "error", err)
		} else {
			flagsCh <- flags
		}
	}()

	wg.Wait()
	close(flagsCh)
}

func (processor *ReportProcessor) ProcessAuthorReport(report reports.ReportUpdateTask) {
	logger := slog.With("report_id", report.Id)

	logger.Info("starting report processing", "author_id", report.AuthorId, "author_name", report.AuthorName, "source", report.Source, "is_university_queued", report.ForUniversityReport)

	workStream, err := processor.getWorkStream(report)
	if err != nil {
		logger.Error("report failed: unable to get author works", "error", err)
		if err := processor.manager.UpdateAuthorReport(report.Id, schema.ReportFailed, time.Time{}, nil); err != nil {
			slog.Error("error updating author report status to failed", "error", err)
		}
		return
	}

	flagsCh := make(chan []api.Flag, 100)

	go processor.processWorks(logger, report.AuthorName, workStream, flagsCh, report.ForUniversityReport)

	seen := make(map[[sha256.Size]byte]struct{})
	flagCounts := make(map[string]int)
	for flags := range flagsCh {
		for _, flag := range flags {
			hash := flag.Hash()
			if _, ok := seen[hash]; !ok {
				seen[hash] = struct{}{}
				flagCounts[flag.Type()]++
			}
		}

		if len(flags) > 0 {
			slog.Info("received batch of flags", "type", flags[0].Type(), "n_flags", len(flags))
			if err := processor.manager.UpdateAuthorReport(report.Id, schema.ReportInProgress, report.EndDate, flags); err != nil {
				slog.Error("error updating author report status for partial flags", "error", err)
			}
		}
	}

	attrs := make([]any, 0, len(flagCounts)+1)
	attrs = append(attrs, slog.Int("n_flags", len(seen)))
	for flagType, count := range flagCounts {
		attrs = append(attrs, slog.Int(flagType, count))
	}

	logger.Info("report complete", attrs...)

	if err := processor.manager.UpdateAuthorReport(report.Id, schema.ReportCompleted, report.EndDate, nil); err != nil {
		slog.Error("error updating author report status to complete", "error", err)
	}
}

func (processor *ReportProcessor) ProcessNextAuthorReport() bool {
	report, err := processor.manager.GetNextAuthorReport()
	if err != nil {
		slog.Error("error checking for next report", "error", err)
		return false
	}
	if report == nil {
		return false
	}

	processor.ProcessAuthorReport(*report)

	return true
}

func (processor *ReportProcessor) getUniversityAuthors(report reports.UniversityReportUpdateTask) ([]reports.UniversityAuthorReport, error) {
	authors, err := processor.openalex.GetInstitutionAuthors(report.UniversityId, time.Now().AddDate(-4, 0, 0), time.Now())
	if err != nil {
		return nil, err
	}

	output := make([]reports.UniversityAuthorReport, 0, len(authors))
	for _, author := range authors {
		output = append(output, reports.UniversityAuthorReport{
			AuthorId:   author.AuthorId,
			AuthorName: author.AuthorName,
			Source:     api.OpenAlexSource,
		})
	}

	return output, nil
}

func (processor *ReportProcessor) ProcessNextUniversityReport() bool {
	nextReport, err := processor.manager.GetNextUniversityReport()
	if err != nil {
		slog.Error("error checking for next report", "error", err)
		return false
	}
	if nextReport == nil {
		return false
	}

	slog.Info("processing university report", "report_id", nextReport.Id, "university_report_id", nextReport.UniversityId, "university_name", nextReport.UniversityName)

	authors, err := processor.getUniversityAuthors(*nextReport)
	if err != nil {
		slog.Error("error processing university report: %w")

		if err := processor.manager.UpdateUniversityReport(nextReport.Id, schema.ReportFailed, time.Time{}, nil); err != nil {
			slog.Error("error updating report status to failed", "error", err)
		}
		return true
	}

	slog.Info("authors found for university report", "n_authors", len(authors))

	if err := processor.manager.UpdateUniversityReport(nextReport.Id, schema.ReportCompleted, nextReport.UpdateDate, authors); err != nil {
		slog.Error("error updating university report status to complete", "error", err)
	}

	slog.Info("university report complete", "report_id", nextReport.Id, "university_report_id", nextReport.UniversityId, "university_name", nextReport.UniversityName)

	return true
}
