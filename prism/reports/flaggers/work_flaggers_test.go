package flaggers

import (
	"log/slog"
	"path/filepath"
	"prism/prism/openalex"
	"prism/prism/reports/flaggers/eoc"
	"testing"
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
				RawText: "special thanks to bad entity xyz",
				SearchableEntities: []Entity{
					{"bad entity xyz", "", 0},
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

	flagger := OpenAlexAcknowledgementIsEOC{
		openalex:     openalex.NewRemoteKnowledgeBase(),
		entityLookup: entityStore,
		authorCache:  authorCache,
		extractor:    &mockAcknowledgmentExtractor{},
		sussyBakas:   []string{"bad entity xyz"},
	}

	flags, err := flagger.Flag(slog.Default(), []openalex.Work{{WorkId: "a/b", DownloadUrl: "n/a"}}, []string{})
	if err != nil {
		t.Fatal(err)
	}

	if len(flags) != 1 {
		t.Fatal("expected 1 flag")
	}
}
