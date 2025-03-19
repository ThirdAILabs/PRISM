package flaggers

import (
	"log/slog"
	"prism/prism/api"
	"prism/prism/llms"
	"prism/prism/openalex"
	"prism/prism/search"
	"slices"
	"testing"
)

func init() {
	const licensePath = "../../../.test_license/thirdai.license"
	if err := search.SetLicensePath(licensePath); err != nil {
		panic(err)
	}
}

func TestAuthorIsFacultyAtEOC(t *testing.T) {
	ndb, err := search.NewNeuralDB(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Free()

	if err := ndb.Insert(
		"doc", "id",
		[]string{"prof 123 456", "7 8 9", "Dr. 10 11"},
		[]map[string]any{{"university": "abc", "url": "abc.com"}, {"university": "xyz", "url": "xyz.com"}, {"university": "qrs", "url": "qrs.com"}},
		nil,
	); err != nil {
		t.Fatal(err)
	}

	flagger := AuthorIsFacultyAtEOCFlagger{universityNDB: ndb}

	flags, err := flagger.Flag(slog.Default(), "7 9")
	if err != nil {
		t.Fatal(err)
	}

	if len(flags) != 1 {
		t.Fatal("expected flag")
	}

	flag := flags[0].(*api.PotentialAuthorAffiliationFlag)
	if flag.University != "xyz" || flag.UniversityUrl != "xyz.com" {
		t.Fatal("incorrect flag")
	}

	noflags, err := flagger.Flag(slog.Default(), "some random name")
	if err != nil {
		t.Fatal(err)
	}

	if len(noflags) != 0 {
		t.Fatal("should be no flags")
	}
}

var (
	mockPressReleaseEntities = [][]string{
		{"abc", "xyz"},
		{"qrs"},
	}

	mockPressReleaseMetadata = []LinkMetadata{
		{Title: "indicted", Url: "indicted.com", Entities: []string{"abc", "xyz"}, Text: "abc and xyz are indicted"},
		{Title: "leaked docs", Url: "leakeddocs.com", Entities: []string{"qrs"}, Text: "qrs is implicated by leaked docs"},
	}

	mockAuxDocEntities = [][]string{
		{"xyz", "123"},
		{"456", "qrs"},
		{"456", "789"},
	}

	mockAuxDocMetadata = []LinkMetadata{
		{Title: "new company", Url: "newcompany.com", Entities: []string{"xyz", "123"}, Text: "xyz and 123 found company"},
		{Title: "graduate students", Url: "graduatestudents.com", Entities: []string{"456", "qrs"}, Text: "456 and qrs are graduate students together"},
		{Title: "best friends", Url: "bestfriends.com", Entities: []string{"456", "789"}, Text: "456 and 789 are best friends"},
	}
)

func TestAuthorAssociationIsEOC(t *testing.T) {
	flagger := AuthorIsAssociatedWithEOCFlagger{docIndex: search.NewManyToOneIndex(mockPressReleaseEntities, mockPressReleaseMetadata), auxIndex: search.NewManyToOneIndex(mockAuxDocEntities, mockAuxDocMetadata)}

	t.Run("test primary connection", func(t *testing.T) {
		works := []openalex.Work{ // Only the author names are used in this flagger
			{Authors: []openalex.Author{{DisplayName: "abc"}, {DisplayName: "def"}}},
		}

		flags, err := flagger.Flag(slog.Default(), works, nil, "abc")
		if err != nil {
			t.Fatal(err)
		}

		if len(flags) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := flags[0].(*api.MiscHighRiskAssociationFlag)

		if flag.DocTitle != "indicted" ||
			flag.DocUrl != "indicted.com" ||
			!slices.Equal(flag.DocEntities, []string{"abc", "xyz"}) ||
			len(flag.Connections) != 0 ||
			flag.EntityMentioned != "abc" {
			t.Fatalf("incorrect flag: %v", *flag)
		}
	})

	t.Run("test frequent coauthor connection", func(t *testing.T) {
		works := []openalex.Work{ // Only the author names are used in this flagger
			{Authors: []openalex.Author{{DisplayName: "abc"}, {DisplayName: "def"}}},
		}

		flags, err := flagger.Flag(slog.Default(), works, nil, "def")
		if err != nil {
			t.Fatal(err)
		}

		if len(flags) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := flags[0].(*api.MiscHighRiskAssociationFlag)

		if flag.DocTitle != "indicted" ||
			flag.DocUrl != "indicted.com" ||
			!slices.Equal(flag.DocEntities, []string{"abc", "xyz"}) ||
			flag.EntityMentioned != "abc" ||
			flag.FrequentCoauthor == nil ||
			*flag.FrequentCoauthor != "abc" ||
			len(flag.Connections) != 1 ||
			flag.Connections[0].DocTitle != "abc (frequent coauthor)" {
			t.Fatalf("incorrect flag: %v", *flag)
		}
	})

	t.Run("test secondary connection", func(t *testing.T) {
		flags, err := flagger.Flag(slog.Default(), []openalex.Work{}, nil, "123")
		if err != nil {
			t.Fatal(err)
		}

		if len(flags) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := flags[0].(*api.MiscHighRiskAssociationFlag)

		if len(flag.Connections) != 1 ||
			flag.DocTitle != "indicted" ||
			flag.DocUrl != "indicted.com" ||
			!slices.Equal(flag.DocEntities, []string{"abc", "xyz"}) ||
			flag.EntityMentioned != "xyz" ||
			flag.FrequentCoauthor != nil ||
			len(flag.Connections) != 1 ||
			flag.Connections[0].DocTitle != "new company" ||
			flag.Connections[0].DocUrl != "newcompany.com" {
			t.Fatalf("incorrect flag: %v", *flag)
		}
	})

	t.Run("test tertiary connection", func(t *testing.T) {
		flags, err := flagger.Flag(slog.Default(), []openalex.Work{}, nil, "789")
		if err != nil {
			t.Fatal(err)
		}

		if len(flags) != 1 {
			t.Fatal("expected 1 flag")
		}

		flag := flags[0].(*api.MiscHighRiskAssociationFlag)

		if flag.DocTitle != "leaked docs" ||
			flag.DocUrl != "leakeddocs.com" ||
			!slices.Equal(flag.DocEntities, []string{"qrs"}) ||
			flag.EntityMentioned != "qrs" ||
			flag.FrequentCoauthor != nil ||
			len(flag.Connections) != 2 ||
			flag.Connections[0].DocTitle != "best friends" ||
			flag.Connections[0].DocUrl != "bestfriends.com" ||
			flag.Connections[1].DocTitle != "graduate students" ||
			flag.Connections[1].DocUrl != "graduatestudents.com" {
			t.Fatalf("incorrect flag: %v", *flag)
		}
	})
}

func TestAuthorNewsArticleFlagger(t *testing.T) {
	flagger := AuthorNewsArticlesFlagger{
		llm: llms.New(),
	}

	flags, err := flagger.Flag(slog.Default(), "charles lieber")
	if err != nil {
		t.Fatal(err)
	}

	if len(flags) < 1 {
		t.Fatal("expected flags")
	}
}
