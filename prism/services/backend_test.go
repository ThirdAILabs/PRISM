package services_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"prism/api"
	"prism/openalex"
	"prism/schema"
	"prism/search"
	"prism/services"
	"prism/services/licensing"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	const licensePath = "../../.test_license/thirdai.license"
	if err := search.SetLicensePath(licensePath); err != nil {
		panic(err)
	}
}

const (
	userPrefix  = "user"
	adminPrefix = "admin"
)

func newUser() string {
	return userPrefix + uuid.NewString()
}

func newAdmin() string {
	return adminPrefix + uuid.NewString()
}

type MockTokenVerifier struct {
	prefix string
}

func (m *MockTokenVerifier) VerifyToken(token string) (uuid.UUID, error) {
	if !strings.HasPrefix(token, m.prefix) {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	id, err := uuid.Parse(strings.TrimPrefix(token, m.prefix))
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func createBackend(t *testing.T) http.Handler {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(
		&schema.Report{}, &schema.ReportContent{}, &schema.License{},
		&schema.LicenseUser{}, &schema.LicenseUsage{},
	); err != nil {
		t.Fatal(err)
	}

	ndbPath := filepath.Join(t.TempDir(), "entity.ndb")
	ndb, err := search.NewNeuralDB(ndbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ndb.Free()
	})

	entities := []string{"abc university", "institute of xyz", "123 org"}
	if err := ndb.Insert("doc", "id", entities, nil, nil); err != nil {
		t.Fatal(err)
	}

	backend := services.NewBackend(
		db, openalex.NewRemoteKnowledgeBase(), ndb, &MockTokenVerifier{prefix: userPrefix}, &MockTokenVerifier{prefix: adminPrefix},
	)

	return backend.Routes()
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
		report.StartYear != 10 ||
		report.EndYear != 12 ||
		report.Status != "queued" {
		t.Fatal("invalid reports returned")
	}
}

func checkListReports(t *testing.T, backend http.Handler, user string, expected []string) {
	var reports []api.Report
	if err := Get(backend, "/report/list", user, &reports); err != nil {
		t.Fatal(err)
	}

	slices.SortFunc(reports, func(a, b api.Report) int {
		if a.AuthorId < b.AuthorId {
			return -1
		}
		if a.AuthorId > b.AuthorId {
			return 1
		}
		return 0
	})

	slices.Sort(expected)

	if len(reports) != len(expected) {
		t.Fatal("incorrect number of reports returned")
	}
	for i := range expected {
		compareReport(t, reports[i], expected[i])
	}
}

func getReport(backend http.Handler, user string, id uuid.UUID) (api.Report, error) {
	var res api.Report
	err := Get(backend, "/report/"+id.String(), user, &res)
	return res, err
}

func createReport(backend http.Handler, user, name string) (api.CreateReportResponse, error) {
	req := api.CreateReportRequest{
		AuthorId:   name + "-id",
		AuthorName: name + "-name",
		Source:     api.OpenAlexSource,
		StartYear:  10,
		EndYear:    12,
	}

	var res api.CreateReportResponse
	err := Post(backend, "/report/create", user, req, &res)
	return res, err
}

func createLicense(backend http.Handler, name, user string) (api.CreateLicenseResponse, error) {
	return createLicenseWithExp(backend, name, user, time.Now().UTC().Add(5*time.Minute))
}

func createLicenseWithExp(backend http.Handler, name, user string, exp time.Time) (api.CreateLicenseResponse, error) {
	req := api.CreateLicenseRequest{
		Name:       name,
		Expiration: exp,
	}
	var res api.CreateLicenseResponse
	err := Post(backend, "/license/create", user, req, &res)
	return res, err
}

func listLicenses(backend http.Handler, user string) ([]api.License, error) {
	var res []api.License
	err := Get(backend, "/license/list", user, &res)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(res, func(a, b api.License) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})

	return res, nil
}

func activateLicense(backend http.Handler, user, license string) error {
	req := api.AddLicenseUserRequest{
		License: license,
	}
	return Post(backend, "/report/activate-license", user, req, nil)
}

func deactivateLicense(backend http.Handler, user string, id uuid.UUID) error {
	return Delete(backend, "/license/"+id.String(), user)
}

func TestLicenseEndpoints(t *testing.T) {
	backend := createBackend(t)

	admin := newAdmin()
	user1, user2 := newUser(), newUser()

	if _, err := createLicense(backend, "license-0", user1); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users cannot create licenses")
	}

	license1, err := createLicense(backend, "xyz", admin)
	if err != nil {
		t.Fatal(err)
	}

	license2, err := createLicense(backend, "abc", admin)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := listLicenses(backend, user1); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users cannot list licenses")
	}

	licenses, err := listLicenses(backend, admin)
	if err != nil {
		t.Fatal(err)
	}

	if len(licenses) != 2 ||
		licenses[0].Name != "abc" || licenses[0].Id != license2.Id ||
		licenses[1].Name != "xyz" || licenses[1].Id != license1.Id {
		t.Fatalf("incorrect licenses: %v", licenses)
	}

	if err := activateLicense(backend, user1, license1.License); err != nil {
		t.Fatal(err)
	}

	if err := deactivateLicense(backend, admin, license2.Id); err != nil {
		t.Fatal(err)
	}

	if err := activateLicense(backend, user2, license2.License); err == nil || !strings.Contains(err.Error(), "license is deactivated") {
		t.Fatal("cannot use deactivated license")
	}

	if _, err := createReport(backend, user1, "report1"); err != nil {
		t.Fatal(err)
	}

	if err := deactivateLicense(backend, user2, license1.Id); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("users cannot deactivate licenses")
	}

	if _, err := createReport(backend, user1, "report1"); err != nil {
		t.Fatal(err)
	}

	if err := deactivateLicense(backend, admin, license1.Id); err != nil {
		t.Fatal(err)
	}

	if _, err := createReport(backend, user1, "report1"); err == nil || !strings.Contains(err.Error(), "license is deactivated") {
		t.Fatal("cannot use deactivated license")
	}
}

func TestLicenseExpiration(t *testing.T) {
	backend := createBackend(t)

	admin := newAdmin()
	user1, user2 := newUser(), newUser()

	license, err := createLicenseWithExp(backend, "xyz", admin, time.Now().UTC().Add(500*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	if err := activateLicense(backend, user1, license.License); err != nil {
		t.Fatal(err)
	}

	if _, err := createReport(backend, user1, "report1"); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	if _, err := createReport(backend, user1, "report1"); err == nil || !strings.Contains(err.Error(), "expired license") {
		t.Fatal(err)
	}

	if err := activateLicense(backend, user2, license.License); err == nil || !strings.Contains(err.Error(), "expired license") {
		t.Fatal(err)
	}
}

func TestLicenseInvalidLicense(t *testing.T) {
	backend := createBackend(t)

	admin := newAdmin()
	user := newUser()

	license, err := createLicense(backend, "xyz", admin)
	if err != nil {
		t.Fatal(err)
	}

	badSecretLicense := licensing.LicensePayload{Id: license.Id, Secret: []byte("invalid secret")}
	badSecretKey, err := badSecretLicense.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if err := activateLicense(backend, user, badSecretKey); !strings.Contains(err.Error(), "invalid license") {
		t.Fatal(err)
	}

	badIdLicense := licensing.LicensePayload{Id: uuid.New(), Secret: []byte("invalid secret")}
	badIdKey, err := badIdLicense.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if err := activateLicense(backend, user, badIdKey); !strings.Contains(err.Error(), "license not found") {
		t.Fatal(err)
	}
}

func TestReportEndpoints(t *testing.T) {
	backend := createBackend(t)

	admin := newAdmin()
	user1, user2 := newUser(), newUser()

	checkListReports(t, backend, user1, []string{})
	checkListReports(t, backend, user2, []string{})

	license, err := createLicense(backend, "test-license", admin)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := createReport(backend, user1, "report1"); err == nil || !strings.Contains(err.Error(), "user does not have registered license") {
		t.Fatalf("report should fail %v", err)
	}

	if err := activateLicense(backend, user1, license.License); err != nil {
		t.Fatal(err)
	}
	if err := activateLicense(backend, user2, license.License); err != nil {
		t.Fatal(err)
	}

	report, err := createReport(backend, user1, "report1")
	if err != nil {
		t.Fatal(err)
	}

	checkListReports(t, backend, user1, []string{"report1"})
	checkListReports(t, backend, user2, []string{})

	if _, err := getReport(backend, user2, report.Id); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("should be unauthorized: %v", err)
	}

	reportData, err := getReport(backend, user1, report.Id)
	if err != nil {
		t.Fatal(err)
	}

	compareReport(t, reportData, "report1")

	if _, err := createReport(backend, user2, "report2"); err != nil {
		t.Fatal(err)
	}

	if _, err := createReport(backend, user1, "report3"); err != nil {
		t.Fatal(err)
	}

	checkListReports(t, backend, user1, []string{"report1", "report3"})
	checkListReports(t, backend, user2, []string{"report2"})
}

func TestAutocompleteAuthor(t *testing.T) {
	backend := createBackend(t)

	user := newUser()

	var results []api.Author
	err := mockRequest(backend, "GET", "/autocomplete/author?query="+url.QueryEscape("anshumali shriva"), user, nil, &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("should have some results")
	}

	for _, res := range results {
		if !strings.HasPrefix(res.AuthorId, "https://openalex.org/") ||
			!strings.EqualFold(res.AuthorName, "Anshumali Shrivastava") ||
			res.Source != "openalex" {
			t.Fatal("invalid result")
		}
	}
}

func TestAutocompleteInstution(t *testing.T) {
	backend := createBackend(t)

	user := newUser()

	var results []api.Institution
	err := mockRequest(backend, "GET", "/autocomplete/institution?query="+url.QueryEscape("rice univer"), user, nil, &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("should have some results")
	}

	for _, res := range results {
		if !strings.HasPrefix(res.InstitutionId, "https://openalex.org/") ||
			strings.EqualFold(res.InstitutionName, "Rice University, USA") {
			t.Fatal("invalid result")
		}
	}
}

func TestMatchEntities(t *testing.T) {
	backend := createBackend(t)

	user := newUser()

	var results api.MatchEntitiesResponse
	err := mockRequest(backend, "GET", "/search/match-entities?query="+url.QueryEscape("xyz"), user, nil, &results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results.Entities) != 1 || results.Entities[0] != "institute of xyz" {
		t.Fatalf("incorrect entities matched: %v", results.Entities)
	}
}

func TestSearchOpenalexAuthors(t *testing.T) {
	backend := createBackend(t)

	user := newUser()

	authorName := "anshumali shrivastava"
	insitutionId := "https://openalex.org/I74775410"

	url := fmt.Sprintf("/search/regular?author_name=%s&institution_id=%s", url.QueryEscape(authorName), url.QueryEscape(insitutionId))
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
}

func TestSearchGoogleScholarAuthors(t *testing.T) {
	backend := createBackend(t)

	user := newUser()

	authorName := "anshumali shrivastava"

	url := fmt.Sprintf("/search/advanced?query=%s", url.QueryEscape(authorName))
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
	backend := createBackend(t)

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

	url1 := fmt.Sprintf("/search/advanced?query=%s", url.QueryEscape(authorName))
	var results1 api.GScholarSearchResults
	if err := mockRequest(backend, "GET", url1, user, nil, &results1); err != nil {
		t.Fatal(err)
	}

	checkQuery(results1.Authors)

	url2 := fmt.Sprintf("/search/advanced?query=%s&cursor=%s", url.QueryEscape(authorName), results1.Cursor)
	var results2 api.GScholarSearchResults
	if err := mockRequest(backend, "GET", url2, user, nil, &results2); err != nil {
		t.Fatal(err)
	}

	checkQuery(results2.Authors)
}
