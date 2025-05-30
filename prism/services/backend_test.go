package services_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"prism/prism/api"
	"prism/prism/licensing"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/schema"
	"prism/prism/search"
	"prism/prism/services"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func init() {
	const licensePath = "../../.test_license/thirdai.license"
	if err := search.SetLicensePath(licensePath); err != nil {
		panic(err)
	}
}

func shouldSkip(t *testing.T) {
	if os.Getenv("SKIP_SERP_TESTS") != "" {
		t.Skip("Skipping SERP tests due to SKIP_SERP_TESTS env var")
	}
}

const (
	userPrefix = "user"
)

func newUser() string {
	return userPrefix + uuid.NewString()
}

type MockTokenVerifier struct {
	prefix string
}

func (m *MockTokenVerifier) VerifyToken(token string) (uuid.UUID, string, error) {
	if !strings.HasPrefix(token, m.prefix) {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	id, err := uuid.Parse(strings.TrimPrefix(token, m.prefix))
	if err != nil {
		return uuid.Nil, "", err
	}
	return id, id.String() + "@mock.com", nil
}

type mockOpenAlex struct{}

func (m *mockOpenAlex) AutocompleteAuthor(query string) ([]api.Autocompletion, error) {
	return nil, nil
}

func (m *mockOpenAlex) AutocompleteInstitution(query string) ([]api.Autocompletion, error) {
	return nil, nil
}

func (m *mockOpenAlex) AutocompletePaper(query string) ([]api.Autocompletion, error) {
	return nil, nil
}

func (m *mockOpenAlex) FindAuthors(authorName, institutionId string) ([]openalex.Author, error) {
	return nil, nil
}

func (m *mockOpenAlex) FindAuthorByOrcidId(orcidId string) (openalex.Author, error) {
	return openalex.Author{}, nil
}

func (m *mockOpenAlex) StreamWorks(authorId string, startDate, endDate time.Time) chan openalex.WorkBatch {
	return nil
}

func (m *mockOpenAlex) FindWorksByTitle(titles []string, startDate, endDate time.Time) ([]openalex.Work, error) {
	return nil, nil
}

func (m *mockOpenAlex) GetAuthor(authorId string) (openalex.Author, error) {
	return openalex.Author{
		Institutions: []openalex.Institution{{InstitutionName: authorId + "-affiliation1"}, {InstitutionName: authorId + "-affiliation2"}},
		Concepts:     []string{authorId + "-interest1", authorId + "-interest2"},
	}, nil
}

func (m *mockOpenAlex) GetInstitutionAuthors(institutionId string, startDate, endDate time.Time) ([]openalex.InstitutionAuthor, error) {
	return nil, nil
}

func createBackend(t *testing.T) (http.Handler, *gorm.DB) {
	db := schema.SetupTestDB(t)

	entities := []api.MatchedEntity{{Names: "abc university"}, {Names: "institute of xyz"}, {Names: "123 org"}}

	licensing, err := licensing.NewLicenseVerifier("AC013F-FD0B48-00B160-64836E-76E88D-V3")
	if err != nil {
		t.Fatal(err)
	}

	oa := openalex.NewRemoteKnowledgeBase()

	backend := services.NewBackend(
		services.NewReportService(reports.NewManager(db), licensing, &mockOpenAlex{}, "./resources"),
		services.NewSearchService(oa, entities),
		services.NewAutoCompleteService(oa),
		services.NewHookService(db, map[string]services.Hook{}, 1*time.Second),
		&MockTokenVerifier{prefix: userPrefix},
	)

	return backend.Routes(), db
}

var ErrUnauthorized = errors.New("unauthorized")

func mockRequest(backend http.Handler, method, endpoint, token string, jsonBody any, result any) error {
	var body io.Reader
	if jsonBody != nil {
		data := new(bytes.Buffer)
		err := json.NewEncoder(data).Encode(jsonBody)
		if err != nil {
			return fmt.Errorf("error encoding json body for endpoint %v: %w", endpoint, err)
		}
		body = data
	}

	req := httptest.NewRequest(method, endpoint, body)
	req.Header.Add("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	backend.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("%v request to endpoint %v returned status %d, content '%v'", method, endpoint, res.StatusCode, w.Body.String())
		if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
			return errors.Join(ErrUnauthorized, err)
		}
		return err
	}

	if result != nil {
		err := json.NewDecoder(res.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("error parsing %v response from endpoint %v: %w", method, endpoint, err)
		}
	}

	return nil
}

func Get(backend http.Handler, endpoint, token string, result any) error {
	return mockRequest(backend, "GET", endpoint, token, nil, result)
}

func Post(backend http.Handler, endpoint, token string, jsonBody any, result any) error {
	return mockRequest(backend, "POST", endpoint, token, jsonBody, result)
}

func Delete(backend http.Handler, endpoint, token string) error {
	return mockRequest(backend, "DELETE", endpoint, token, nil, nil)
}

func compareReport(t *testing.T, report api.Report, expected string) {
	if report.AuthorId != expected+"-id" ||
		report.AuthorName != expected+"-name" ||
		report.Source != api.OpenAlexSource ||
		report.Affiliations != expected+"-id-affiliation1, "+expected+"-id-affiliation2" ||
		report.ResearchInterests != expected+"-id-interest1, "+expected+"-id-interest2" ||
		report.Status != "queued" {
		t.Fatal("invalid reports returned")
	}
}

func checkListAuthorReports(t *testing.T, backend http.Handler, user string, expected []string) {
	var reports []api.Report
	if err := Get(backend, "/report/author/list", user, &reports); err != nil {
		t.Fatal(err)
	}

	slices.SortFunc(reports, func(a, b api.Report) int {
		return strings.Compare(a.AuthorId, b.AuthorId)
	})

	slices.Sort(expected)

	if len(reports) != len(expected) {
		t.Fatal("incorrect number of reports returned")
	}
	for i := range expected {
		compareReport(t, reports[i], expected[i])
	}
}

func getAuthorReport(backend http.Handler, user string, id uuid.UUID) (api.Report, error) {
	var res api.Report
	err := Get(backend, "/report/author/"+id.String(), user, &res)
	return res, err
}

func createAuthorReport(backend http.Handler, user, name string) (api.CreateReportResponse, error) {
	req := api.CreateAuthorReportRequest{
		AuthorId:   name + "-id",
		AuthorName: name + "-name",
		Source:     api.OpenAlexSource,
	}

	var res api.CreateReportResponse
	err := Post(backend, "/report/author/create", user, req, &res)
	return res, err
}

func deleteAuthorReport(backend http.Handler, user string, id uuid.UUID) error {
	return Delete(backend, "/report/author/"+id.String(), user)
}

func checkListUniversityReports(t *testing.T, backend http.Handler, user string, expected []string) {
	var reports []api.UniversityReport
	if err := Get(backend, "/report/university/list", user, &reports); err != nil {
		t.Fatal(err)
	}

	slices.SortFunc(reports, func(a, b api.UniversityReport) int {
		return strings.Compare(a.UniversityId, b.UniversityId)
	})

	slices.Sort(expected)

	if len(reports) != len(expected) {
		t.Fatal("incorrect number of reports returned")
	}
	for i := range expected {
		if reports[i].UniversityId != expected[i]+"-id" ||
			reports[i].UniversityName != expected[i]+"-name" ||
			reports[i].UniversityLocation != expected[i]+"-location" ||
			reports[i].Status != "complete" {
			t.Fatalf("invalid reports returned: %+v, %s", reports[i], expected[i])
		}
	}
}

func getUniversityReport(backend http.Handler, user string, id uuid.UUID) (api.UniversityReport, error) {
	var res api.UniversityReport
	err := Get(backend, "/report/university/"+id.String(), user, &res)
	return res, err
}

func createUniversityReport(backend http.Handler, user, name string) (api.CreateReportResponse, error) {
	req := api.CreateUniversityReportRequest{
		UniversityId:       name + "-id",
		UniversityName:     name + "-name",
		UniversityLocation: name + "-location",
	}

	var res api.CreateReportResponse
	err := Post(backend, "/report/university/create", user, req, &res)
	return res, err
}

func deleteUniversityReport(backend http.Handler, user string, id uuid.UUID) error {
	return Delete(backend, "/report/university/"+id.String(), user)
}

func TestCheckDisclosure(t *testing.T) {
	backend, db := createBackend(t)

	user := newUser()

	manager := reports.NewManager(db)

	reportResp, err := createAuthorReport(backend, user, "disclosure-report")
	if err != nil {
		t.Fatal(err)
	}

	content := []api.Flag{
		&api.TalentContractFlag{
			DisclosableFlag: api.DisclosableFlag{},
			Message:         "Test disclosure flag - TalentContract",
			Work: api.WorkSummary{
				WorkId:          "work-1",
				DisplayName:     "Test Work",
				WorkUrl:         "http://example.com/work-1",
				OaUrl:           "http://example.com/oa/work-1",
				PublicationDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Entities: []api.AcknowledgementEntity{{Entity: "discloseme"}},
		},
		&api.AssociationWithDeniedEntityFlag{
			DisclosableFlag: api.DisclosableFlag{},
			Message:         "Test disclosure flag - Association",
			Work: api.WorkSummary{
				WorkId:          "work-2",
				DisplayName:     "Test Work 2",
				WorkUrl:         "http://example.com/work-2",
				OaUrl:           "http://example.com/oa/work-2",
				PublicationDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Entities: []api.AcknowledgementEntity{{Entity: "nonmatching"}},
		},
	}

	nextReport, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	if nextReport.ForUniversityReport {
		t.Fatal("next report should not be for university report")
	}
	if nextReport == nil {
		t.Fatal("next report should not be nil")
	}

	if err := manager.UpdateAuthorReport(nextReport.Id, schema.ReportCompleted, time.Now(), content); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("files", "sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte("This sample file contains discloseme which should trigger disclosure."))
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	endpoint := fmt.Sprintf("/report/author/%s/check-disclosure", reportResp.Id.String())
	req := httptest.NewRequest("POST", endpoint, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", "Bearer "+user)

	w := httptest.NewRecorder()
	backend.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200; got %d, response: %s", res.StatusCode, w.Body.String())
	}

	var updatedReport api.Report
	if err := json.NewDecoder(res.Body).Decode(&updatedReport); err != nil {
		t.Fatal(err)
	}

	if len(updatedReport.Content["TalentContracts"]) != 1 {
		t.Fatalf("expected 1 TalentContract flag; got %d", len(updatedReport.Content["TalentContracts"]))
	}
	if !updatedReport.Content["TalentContracts"][0].(*api.TalentContractFlag).Disclosed {
		t.Fatal("expected TalentContract flag to be marked as disclosed")
	}

	if len(updatedReport.Content["AssociationsWithDeniedEntities"]) != 1 {
		t.Fatalf("expected 1 AssociationWithDeniedEntity flag; got %d", len(updatedReport.Content["AssociationsWithDeniedEntities"]))
	}
	if updatedReport.Content["AssociationsWithDeniedEntities"][0].(*api.AssociationWithDeniedEntityFlag).Disclosed {
		t.Fatal("expected AssociationWithDeniedEntity flag to remain undisclosed")
	}
}

func TestDownloadReportAllFormats(t *testing.T) {
	backend, db := createBackend(t)

	user := newUser()

	reportResp, err := createAuthorReport(backend, user, "download-report")
	if err != nil {
		t.Fatal(err)
	}

	manager := reports.NewManager(db)

	content := []api.Flag{
		&api.TalentContractFlag{
			DisclosableFlag: api.DisclosableFlag{},
			Message:         "Test Talent Contract",
			Work: api.WorkSummary{
				WorkId:          "work-1",
				DisplayName:     "Test Work",
				WorkUrl:         "http://example.com/work-1",
				OaUrl:           "http://example.com/oa/work-1",
				PublicationDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			RawAcknowledgements: []string{"flag-content"},
		},
	}

	nextReport, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}
	if nextReport.ForUniversityReport {
		t.Fatal("next report should not be for university report")
	}
	if nextReport == nil {
		t.Fatal("next report should not be nil")
	}

	if err := manager.UpdateAuthorReport(nextReport.Id, schema.ReportCompleted, time.Now(), content); err != nil {
		t.Fatal(err)
	}

	formats := []string{"csv", "pdf", "excel"}
	for _, format := range formats {
		endpoint := fmt.Sprintf("/report/author/%s/download?format=%s", reportResp.Id.String(), format)
		req := httptest.NewRequest("POST", endpoint, io.Reader(bytes.NewBuffer([]byte("{}"))))
		req.Header.Add("Authorization", "Bearer "+user)
		w := httptest.NewRecorder()
		backend.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		fileBytes, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("error reading response body for format %s: %v", format, err)
		}

		var expectedContentType, expectedFilename string
		switch format {
		case "csv":
			expectedContentType = "text/csv"
			expectedFilename = "download-report-name Report.csv"
		case "pdf":
			expectedContentType = "application/pdf"
			expectedFilename = "download-report-name Report.pdf"
		case "excel":
			expectedContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
			expectedFilename = "download-report-name Report.xlsx"
		}

		if ct := res.Header.Get("Content-Type"); ct != expectedContentType {
			t.Fatalf("expected content type %s for format %s, got %s", expectedContentType, format, ct)
		}
		contentDisp := res.Header.Get("Content-Disposition")
		if !strings.Contains(contentDisp, expectedFilename) {
			t.Fatalf("expected filename '%s' in Content-Disposition for format %s, got %s", expectedFilename, format, contentDisp)
		}

		if len(fileBytes) == 0 {
			t.Fatalf("downloaded file for format %s is empty", format)
		}

		switch format {
		case "pdf":
			if !strings.HasPrefix(string(fileBytes), "%PDF") {
				t.Fatalf("downloaded file for format %s does not appear to be a valid PDF", format)
			}
		case "excel":
			f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
			if err != nil {
				t.Fatalf("error opening excel file for format %s: %v", format, err)
			}
			sheets := f.GetSheetList()
			found := false
			for _, sheet := range sheets {
				if sheet == "Talent Contracts" {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("excel file does not contain sheet 'Talent Contracts'; sheets: %v", sheets)
			}

			val, err := f.GetCellValue("Talent Contracts", "B2")
			if err != nil {
				t.Fatalf("error reading cell value from 'Talent Contracts' sheet: %v", err)
			}
			if !strings.Contains(val, "Test Work") {
				t.Fatalf("expected flag message 'Test Work' in cell B2, got %s", val)
			}
		case "csv":
			reader := csv.NewReader(bytes.NewReader(fileBytes))
			records, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("error reading CSV for format %s: %v", format, err)
			}
			if len(records) == 0 {
				t.Fatalf("no records found in CSV for format %s", format)
			}
			expectedHeader := []string{"Field", "Value"}
			for i, v := range expectedHeader {
				if records[0][i] != v {
					t.Fatalf("expected header %v, got %v", expectedHeader, records[0])
				}
			}
			if len(records) < 2 || !strings.Contains(records[1][1], reportResp.Id.String()) {
				t.Fatalf("expected report id %s in CSV summary, got %v", reportResp.Id.String(), records[1])
			}
		}
	}
}

func TestAuthorReportEndpoints(t *testing.T) {
	backend, _ := createBackend(t)

	user1, user2 := newUser(), newUser()

	checkListAuthorReports(t, backend, user1, []string{})
	checkListAuthorReports(t, backend, user2, []string{})

	report, err := createAuthorReport(backend, user1, "report1")
	if err != nil {
		t.Fatal(err)
	}

	checkListAuthorReports(t, backend, user1, []string{"report1"})
	checkListAuthorReports(t, backend, user2, []string{})

	if _, err := getAuthorReport(backend, user2, report.Id); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("should be unauthorized: %v", err)
	}

	reportData, err := getAuthorReport(backend, user1, report.Id)
	if err != nil {
		t.Fatal(err)
	}

	compareReport(t, reportData, "report1")

	if _, err := createAuthorReport(backend, user2, "report2"); err != nil {
		t.Fatal(err)
	}

	if _, err := createAuthorReport(backend, user1, "report3"); err != nil {
		t.Fatal(err)
	}

	checkListAuthorReports(t, backend, user1, []string{"report1", "report3"})
	checkListAuthorReports(t, backend, user2, []string{"report2"})

	if err := deleteAuthorReport(backend, user2, report.Id); err == nil || !strings.Contains(err.Error(), "report not found") {
		t.Fatal(err)
	}

	checkListAuthorReports(t, backend, user1, []string{"report1", "report3"})
	checkListAuthorReports(t, backend, user2, []string{"report2"})

	if err := deleteAuthorReport(backend, user1, report.Id); err != nil {
		t.Fatal(err)
	}

	checkListAuthorReports(t, backend, user1, []string{"report3"})
	checkListAuthorReports(t, backend, user2, []string{"report2"})
}

func TestUniversityReportEndpoints(t *testing.T) {
	backend, db := createBackend(t)
	manager := reports.NewManager(db)

	user1, user2 := newUser(), newUser()

	uniReportId, err := createUniversityReport(backend, user1, "uni-report1")
	if err != nil {
		t.Fatal(err)
	}

	nextUniReport, err := manager.GetNextUniversityReport()
	if err != nil {
		t.Fatal(err)
	}
	if nextUniReport == nil {
		t.Fatal("next report should not be nil")
	}

	if err := manager.UpdateUniversityReport(nextUniReport.Id, schema.ReportCompleted, nextUniReport.UpdateDate, []reports.UniversityAuthorReport{
		{AuthorId: "1", AuthorName: "author1", Source: api.OpenAlexSource},
	}); err != nil {
		t.Fatal(err)
	}

	nextAuthorReport, err := manager.GetNextAuthorReport()
	if err != nil {
		t.Fatal(err)
	}

	if !nextAuthorReport.ForUniversityReport {
		t.Fatal("next author report should be for university report")
	}

	if err := manager.UpdateAuthorReport(nextAuthorReport.Id, schema.ReportCompleted, time.Now(), []api.Flag{
		&api.HighRiskFunderFlag{Work: api.WorkSummary{WorkId: "abc", PublicationDate: time.Now()}},
	}); err != nil {
		t.Fatal(err)
	}

	uniReport, err := getUniversityReport(backend, user1, uniReportId.Id)
	if err != nil {
		t.Fatal(err)
	}

	if uniReport.UniversityId != "uni-report1-id" || uniReport.UniversityName != "uni-report1-name" || uniReport.UniversityLocation != "uni-report1-location" || uniReport.Status != "complete" || len(uniReport.Content.Flags[api.HighRiskFunderType]) != 1 {
		t.Fatal("invalid university report returned")
	}

	if _, err := getUniversityReport(backend, user2, uniReportId.Id); err == nil || !strings.Contains(err.Error(), "user cannot access report") {
		t.Fatal(err)
	}

	checkListUniversityReports(t, backend, user1, []string{"uni-report1"})

	if err := deleteUniversityReport(backend, user2, uniReport.Id); err == nil || !strings.Contains(err.Error(), "report not found") {
		t.Fatal(err)
	}
	checkListUniversityReports(t, backend, user1, []string{"uni-report1"})

	if err := deleteUniversityReport(backend, user1, uniReport.Id); err != nil {
		t.Fatal(err)
	}

	checkListUniversityReports(t, backend, user1, []string{})
}

func TestAutocompleteAuthor(t *testing.T) {
	backend, _ := createBackend(t)

	user := newUser()

	var results []api.Autocompletion
	err := mockRequest(backend, "GET", "/autocomplete/author?query="+url.QueryEscape("anshumali shriva"), user, nil, &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("should have some results")
	}

	for _, res := range results {
		if !strings.HasPrefix(res.Id, "https://openalex.org/") ||
			!strings.EqualFold(res.Name, "Anshumali Shrivastava") {
			t.Fatal("invalid result")
		}
	}
}

func TestAutocompleteInstution(t *testing.T) {
	backend, _ := createBackend(t)

	user := newUser()

	var results []api.Autocompletion
	err := mockRequest(backend, "GET", "/autocomplete/institution?query="+url.QueryEscape("rice univer"), user, nil, &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("should have some results")
	}

	for _, res := range results {
		if !strings.HasPrefix(res.Id, "https://openalex.org/") ||
			!strings.EqualFold(res.Name, "Rice University") ||
			!strings.EqualFold(res.Hint, "Houston, USA") {
			t.Fatal("invalid result")
		}
	}
}

func TestAutocompletePaper(t *testing.T) {
	backend, _ := createBackend(t)

	user := newUser()

	query := "From Research to Production: Towards Scalable and Sustainable Neural Recommendation"
	var results []api.Autocompletion
	err := mockRequest(backend, "GET", "/autocomplete/paper?query="+url.QueryEscape(query), user, nil, &results)
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

func TestMatchEntities(t *testing.T) {
	backend, _ := createBackend(t)

	user := newUser()

	var results []api.MatchedEntity
	err := mockRequest(backend, "GET", "/search/match-entities?query="+url.QueryEscape("xyz"), user, nil, &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 || results[0].Names != "institute of xyz" {
		t.Fatalf("incorrect entities matched: %v", results)
	}
}

func TestSearchAuthors(t *testing.T) {
	backend, _ := createBackend(t)

	user := newUser()

	t.Run("Search By Author Name", func(t *testing.T) {
		authorName := "anshumali shrivastava"
		insitutionId := "https://openalex.org/I74775410"
		institutionName := "Rice University"

		url := fmt.Sprintf("/search/authors?author_name=%s&institution_id=%s&institution_name=%s", url.QueryEscape(authorName), url.QueryEscape(insitutionId), url.QueryEscape(institutionName))
		var results []api.Author
		err := mockRequest(backend, "GET", url, user, nil, &results)
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 1 || !strings.HasPrefix(results[0].AuthorId, "https://openalex.org/") ||
			results[0].AuthorName != "Anshumali Shrivastava" ||
			len(results[0].Institutions) == 0 || !slices.Contains(results[0].Institutions, "Rice University") ||
			results[0].Source != "openalex" {
			t.Fatal("incorrect authors returned")
		}
	})

	t.Run("Search By ORCID", func(t *testing.T) {
		orcidId := "0000-0002-5042-2856"

		url := fmt.Sprintf("/search/authors?orcid=%s", url.QueryEscape(orcidId))
		var results []api.Author
		err := mockRequest(backend, "GET", url, user, nil, &results)
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 1 || !strings.HasPrefix(results[0].AuthorId, "https://openalex.org/") ||
			results[0].AuthorName != "Anshumali Shrivastava" ||
			len(results[0].Institutions) == 0 || !slices.Contains(results[0].Institutions, "Rice University") ||
			results[0].Source != "openalex" {
			t.Fatal("incorrect authors returned")
		}
	})

	t.Run("Search By Paper Title", func(t *testing.T) {
		title := "From Research to Production: Towards Scalable and Sustainable Neural Recommendation Models on Commodity CPU Hardware"

		url := fmt.Sprintf("/search/authors?paper_title=%s", url.QueryEscape(title))
		var results []api.Author
		err := mockRequest(backend, "GET", url, user, nil, &results)
		if err != nil {
			t.Fatal(err)
		}

		if len(results) < 1 || !strings.HasPrefix(results[0].AuthorId, "https://openalex.org/") ||
			results[0].AuthorName == "" ||
			len(results[0].Institutions) == 0 ||
			results[0].Source != "openalex" {
			t.Fatal("expected > 0 results")
		}
	})
}

func TestSearchGoogleScholarAuthors(t *testing.T) {
	shouldSkip(t)

	backend, _ := createBackend(t)

	user := newUser()

	authorName := "anshumali shrivastava"

	url := fmt.Sprintf("/search/authors-advanced?author_name=%s&institution_name=%s", url.QueryEscape(authorName), url.QueryEscape("rice university"))
	var results api.GScholarSearchResults
	err := mockRequest(backend, "GET", url, user, nil, &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Authors) != 1 || results.Authors[0].AuthorId != "SGT23RAAAAAJ" ||
		results.Authors[0].AuthorName != "Anshumali Shrivastava" ||
		len(results.Authors[0].Institutions) == 0 || !slices.Contains(results.Authors[0].Institutions, "Rice University") ||
		results.Authors[0].Source != "google-scholar" {
		t.Fatal("incorrect authors returned")
	}
}

func TestSearchGoogleScholarAuthorsWithCursor(t *testing.T) {
	shouldSkip(t)

	backend, _ := createBackend(t)

	user := newUser()

	checkQuery := func(authors []api.Author) {
		if len(authors) == 0 {
			t.Fatal("expect > 0 results for query")
		}

		for _, author := range authors {
			if len(author.AuthorId) == 0 || len(author.AuthorName) == 0 || len(author.Institutions) == 0 || author.Source != "google-scholar" {
				t.Fatal("author attributes should not be empty")
			}
		}
	}

	authorName := "bill zhang"

	url1 := fmt.Sprintf("/search/authors-advanced?author_name=%s&institution_name=any", url.QueryEscape(authorName))
	var results1 api.GScholarSearchResults
	if err := mockRequest(backend, "GET", url1, user, nil, &results1); err != nil {
		t.Fatal(err)
	}

	checkQuery(results1.Authors)

	url2 := fmt.Sprintf("/search/authors-advanced?author_name=%s&institution_name=any&cursor=%s", url.QueryEscape(authorName), results1.Cursor)
	var results2 api.GScholarSearchResults
	if err := mockRequest(backend, "GET", url2, user, nil, &results2); err != nil {
		t.Fatal(err)
	}

	checkQuery(results2.Authors)
}

type hookInvocation struct {
	reportId  uuid.UUID
	data      []byte
	lastRanAt time.Time
}

type testHook struct {
	invoked *hookInvocation
}

func (t *testHook) Validate(data []byte, reportId uuid.UUID, interval int) error {
	return nil
}

func (t *testHook) Run(report api.Report, data []byte, lastRanAt time.Time) error {
	t.invoked = &hookInvocation{reportId: report.Id, data: data, lastRanAt: lastRanAt}
	return nil
}

func (t *testHook) Type() string {
	return "test"
}

func (t *testHook) CreateHookData(r *http.Request, payload []byte, interval int) (hookData []byte, err error) {
	return payload, nil
}

func createReportHook(backend http.Handler, reportId uuid.UUID, user string, payload string, interval int) error {
	return Post(backend, "/hooks/"+reportId.String(), user, api.CreateHookRequest{
		Action: "test", Data: []byte(payload), Interval: interval,
	}, nil)
}

func TestHooks(t *testing.T) {
	db := schema.SetupTestDB(t)

	licensing, err := licensing.NewLicenseVerifier("AC013F-FD0B48-00B160-64836E-76E88D-V3")
	if err != nil {
		t.Fatal(err)
	}

	oa := openalex.NewRemoteKnowledgeBase()

	mockHook := &testHook{invoked: nil}

	manager := reports.NewManager(db).SetAuthorReportUpdateInterval(2 * time.Second)

	hookService := services.NewHookService(db, map[string]services.Hook{"test": mockHook}, 1*time.Second)

	backend := services.NewBackend(
		services.NewReportService(manager, licensing, &mockOpenAlex{}, "./resources"),
		services.NewSearchService(oa, nil),
		services.NewAutoCompleteService(oa),
		hookService,
		&MockTokenVerifier{prefix: userPrefix},
	)

	api := backend.Routes()

	user := newUser()

	report, err := createAuthorReport(api, user, "hook-report")
	if err != nil {
		t.Fatal(err)
	}

	if err := createReportHook(api, report.Id, user, "hook-1", 3); err != nil {
		t.Fatal(err)
	}

	completeNextReport := func() {
		next, err := manager.GetNextAuthorReport()
		if err != nil {
			t.Fatal(err)
		}
		if next == nil {
			t.Fatal("should be next report")
		}

		if err := manager.UpdateAuthorReport(next.Id, schema.ReportCompleted, time.Now(), nil); err != nil {
			t.Fatal(err)
		}
	}

	if mockHook.invoked != nil {
		t.Fatal("hook should not be invoked")
	}

	checkNoQueuedReport := func() {
		next, err := manager.GetNextAuthorReport()
		if err != nil {
			t.Fatal(err)
		}
		if next != nil {
			t.Fatal("should be no next report")
		}
	}

	time.Sleep(3 * time.Second)
	completeNextReport()
	checkNoQueuedReport()

	hookRun := time.Now()
	hookService.RunNextHook()

	if mockHook.invoked == nil ||
		mockHook.invoked.reportId != report.Id ||
		string(mockHook.invoked.data) != "hook-1" {
		t.Fatal("hook should be invoked")
	}

	mockHook.invoked = nil

	time.Sleep(2 * time.Second)

	if err := manager.CheckForStaleAuthorReports(); err != nil {
		t.Fatal(err)
	}

	completeNextReport()
	checkNoQueuedReport()

	hookService.RunNextHook()

	if mockHook.invoked != nil {
		t.Fatal("hook should not be invoked")
	}

	time.Sleep(time.Second)

	if err := manager.CheckForStaleAuthorReports(); err != nil {
		t.Fatal(err)
	}

	completeNextReport()
	checkNoQueuedReport()

	hookService.RunNextHook()

	if mockHook.invoked == nil ||
		mockHook.invoked.reportId != report.Id ||
		string(mockHook.invoked.data) != "hook-1" ||
		mockHook.invoked.lastRanAt.Sub(hookRun).Abs() > 100*time.Millisecond {
		t.Fatal("hook should be invoked")
	}

	mockHook.invoked = nil

	hookService.RunNextHook()
	if mockHook.invoked != nil {
		t.Fatal("hook should not be invoked")
	}
}
