package flaggers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"prism/prism/search"
	"time"
)

const (
	insertionBatchSize = 10000
)

func parseJsonData(filename string, dest any) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening '%s': %v", filepath.Base(filename), err)
	}

	if err := json.NewDecoder(file).Decode(&dest); err != nil {
		log.Fatalf("error parsing '%s': %v", filepath.Base(filename), err)
	}
}

type universityDataRecord struct {
	Entity  string `json:"entity"`
	Url     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func BuildUniversityNDB(dataPath string, ndbPath string) search.NeuralDB {
	log.Printf("creating university ndb %s from data %s", ndbPath, dataPath)

	var records []universityDataRecord
	parseJsonData(dataPath, &records)

	log.Printf("loaded %d records", len(records))

	ndb, err := search.NewNeuralDB(ndbPath)
	if err != nil {
		log.Fatalf("error creating ndb: %v", err)
	}

	s := time.Now()

	for start := 0; start < len(records); start += insertionBatchSize {
		end := min(start+insertionBatchSize, len(records))
		recordsBatch := records[start:end]

		chunks := make([]string, 0, len(recordsBatch))
		metadata := make([]map[string]any, 0, len(recordsBatch))
		for _, record := range recordsBatch {
			chunks = append(chunks, record.Content)
			metadata = append(metadata, map[string]any{"university": record.Entity, "url": record.Url})
		}

		if err := ndb.Insert("university_data", "0", chunks, metadata, nil); err != nil {
			log.Fatalf("error inserting into ndb: %v", err)
		}

		log.Printf("processed %d/%d %.2f%% complete", end, len(records), 100*float64(end)/float64(len(records)))
	}

	e := time.Now()

	log.Printf("ndb created successfully time %.3f s", e.Sub(s).Seconds())

	return ndb
}

func BuildDocIndex(dataPath string) *search.ManyToOneIndex[DojArticleRecord] {
	log.Printf("creating doc index from data %s", dataPath)

	var countryToArticles map[string][]DojArticleRecord
	parseJsonData(dataPath, &countryToArticles)

	data := make([]DojArticleRecord, 0)
	for _, articles := range countryToArticles {
		data = append(data, articles...)
	}

	log.Printf("loaded %d records", len(data))

	entities := make([][]string, 0, len(data))
	metadata := make([]DojArticleRecord, 0, len(data))
	for _, record := range data {
		entities = append(entities, record.getEntitiesForIndexing())
		metadata = append(metadata, DojArticleRecord{
			Title:        record.Title,
			Url:          record.Url,
			Text:         record.Text,
			Individuals:  record.Individuals,
			Institutions: record.Institutions,
		})
	}

	s := time.Now()

	index := search.NewManyToOneIndex(entities, metadata)

	e := time.Now()

	log.Printf("index created successfully time %.3f s", e.Sub(s).Seconds())

	return index
}

func BuildAuxIndex(dataPath string) *search.ManyToOneIndex[ReleveantWebpageRecord] {
	log.Printf("creating aux index from data %s", dataPath)

	var data []ReleveantWebpageRecord
	parseJsonData(dataPath, &data)

	log.Printf("loaded %d records", len(data))

	entities := make([][]string, 0, len(data))
	metadata := make([]ReleveantWebpageRecord, 0, len(data))
	for _, record := range data {
		entities = append(entities, record.getEntitiesForIndexing())
		metadata = append(metadata, ReleveantWebpageRecord{
			Title:        record.Title,
			Url:          record.Url,
			Individuals:  record.Individuals,
			Institutions: record.Institutions,
			Text:         record.Text,
			ReferredFrom: record.ReferredFrom,
		})
	}

	s := time.Now()

	index := search.NewManyToOneIndex(entities, metadata)

	e := time.Now()

	log.Printf("index created successfully time %.3f s", e.Sub(s).Seconds())

	return index
}
