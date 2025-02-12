package flaggers_test

import (
	"log/slog"
	"prism/reports/flaggers"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEntityLookup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&flaggers.AliasRecord{}, &flaggers.EntityRecord{}, &flaggers.SourceRecord{}); err != nil {
		t.Fatal(err)
	}

	source1, source2 := uuid.New(), uuid.New()
	if err := db.Create([]flaggers.SourceRecord{
		{Id: source1, Name: "source_a", Link: "a.com"},
		{Id: source2, Name: "source_b", Link: "b.com"},
	}).Error; err != nil {
		t.Fatal(err)
	}

	entity1, entity2 := uuid.New(), uuid.New()
	if err := db.Create([]flaggers.EntityRecord{
		{Id: entity1, Name: "entity_a", SourceId: source1},
		{Id: entity2, Name: "entity_b", SourceId: source2},
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.Create([]flaggers.AliasRecord{
		{Id: uuid.New(), Alias: "1 2 3 4", EntityId: entity1},
		{Id: uuid.New(), Alias: "5 6 7", EntityId: entity1},
		{Id: uuid.New(), Alias: "8 9 10 11", EntityId: entity2},
		{Id: uuid.New(), Alias: "12 13 14", EntityId: entity2},
	}).Error; err != nil {
		t.Fatal(err)
	}

	entityStore, err := flaggers.NewEntityStore(t.TempDir(), db)
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
