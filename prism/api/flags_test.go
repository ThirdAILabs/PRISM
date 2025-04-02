package api_test

import (
	"encoding/json"
	"prism/prism/api"
	"prism/prism/schema"
	"testing"
	"time"

	"github.com/google/uuid"
)

func createDummyFlags() map[string][]api.Flag {
	var flags = make(map[string][]api.Flag)
	TcFlag, _ := api.CreateFlag(
		api.TalentContractType,
		map[string]interface{}{
			"work": api.WorkSummary{
				WorkId: "work-id-1",
			},
			"raw_acknowledgements": []string{"Raw acks 1"},
		},
	)
	flags[api.TalentContractType] = []api.Flag{TcFlag}

	AwDeFlag, _ := api.CreateFlag(
		api.AssociationsWithDeniedEntityType,
		map[string]interface{}{
			"work": api.WorkSummary{
				WorkId: "work-id-1",
			},
			"raw_acknowledgements": []string{"Raw acks 1"},
		},
	)
	flags[api.AssociationsWithDeniedEntityType] = []api.Flag{AwDeFlag}

	HrFtFlag, _ := api.CreateFlag(
		api.HighRiskFunderType,
		map[string]interface{}{
			"work": api.WorkSummary{
				WorkId: "work-id-1",
			},
			"funders": []string{"Funder 1"},
		},
	)
	flags[api.HighRiskFunderType] = []api.Flag{HrFtFlag}

	AaFlag, _ := api.CreateFlag(
		api.AuthorAffiliationType,
		map[string]interface{}{
			"work": api.WorkSummary{
				WorkId: "work-id-1",
			},
			"affiliations": []string{"Affiliation 1"},
		},
	)
	flags[api.AuthorAffiliationType] = []api.Flag{AaFlag}

	PaFlag, _ := api.CreateFlag(
		api.PotentialAuthorAffiliationType,
		map[string]interface{}{
			"university": "uni-1",
		},
	)
	flags[api.PotentialAuthorAffiliationType] = []api.Flag{PaFlag}

	MhRaFlag, _ := api.CreateFlag(
		api.MiscHighRiskAssociationType,
		map[string]interface{}{
			"doc_title":        "doc-title-1",
			"entity_mentioned": "some entity",
		},
	)
	flags[api.MiscHighRiskAssociationType] = []api.Flag{MhRaFlag}

	CaFlag, _ := api.CreateFlag(
		api.CoauthorAffiliationType,
		map[string]interface{}{
			"message": "talent contract message",
			"work": api.WorkSummary{
				WorkId: "work-id-1",
			},
			"coauthors": []string{"coauthor-1"},
		},
	)
	flags[api.CoauthorAffiliationType] = []api.Flag{CaFlag}

	return flags

}

func TestFlagParsing(t *testing.T) {
	DummyFlags := createDummyFlags()

	report := api.Report{
		Id:             uuid.New(),
		LastAccessedAt: time.Now(),
		AuthorId:       "author-id-1",
		AuthorName:     "Author Name",
		Source:         api.OpenAlexSource,
		Status:         schema.ReportInProgress,
		Content:        DummyFlags,
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}

	var parsedReport api.Report
	if err := json.Unmarshal(data, &parsedReport); err != nil {
		t.Fatal(err)
	}

	if report.Id != parsedReport.Id {
		t.Errorf("expected Id %v, got %v", report.Id, parsedReport.Id)
	}
	if !report.LastAccessedAt.Equal(parsedReport.LastAccessedAt) {
		t.Errorf("expected LastAccessedAt %v, got %v", report.LastAccessedAt, parsedReport.LastAccessedAt)
	}
	if report.AuthorId != parsedReport.AuthorId {
		t.Errorf("expected AuthorId %v, got %v", report.AuthorId, parsedReport.AuthorId)
	}
	if report.AuthorName != parsedReport.AuthorName {
		t.Errorf("expected AuthorName %v, got %v", report.AuthorName, parsedReport.AuthorName)
	}
	if report.Source != parsedReport.Source {
		t.Errorf("expected Source %v, got %v", report.Source, parsedReport.Source)
	}
	if report.Status != parsedReport.Status {
		t.Errorf("expected Status %v, got %v", report.Status, parsedReport.Status)
	}

	if len(report.Content) != len(parsedReport.Content) {
		t.Fatal("invalid content")
	}

	for ftype, flags := range report.Content {
		parsedFlags, ok := parsedReport.Content[ftype]
		if !ok {
			t.Errorf("expected flag type %v to be present", ftype)
			continue
		}
		if len(flags) != len(parsedFlags) {
			t.Errorf("expected %d flags for type %v, got %d", len(flags), ftype, len(parsedFlags))
			continue
		}
		for i, flag := range parsedFlags {
			if flag.Type() != ftype {
				t.Fatal("invalid type")
			}

			if flags[i].CalculateHash() != flag.CalculateHash() {
				t.Fatal("invalid hash")
			}
		}
	}
}
