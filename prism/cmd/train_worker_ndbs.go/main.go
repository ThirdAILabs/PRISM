package main

import (
	"log"
	"path/filepath"
	"prism/prism/cmd"
	"prism/prism/train_ndb"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	WorkDir string `env:"WORK_DIR,notEmpty" envDefault:"./work"`

	UniversityData string `env:"UNIVERSITY_DATA,notEmpty,required"`
	DocData        string `env:"DOC_DATA,notEmpty,required"`
	AuxData        string `env:"AUX_DATA,notEmpty,required"`
}

func main() {
	cmd.LoadEnvFile()

	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	ndbDir := filepath.Join(config.WorkDir, "ndbs")
	train_ndb.RetrainWorkerNDBs(ndbDir, config.UniversityData, config.DocData, config.AuxData)

}
