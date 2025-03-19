package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"prism/prism/llms"
	"strings"
	"sync"
)

const prompt = `I will give you the name of a person or organization. Please provide at most 3 alternative forms of this name. For example Joe Smith could be J. Smith, Joseph Smith, or Joe S. You are not allowed to replace first names with different names, only with initials or alternative forms of the original name
If there are no alternative forms, you must reply with just the word "none". Otherwise the output should be a comma separated list containing just the names.
Name: %s`

func generateQueries(llm llms.LLM, entity string) []string {
	res, err := llm.Generate(fmt.Sprintf(prompt, entity), nil)
	if err != nil {
		log.Printf("generation failed: %v", err)
		return nil
	}

	if strings.Contains(strings.ToLower(res), "none") {
		return nil
	}

	samples := strings.Split(res, ",")
	for i := range samples {
		samples[i] = strings.TrimSpace(samples[i])
	}
	return samples
}

func loadMultihopEntities() []string {
	type record struct {
		Entities []string `json:"entities"`
	}

	var entities []string

	var countryToArticles map[string][]record
	ParseJsonData("../../data/docs_and_press_releases.json", &countryToArticles)

	for _, articles := range countryToArticles {
		for _, article := range articles {
			entities = append(entities, article.Entities...)
		}
	}

	var articles []record
	ParseJsonData("../../data/auxiliary_webpages.json", &articles)

	for _, article := range articles {
		entities = append(entities, article.Entities...)
	}

	log.Printf("loaded %d entities from docs_and_press_releases.json and auxiliary_webpages.json", len(entities))

	return entities
}

func loadWatchlistEntities() []string {
	var aliasToSource map[string]string
	ParseJsonData("../../prism/reports/flaggers/eoc/data/alias_to_source.json", &aliasToSource)

	keys := make([]string, 0, len(aliasToSource))
	for k := range aliasToSource {
		keys = append(keys, k)
	}

	log.Printf("loaded %d entities from alias_to_source.json", len(keys))

	return keys
}

func generateData(queries []string, outputFile string) {
	llm := llms.New()

	rand.Shuffle(len(queries), func(i, j int) { queries[i], queries[j] = queries[j], queries[i] })
	queries = queries[:min(len(queries), 1000)]

	samples := make(chan Sample, len(queries))

	wg := sync.WaitGroup{}

	for i := 0; i < len(queries); i += 100 {
		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()

			for _, query := range batch {
				res := generateQueries(llm, query)
				samples <- Sample{Entity: query, Queries: res}
			}
		}(queries[i:min(i+100, len(queries))])
	}

	wg.Wait()

	close(samples)

	samplesList := make([]Sample, 0, len(queries))
	for s := range samples {
		samplesList = append(samplesList, s)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("error creating output file: %v", err)
	}
	defer file.Close()

	data, err := json.MarshalIndent(samplesList, "", "  ")
	if err != nil {
		log.Fatalf("error formatting output: %v", err)
	}

	if _, err := file.Write(data); err != nil {
		log.Fatalf("error writing output: %v", err)
	}
}

func GenerateMultiHopData() {
	generateData(loadMultihopEntities(), "./multihop_queries.json")
}

func GenerateWatchlistData() {
	generateData(loadWatchlistEntities(), "./watchlist_queries.json")
}
