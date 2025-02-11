package services_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"prism/api"
	"prism/openalex"
	"prism/schema"
	"prism/search"
	"prism/services"
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

func createLicense(backend http.Handler, user string) (string, error) {
	req := api.CreateLicenseRequest{
		Name:       "test-license",
		Expiration: time.Now().UTC().Add(5 * time.Minute),
	}
	var res api.CreateLicenseResponse
	err := Post(backend, "/license/create", user, req, &res)
	return res.License, err
}

func useLicense(backend http.Handler, user, license string) error {
	req := api.AddLicenseUserRequest{
		License: license,
	}
	return Post(backend, "/report/use-license", user, req, nil)
}

func TestReportEndpoints(t *testing.T) {
	backend := createBackend(t)

	admin := newAdmin()
	user1, user2 := newUser(), newUser()

	checkListReports(t, backend, user1, []string{})
	checkListReports(t, backend, user2, []string{})

	license, err := createLicense(backend, admin)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := createReport(backend, user1, "report1"); err == nil || !strings.Contains(err.Error(), "user does not have registered license") {
		t.Fatalf("report should fail %v", err)
	}

	if err := useLicense(backend, user1, license); err != nil {
		t.Fatal(err)
	}
	if err := useLicense(backend, user2, license); err != nil {
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
}
