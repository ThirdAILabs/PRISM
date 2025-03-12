package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"prism/prism/cmd"
	"prism/prism/licensing"
	"prism/prism/reports"
	"prism/prism/reports/flaggers"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/search"
	"prism/prism/train_ndb"
	"prism/prism/triangulation"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	PostgresUri              string `env:"DB_URI,notEmpty,required"`
	FundcodeTriangulationUri string `env:"FUNDCODE_TRIANGULATION_DB_URI,notEmpty,required"`
	Logfile                  string `env:"LOGFILE,notEmpty" envDefault:"prism_worker.log"`
	PrismLicense             string `env:"PRISM_LICENSE,notEmpty,required"`

	WorkDir string `env:"WORK_DIR,notEmpty" envDefault:"./work"`

	UniversityData string `env:"UNIVERSITY_DATA,notEmpty,required"`
	DocData        string `env:"DOC_DATA,notEmpty,required"`
	AuxData        string `env:"AUX_DATA,notEmpty,required"`

	GrobidEndpoint string `env:"GROBID_ENDPOINT,notEmpty,required"`

	// This variable is directly loaded by the openai client library, it is just
	// listed here so that and error is raised if it's missing.
	OpenaiKey string `env:"OPENAI_API_KEY,notEmpty,required"`

	MaxDownloadThreads int `env:"MAX_DOWNLOAD_THREADS" envDefault:"40"`
	MaxGrobidThreads   int `env:"MAX_GROBID_THREADS" envDefault:"10"`

	RetrainNDBs bool `env:"RETRAIN_NDBS" envDefault:"true"`
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

	logFile, err := os.OpenFile(config.logfile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	cmd.InitLogging(logFile)

	licensing, err := licensing.NewLicenseVerifier(config.PrismLicense)
	if err != nil {
		log.Fatalf("error initializing licensing: %v", err)
	}

	if err := search.SetLicenseKey(config.PrismLicense); err != nil {
		log.Fatalf("error activating license key: %v", err)
	}

	ndbDir := filepath.Join(config.WorkDir, "ndbs")

	var universityNDB search.NeuralDB
	var docNDB search.NeuralDB
	var auxNDB search.NeuralDB

	if config.RetrainNDBs {
		universityNDB, docNDB, auxNDB = train_ndb.RetrainWorkerNDBs(ndbDir, config.UniversityData, config.DocData, config.AuxData)
	} else {
		universityNDB, err = search.NewNeuralDB(filepath.Join(ndbDir, "university.ndb"))
		if err != nil {
			log.Fatalf("error initializing university ndb: %v", err)
		}

		docNDB, err = search.NewNeuralDB(filepath.Join(ndbDir, "doc.ndb"))
		if err != nil {
			log.Fatalf("error initializing doc ndb: %v", err)
		}

		auxNDB, err = search.NewNeuralDB(filepath.Join(ndbDir, "aux.ndb"))
		if err != nil {
			log.Fatalf("error initializing aux ndb: %v", err)
		}
	}

	entityStore, err := flaggers.NewEntityStore(filepath.Join(ndbDir, "entity_lookup.ndb"), eoc.LoadSourceToAlias())
	if err != nil {
		log.Fatalf("error creating entity store: %v", err)
	}
	defer entityStore.Free()

	opts := flaggers.ReportProcessorOptions{
		UniversityNDB:   universityNDB,
		DocNDB:          docNDB,
		AuxNDB:          auxNDB,
		TriangulationDB: triangulation.CreateTriangulationDB(cmd.OpenDB(config.FundcodeTriangulationUri)),

		EntityLookup: entityStore,

		ConcerningEntities:     eoc.LoadGeneralEOC(),
		ConcerningInstitutions: eoc.LoadInstitutionEOC(),
		ConcerningFunders:      eoc.LoadFunderEOC(),
		ConcerningPublishers:   eoc.LoadPublisherEOC(),
		SussyBakas:             eoc.LoadSussyBakas(),

		GrobidEndpoint: config.GrobidEndpoint,
		WorkDir:        config.WorkDir,

		MaxDownloadThreads: config.MaxDownloadThreads,
		MaxGrobidThreads:   config.MaxGrobidThreads,
	}

	db := cmd.OpenDB(config.PostgresUri)

	reportManager := reports.NewManager(db, reports.StaleReportThreshold)

	processor, err := flaggers.NewReportProcessor(reportManager, opts)
	if err != nil {
		log.Fatalf("error creating work processor: %v", err)
	}

	lastLicenseCheck := time.Now()
	for {
		if time.Since(lastLicenseCheck) > 10*time.Minute {
			if err := licensing.VerifyLicense(); err != nil {
				slog.Error("error verifying license", "error", err)
				time.Sleep(5 * time.Minute)
				continue
			}
			lastLicenseCheck = time.Now()
		}

		foundAuthorReport := processor.ProcessNextAuthorReport()
		foundUniversityReport := processor.ProcessNextUniversityReport()

		if !foundAuthorReport && !foundUniversityReport {
			time.Sleep(10 * time.Second)
		}
	}
}
