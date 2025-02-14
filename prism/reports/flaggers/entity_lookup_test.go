package flaggers_test

import (
	"log/slog"
	"prism/reports/flaggers"
	"testing"
)

func TestEntityLookup(t *testing.T) {
	aliasToSource := map[string]string{
		"1 2 3 4":   "source_a",
		"5 6 7":     "source_a",
		"8 9 10 11": "source_b",
		"12 13 14":  "source_b",
	}

	entityStore, err := flaggers.NewEntityStore(t.TempDir(), aliasToSource)
	if err != nil {
		t.Fatal(err)
	}
	defer entityStore.Free()

	results, err := entityStore.SearchEntities(slog.Default(), []string{"1 2 3 4", "8 9 10 11"})
	if err != nil {
		t.Fatal(err)
	}

	matches1, matches2 := results["1 2 3 4"], results["8 9 10 11"]

	if len(matches1) != 1 || len(matches1["source_a"]) != 1 || matches1["source_a"][0] != "1 2 3 4" {
		t.Fatalf("incorrect match for first query: %v", matches1)
	}

	if len(matches2) != 1 || len(matches2["source_b"]) != 1 || matches2["source_b"][0] != "8 9 10 11" {
		t.Fatalf("incorrect match for second query: %v", matches2)
	}
}
