package flaggers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"prism/prism/search"
	"time"
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

func BuildUniversityIndex(dataPath string) *search.EntityIndex[UniversityInfo] {
	log.Printf("creating university index from data %s", dataPath)

	var records []universityDataRecord
	parseJsonData(dataPath, &records)

	log.Printf("loaded %d records", len(records))

	s := time.Now()

	entities := make([]search.Record[UniversityInfo], 0, len(records))
	for _, record := range records {
		entities = append(entities, search.Record[UniversityInfo]{
			Entity: record.Content,
			Metadata: UniversityInfo{
				University: record.Entity,
				Url:        record.Url,
			},
		})
	}

	index := search.NewIndex(entities)

	e := time.Now()

	log.Printf("index created successfully time %.3f s", e.Sub(s).Seconds())

	return index
}

type dojArticleRecord struct {
	Title    string   `json:"title"`
	Url      string   `json:"link"`
	Text     string   `json:"entities_as_text"`
	Entities []string `json:"entities"`
}

func BuildDocIndex(dataPath string) *search.ManyToOneIndex[LinkMetadata] {
	log.Printf("creating doc index from data %s", dataPath)

	var countryToArticles map[string][]dojArticleRecord
	parseJsonData(dataPath, &countryToArticles)

	data := make([]dojArticleRecord, 0)
	for _, articles := range countryToArticles {
		data = append(data, articles...)
	}

	log.Printf("loaded %d records", len(data))

	entities := make([][]string, 0, len(data))
	metadata := make([]LinkMetadata, 0, len(data))
	for _, record := range data {
		entities = append(entities, record.Entities)
		metadata = append(metadata, LinkMetadata{
			Title:    record.Title,
			Url:      record.Url,
			Entities: record.Entities,
			Text:     record.Text,
		})
	}

	s := time.Now()

	index := search.NewManyToOneIndex(entities, metadata)

	e := time.Now()

	log.Printf("index created successfully time %.3f s", e.Sub(s).Seconds())

	return index
}

type releveantWebpageRecord struct {
	Title    string   `json:"title"`
	Url      string   `json:"url"`
	DojTitle string   `json:"doj_title"`
	DojUrl   string   `json:"doj_url"`
	Content  string   `json:"content"`
	Entities []string `json:"entities"`
}

func BuildAuxIndex(dataPath string) *search.ManyToOneIndex[LinkMetadata] {
	log.Printf("creating aux index from data %s", dataPath)

	var data []releveantWebpageRecord
	parseJsonData(dataPath, &data)

	log.Printf("loaded %d records", len(data))

	entities := make([][]string, 0, len(data))
	metadata := make([]LinkMetadata, 0, len(data))
	for _, record := range data {
		entities = append(entities, record.Entities)
		metadata = append(metadata, LinkMetadata{
			Title:    record.Title,
			Url:      record.Url,
			Entities: record.Entities,
			Text:     record.Content,
		})
	}

	s := time.Now()

	index := search.NewManyToOneIndex(entities, metadata)

	e := time.Now()

	log.Printf("index created successfully time %.3f s", e.Sub(s).Seconds())

	return index
}
