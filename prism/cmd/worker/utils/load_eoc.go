package utils

import (
	"embed"
	"encoding/json"
	"log"
)

//go:embed eoc/*
var eoc embed.FS

func parseFile(filename string, dest any) {
	contents, err := eoc.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading '%s': %v", filename, err)
	}

	if err := json.Unmarshal(contents, dest); err != nil {
		log.Fatalf("error parsing '%s': %v", filename, err)
	}
}

func LoadGeneralEOC() []string {
	var entities []string
	parseFile("eoc/entities.json", &entities)

	return entities
}

type eocEntity struct {
	Match struct {
		Id string `json:"id"`
	} `json:"match"`
}

func loadMatchIds(filename string) []string {
	var entities []eocEntity
	parseFile(filename, &entities)

	ids := make([]string, 0, len(entities))
	for _, entity := range entities {
		ids = append(ids, entity.Match.Id)
	}
	return ids
}

func LoadFunderEOC() []string {
	return loadMatchIds("eoc/funders.json")
}

func LoadInstitutionEOC() []string {
	return loadMatchIds("eoc/institutions.json")
}

func LoadPublisherEOC() []string {
	return loadMatchIds("eoc/publishers.json")
}

func LoadSussyBakas() []string {
	var entities []string
	parseFile("eoc/sussy_bakas.json", &entities)

	return entities
}

func LoadSourceToAlias() map[string]string {
	var entities map[string]string
	parseFile("eoc/alias_to_source.json", &entities)

	return entities
}
