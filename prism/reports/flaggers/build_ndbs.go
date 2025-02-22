package flaggers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"prism/prism/search"
	"strings"
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

type dojArticleRecord struct {
	Title    string   `json:"title"`
	Url      string   `json:"link"`
	Text     string   `json:"entities_as_text"`
	Entities []string `json:"entities"`
}

func BuildDocNDB(dataPath string, ndbPath string) search.NeuralDB {
	log.Printf("creating doc ndb %s from data %s", ndbPath, dataPath)

	var countryToArticles map[string][]dojArticleRecord
	parseJsonData(dataPath, &countryToArticles)

	records := make([]dojArticleRecord, 0)
	for _, articles := range countryToArticles {
		records = append(records, articles...)
	}

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
			chunks = append(chunks, record.Text)
			metadata = append(metadata, map[string]any{"title": record.Title, "url": record.Url, "entities": strings.Join(record.Entities, ";")})
		}

		if err := ndb.Insert("press_release_docs", "0", chunks, metadata, nil); err != nil {
			log.Fatalf("error inserting into ndb: %v", err)
		}

		log.Printf("processed %d/%d %.2f%% complete", end, len(records), 100*float64(end)/float64(len(records)))
	}

	e := time.Now()

	log.Printf("ndb created successfully time %.3f s", e.Sub(s).Seconds())

	return ndb
}

type releveantWebpageRecord struct {
	Title    string   `json:"title"`
	Url      string   `json:"url"`
	DojTitle string   `json:"doj_title"`
	DojUrl   string   `json:"doj_url"`
	Content  string   `json:"content"`
	Entities []string `json:"entities"`
}

func BuildAuxNDB(dataPath string, ndbPath string) search.NeuralDB {
	log.Printf("creating aux ndb %s from data %s", ndbPath, dataPath)

	var records []releveantWebpageRecord
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
			metadata = append(metadata, map[string]any{
				"title":     record.Title,
				"url":       record.Url,
				"entities":  strings.Join(record.Entities, ";"),
				"doj_title": record.DojTitle,
				"doj_url":   record.DojUrl,
			})
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
