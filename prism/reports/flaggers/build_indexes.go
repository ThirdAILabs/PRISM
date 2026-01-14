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
		entities = append(entities, record.getEntities())
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
		entities = append(entities, record.getEntities())
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
