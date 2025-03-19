package search_test

import (
	"prism/prism/search"
	"testing"
)

type sample struct {
	Entity  string
	Queries []string
}

func TestEntitySearch(t *testing.T) {
	samples := []sample{
		{Entity: "United States Department of Justice", Queries: []string{"U.S. Department of Justice", "US Department of Justice", "Justice Department"}},
		{Entity: "Maxim Integrated", Queries: []string{"Maxim Int.", "Maxim Inc.", "Maxim Integrated Products"}},
		{Entity: "Group1 Automotive Inc", Queries: []string{"Group 1 Automotive", "Group One Automotive", "Group1 Automotive"}},
		{Entity: "Nicole Boeckmann", Queries: []string{"N. Boeckmann", "Nicole B.", "N. Boeckmann"}},
		{Entity: "Ford Motor Company", Queries: []string{"Ford", "Ford Motors", "Ford Motor Co."}},
		{Entity: "Huawei", Queries: []string{"Huawei Technologies Co.", "Ltd.", "Huawei Technologies", "Huawei Co."}},
		{Entity: "Wanzhou Meng", Queries: []string{"Meng W.", "Meng Wanzhou", "M. Wanzhou"}},
		{Entity: "Mark J. Lesko", Queries: []string{"M. J. Lesko", "Mark Lesko", "M. Lesko"}},
		{Entity: "Office of Public Affairs", Queries: []string{"Office of Public Affairs", "OPA", "Public Affairs Office"}},
	}

	records := make([]search.Record[int], 0, len(samples))
	for i, s := range samples {
		records = append(records, search.Record[int]{Entity: s.Entity, Metadata: i})
	}

	index := search.NewIndex(records)

	correct, total := 0, 0
	for _, sample := range samples {
		for _, query := range sample.Queries {
			results := index.Query(query, 10)

			if len(results) == 0 {
				t.Errorf("no results found for query %s", query)
			}

			if samples[results[0].Metadata].Entity == sample.Entity {
				correct++
			}
			total++
		}
	}

	if float64(correct)/float64(total) < 0.95 {
		t.Errorf("accuracy below 95%%: %.3f", float64(correct)/float64(total))
	}
}

func TestManyToOneIndex(t *testing.T) {
	docA := []string{"abc", "xyz", "qrs"}
	docB := []string{"123", "456", "78", "90"}

	index := search.NewManyToOneIndex([][]string{docA, docB}, []string{"doc_a", "doc_b"})

	for _, query := range docA {
		results := index.Query(query, 10)

		if len(results) == 0 {
			t.Errorf("no results found for query %s", query)
		}

		if results[0].Metadata != "doc_a" {
			t.Errorf("incorrect metadata for query %s", query)
		}
	}

	for _, query := range docB {
		results := index.Query(query, 10)

		if len(results) == 0 {
			t.Errorf("no results found for query %s", query)
		}

		if results[0].Metadata != "doc_b" {
			t.Errorf("incorrect metadata for query %s", query)
		}
	}
}
