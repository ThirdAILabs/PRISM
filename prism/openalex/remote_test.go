package openalex_test

import (
	"prism/prism/openalex"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestAutocompleteAuthor(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	results, err := oa.AutocompleteAuthor("anshumali shriva", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("should have some results")
	}

	found := false
	for _, res := range results {
		if !strings.HasPrefix(res.Id, "https://openalex.org/") ||
			!strings.EqualFold(res.Name, "Anshumali Shrivastava") {
			t.Fatal("invalid result")
		}
		if strings.EqualFold(res.Hint, "Rice University, USA") {
			found = true
		}
	}

	if !found {
		t.Fatal("didn't find correct result")
	}
}

func TestAutocompleteInstitution(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	results, err := oa.AutocompleteInstitution("rice univer")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatal("should have 1 result")
	}

	if !strings.HasPrefix(results[0].Id, "https://openalex.org/") ||
		!strings.EqualFold(results[0].Name, "Rice University") ||
		!strings.EqualFold(results[0].Hint, "Houston, USA") {
		t.Fatal("invalid result")
	}
}

func TestAutocompletePaper(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	results, err := oa.AutocompletePaper("From Research to Production: Towards Scalable and Sustainable Neural Recommendation")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatal("should have 1 result")
	}

	expectedTitle := "From Research to Production: Towards Scalable and Sustainable Neural Recommendation Models on Commodity CPU Hardware"

	if !strings.HasPrefix(results[0].Id, "https://openalex.org/") ||
		!strings.EqualFold(results[0].Name, expectedTitle) ||
		!strings.HasPrefix(results[0].Hint, "Anshumali Shrivastava") {
		t.Fatal("invalid result")
	}
}

func TestFindAuthors(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	authorName := "anshumali shrivastava"
	insitutionId := "https://openalex.org/I74775410"

	results, err := oa.FindAuthors(authorName, insitutionId)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 || !strings.HasPrefix(results[0].AuthorId, "https://openalex.org/") ||
		results[0].DisplayName != "Anshumali Shrivastava" ||
		len(results[0].Institutions) == 0 ||
		!slices.Contains(results[0].InstitutionNames(), "Rice University") ||
		!strings.HasPrefix(results[0].Institutions[0].InstitutionId, "https://openalex.org/") ||
		len(results[0].DisplayNameAlternatives) == 0 {
		t.Fatal("incorrect authors returned")
	}
}

func TestFindAuthorsOrcidId(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	orcidId := "0000-0002-5042-2856"
	result, err := oa.FindAuthorByOrcidId(orcidId)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(result.AuthorId, "https://openalex.org/") ||
		result.DisplayName != "Anshumali Shrivastava" ||
		len(result.Institutions) == 0 ||
		!slices.Contains(result.InstitutionNames(), "Rice University") {
		t.Fatal("incorrect author returned")
	}
}

func yearStart(year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
}

func yearEnd(year int) time.Time {
	return time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
}

func TestStreamWorks(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	authorId := "https://openalex.org/A5024993683"
	stream := oa.StreamWorks(authorId, yearStart(2024), yearEnd(2024))

	results := make([]openalex.Work, 0)
	for result := range stream {
		if result.Error != nil {
			t.Fatal(result.Error)
		}

		if !slices.Contains(result.TargetAuthorIds, authorId) {
			t.Fatal("missing author id")
		}

		results = append(results, result.Works...)
	}

	if len(results) < 2 {
		t.Fatal("expected > 1 results")
	}

	found := false
	for _, work := range results {
		if !strings.HasPrefix(work.WorkId, "https://openalex.org/") ||
			work.DisplayName == "" ||
			work.WorkUrl == "" ||
			work.OaUrl == "" ||
			work.DownloadUrl == "" ||
			work.PublicationDate.Year() != 2024 ||
			len(work.Authors) == 0 ||
			len(work.Locations) == 0 {
			t.Fatal("invalid work")
		}

		for _, author := range work.Authors {
			if !strings.HasPrefix(author.AuthorId, "https://openalex.org/") ||
				author.DisplayName == "" ||
				author.RawAuthorName == nil || *author.RawAuthorName == "" {
				t.Fatal("invalid author")
			}
			if work.WorkId == "https://openalex.org/W4396722559" {
				// This work gives institutions for the authors, others don't
				if len(author.Institutions) == 0 {
					t.Fatal("missing institutions")
				}
				found = true
				for _, inst := range author.Institutions {
					if !strings.HasPrefix(inst.InstitutionId, "https://openalex.org/") ||
						inst.InstitutionName == "" {
						t.Fatal("invalid institution")
					}
				}
			}
		}
	}

	if !found {
		t.Fatal("missing work")
	}
}

func TestStreamWorksWithFundersAndLocations(t *testing.T) {
	// This test is to check that the grants and locations are correctly processed, if they are missing
	// then something is wrong with the response parsing.
	authorId := "https://openalex.org/A5084836278"
	workId := "https://openalex.org/W2910300516"

	oa := openalex.NewRemoteKnowledgeBase()
	stream := oa.StreamWorks(authorId, yearStart(2019), yearEnd(2019))

	results := make([]openalex.Work, 0)
	for result := range stream {
		if result.Error != nil {
			t.Fatal(result.Error)
		}

		if !slices.Contains(result.TargetAuthorIds, authorId) {
			t.Fatal("missing author id")
		}

		results = append(results, result.Works...)
	}

	found := false
	for _, work := range results {
		if work.WorkId != workId {
			continue
		}
		found = true

		for _, author := range work.Authors {
			if !strings.HasPrefix(author.AuthorId, "https://openalex.org/") ||
				author.DisplayName == "" ||
				author.RawAuthorName == nil || *author.RawAuthorName == "" {
				t.Fatal("invalid author")
			}
			// This work gives institutions for the authors, others don't
			if len(author.Institutions) == 0 {
				t.Fatal("missing institutions")
			}
			for _, inst := range author.Institutions {
				if !strings.HasPrefix(inst.InstitutionId, "https://openalex.org/") ||
					inst.InstitutionName == "" || inst.Location == "" {
					t.Fatal("invalid institution")
				}
			}
		}

		if len(work.Grants) == 0 {
			t.Fatal("missing grants")
		}

		for _, grant := range work.Grants {
			if grant.FunderId == "" || grant.FunderName == "" {
				t.Fatal("missing funder info")
			}
		}

		for _, loc := range work.Locations {
			if loc.OrganizationId == "" || loc.OrganizationName == "" {
				t.Fatal("missing location info")
			}
		}
	}

	if !found {
		t.Fatal("missing work")
	}
}

func TestFindWorksByTitle(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	titles := []string{
		"From Research to Production: Towards Scalable and Sustainable Neural Recommendation Models on Commodity CPU Hardware",
		"Near Neighbor Search for Constraint Queries",
		"Learning Scalable Structural Representations for Link Prediction with Bloom Signatures",
	}

	results, err := oa.FindWorksByTitle(titles, yearStart(2023), yearEnd(2024))
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != len(titles) {
		t.Fatal("invalid results")
	}

	for i, work := range results {
		if !strings.HasPrefix(work.WorkId, "https://openalex.org/") ||
			work.DisplayName != titles[i] ||
			work.WorkUrl == "" || work.OaUrl == "" ||
			len(work.Authors) == 0 ||
			len(work.Locations) == 0 {
			t.Fatal("invalid work")
		}

		for _, author := range work.Authors {
			if !strings.HasPrefix(author.AuthorId, "https://openalex.org/") ||
				author.DisplayName == "" ||
				author.RawAuthorName == nil || *author.RawAuthorName == "" {
				t.Fatal("invalid author")
			}
			for _, inst := range author.Institutions {
				if !strings.HasPrefix(inst.InstitutionId, "https://openalex.org/") ||
					inst.InstitutionName == "" {
					t.Fatal("invalid institution")
				}
			}
		}
	}
}

func TestGetAuthor(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	authorId := "https://openalex.org/A5024993683"
	author, err := oa.GetAuthor(authorId)
	if err != nil {
		t.Fatal(err)
	}

	if author.AuthorId != authorId || author.DisplayName != "Anshumali Shrivastava" ||
		len(author.DisplayNameAlternatives) == 0 || len(author.Institutions) == 0 {
		t.Fatal("incorrect author")
	}
}

func TestGetInstitutionAuthors(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	institutionId := "https://openalex.org/I74775410" // Rice university

	startDate, endDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)
	authors, err := oa.GetInstitutionAuthors(institutionId, startDate, endDate)
	if err != nil {
		t.Fatal(err)
	}

	if len(authors) == 0 {
		t.Fatal("expected > 0 authors")
	}

	for _, author := range authors {
		if author.AuthorId == "" || author.AuthorName == "" {
			t.Fatal("author info cannot be empty")
		}
	}
}

func TestAutoCompleteInstituteAuthor(t *testing.T) {
	oa := openalex.NewRemoteKnowledgeBase()

	results, err := oa.AutocompleteAuthor("M kather", "https://openalex.org/I91045830")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("should have some results")
	}

	found := false
	for _, res := range results {
		if strings.EqualFold(res.Name, "M. Katherine Banks") && strings.EqualFold(res.Hint, "Texas A&M University, USA") {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("didn't find correct result")
	}
}
