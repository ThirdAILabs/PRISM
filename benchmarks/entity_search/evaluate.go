package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"prism/benchmarks/entity_search/utils"
	"prism/prism/search"
	"strings"
)

func evaluateEntitySearch(queries []utils.Sample) {
	entities := make([]search.Record[struct{}], 0, len(queries))
	for _, query := range queries {
		entities = append(entities, search.Record[struct{}]{Entity: query.Entity})
	}

	index := search.NewIndex(entities)

	p_at_1, p_at_10, total := 0, 0, 0
	for _, sample := range queries {
		if len(sample.Queries) == 0 || (len(sample.Queries) == 1 && strings.ToLower(strings.TrimSpace(sample.Queries[0])) == "none") {
			continue
		}
		for _, query := range sample.Queries {
			results := index.Query(query, 10)

			if len(results) > 0 && results[0].Entity == sample.Entity {
				p_at_1++
			}

			for _, res := range results {
				if res.Entity == sample.Entity {
					p_at_10++
					break
				}
			}

			total++
		}
	}

	fmt.Printf("p@1 = %.3f p@10 = %.3f\n", float64(p_at_1)/float64(total), float64(p_at_10)/float64(total))
}

func evaluateNDB(queries []utils.Sample) {
	if err := search.SetLicensePath("../../.test_license/thirdai.license"); err != nil {
		log.Fatalf("error setting license key: %v", err)
	}

	entities := make([]string, 0, len(queries))
	for _, query := range queries {
		entities = append(entities, query.Entity)
	}

	ndb, err := search.NewNeuralDB("./test.ndb")
	if err != nil {
		panic(err)
	}

	defer ndb.Free()
	defer func() {
		os.RemoveAll("./test.ndb")
	}()

	if err := ndb.Insert("doc", "id", entities, nil, nil); err != nil {
		panic(err)
	}

	p_at_1, p_at_10, total := 0, 0, 0
	for _, sample := range queries {
		if len(sample.Queries) == 0 || (len(sample.Queries) == 1 && strings.ToLower(strings.TrimSpace(sample.Queries[0])) == "none") {
			continue
		}
		for _, query := range sample.Queries {
			results, err := ndb.Query(query, 10, nil)
			if err != nil {
				panic(err)
			}

			if len(results) > 0 && results[0].Text == sample.Entity {
				p_at_1++
			}

			for _, res := range results {
				if res.Text == sample.Entity {
					p_at_10++
					break
				}
			}

			total++
		}
	}

	fmt.Printf("p@1 = %.3f p@10 = %.3f\n", float64(p_at_1)/float64(total), float64(p_at_10)/float64(total))
}

func createCsv(aliases []string) (string, error) {
	filename := "entities.csv"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return "", fmt.Errorf("error creating csv file for training: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	rows := make([][]string, 0, len(aliases)+1)
	rows = append(rows, []string{"entity"})
	for _, alias := range aliases {
		rows = append(rows, []string{alias})
	}

	if err := writer.WriteAll(rows); err != nil {
		return "", fmt.Errorf("error writing rows: %w", err)
	}

	return filename, nil
}

func evaluateFlash(queries []utils.Sample) {
	if err := search.SetLicensePath("../../.test_license/thirdai.license"); err != nil {
		log.Fatalf("error setting license key: %v", err)
	}

	entities := make([]string, 0, len(queries))
	for _, query := range queries {
		entities = append(entities, query.Entity)
	}

	csv, err := createCsv(entities)
	if err != nil {
		panic(err)
	}

	defer func() {
		os.Remove(csv)
	}()

	flash, err := search.NewFlash()
	if err != nil {
		panic(err)
	}
	defer flash.Free()

	if err := flash.Train(csv); err != nil {
		panic(err)
	}

	p_at_1, p_at_10, total := 0, 0, 0
	for _, sample := range queries {
		if len(sample.Queries) == 0 || (len(sample.Queries) == 1 && strings.ToLower(strings.TrimSpace(sample.Queries[0])) == "none") {
			continue
		}
		for _, query := range sample.Queries {
			results, err := flash.Predict(query, 10)
			if err != nil {
				panic(err)
			}

			if len(results) > 0 && results[0] == sample.Entity {
				p_at_1++
			}

			for _, res := range results {
				if res == sample.Entity {
					p_at_10++
					break
				}
			}

			total++
		}
	}

	fmt.Printf("p@1 = %.3f p@10 = %.3f\n", float64(p_at_1)/float64(total), float64(p_at_10)/float64(total))
}

func main() {
	generateData := flag.Bool("generate-data", false, "generate data for evaluation")
	flag.Parse()

	if *generateData {
		utils.GenerateMultiHopData()
		utils.GenerateWatchlistData()
		return
	}

	for _, dataset := range []string{"./multihop_queries.json", "./watchlist_queries.json"} {
		fmt.Println("Evaluating on dataset:", dataset)
		var samples []utils.Sample
		utils.ParseJsonData(dataset, &samples)

		fmt.Println("new entity lookup")
		evaluateEntitySearch(samples)

		fmt.Println("NDB")
		evaluateNDB(samples)

		fmt.Println("Flash")
		evaluateFlash(samples)

		fmt.Println("=====================================")
	}
}
