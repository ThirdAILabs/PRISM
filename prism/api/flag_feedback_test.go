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
		Feedback: api.FlagFeedback{
			IncorrectAuthor: true,
			EntityNotFound:  true,
		},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.HighRiskFunderType][0],
		Feedback: api.FlagFeedback{
			EntityNotEoc: true,
		},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.AuthorAffiliationType][0],
		Feedback: api.FlagFeedback{
			AuthorNotFound: true,
		},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.CoauthorAffiliationType][0],
		Feedback: api.FlagFeedback{
			IncorrectAffiliations: []string{"Author2", "Author4"},
		},
	})

	feedbacks = append(feedbacks, reports.FlagFeedbackTask{
		Flag: flagMap[api.HighRiskPublisherType][0],
		Feedback: api.FlagFeedback{
			IncorrectDoc: api.KeyValue{
				Key:   "doc1",
				Value: "https://doc1.article.com",
			},
		},
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

	manager.UpdateAuthorReport(reportId, schema.ReportCompleted, time.Now(), flattendFlags)

	//save the flag feedback
	feedbacks := AddDummyFeedback(dummyFlag)
	for _, feedback := range feedbacks {
		err := manager.SaveFlagFeedback(reportId, user1, feedback.Flag.GetHash(), feedback.Feedback)
		if err != nil {
			t.Fatalf("error inserting feedback: %v", err)
		}
	}

	// check the feedbacks
	savedFeedbacks, err := manager.GetFlagFeedback(reportId, user1)
	if err != nil {
		t.Fatalf("error getting feedback: %v", err)
	}

	if len(savedFeedbacks) != len(feedbacks) {
		t.Fatalf("expected %d feedbacks, got %d", len(feedbacks), len(savedFeedbacks))
	}

	// sorting the feedbacks and savedFeedbacks	for corresponding comparison
	sort.Slice(feedbacks, func(i, j int) bool {
		return feedbacks[i].Flag.Type() < feedbacks[j].Flag.Type()
	})

	sort.Slice(savedFeedbacks, func(i, j int) bool {
		return savedFeedbacks[i].Flag.Type() < savedFeedbacks[j].Flag.Type()
	})

	for i := range savedFeedbacks {
		if savedFeedbacks[i].Flag.GetHash() != feedbacks[i].Flag.GetHash() {
			t.Fatalf("hash mismatch")
		}

		// Incorrect Author
		if savedFeedbacks[i].Feedback.AuthorNotFound != feedbacks[i].Feedback.AuthorNotFound {
			t.Fatalf("AuthorNotFound mismatch")
		}

		// EntityNotFound
		if savedFeedbacks[i].Feedback.EntityNotFound != feedbacks[i].Feedback.EntityNotFound {
			t.Fatalf("EntityNotFound mismatch")
		}

		// EntityNotEoc
		if savedFeedbacks[i].Feedback.EntityNotEoc != feedbacks[i].Feedback.EntityNotEoc {
			t.Fatalf("EntityNotEoc mismatch")
		}

		// IncorrectDoc
		if savedFeedbacks[i].Feedback.IncorrectDoc.Key != feedbacks[i].Feedback.IncorrectDoc.Key {
			t.Fatalf("IncorrectDoc.Key mismatch")
		}

		if savedFeedbacks[i].Feedback.IncorrectDoc.Value != feedbacks[i].Feedback.IncorrectDoc.Value {
			t.Fatalf("IncorrectDoc.Value mismatch")
		}

		// IncorrectAffiliations
		if len(savedFeedbacks[i].Feedback.IncorrectAffiliations) != len(feedbacks[i].Feedback.IncorrectAffiliations) {
			t.Fatalf("IncorrectAffiliations mismatch")
		}

		for j := range savedFeedbacks[i].Feedback.IncorrectAffiliations {
			if savedFeedbacks[i].Feedback.IncorrectAffiliations[j] != feedbacks[i].Feedback.IncorrectAffiliations[j] {
				t.Fatalf("IncorrectAffiliations mismatch")
			}
		}

	}

}
