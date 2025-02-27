package api

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type baseClient struct {
	backend  *resty.Client
	keycloak *resty.Client
	realm    string
}

func (client *baseClient) Login(username, password string) error {
	type loginRes struct {
		AccessToken string `json:"access_token"`
	}

	res, err := client.keycloak.R().
		SetFormData(map[string]string{
			"grant_type":    "password",
			"username":      username,
			"password":      password,
			"client_id":     fmt.Sprintf("%s-login-client", client.realm),
			"lcient_secret": "",
			"scope":         "email profile openid",
		}).
		SetResult(&loginRes{}).
		Post(fmt.Sprintf("/realms/%s/protocol/openid-connect/token", client.realm))

	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if !res.IsSuccess() {
		return fmt.Errorf("login request returned status=%d body=%s", res.StatusCode(), res.String())
	}

	client.backend.SetAuthToken(res.Result().(*loginRes).AccessToken)

	return nil
}

type AdminClient struct {
	baseClient
}

func NewAdminClient(backendUrl, keycloakUrl string) *AdminClient {
	return &AdminClient{
		baseClient{
			backend:  resty.New().SetBaseURL(backendUrl).SetAuthScheme("Bearer"),
			keycloak: resty.New().SetBaseURL(keycloakUrl),
			realm:    "prism-admin",
		},
	}
}

func (client *AdminClient) CreateLicense(name string, expiration time.Time) (string, error) {
	res, err := client.backend.R().
		SetBody(CreateLicenseRequest{
			Name:       name,
			Expiration: expiration,
		}).
		SetResult(&CreateLicenseResponse{}).
		Post("/api/v1/license/create")
	if err != nil {
		return "", fmt.Errorf("create license request failed: %w", err)
	}

	if !res.IsSuccess() {
		return "", fmt.Errorf("create license returned status=%d, error=%v", res.StatusCode(), res.String())
	}

	return res.Result().(*CreateLicenseResponse).License, nil
}

type UserClient struct {
	baseClient
}

func NewUserClient(backendUrl, keycloakUrl string) *UserClient {
	return &UserClient{
		baseClient{
			backend:  resty.New().SetBaseURL(backendUrl).SetAuthScheme("Bearer"),
			keycloak: resty.New().SetBaseURL(keycloakUrl),
			realm:    "prism-user",
		},
	}
}

func (client *UserClient) CreateReport(report CreateAuthorReportRequest) (uuid.UUID, error) {
	res, err := client.backend.R().
		SetBody(report).
		SetResult(&CreateReportResponse{}).
		Post("/api/v1/report/author/create")
	if err != nil {
		return uuid.Nil, fmt.Errorf("create report request failed: %w", err)
	}

	if !res.IsSuccess() {
		return uuid.Nil, fmt.Errorf("create report returned status=%d, error=%v", res.StatusCode(), res.String())
	}

	return res.Result().(*CreateReportResponse).Id, nil
}

func (client *UserClient) GetReport(reportId uuid.UUID) (*Report, error) {
	res, err := client.backend.R().
		SetResult(&Report{}).
		SetPathParam("report_id", reportId.String()).
		Get("/api/v1/report/author/{report_id}")
	if err != nil {
		return nil, fmt.Errorf("get report request failed: %w", err)
	}

	if !res.IsSuccess() {
		return nil, fmt.Errorf("get report returned status=%d, error=%v", res.StatusCode(), res.String())
	}

	return res.Result().(*Report), nil
}

func (client *UserClient) WaitForReport(reportId uuid.UUID, timeout time.Duration) (*Report, error) {
	interval := time.Tick(time.Second)
	stop := time.After(timeout)
	for {
		select {
		case <-interval:
			report, err := client.GetReport(reportId)
			if err != nil {
				return nil, err
			}
			if report.Status == "complete" || report.Status == "failed" {
				return report, nil
			}
		case <-stop:
			return nil, fmt.Errorf("timeout reached before report completed")
		}
	}
}

func (client *UserClient) ActivateLicense(license string) error {
	res, err := client.backend.R().
		SetBody(ActivateLicenseRequest{
			License: license,
		}).
		Post("/api/v1/report/activate-license")
	if err != nil {
		return fmt.Errorf("activate license request failed: %w", err)
	}

	if !res.IsSuccess() {
		return fmt.Errorf("activate license returned status=%d, error=%v", res.StatusCode(), res.String())
	}

	return nil
}
