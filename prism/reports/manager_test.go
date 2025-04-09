package reports_test

import (
	"prism/prism/api"
	"prism/prism/reports"
	"prism/prism/schema"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setup(t *testing.T) *reports.ReportManager {
	db := schema.SetupTestDB(t)

	return reports.NewManager(db).
		SetAuthorReportUpdateInterval(time.Second).
		SetUniversityReportUpdateInterval(time.Second)
}

func checkNextAuthorReport(t *testing.T, next *reports.ReportUpdateTask, authorId, authorName, source string, startDate, endDate time.Time, forUniReport bool) {
	if next == nil ||
		next.AuthorId != authorId ||
		next.AuthorName != authorName ||
		next.Source != source ||
		next.StartDate.Sub(startDate).Abs() > 100*time.Millisecond ||
		next.EndDate.Sub(endDate).Abs() > 100*time.Millisecond ||
		next.ForUniversityReport != forUniReport {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: incorrect next report: %v", file, line, next)
	}
}

func checkAuthorReport(t *testing.T, manager *reports.ReportManager, userId, reportId uuid.UUID, authorId, authorName, source, status string, nflags int) {
	report, err := manager.GetAuthorReport(userId, reportId)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: %v", file, line, err)
	}

	if report.Id != reportId ||
		report.AuthorId != authorId ||
		report.AuthorName != authorName ||
		report.Source != source ||
		report.Status != status ||
		(nflags > 0 && len(report.Content) == 0) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: incorrect report: %v", file, line, report)
	}
	for ftype, flags := range report.Content {
		if len(flags) != nflags {
			t.Fatalf("expected %d flags for %v, got %d", nflags, ftype, len(flags))
		}
	}
}

func checkNoNextAuthorReport(t *testing.T, manager *reports.ReportManager) {
	next, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	if next != nil {
		t.Fatal("should be no report")
	}
}

func dummyReportUpdate() []api.Flag {
	return []api.Flag{
		&api.TalentContractFlag{Work: api.WorkSummary{WorkId: uuid.NewString(), PublicationDate: time.Now()}},
		&api.AssociationWithDeniedEntityFlag{Work: api.WorkSummary{WorkId: uuid.NewString(), PublicationDate: time.Now()}},
		&api.HighRiskFunderFlag{Work: api.WorkSummary{WorkId: uuid.NewString(), PublicationDate: time.Now()}},
		&api.AuthorAffiliationFlag{Work: api.WorkSummary{WorkId: uuid.NewString(), PublicationDate: time.Now()}},
		&api.PotentialAuthorAffiliationFlag{University: uuid.NewString()},
		&api.MiscHighRiskAssociationFlag{DocTitle: uuid.NewString()},
		&api.CoauthorAffiliationFlag{Work: api.WorkSummary{WorkId: uuid.NewString(), PublicationDate: time.Now()}},
	}
}

func TestCreateGetAuthorReports(t *testing.T) {
	// This test checks report creation and access, as well as verifing that the report
	// processing queue behaves as expected. It also verifies the caching behavior,
	// it checks that reports content can be reused when possible, and also verifies
	// that reports are requeued when they become stale and are accessed.
	manager := setup(t)

	user1, user2 := uuid.New(), uuid.New()

	checkNoNextAuthorReport(t, manager)

	reportId1, err := manager.CreateAuthorReport(user1, "1", "author1", api.OpenAlexSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	checkAuthorReport(t, manager, user1, reportId1, "1", "author1", api.OpenAlexSource, "queued", 0)

	next1, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	report1End := time.Now()
	checkNextAuthorReport(t, next1, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, report1End, false)

	checkAuthorReport(t, manager, user1, reportId1, "1", "author1", api.OpenAlexSource, "in-progress", 0)

	if err := manager.UpdateAuthorReport(next1.Id, "complete", next1.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	checkAuthorReport(t, manager, user1, reportId1, "1", "author1", api.OpenAlexSource, "complete", 1)

	time.Sleep(500 * time.Millisecond) // This is so that we can wait 1 sec and only the first report is timed out.

	reportId2, err := manager.CreateAuthorReport(user2, "2", "author2", api.GoogleScholarSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	reportId3, err := manager.CreateAuthorReport(user1, "3", "author3", api.OpenAlexSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	checkAuthorReport(t, manager, user2, reportId2, "2", "author2", api.GoogleScholarSource, "queued", 0)
	checkAuthorReport(t, manager, user1, reportId3, "3", "author3", api.OpenAlexSource, "queued", 0)

	next2, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	report2End := time.Now()
	checkNextAuthorReport(t, next2, "2", "author2", api.GoogleScholarSource, reports.EarliestReportDate, report2End, false)

	checkAuthorReport(t, manager, user2, reportId2, "2", "author2", api.GoogleScholarSource, "in-progress", 0)

	if err := manager.UpdateAuthorReport(next2.Id, "complete", next2.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	checkAuthorReport(t, manager, user2, reportId2, "2", "author2", api.GoogleScholarSource, "complete", 1)

	next3, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	report3End := time.Now()
	checkNextAuthorReport(t, next3, "3", "author3", api.OpenAlexSource, reports.EarliestReportDate, report3End, false)

	checkAuthorReport(t, manager, user1, reportId3, "3", "author3", api.OpenAlexSource, "in-progress", 0)

	if err := manager.UpdateAuthorReport(next3.Id, "complete", next3.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	checkAuthorReport(t, manager, user1, reportId3, "3", "author3", api.OpenAlexSource, "complete", 1)

	// Both reports should be removed from queue now
	checkNoNextAuthorReport(t, manager)

	// Check that reports are reused
	reportId4, err := manager.CreateAuthorReport(user2, "1", "author1", api.OpenAlexSource, "", "")
	if err != nil {
		t.Fatal(err)
	}
	// Content should be reused so report is immediately available
	checkAuthorReport(t, manager, user2, reportId4, "1", "author1", api.OpenAlexSource, "complete", 1)
	// Check that no report update is queued
	checkNoNextAuthorReport(t, manager)

	// This is to ensure the first report is stale
	time.Sleep(500 * time.Millisecond)

	// Reports should be stale now
	if err := manager.CheckForStaleAuthorReports(); err != nil {
		t.Fatal(err)
	}

	// The stale report check should add the author2 report to the queue.
	checkAuthorReport(t, manager, user1, reportId1, "1", "author1", api.OpenAlexSource, "queued", 1)
	checkAuthorReport(t, manager, user2, reportId2, "2", "author2", api.GoogleScholarSource, "complete", 1)
	checkAuthorReport(t, manager, user1, reportId3, "3", "author3", api.OpenAlexSource, "complete", 1)
	checkAuthorReport(t, manager, user2, reportId4, "1", "author1", api.OpenAlexSource, "queued", 1)

	next4, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, next4, "1", "author1", api.OpenAlexSource, report1End, time.Now(), false)
	// Check report was only queued once
	checkNoNextAuthorReport(t, manager)

	if err := manager.UpdateAuthorReport(next4.Id, "complete", next4.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	// Check that report is updated
	checkAuthorReport(t, manager, user1, reportId1, "1", "author1", api.OpenAlexSource, "complete", 2)
	checkAuthorReport(t, manager, user2, reportId2, "2", "author2", api.GoogleScholarSource, "complete", 1)
	checkAuthorReport(t, manager, user1, reportId3, "3", "author3", api.OpenAlexSource, "complete", 1)
	checkAuthorReport(t, manager, user2, reportId4, "1", "author1", api.OpenAlexSource, "complete", 2)
}

func TestAuthorReportAccessErrors(t *testing.T) {
	manager := setup(t)

	user1, user2 := uuid.New(), uuid.New()

	reportId, err := manager.CreateAuthorReport(user1, "1", "author1", api.OpenAlexSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := manager.GetAuthorReport(user1, reportId); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.GetAuthorReport(user2, reportId); err != reports.ErrUserCannotAccessReport {
		t.Fatal(err)
	}

	if _, err := manager.GetAuthorReport(user1, uuid.New()); err != reports.ErrReportNotFound {
		t.Fatal(err)
	}
}

func TestListAuthorReports(t *testing.T) {
	manager := setup(t)

	user1, user2 := uuid.New(), uuid.New()

	report1, err := manager.CreateAuthorReport(user1, "1", "author1", api.OpenAlexSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	noReports, err := manager.ListAuthorReports(user2)
	if err != nil {
		t.Fatal(err)
	}
	if len(noReports) != 0 {
		t.Fatal("should be no reports for user2")
	}

	report2, err := manager.CreateAuthorReport(user2, "2", "author2", api.OpenAlexSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	report3, err := manager.CreateAuthorReport(user1, "3", "author3", api.GoogleScholarSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	reports1, err := manager.ListAuthorReports(user1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports1) != 2 {
		t.Fatal("should be 2 reports")
	}

	if reports1[0].Id != report3 || reports1[1].Id != report1 ||
		reports1[0].AuthorId != "3" || reports1[1].AuthorId != "1" ||
		reports1[0].AuthorName != "author3" || reports1[1].AuthorName != "author1" ||
		reports1[0].Source != api.GoogleScholarSource || reports1[1].Source != api.OpenAlexSource ||
		reports1[0].Status != "queued" || reports1[1].Status != "queued" {
		t.Fatal("incorrect reports")
	}

	reports2, err := manager.ListAuthorReports(user2)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports2) != 1 {
		t.Fatal("should be 1 report")
	}

	if reports2[0].Id != report2 ||
		reports2[0].AuthorId != "2" ||
		reports2[0].AuthorName != "author2" ||
		reports2[0].Source != api.OpenAlexSource ||
		reports2[0].Status != "queued" {
		t.Fatal("incorrect report")
	}

	// Update a report to complete
	next, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	if next == nil {
		t.Fatal("report should not be nil")
	}
	checkNextAuthorReport(t, next, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), false)

	if err := manager.UpdateAuthorReport(next.Id, "complete", time.Now(), nil); err != nil {
		t.Fatal(err)
	}

	if err := manager.DeleteAuthorReport(user2, report1); err != reports.ErrReportNotFound {
		t.Fatal("user cannot delete another user's report")
	}

	reports1, err = manager.ListAuthorReports(user1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports1) != 2 {
		t.Fatal("should be 2 reports")
	}

	if reports1[0].Id != report3 || reports1[1].Id != report1 ||
		reports1[0].AuthorId != "3" || reports1[1].AuthorId != "1" ||
		reports1[0].AuthorName != "author3" || reports1[1].AuthorName != "author1" ||
		reports1[0].Source != api.GoogleScholarSource || reports1[1].Source != api.OpenAlexSource ||
		reports1[0].Status != "queued" || reports1[1].Status != "complete" {
		t.Fatal("incorrect reports")
	}

	if err := manager.DeleteAuthorReport(user1, report1); err != nil {
		t.Fatal(err)
	}

	reports1, err = manager.ListAuthorReports(user1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports1) != 1 {
		t.Fatal("should be 1 report")
	}

	if reports1[0].Id != report3 ||
		reports1[0].AuthorId != "3" ||
		reports1[0].AuthorName != "author3" ||
		reports1[0].Source != api.GoogleScholarSource ||
		reports1[0].Status != "queued" {
		t.Fatal("incorrect reports")
	}

}

func checkNextUniversityReport(t *testing.T, next *reports.UniversityReportUpdateTask, universityId, universityName string, updateDate time.Time) {
	if next == nil ||
		next.UniversityId != universityId ||
		next.UniversityName != universityName ||
		next.UpdateDate.Sub(updateDate).Abs() > 100*time.Millisecond {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: incorrect next report: %v", file, line, next)
	}
}

func checkUniversityReport(t *testing.T, manager *reports.ReportManager, userId, reportId uuid.UUID, universityId, universityName, status string, nauthorsFlagged, nflags, authorsReviewed, totalAuthors int) {
	report, err := manager.GetUniversityReport(userId, reportId)
	if err != nil {
		t.Fatal(err)
	}

	if report.Id != reportId ||
		report.UniversityId != universityId ||
		report.UniversityName != universityName ||
		report.Status != status ||
		report.Content.AuthorsReviewed != authorsReviewed ||
		report.Content.TotalAuthors != totalAuthors ||
		(nflags > 0 && len(report.Content.Flags) == 0) {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("%s:%d: incorrect report: %+v", file, line, report)
	}
	for ftype, flags := range report.Content.Flags {
		if len(flags) != nauthorsFlagged {
			_, file, line, _ := runtime.Caller(1)
			t.Fatalf("%s:%d: expected %d authors to be flagged for %v, got %d", file, line, nflags, ftype, len(flags))
		}
		total := 0
		for _, flag := range flags {
			total += flag.FlagCount
		}
		if total != nflags {
			_, file, line, _ := runtime.Caller(1)
			t.Fatalf("%s:%d: expected %d flags for %v, got %d", file, line, nflags, ftype, total)
		}
	}
}

func checkNoNextUniversityReport(t *testing.T, manager *reports.ReportManager) {
	next, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	if next != nil {
		t.Fatalf("should be no report: %+v", next)
	}
}

func TestUserAuthorReportsAreNotUsedInUniversityReports(t *testing.T) {
	manager := setup(t)

	user1 := uuid.New()

	_, err := manager.CreateAuthorReport(user1, "1", "author1", api.OpenAlexSource, "", "")
	if err != nil {
		t.Fatal(err)
	}

	nextAuthor1, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor1, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), false)
	if err := manager.UpdateAuthorReport(nextAuthor1.Id, "complete", nextAuthor1.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	checkNoNextAuthorReport(t, manager)

	if _, err := manager.CreateUniversityReport(user1, "1", "university1"); err != nil {
		t.Fatal(err)
	}

	nextUni1, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextUniversityReport(t, nextUni1, "1", "university1", time.Now())
	if err := manager.UpdateUniversityReport(nextUni1.Id, "complete", nextUni1.UpdateDate, []reports.UniversityAuthorReport{
		{AuthorId: "1", AuthorName: "author1", Source: api.OpenAlexSource},
	}); err != nil {
		t.Fatal(err)
	}

	nextAuthor2, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor2, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), true)

	checkNoNextAuthorReport(t, manager)
}

func TestCreateGetUniversityReports(t *testing.T) {
	// This test checks report creation and access, as well as verifing that the report
	// processing queue behaves as expected. It also verifies the caching behavior,
	// it checks that reports content can be reused when possible, and also verifies
	// that reports are requeued when they become stale and are accessed.
	manager := setup(t)

	user1 := uuid.New()

	uniId1, err := manager.CreateUniversityReport(user1, "1", "university1")
	if err != nil {
		t.Fatal(err)
	}

	checkNoNextAuthorReport(t, manager)

	nextUni1, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextUniversityReport(t, nextUni1, "1", "university1", time.Now())
	if err := manager.UpdateUniversityReport(nextUni1.Id, "complete", nextUni1.UpdateDate, []reports.UniversityAuthorReport{
		{AuthorId: "1", AuthorName: "author1", Source: api.OpenAlexSource},
		{AuthorId: "2", AuthorName: "author2", Source: api.OpenAlexSource},
	}); err != nil {
		t.Fatal(err)
	}

	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 0, 0, 0, 2)

	nextAuthor1Univ, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	report1EndUniv := time.Now()
	checkNextAuthorReport(t, nextAuthor1Univ, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, report1EndUniv, true)
	if err := manager.UpdateAuthorReport(nextAuthor1Univ.Id, "complete", nextAuthor1Univ.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	// Now First Author flags should appear in the Univeristy report
	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 1, 1, 1, 2)

	nextAuthor2, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	if !nextAuthor2.ForUniversityReport {
		t.Fatal("should be university report")
	}
	report2End := time.Now()
	checkNextAuthorReport(t, nextAuthor2, "2", "author2", api.OpenAlexSource, reports.EarliestReportDate, report2End, true)
	if err := manager.UpdateAuthorReport(nextAuthor2.Id, "complete", nextAuthor2.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	// University report should be updated to have flags from second author
	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 2, 2, 2, 2)

	// No other author reports should be queued
	checkNoNextAuthorReport(t, manager)
	checkNoNextUniversityReport(t, manager)

	// Create another university report
	uniId2, err := manager.CreateUniversityReport(user1, "2", "university2")
	if err != nil {
		t.Fatal(err)
	}

	nextUni2, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextUniversityReport(t, nextUni2, "2", "university2", time.Now())
	if err := manager.UpdateUniversityReport(nextUni2.Id, "complete", nextUni2.UpdateDate, []reports.UniversityAuthorReport{
		{AuthorId: "1", AuthorName: "author1", Source: api.OpenAlexSource},
		{AuthorId: "3", AuthorName: "author3", Source: api.OpenAlexSource},
	}); err != nil {
		t.Fatal(err)
	}

	// Author 2 is cached from the first university report so only author 3 needs to be completed
	checkUniversityReport(t, manager, user1, uniId2, "2", "university2", "complete", 1, 1, 1, 2)

	nextAuthor3, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	report3End := time.Now()
	checkNextAuthorReport(t, nextAuthor3, "3", "author3", api.OpenAlexSource, reports.EarliestReportDate, report3End, true)
	if err := manager.UpdateAuthorReport(nextAuthor3.Id, "complete", nextAuthor3.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	// No other author reports should be queued
	checkNoNextAuthorReport(t, manager)
	checkNoNextUniversityReport(t, manager)

	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 2, 2, 2, 2)
	checkUniversityReport(t, manager, user1, uniId2, "2", "university2", "complete", 2, 2, 2, 2)

	// Check that reports are reused
	uniId3, err := manager.CreateUniversityReport(user1, "1", "university1")
	if err != nil {
		t.Fatal(err)
	}
	// Content should be reused so report is immediately available
	checkUniversityReport(t, manager, user1, uniId3, "1", "university1", "complete", 2, 2, 2, 2)
	// Check that no report update is queued
	checkNoNextAuthorReport(t, manager)
	checkNoNextUniversityReport(t, manager)

	// This is to ensure reports are stale
	time.Sleep(time.Second)

	if err := manager.CheckForStaleUniversityReports(); err != nil {
		t.Fatal(err)
	}

	// Author reports are only queued by the CheckForStaleAuthorReports function
	checkNoNextAuthorReport(t, manager)

	nextUni3, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextUniversityReport(t, nextUni3, "1", "university1", time.Now())

	if err := manager.UpdateUniversityReport(nextUni3.Id, "complete", nextUni3.UpdateDate, []reports.UniversityAuthorReport{
		{AuthorId: "1", AuthorName: "author1", Source: api.OpenAlexSource},
		{AuthorId: "4", AuthorName: "author4", Source: api.OpenAlexSource}, // Author 2 is replaced with 4
	}); err != nil {
		t.Fatal(err)
	}

	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 1, 1, 1, 2)

	// Author report should be queued too since it's stale
	nextAuthor4, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor4, "4", "author4", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), true)
	checkNoNextAuthorReport(t, manager)
	if err := manager.UpdateAuthorReport(nextAuthor4.Id, "complete", nextAuthor4.EndDate, nil); err != nil {
		t.Fatal(err)
	}

	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 1, 1, 2, 2)

	if err := manager.CheckForStaleAuthorReports(); err != nil {
		t.Fatal(err)
	}

	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 1, 1, 1, 2)
	checkUniversityReport(t, manager, user1, uniId2, "2", "university2", "queued", 2, 2, 0, 2)

	// Stale author reports in the university report should now be queued
	nextAuthor5, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor5, "1", "author1", api.OpenAlexSource, report1EndUniv, time.Now(), true)

	if err := manager.UpdateAuthorReport(nextAuthor5.Id, "complete", nextAuthor5.EndDate, dummyReportUpdate()); err != nil {
		t.Fatal(err)
	}

	checkUniversityReport(t, manager, user1, uniId1, "1", "university1", "complete", 1, 2, 2, 2)
	checkUniversityReport(t, manager, user1, uniId2, "2", "university2", "queued", 2, 3, 1, 2)
}

func TestListUniversityReport(t *testing.T) {
	manager := setup(t)

	user1, user2 := uuid.New(), uuid.New()

	uniId1, err := manager.CreateUniversityReport(user1, "1", "university1")
	if err != nil {
		t.Fatal(err)
	}

	noReports, err := manager.ListUniversityReports(user2)
	if err != nil {
		t.Fatal(err)
	}
	if len(noReports) != 0 {
		t.Fatal("should be no reports for user2")
	}

	uniId2, err := manager.CreateUniversityReport(user2, "2", "university2")
	if err != nil {
		t.Fatal(err)
	}

	uniId3, err := manager.CreateUniversityReport(user1, "3", "university3")
	if err != nil {
		t.Fatal(err)
	}

	reports1, err := manager.ListUniversityReports(user1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports1) != 2 {
		t.Fatal("should be 2 reports")
	}

	if reports1[0].Id != uniId3 || reports1[1].Id != uniId1 ||
		reports1[0].UniversityId != "3" || reports1[1].UniversityId != "1" ||
		reports1[0].UniversityName != "university3" || reports1[1].UniversityName != "university1" ||
		reports1[0].Status != "queued" || reports1[1].Status != "queued" {
		t.Fatalf("incorrect reports: %+v", reports1)
	}

	reports2, err := manager.ListUniversityReports(user2)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports2) != 1 {
		t.Fatal("should be 1 report")
	}

	if reports2[0].Id != uniId2 ||
		reports2[0].UniversityId != "2" ||
		reports2[0].UniversityName != "university2" ||
		reports2[0].Status != "queued" {
		t.Fatal("incorrect report")
	}

	if err := manager.DeleteUniversityReport(user2, uniId3); err != reports.ErrReportNotFound {
		t.Fatal("user cannot delete another user report")
	}

	reportsAfterInvalidDelete, err := manager.ListUniversityReports(user1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reportsAfterInvalidDelete) != 2 {
		t.Fatal("should be 2 reports")
	}

	if reportsAfterInvalidDelete[0].Id != uniId3 || reportsAfterInvalidDelete[1].Id != uniId1 ||
		reportsAfterInvalidDelete[0].UniversityId != "3" || reportsAfterInvalidDelete[1].UniversityId != "1" ||
		reportsAfterInvalidDelete[0].UniversityName != "university3" || reportsAfterInvalidDelete[1].UniversityName != "university1" ||
		reportsAfterInvalidDelete[0].Status != "queued" || reportsAfterInvalidDelete[1].Status != "queued" {
		t.Fatalf("incorrect reports: %+v", reports1)
	}

	if err := manager.DeleteUniversityReport(user1, uniId3); err != nil {
		t.Fatal(err)
	}

	reportsAfterDelete, err := manager.ListUniversityReports(user1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reportsAfterDelete) != 1 {
		t.Fatal("should be 1 report")
	}

	if reportsAfterDelete[0].Id != uniId1 ||
		reportsAfterDelete[0].UniversityId != "1" ||
		reportsAfterDelete[0].UniversityName != "university1" ||
		reportsAfterDelete[0].Status != "queued" {
		t.Fatal("incorrect reports")
	}
}

func TestUniversityReportsFilterFlagsByDate(t *testing.T) {
	manager := setup(t)

	user := uuid.New()

	uniId, err := manager.CreateUniversityReport(user, "1", "university1")
	if err != nil {
		t.Fatal(err)
	}

	nextUni, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}

	if err := manager.UpdateUniversityReport(nextUni.Id, "complete", nextUni.UpdateDate, []reports.UniversityAuthorReport{
		{AuthorId: "1", AuthorName: "author1", Source: api.OpenAlexSource},
	}); err != nil {
		t.Fatal(err)
	}

	nextAuthor, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	if !nextAuthor.ForUniversityReport {
		t.Fatal("should be university report")
	}

	content := []api.Flag{
		// This should not be filtered because the date is recent
		&api.TalentContractFlag{Work: api.WorkSummary{WorkId: uuid.NewString(), PublicationDate: time.Now()}},
		// This should be filtered because the date is too old
		&api.HighRiskFunderFlag{Work: api.WorkSummary{WorkId: uuid.NewString(), PublicationDate: time.Now().AddDate(-6, 0, 0)}},
		// This should not be filtered because there is no date
		&api.MiscHighRiskAssociationFlag{DocTitle: uuid.NewString()},
	}

	if err := manager.UpdateAuthorReport(nextAuthor.Id, "complete", nextAuthor.EndDate, content); err != nil {
		t.Fatal(err)
	}

	report, err := manager.GetUniversityReport(user, uniId)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Content.Flags) != 2 ||
		len(report.Content.Flags[api.TalentContractType]) != 1 ||
		len(report.Content.Flags[api.HighRiskFunderType]) != 0 ||
		len(report.Content.Flags[api.MiscHighRiskAssociationType]) != 1 {
		t.Fatalf("incorrect flags: %+v", report.Content.Flags)
	}
}

func TestUserQueuedReportsArePrioritizedOverUniversityReports(t *testing.T) {
	manager := setup(t)

	user := uuid.New()

	if _, err := manager.CreateAuthorReport(user, "1", "author1", api.OpenAlexSource, "", ""); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.CreateUniversityReport(user, "1", "university1"); err != nil {
		t.Fatal(err)
	}

	nextUni, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.UpdateUniversityReport(nextUni.Id, "complete", nextUni.UpdateDate, []reports.UniversityAuthorReport{
		{AuthorId: "1", AuthorName: "author1", Source: api.OpenAlexSource},
		{AuthorId: "2", AuthorName: "author2", Source: api.OpenAlexSource},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.CreateAuthorReport(user, "3", "author3", api.OpenAlexSource, "", ""); err != nil {
		t.Fatal(err)
	}

	// Author 1 report was queued by user
	nextAuthor1, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor1, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), false)

	// University report was queued first, but author report 3 was queued by user
	nextAuthor3, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor3, "3", "author3", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), false)

	// Author 2 report was queued by university
	nextAuthor1Univ, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor1Univ, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), true)

	nextAuthor2, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, nextAuthor2, "2", "author2", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), true)
}

func TestAuthorReportRetry(t *testing.T) {
	manager := setup(t).SetAuthorReportUpdateInterval(reports.AuthorReportUpdateInterval).SetAuthorReportTimeout(time.Second)

	user := uuid.New()
	if _, err := manager.CreateAuthorReport(user, "1", "author1", api.OpenAlexSource, "", ""); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.CreateAuthorReport(user, "2", "author2", api.OpenAlexSource, "", ""); err != nil {
		t.Fatal(err)
	}

	next1, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, next1, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), false)

	next2, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextAuthorReport(t, next2, "2", "author2", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), false)

	time.Sleep(time.Second)

	checkNoNextAuthorReport(t, manager)

	if err := manager.CheckForStaleAuthorReports(); err != nil {
		t.Fatal(err)
	}

	next3, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}

	// Author report 1 should be retried because the timeout is expired, and its status is left as in-progress.
	checkNextAuthorReport(t, next3, "1", "author1", api.OpenAlexSource, reports.EarliestReportDate, time.Now(), false)
}

func TestUniversityReportRetry(t *testing.T) {
	manager := setup(t).SetUniversityReportTimeout(time.Second).SetUniversityReportUpdateInterval(reports.UniversityReportUpdateInterval)

	user := uuid.New()
	if _, err := manager.CreateUniversityReport(user, "1", "university1"); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.CreateUniversityReport(user, "2", "university2"); err != nil {
		t.Fatal(err)
	}

	next1, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextUniversityReport(t, next1, "1", "university1", time.Now())

	next2, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	checkNextUniversityReport(t, next2, "2", "university2", time.Now())

	checkNoNextUniversityReport(t, manager)

	time.Sleep(time.Second)

	if err := manager.CheckForStaleUniversityReports(); err != nil {
		t.Fatal(err)
	}

	next3, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}

	// University report 1 should be retried because the timeout is expired and its status is left as in-progress.
	checkNextUniversityReport(t, next3, "1", "university1", time.Now())
}
