package main

import (
	"log"
	"os"
	"path/filepath"

	"prism/prism/cmd"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/reports/flaggers"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/reports/utils"
	"prism/prism/triangulation"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	PostgresUri              string `env:"DB_URI,notEmpty,required"`
	FundcodeTriangulationUri string `env:"FUNDCODE_TRIANGULATION_DB_URI,notEmpty,required"`
	Logfile                  string `env:"LOGFILE,notEmpty" envDefault:"prism_worker.log"`

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

	S3Bucket string `env:"S3_BUCKET" envDefault:"thirdai-prism"`

	PpxApiKey string `env:"PPX_API_KEY" envDefault:""`
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

	entityStore := flaggers.BuildWatchlistEntityIndex(eoc.LoadSourceToAlias())

	authorCache, err := utils.NewCache[openalex.Author]("authors", filepath.Join(config.WorkDir, "authors.cache"))
	if err != nil {
		log.Fatalf("error creating author info cache: %v", err)
	}
	ackCache, err := utils.NewCache[flaggers.Acknowledgements]("acks", filepath.Join(config.WorkDir, "acks.cache"))
	if err != nil {
		log.Fatalf("error creating ack cache: %v", err)
	}

	db := cmd.OpenDB(config.PostgresUri)

	reportManager := reports.NewManager(db)

	concerningEntities := eoc.LoadGeneralEOC()
	concerningFunders := eoc.LoadFunderEOC()
	concerningInstitutions := eoc.LoadInstitutionEOC()

	authorFlaggers := []reports.AuthorFlagger{
		flaggers.NewAuthorIsFacultyAtEOCFlagger(
			flaggers.BuildUniversityIndex(config.UniversityData),
		),
	}
	if config.PpxApiKey != "" {
		authorFlaggers = append(authorFlaggers, flaggers.NewAuthorNewsArticlesFlagger(config.PpxApiKey))
	}

	processor := reports.NewProcessor(
		[]reports.WorkFlagger{
			flaggers.NewOpenAlexFunderIsEOC(
				concerningFunders, concerningEntities,
			),
			flaggers.NewOpenAlexAuthorAffiliationIsEOC(
				concerningEntities, concerningInstitutions,
			),
			flaggers.NewOpenAlexCoauthorAffiliationIsEOC(
				concerningEntities, concerningInstitutions,
			),
			flaggers.NewOpenAlexAcknowledgementIsEOC(
				entityStore,
				authorCache,
				flaggers.NewGrobidExtractor(
					ackCache,
					config.GrobidEndpoint,
					config.MaxGrobidThreads,
					config.MaxDownloadThreads,
					config.S3Bucket,
				),
				eoc.LoadSussyBakas(),
				triangulation.CreateTriangulationDB(cmd.OpenDB(config.FundcodeTriangulationUri)),
			),
			flaggers.NewAuthorIsAssociatedWithEOCFlagger(
				flaggers.BuildDocIndex(config.DocData),
				flaggers.BuildAuxIndex(config.AuxData),
			),
		},
		authorFlaggers,
		reportManager,
	)

	for {
		foundAuthorReport := processor.ProcessNextAuthorReport()
		foundUniversityReport := processor.ProcessNextUniversityReport()

		if !foundAuthorReport && !foundUniversityReport {
			time.Sleep(10 * time.Second)
		}
	}
}
