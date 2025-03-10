package api_test

import (
	"encoding/json"
	"prism/prism/api"
	"prism/prism/schema"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFlagParsing(t *testing.T) {
	report := api.Report{
		Id:             uuid.New(),
		LastAccessedAt: time.Now(),
		AuthorId:       "author-id-1",
		AuthorName:     "Author Name",
		Source:         api.OpenAlexSource,
		Status:         schema.ReportInProgress,
		Content: map[string][]api.Flag{
			api.TalentContractType: {
				&api.TalentContractFlag{
					Work: api.WorkSummary{
						WorkId: "work-id-1",
					},
					RawAcknowledements: []string{"Raw acks 1"},
				},
			},
			api.AssociationsWithDeniedEntityType: {
				&api.AssociationWithDeniedEntityFlag{
					Work: api.WorkSummary{
						WorkId: "work-id-1",
					},
					RawAcknowledements: []string{"Raw acks 1"},
				},
			},
			api.HighRiskFunderType: {
				&api.HighRiskFunderFlag{
					Work: api.WorkSummary{
						WorkId: "work-id-1",
					},
					Funders: []string{"Funder 1"},
				},
			},
			api.AuthorAffiliationType: {
				&api.AuthorAffiliationFlag{
					Work: api.WorkSummary{
						WorkId: "work-id-1",
					},
					Affiliations: []string{"Affiliation 1"},
				},
			},
			api.PotentialAuthorAffiliationType: {
				&api.PotentialAuthorAffiliationFlag{
					University: "uni-1",
				},
			},
			api.MiscHighRiskAssociationType: {
				&api.MiscHighRiskAssociationFlag{
					DocTitle:        "doc-title-1",
					EntityMentioned: "some entity",
				},
			},
			api.CoauthorAffiliationType: {
				&api.CoauthorAffiliationFlag{
					Message: "talent contract message",
					Work: api.WorkSummary{
						WorkId: "work-id-1",
					},
					Coauthors: []string{"coauthor-1"},
				},
			},
		},
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

			if flags[i].Hash() != flag.Hash() {
				t.Fatal("invalid hash")
			}
		}
	}
}
