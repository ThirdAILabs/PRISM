package train_ndb

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"prism/prism/reports/flaggers"
	"prism/prism/search"
)

func RetrainWorkerNDBs(ndbDir, universityData, docData, auxData string) (search.NeuralDB, search.NeuralDB, search.NeuralDB) {

	var universityNDB search.NeuralDB
	var docNDB search.NeuralDB
	var auxNDB search.NeuralDB

	if err := os.RemoveAll(ndbDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("error deleting existing ndb dir '%s': %v", ndbDir, err)
	}

	if err := os.MkdirAll(ndbDir, 0777); err != nil {
		log.Fatalf("error creating work dir: %v", err)
	}

	universityNDB = flaggers.BuildUniversityNDB(universityData, filepath.Join(ndbDir, "university.ndb"))
	docNDB = flaggers.BuildDocNDB(docData, filepath.Join(ndbDir, "doc.ndb"))
	auxNDB = flaggers.BuildAuxNDB(auxData, filepath.Join(ndbDir, "aux.ndb"))

	return universityNDB, docNDB, auxNDB
}
