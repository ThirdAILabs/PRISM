package gscholar_test

import (
	"prism/prism/api"
	"prism/prism/gscholar"
	"slices"
	"testing"
)

func TestAuthorSearch(t *testing.T) {
	authorName := "anshumali shrivastava"

	authors, _, err := gscholar.NextGScholarPage(authorName, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(authors) != 1 || authors[0].AuthorId != "SGT23RAAAAAJ" ||
		authors[0].AuthorName != "Anshumali Shrivastava" ||
		len(authors[0].Institutions) == 0 ||
		!slices.Contains(authors[0].Institutions, "Rice University") ||
		authors[0].Source != "google-scholar" {
		t.Fatal("incorrect authors returned")
	}
}

func TestAuthorSearchWithCursor(t *testing.T) {
	checkQuery := func(authors []api.Author) {
		if len(authors) == 0 {
			t.Fatal("expect > 0 results for query")
		}
		for _, author := range authors {
			if len(author.AuthorId) == 0 || len(author.AuthorName) == 0 || len(author.Institutions) == 0 || author.Source != "google-scholar" {
				t.Fatal("author attributes should not be empty")
			}
		}
	}

	authorName := "bill zhang"

	authors1, cursor, err := gscholar.NextGScholarPage(authorName, "")
	if err != nil {
		t.Fatal(err)
	}

	checkQuery(authors1)

	authors2, _, err := gscholar.NextGScholarPage(authorName, cursor)
	if err != nil {
		t.Fatal(err)
	}

	checkQuery(authors2)

	ids1 := make(map[string]bool)
	for _, author := range authors1 {
		ids1[author.AuthorId] = true
	}

	for _, author := range authors2 {
		if ids1[author.AuthorId] {
			t.Fatal("duplicate author with cursor")
		}
	}
}

func TestAuthorPaperIterator(t *testing.T) {
	iter := gscholar.NewAuthorPaperIterator("SGT23RAAAAAJ")

	seen := make(map[string]bool)
	for i := 0; i < 2; i++ {
		papers, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if len(papers) == 0 {
			t.Fatal("got 0 papers")
		}
		for _, paper := range papers {
			if seen[paper] {
				t.Fatal("duplicate paper returned")
			}
			seen[paper] = true
		}
	}
}
