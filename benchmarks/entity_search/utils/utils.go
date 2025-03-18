package utils

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

func ParseJsonData(filename string, dest any) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening '%s': %v", filepath.Base(filename), err)
	}

	if err := json.NewDecoder(file).Decode(&dest); err != nil {
		log.Fatalf("error parsing '%s': %v", filepath.Base(filename), err)
	}
}

type Sample struct {
	Entity  string
	Queries []string
}
