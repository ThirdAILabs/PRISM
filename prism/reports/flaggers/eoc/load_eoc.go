package eoc

import (
	"embed"
	"encoding/json"
	"log"
)

type EocSet map[string]struct{}

func (s *EocSet) Contains(entity string) bool {
	_, exists := (*s)[entity]
	return exists
}

func convertToSet(list []string) EocSet {
	set := make(EocSet)
	for _, item := range list {
		set[item] = struct{}{}
	}
	return set
}

//go:embed data/*
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

func LoadGeneralEOC() EocSet {
	var entities []string
	parseFile("data/entities.json", &entities)

	return convertToSet(entities)
}

type eocEntity struct {
	Match struct {
		Id string `json:"id"`
	} `json:"match"`
}

func loadMatchIds(filename string) EocSet {
	var entities []eocEntity
	parseFile(filename, &entities)

	ids := make(EocSet)
	for _, entity := range entities {
		ids[entity.Match.Id] = struct{}{}
	}
	return ids
}

func LoadFunderEOC() EocSet {
	return loadMatchIds("data/funders.json")
}

func LoadInstitutionEOC() EocSet {
	return loadMatchIds("data/institutions.json")
}

func LoadPublisherEOC() EocSet {
	return loadMatchIds("data/publishers.json")
}

func LoadSussyBakas() []string {
	var entities []string
	parseFile("data/sussy_bakas.json", &entities)

	return entities
}

func LoadSourceToAlias() map[string]string {
	var entities map[string]string
	parseFile("data/alias_to_source.json", &entities)

	return entities
}
