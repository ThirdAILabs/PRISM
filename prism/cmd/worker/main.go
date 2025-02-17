package main

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"prism/prism/cmd"
	"prism/prism/cmd/worker/utils"
	"prism/prism/reports"
	"prism/prism/reports/flaggers"
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

	entityStore, err := flaggers.NewEntityStore(filepath.Join(ndbDir, "entity_lookup.ndb"), utils.LoadSourceToAlias())
	if err != nil {
		log.Fatalf("error creating entity store: %v", err)
	}
	defer entityStore.Free()

	opts := flaggers.ReportProcessorOptions{
		UniversityNDB: utils.BuildUniversityNDB(config.NDBData.University, filepath.Join(ndbDir, "university.ndb")),
		DocNDB:        utils.BuildDocNDB(config.NDBData.Doc, filepath.Join(ndbDir, "doc.ndb")),
		AuxNDB:        utils.BuildAuxNDB(config.NDBData.Aux, filepath.Join(ndbDir, "aux.ndb")),

		EntityLookup: entityStore,

		ConcerningEntities:     utils.LoadGeneralEOC(),
		ConcerningInstitutions: utils.LoadInstitutionEOC(),
		ConcerningFunders:      utils.LoadFunderEOC(),
		ConcerningPublishers:   utils.LoadPublisherEOC(),
		SussyBakas:             utils.LoadSussyBakas(),

		GrobidEndpoint: config.GrobidEndpoint,
		WorkDir:        config.WorkDir,
	}

	processor, err := flaggers.NewReportProcessor(opts)
	if err != nil {
		log.Fatalf("error creating work processor: %v", err)
	}

	db := cmd.InitDb(config.PostgresUri)

	reportManager := reports.NewManager(db)

	for {
		time.Sleep(10 * time.Second)

		nextReport, err := reportManager.GetNextReport()
		if err != nil {
			slog.Error("error checking for next report", "error", err)
			continue
		}
		if nextReport == nil {
			continue
		}

		flags, err := processor.ProcessReport(*nextReport)
		if err != nil {
			slog.Error("error processing report: %w")

			if err := reportManager.UpdateReport(nextReport.Id, "failed", []byte(err.Error())); err != nil {
				slog.Error("error updating report status to failed", "error", err)
			}
			continue
		}

		content := reports.FormatReport(nextReport.AuthorName, flags)

		contentBytes, err := json.Marshal(content)
		if err != nil {
			slog.Error("error serializing report", "error", err)

			if err := reportManager.UpdateReport(nextReport.Id, "failed", []byte(err.Error())); err != nil {
				slog.Error("error updating report status to failed", "error", err)
			}
			continue
		}

		if err := reportManager.UpdateReport(nextReport.Id, "complete", contentBytes); err != nil {
			slog.Error("error updating report status to complete", "error", err)
		}
	}
}
