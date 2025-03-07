package tests

import (
	"prism/prism/api"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReportGeneration(t *testing.T) {
	user := setupTestEnv(t)

	reportRequests := []api.CreateAuthorReportRequest{
		{
			AuthorId:   "https://openalex.org/A5084836278",
			AuthorName: "Charles M. Lieber",
			Source:     api.OpenAlexSource,
		},
		{
			AuthorId:   "https://openalex.org/A5016320004",
			AuthorName: "David Zhang",
			Source:     api.OpenAlexSource,
		},
	}

	reportIds := make([]uuid.UUID, 0, len(reportRequests))

	for _, report := range reportRequests {
		reportId, err := user.CreateReport(report)
		if err != nil {
			t.Fatal(err)
		}
		reportIds = append(reportIds, reportId)
	}

	expectedFlagCounts := []map[string]int{
		{
			api.TalentContractType:               3,
			api.AssociationsWithDeniedEntityType: 0,
			api.HighRiskFunderType:               3,
			api.AuthorAffiliationType:            4,
			api.PotentialAuthorAffiliationType:   0,
			api.MiscHighRiskAssociationType:      5,
			api.CoauthorAffiliationType:          10,
		}, {
			api.TalentContractType:               0,
			api.AssociationsWithDeniedEntityType: 0,
			api.HighRiskFunderType:               12,
			api.AuthorAffiliationType:            2,
			api.PotentialAuthorAffiliationType:   0,
			api.MiscHighRiskAssociationType:      6,
			api.CoauthorAffiliationType:          31,
		},
	}

	for i, reportId := range reportIds {
		report, err := user.WaitForReport(reportId, 200*time.Second)
		if err != nil {
			t.Fatal(err)
		}

		if report.Status != "complete" ||
			report.AuthorId != reportRequests[i].AuthorId ||
			report.AuthorName != reportRequests[i].AuthorName ||
			report.Source != reportRequests[i].Source {
			t.Fatal("incorrect report returned")
		}

		expectedFlags := expectedFlagCounts[i]

		for flagType, expectedCount := range expectedFlags {
			if len(report.Content[flagType]) != expectedCount {
				t.Fatalf("expected %d flags of type %s, got %d", expectedCount, flagType, len(report.Content[flagType]))
			}
		}
	}
}
