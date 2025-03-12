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
			AuthorId:   "https://openalex.org/A5100327325",
			AuthorName: "Xin Zhang",
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
			api.TalentContractType:               4,
			api.AssociationsWithDeniedEntityType: 0,
			api.HighRiskFunderType:               18,
			api.AuthorAffiliationType:            13,
			api.PotentialAuthorAffiliationType:   0,
			api.MiscHighRiskAssociationType:      0,
			api.CoauthorAffiliationType:          21,
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
		report, err := user.WaitForReport(reportId, 300*time.Second)
		if err != nil {
			t.Errorf("Report %s (%s): failed to get report %v", reportRequests[i].AuthorName, reportRequests[i].AuthorId, err)
			continue
		}

		if report.Status != "complete" ||
			report.AuthorId != reportRequests[i].AuthorId ||
			report.AuthorName != reportRequests[i].AuthorName ||
			report.Source != reportRequests[i].Source {
			t.Errorf("Report %s (%s): incorrect report status=%s, author_id=%s, author_name=%s, source=%s", reportRequests[i].AuthorName, reportRequests[i].AuthorId, report.Status, report.AuthorId, report.AuthorName, report.Source)
			continue
		}

		expectedFlags := expectedFlagCounts[i]

		for flagType, expectedCount := range expectedFlags {
			if len(report.Content[flagType]) < expectedCount-3 {
				t.Errorf("Report %s (%s): expected >= %d flags of type %s, got %d", reportRequests[i].AuthorName, reportRequests[i].AuthorId, expectedCount-2, flagType, len(report.Content[flagType]))
			}
		}
	}
}
