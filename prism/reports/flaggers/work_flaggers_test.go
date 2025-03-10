package flaggers

import (
	"log/slog"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/openalex"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/triangulation"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMultipleAssociations(t *testing.T) {
	flagger := OpenAlexMultipleAffiliationsFlagger{}

	works := []openalex.Work{
		{Authors: []openalex.Author{
			{AuthorId: "1", Institutions: []openalex.Institution{{}, {}}},
		}},
		{Authors: []openalex.Author{
			{AuthorId: "2", Institutions: []openalex.Institution{{}, {}}},
			{AuthorId: "4", Institutions: []openalex.Institution{{}, {}}},
		}},
	}

	flags, err := flagger.Flag(slog.Default(), works, []string{"2", "3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(flags) != 1 {
		t.Fatal("expected 1 flag")
	}

	noflags, err := flagger.Flag(slog.Default(), works, []string{"5", "6"})
	if err != nil {
		t.Fatal(err)
	}
	if len(noflags) != 0 {
		t.Fatal("expected 0 flags")
	}
}

func makeSet(entities ...string) eoc.EocSet {
	set := make(eoc.EocSet)
	for _, e := range entities {
		set[e] = struct{}{}
	}
	return set
}

func TestFunderEOC(t *testing.T) {
	flagger := OpenAlexFunderIsEOC{
		concerningFunders:  makeSet("bad-abc", "bad-xyz"),
		concerningEntities: makeSet("bad-123", "bad-456"),
	}

	for funder, nflags := range map[string]int{"bad-xyz": 1, "bad-456": 1, "abc": 0, "123": 0} {
		works := []openalex.Work{
			{Grants: []openalex.Grant{
				{FunderId: "some funder"},
				{FunderId: funder},
			}},
		}

		flags, err := flagger.Flag(slog.Default(), works, []string{"a", "b"})
		if err != nil {
			t.Fatal(err)
		}

		if len(flags) != nflags {
			t.Fatal("incorrect number of flags")
		}
	}
}

func TestPublisherEOC(t *testing.T) {
	flagger := OpenAlexPublisherIsEOC{
		concerningPublishers: makeSet("bad-abc", "bad-xyz"),
	}

	for publisher, nflags := range map[string]int{"abc": 0, "bad-xyz": 1} {
		works := []openalex.Work{
			{Locations: []openalex.Location{
				{OrganizationId: "some publisher"},
				{OrganizationId: publisher},
			}},
		}

		flags, err := flagger.Flag(slog.Default(), works, []string{"a", "b"})
		if err != nil {
			t.Fatal(err)
		}
		if len(flags) != nflags {
			t.Fatal("incorrect number of flags")
		}
	}
}

func TestCoauthorEOC(t *testing.T) {
	flagger := OpenAlexCoauthorIsEOC{
		concerningEntities: makeSet("bad-abc", "bad-xyz"),
	}

	for coauthor, nflags := range map[string]int{"abc": 0, "bad-xyz": 1} {
		works := []openalex.Work{
			{Authors: []openalex.Author{
				{AuthorId: "some author"},
				{AuthorId: coauthor},
			}},
		}

		flags, err := flagger.Flag(slog.Default(), works, []string{"a", "b"})
		if err != nil {
			t.Fatal(err)
		}
		if len(flags) != nflags {
			t.Fatal("incorrect number of flags")
		}
	}
}

func TestAuthorAffiliationEOC(t *testing.T) {
	flagger := OpenAlexAuthorAffiliationIsEOC{
		concerningEntities:     makeSet("bad-abc", "bad-xyz"),
		concerningInstitutions: makeSet("bad-123", "bad-456"),
	}

	for author, isTarget := range map[string]int{"a": 1, "c": 0} {
		for institution, isBad := range map[string]int{"abc": 0, "bad-xyz": 1, "bad-123": 1} {
			works := []openalex.Work{
				{Authors: []openalex.Author{
					{AuthorId: "some author", Institutions: []openalex.Institution{{InstitutionId: "university"}}},
					{AuthorId: author, Institutions: []openalex.Institution{{InstitutionId: institution}}},
				}},
			}

			flags, err := flagger.Flag(slog.Default(), works, []string{"a", "b"})
			if err != nil {
				t.Fatal(err)
			}
			if len(flags) != isBad*isTarget {
				t.Fatal("incorrect number of flags")
			}
		}
	}
}

func TestCoauthorAffiliationEOC(t *testing.T) {
	flagger := OpenAlexCoauthorAffiliationIsEOC{
		concerningEntities:     makeSet("bad-abc", "bad-xyz"),
		concerningInstitutions: makeSet("bad-123", "bad-456"),
	}

	for author, isTarget := range map[string]int{"a": 1, "c": 0} {
		for institution, isBad := range map[string]int{"abc": 0, "bad-xyz": 1, "bad-123": 1} {
			works := []openalex.Work{
				{Authors: []openalex.Author{
					{AuthorId: "some author", Institutions: []openalex.Institution{{InstitutionId: "university"}}},
					{AuthorId: author, Institutions: []openalex.Institution{{InstitutionId: institution}}},
				}},
			}

			flags, err := flagger.Flag(slog.Default(), works, []string{"a", "b"})
			if err != nil {
				t.Fatal(err)
			}
			if len(flags) != isBad*(1-isTarget) {
				t.Fatal("incorrect number of flags")
			}
		}
	}
}

type mockAcknowledgmentExtractor struct{}

func (m *mockAcknowledgmentExtractor) GetAcknowledgements(logger *slog.Logger, works []openalex.Work) chan CompletedTask[Acknowledgements] {
	output := make(chan CompletedTask[Acknowledgements], 1)

	output <- CompletedTask[Acknowledgements]{
		Result: Acknowledgements{
			WorkId: works[0].WorkId,
			Acknowledgements: []Acknowledgement{{
				RawText: "special thanks to bad entity xyz for grants ABC-123456 and XYZ-9876",
				SearchableEntities: []Entity{
					{EntityText: "bad entity xyz", EntityType: "", StartPosition: 0, FundCodes: []string{"ABC-123456", "XYZ-9876"}},
				},
			}},
		},
	}

	close(output)

	return output
}

func TestAcknowledgementEOC(t *testing.T) {
	testDir := t.TempDir()

	authorCache, err := NewCache[openalex.Author]("authors", filepath.Join(testDir, "author.cache"))
	if err != nil {
		t.Fatal(err)
	}
	defer authorCache.Close()

	aliasToSource := map[string]string{
		"bad entity xzy": "source_a", "a worse entity": "source_b",
	}

	entityStore, err := NewEntityStore(t.TempDir(), aliasToSource)
	if err != nil {
		t.Fatal(err)
	}
	defer entityStore.Free()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&triangulation.Author{},
		&triangulation.FundCode{},
	); err != nil {
		t.Fatal(err)
	}

	triangulationDB := triangulation.CreateTriangulationDB(db)

	flagger := OpenAlexAcknowledgementIsEOC{
		openalex:        openalex.NewRemoteKnowledgeBase(),
		entityLookup:    entityStore,
		authorCache:     authorCache,
		extractor:       &mockAcknowledgmentExtractor{},
		sussyBakas:      []string{"bad entity xyz"},
		triangulationDB: triangulationDB,
	}

	flags, err := flagger.Flag(slog.Default(), []openalex.Work{{WorkId: "a/b", DownloadUrl: "n/a"}}, []string{})
	if err != nil {
		t.Fatal(err)
	}

	if len(flags) != 1 {
		t.Fatal("expected 1 flag")
	}
}

func TestFundCodeTriangulation(t *testing.T) {
	testDir := t.TempDir()

	authorCache, err := NewCache[openalex.Author]("authors", filepath.Join(testDir, "author.cache"))
	if err != nil {
		t.Fatal(err)
	}
	authorCache.Update("1", openalex.Author{AuthorId: "1", DisplayName: "Jane Smith"})
	defer authorCache.Close()

	aliasToSource := map[string]string{
		"bad entity xyz": "source_a",
	}

	entityStore, err := NewEntityStore(t.TempDir(), aliasToSource)
	if err != nil {
		t.Fatal(err)
	}
	defer entityStore.Free()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&triangulation.Author{},
		&triangulation.FundCode{},
	); err != nil {
		t.Fatal(err)
	}

	fundCodes := []triangulation.FundCode{
		{
			ID:        1,
			FundCode:  "ABC-123456",
			NumPapers: 100,
		},
		{
			ID:        2,
			FundCode:  "XYZ-9876",
			NumPapers: 100,
		},
	}

	for _, fundCode := range fundCodes {
		if err := db.Create(&fundCode).Error; err != nil {
			t.Fatal(err)
		}
	}

	authors := []triangulation.Author{
		{
			ID:                1,
			FundCodeID:        &fundCodes[0].ID,
			AuthorName:        "Jane Smith",
			NumPapersByAuthor: 52,
		},
		{
			ID:                2,
			FundCodeID:        &fundCodes[1].ID,
			AuthorName:        "Jane Smith",
			NumPapersByAuthor: 5,
		},
	}

	for _, author := range authors {
		if err := db.Create(&author).Error; err != nil {
			t.Fatal(err)
		}
	}

	triangulationDB := triangulation.CreateTriangulationDB(db)

	flagger := OpenAlexAcknowledgementIsEOC{
		openalex:        openalex.NewRemoteKnowledgeBase(),
		entityLookup:    entityStore,
		authorCache:     authorCache,
		extractor:       &mockAcknowledgmentExtractor{},
		sussyBakas:      []string{"bad entity xyz"},
		triangulationDB: triangulationDB,
	}

	works := []openalex.Work{
		{WorkId: "a/b", DownloadUrl: "n/a", Authors: []openalex.Author{
			{AuthorId: "1", Institutions: []openalex.Institution{{}, {}}, DisplayName: "Jane Smith"},
		}},
	}

	flags, err := flagger.Flag(slog.Default(), works, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}

	if len(flags) != 1 {
		t.Fatal("expected 1 flag")
	}

	if flag, ok := flags[0].(*api.HighRiskFunderFlag); ok {
		if flag.FundCodeTriangulation["bad entity xyz"][fundCodes[0].FundCode] != true || flag.FundCodeTriangulation["bad entity xyz"][fundCodes[1].FundCode] != false {
			t.Fatal("incorrect fund code triangulation")
		}
	} else {
		t.Fatal("expected HighRiskFunderFlag")
	}
}
