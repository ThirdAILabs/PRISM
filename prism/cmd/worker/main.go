package main

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/cmd"
	"prism/prism/reports"
	"prism/prism/reports/flaggers"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/search"
	"strings"
	"time"
)

type Config struct {
	PostgresUri string `yaml:"postgres_uri"`
	Logfile     string `yaml:"logfile"`
	NdbLicense  string `yaml:"ndb_license"`

	WorkDir string `yaml:"work_dir"`

	NDBData struct {
		University string `yaml:"university"`
		Doc        string `yaml:"doc"`
		Aux        string `yaml:"aux"`
	} `yaml:"ndb_data"`

	GrobidEndpoint string `yaml:"grobid_endpoint"`
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

	slog.Info("processing university report", "report_id", nextReport.ReportId, "university_report_id", nextReport.UniversityId, "university_name", nextReport.UniversityName)

	authors, err := processor.GetUniversityAuthors(*nextReport)
	if err != nil {
		slog.Error("error processing university report: %w")

		if err := reportManager.UpdateUnivesityReport(nextReport.ReportId, "failed", time.Time{}, nil); err != nil {
			slog.Error("error updating report status to failed", "error", err)
		}
		return true
	}

	slog.Info("authors found for university report", "n_authors", len(authors))

	if err := reportManager.UpdateUnivesityReport(nextReport.ReportId, "complete", time.Now(), authors); err != nil {
		slog.Error("error updating university report status to complete", "error", err)
	}

	slog.Info("university report complete", "report_id", nextReport.ReportId, "university_report_id", nextReport.UniversityId, "university_name", nextReport.UniversityName)

	return true
}

func main() {
	var config Config
	cmd.LoadConfig(&config)

	logFile, err := os.OpenFile(config.logfile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
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

	if err := os.MkdirAll(ndbDir, 0777); err != nil {
		log.Fatalf("error creating work dir: %v", err)
	}

	entityStore, err := flaggers.NewEntityStore(filepath.Join(ndbDir, "entity_lookup.ndb"), eoc.LoadSourceToAlias())
	if err != nil {
		log.Fatalf("error creating entity store: %v", err)
	}
	defer entityStore.Free()

	opts := flaggers.ReportProcessorOptions{
		UniversityNDB: flaggers.BuildUniversityNDB(config.NDBData.University, filepath.Join(ndbDir, "university.ndb")),
		DocNDB:        flaggers.BuildDocNDB(config.NDBData.Doc, filepath.Join(ndbDir, "doc.ndb")),
		AuxNDB:        flaggers.BuildAuxNDB(config.NDBData.Aux, filepath.Join(ndbDir, "aux.ndb")),

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
		if !processNextAuthorReport(reportManager, processor) &&
			!processNextUniversityReport(reportManager, processor) {
			time.Sleep(10 * time.Second)
		}
	}
}
