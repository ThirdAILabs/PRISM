package main

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"prism/prism/api"
	"prism/prism/cmd"
	"prism/prism/reports"
	"prism/prism/reports/flaggers"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/search"
	"prism/prism/triangulation"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	PostgresUri              string `env:"DB_URI,notEmpty,required"`
	FundcodeTriangulationUri string `yaml:"fundcode_triangulation_postgres_uri"`
	Logfile                  string `env:"LOGFILE,notEmpty" envDefault:"prism_worker.log"`
	NdbLicense               string `env:"NDB_LICENSE,notEmpty,required"`

	WorkDir string `env:"WORK_DIR,notEmpty" envDefault:"./work"`

	UniversityData string `env:"UNIVERSITY_DATA,notEmpty,required"`
	DocData        string `env:"DOC_DATA,notEmpty,required"`
	AuxData        string `env:"AUX_DATA,notEmpty,required"`

	GrobidEndpoint string `env:"GROBID_ENDPOINT,notEmpty,required"`

	// This variable is directly loaded by the openai client library, it is just
	// listed here so that and error is raised if it's missing.
	OpenaiKey string `env:"OPENAI_API_KEY,notEmpty,required"`
}

func (c *Config) logfile() string {
	if c.Logfile == "" {
		return "prism_backend.log"
	}
	return c.Logfile
}

func processNextAuthorReport(reportManager *reports.ReportManager, processor *flaggers.ReportProcessor) bool {
	nextReport, err := reportManager.GetNextAuthorReport()
	if err != nil {
		slog.Error("error checking for next report", "error", err)
		return false
	}
	if nextReport == nil {
		return false
	}

	content, err := processor.ProcessReport(*nextReport)
	if err != nil {
		slog.Error("error processing report: %w")

		if err := reportManager.UpdateAuthorReport(nextReport.Id, "failed", time.Time{}, api.ReportContent{}); err != nil {
			slog.Error("error updating report status to failed", "error", err)
		}
		return true
	}

	if err := reportManager.UpdateAuthorReport(nextReport.Id, "complete", nextReport.EndDate, content); err != nil {
		slog.Error("error updating report status to complete", "error", err)
	}
	return true
}

func processNextUniversityReport(reportManager *reports.ReportManager, processor *flaggers.ReportProcessor) bool {
	nextReport, err := reportManager.GetNextUniversityReport()
	if err != nil {
		slog.Error("error checking for next report", "error", err)
		return false
	}
	if nextReport == nil {
		return false
	}

	slog.Info("processing university report", "report_id", nextReport.Id, "university_report_id", nextReport.UniversityId, "university_name", nextReport.UniversityName)

	authors, err := processor.GetUniversityAuthors(*nextReport)
	if err != nil {
		slog.Error("error processing university report: %w")

		if err := reportManager.UpdateUniversityReport(nextReport.Id, "failed", time.Time{}, nil); err != nil {
			slog.Error("error updating report status to failed", "error", err)
		}
		return true
	}

	slog.Info("authors found for university report", "n_authors", len(authors))

	if err := reportManager.UpdateUniversityReport(nextReport.Id, "complete", nextReport.UpdateDate, authors); err != nil {
		slog.Error("error updating university report status to complete", "error", err)
	}

	slog.Info("university report complete", "report_id", nextReport.Id, "university_report_id", nextReport.UniversityId, "university_name", nextReport.UniversityName)

	return true
}

func main() {
	cmd.LoadEnvFile()

	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	logFile, err := os.OpenFile(config.logfile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	cmd.InitLogging(logFile)

	if strings.HasPrefix(config.NdbLicense, "file ") {
		err := search.SetLicensePath(strings.TrimPrefix(config.NdbLicense, "file "))
		if err != nil {
			log.Fatalf("error activating license at path '%s': %v", config.NdbLicense, err)
		}
	} else {
		err := search.SetLicenseKey(config.NdbLicense)
		if err != nil {
			log.Fatalf("error activating license: %v", err)
		}
	}

	ndbDir := filepath.Join(config.WorkDir, "ndbs")
	if err := os.RemoveAll(ndbDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("error deleting existing ndb dir '%s': %v", ndbDir, err)
	}

	if err := os.MkdirAll(ndbDir, 0o777); err != nil {
		log.Fatalf("error creating work dir: %v", err)
	}

	entityStore, err := flaggers.NewEntityStore(filepath.Join(ndbDir, "entity_lookup.ndb"), eoc.LoadSourceToAlias())
	if err != nil {
		log.Fatalf("error creating entity store: %v", err)
	}
	defer entityStore.Free()

	opts := flaggers.ReportProcessorOptions{
		UniversityNDB:   flaggers.BuildUniversityNDB(config.UniversityData, filepath.Join(ndbDir, "university.ndb")),
		DocNDB:          flaggers.BuildDocNDB(config.DocData, filepath.Join(ndbDir, "doc.ndb")),
		AuxNDB:          flaggers.BuildAuxNDB(config.AuxData, filepath.Join(ndbDir, "aux.ndb")),
		TriangulationDB: triangulation.CreateTriangulationDB(cmd.InitTriangulationDb(config.FundcodeTriangulationUri)),

		EntityLookup: entityStore,

		ConcerningEntities:     eoc.LoadGeneralEOC(),
		ConcerningInstitutions: eoc.LoadInstitutionEOC(),
		ConcerningFunders:      eoc.LoadFunderEOC(),
		ConcerningPublishers:   eoc.LoadPublisherEOC(),
		SussyBakas:             eoc.LoadSussyBakas(),

		GrobidEndpoint: config.GrobidEndpoint,
		WorkDir:        config.WorkDir,
	}

	processor, err := flaggers.NewReportProcessor(opts)
	if err != nil {
		log.Fatalf("error creating work processor: %v", err)
	}

	db := cmd.InitDb(config.PostgresUri)

	reportManager := reports.NewManager(db, reports.StaleReportThreshold)

	for {
		foundAuthorReport := processNextAuthorReport(reportManager, processor)
		foundUniversityReport := processNextUniversityReport(reportManager, processor)

		if !foundAuthorReport && !foundUniversityReport {
			time.Sleep(10 * time.Second)
		}
	}
}
