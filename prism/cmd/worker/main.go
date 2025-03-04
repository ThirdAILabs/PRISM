package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"prism/prism/cmd"
	"prism/prism/reports"
	"prism/prism/reports/flaggers"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/search"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	PostgresUri string `env:"DB_URI,notEmpty,required"`
	Logfile     string `env:"LOGFILE,notEmpty" envDefault:"prism_worker.log"`
	NdbLicense  string `env:"NDB_LICENSE,notEmpty,required"`

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

func main() {
	cmd.LoadEnvFile()

	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

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
		UniversityNDB: flaggers.BuildUniversityNDB(config.UniversityData, filepath.Join(ndbDir, "university.ndb")),
		DocNDB:        flaggers.BuildDocNDB(config.DocData, filepath.Join(ndbDir, "doc.ndb")),
		AuxNDB:        flaggers.BuildAuxNDB(config.AuxData, filepath.Join(ndbDir, "aux.ndb")),

		EntityLookup: entityStore,

		ConcerningEntities:     eoc.LoadGeneralEOC(),
		ConcerningInstitutions: eoc.LoadInstitutionEOC(),
		ConcerningFunders:      eoc.LoadFunderEOC(),
		ConcerningPublishers:   eoc.LoadPublisherEOC(),
		SussyBakas:             eoc.LoadSussyBakas(),

		GrobidEndpoint: config.GrobidEndpoint,
		WorkDir:        config.WorkDir,
	}

	db := cmd.InitDb(config.PostgresUri)

	reportManager := reports.NewManager(db, reports.StaleReportThreshold)

	processor, err := flaggers.NewReportProcessor(reportManager, opts)
	if err != nil {
		log.Fatalf("error creating work processor: %v", err)
	}

	for {
		foundAuthorReport := processor.ProcessNextAuthorReport()
		foundUniversityReport := processor.ProcessNextUniversityReport()

		if !foundAuthorReport && !foundUniversityReport {
			time.Sleep(10 * time.Second)
		}
	}
}
