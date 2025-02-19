package reports_test

import (
	"prism/prism/reports"
	"prism/prism/schema"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReportManager(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&schema.Report{}, &schema.ReportContent{}, &schema.LicenseUsage{}); err != nil {
		t.Fatal(err)
	}

	manager := reports.NewManager(db)

	user1, user2 := uuid.New(), uuid.New()

	{
		reports, err := manager.ListReports(user1)
		if err != nil {
			t.Fatal(err)
		}
		if len(reports) != 0 {
			t.Fatal("should be no reports")
		}
	}

	license := uuid.New()

	report1, err := manager.CreateReport(license, user1, "1", "author1", "openalex", 1, 3)
	if err != nil {
		t.Fatal(err)
	}

	report2, err := manager.CreateReport(license, user2, "2", "author2", "openalex", 2, 4)
	if err != nil {
		t.Fatal(err)
	}

	report3, err := manager.CreateReport(license, user1, "3", "author3", "google-scholar", 3, 5)
	if err != nil {
		t.Fatal(err)
	}

	if err := manager.UpdateReport(report1, "complete", []byte(`{"key":"value"}`)); err != nil {
		t.Fatal(err)
	}

	{
		reports, err := manager.ListReports(user1)
		if err != nil {
			t.Fatal(err)
		}
		if len(reports) != 2 {
			t.Fatal("should be 2 reports")
		}

		if reports[0].Id != report1 || reports[1].Id != report3 ||
			reports[0].AuthorId != "1" || reports[1].AuthorId != "3" ||
			reports[0].AuthorName != "author1" || reports[1].AuthorName != "author3" ||
			reports[0].Source != "openalex" || reports[1].Source != "google-scholar" ||
			reports[0].StartYear != 1 || reports[1].StartYear != 3 ||
			reports[0].EndYear != 3 || reports[1].EndYear != 5 ||
			reports[0].Status != "complete" || reports[1].Status != "queued" {
			t.Fatal("incorrect reports")
		}
	}

	{
		reports, err := manager.ListReports(user2)
		if err != nil {
			t.Fatal(err)
		}
		if len(reports) != 1 {
			t.Fatal("should be 1 report")
		}

		if reports[0].Id != report2 ||
			reports[0].AuthorId != "2" ||
			reports[0].AuthorName != "author2" ||
			reports[0].Source != "openalex" ||
			reports[0].StartYear != 2 ||
			reports[0].EndYear != 4 ||
			reports[0].Status != "queued" {
			t.Fatal("incorrect report")
		}
	}

	{
		report, err := manager.GetReport(user1, report1)
		if err != nil {
			t.Fatal(err)
		}

		if report.Id != report1 ||
			report.AuthorId != "1" ||
			report.AuthorName != "author1" ||
			report.Source != "openalex" ||
			report.StartYear != 1 ||
			report.EndYear != 3 ||
			report.Status != "complete" {
			t.Fatal("incorrect report")
		}

		if value := report.Content.(map[string]any)["key"]; value != "value" {
			t.Fatal("incorrect report")
		}
	}

	{
		report, err := manager.GetNextReport()
		if err != nil {
			t.Fatal(err)
		}
		if report == nil {
			t.Fatal("report should not be nil")
		}
		if report.Id != report2 || report.AuthorId != "2" {
			t.Fatal("wrong report")
		}
	}
	{
		report, err := manager.GetNextReport()
		if err != nil {
			t.Fatal(err)
		}
		if report == nil {
			t.Fatal("report should not be nil")
		}
		if report.Id != report3 || report.AuthorId != "3" {
			t.Fatal("wrong report")
		}
	}
	{
		report, err := manager.GetNextReport()
		if err != nil {
			t.Fatal(err)
		}
		if report != nil {
			t.Fatal("report should be nil")
		}
	}
}
