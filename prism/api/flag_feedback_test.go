package api_test

import (
	"prism/prism/api"
	"prism/prism/reports"
	"prism/prism/schema"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setup(t *testing.T) *reports.ReportManager {
	db := schema.SetupTestDB(t)

	return reports.NewManager(db)
}

func AddDummyFeedback(flagMap map[string][]api.Flag) []reports.FlagFeedbackTask {
	feedbacks := make([]reports.FlagFeedbackTask, 0)

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.TalentContractType][0],
		Feedbacks: []api.FlagFeedback{{
			IncorrectAuthor: true,
			EntityNotFound:  true,
		}},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.HighRiskFunderType][0],
		Feedbacks: []api.FlagFeedback{{
			EntityNotEoc: true,
		}},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.AuthorAffiliationType][0],
		Feedbacks: []api.FlagFeedback{{
			AuthorNotFound: true,
		}},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.CoauthorAffiliationType][0],
		Feedbacks: []api.FlagFeedback{{
			IncorrectAffiliations: []string{"Author2", "Author4"},
		}},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.HighRiskPublisherType][0],
		Feedbacks: []api.FlagFeedback{{
			IncorrectDoc: api.KeyValue{
				Key:   "doc1",
				Value: "https://doc1.article.com",
			},
		}},
	})

	return feedbacks
}

func TestInsertFeedback(t *testing.T) {
	manager := setup(t)

	user1 := uuid.New()
	reportId, err := manager.CreateAuthorReport(user1, "1", "author1", api.OpenAlexSource)
	if err != nil {
		t.Fatalf("error creating report: %v", err)
	}

	dummyFlag := createDummyFlags()
	flattendFlags := make([]api.Flag, 0)
	for _, flags := range dummyFlag {
		flattendFlags = append(flattendFlags, flags...)
	}

	if err = manager.UpdateAuthorReport(reportId, schema.ReportCompleted, time.Now(), flattendFlags); err != nil {
		t.Fatalf("error updating report: %v", err)
	}

	//save the flag feedback
	DummyFlagsWithfeedbacks := AddDummyFeedback(dummyFlag)
	for _, flagWithFeedbacks := range DummyFlagsWithfeedbacks {
		for _, feedback := range flagWithFeedbacks.Feedbacks {
			err := manager.SaveFlagFeedback(reportId, user1, flagWithFeedbacks.Flag.GetHash(), feedback)
			if err != nil {
				t.Fatalf("error inserting feedback: %v", err)
			}
		}
	}

	// check the feedbacks
	ParsedFlagsWithFeedbacks, err := manager.GetFlagFeedback(reportId, user1)
	if err != nil {
		t.Fatalf("error getting feedback: %v", err)
	}

	if len(ParsedFlagsWithFeedbacks) != len(DummyFlagsWithfeedbacks) {
		t.Fatalf("expected %d feedbacks, got %d", len(DummyFlagsWithfeedbacks), len(ParsedFlagsWithFeedbacks))
	}

	// sorting the feedbacks and savedFeedbacks	for corresponding comparison
	sort.Slice(DummyFlagsWithfeedbacks, func(i, j int) bool {
		return DummyFlagsWithfeedbacks[i].Flag.Type() < DummyFlagsWithfeedbacks[j].Flag.Type()
	})

	sort.Slice(ParsedFlagsWithFeedbacks, func(i, j int) bool {
		return ParsedFlagsWithFeedbacks[i].Flag.Type() < ParsedFlagsWithFeedbacks[j].Flag.Type()
	})

	for i := range ParsedFlagsWithFeedbacks {
		tempParsedFlag := ParsedFlagsWithFeedbacks[i].Flag
		tempParsedFeedbacks := ParsedFlagsWithFeedbacks[i].Feedbacks
		tempDummyFlag := DummyFlagsWithfeedbacks[i].Flag
		tempDummyFeedback := DummyFlagsWithfeedbacks[i].Feedbacks

		if tempParsedFlag.GetHash() != tempDummyFlag.GetHash() {
			t.Fatalf("hash mismatch")
		}

		// verify the feedbacks for this flag
		for j := range tempParsedFeedbacks {
			// Author Not Found
			if tempParsedFeedbacks[j].AuthorNotFound != tempDummyFeedback[j].AuthorNotFound {
				t.Fatalf("AuthorNotFound mismatch")
			}

			// IncorrectAuthor
			if tempParsedFeedbacks[j].IncorrectAuthor != tempDummyFeedback[j].IncorrectAuthor {
				t.Fatalf("IncorrectAuthor mismatch")
			}

			// EntityNotFound
			if tempParsedFeedbacks[j].EntityNotFound != tempDummyFeedback[j].EntityNotFound {
				t.Fatalf("EntityNotFound mismatch")
			}

			// EntityNotEoc
			if tempParsedFeedbacks[j].EntityNotEoc != tempDummyFeedback[j].EntityNotEoc {
				t.Fatalf("EntityNotEoc mismatch")
			}

			// IncorrectDoc
			if tempParsedFeedbacks[j].IncorrectDoc.Key != tempDummyFeedback[j].IncorrectDoc.Key {
				t.Fatalf("IncorrectDoc.Key mismatch")
			}

			if tempParsedFeedbacks[j].IncorrectDoc.Value != tempDummyFeedback[j].IncorrectDoc.Value {
				t.Fatalf("IncorrectDoc.Value mismatch")
			}

			// IncorrectAffiliations
			if len(tempParsedFeedbacks[j].IncorrectAffiliations) != len(tempDummyFeedback[j].IncorrectAffiliations) {
				t.Fatalf("IncorrectAffiliations mismatch")
			}

			for j := range tempParsedFeedbacks[j].IncorrectAffiliations {
				if tempParsedFeedbacks[j].IncorrectAffiliations[j] != tempDummyFeedback[j].IncorrectAffiliations[j] {
					t.Fatalf("IncorrectAffiliations mismatch")
				}
			}
		}

	}

}
