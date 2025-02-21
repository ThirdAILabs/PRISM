package tests

import (
	"encoding/json"
	"fmt"
	"prism/prism/api"
	"testing"
	"time"
)

func createReport(t *testing.T, user *api.UserClient, authorId, authorName string) any {
	reportId, err := user.CreateReport(api.CreateReportRequest{
		AuthorId:   authorId,
		AuthorName: authorName,
		Source:     api.OpenAlexSource,
	})
	if err != nil {
		t.Fatal(err)
	}

	report, err := user.WaitForReport(reportId, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if report.Status != "complete" {
		t.Fatal("report failed")
	}

	return report.Content
}

func TestReportGeneration(t *testing.T) {
	user, _ := setupTestEnv(t)

	report := createReport(t, user, "https://openalex.org/A5084836278", "Charles M. Lieber")

	data, _ := json.MarshalIndent(report, "", "    ")

	fmt.Println(string(data))
}
