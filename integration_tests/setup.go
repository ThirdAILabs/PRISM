package tests

import (
	"testing"

	"prism/prism/api"
	"prism/prism/services/auth"

	"github.com/caarlos0/env/v11"
)

type testEnv struct {
	BackendUrl            string `env:"BACKEND_URL" envDefault:"http://localhost"`
	KeycloakUrl           string `env:"KEYCLOAK_URL" envDefault:"http://localhost/keycloak"`
	KeycloakAdminUsername string `env:"KEYCLOAK_ADMIN_USERNAME" envDefault:"kc-admin"`
	KeycloakAdminPassword string `env:"KEYCLOAK_ADMIN_PASSWORD" envDefault:"KC-admin-pwd@1"`
}

func setupTestEnv(t *testing.T) *api.PrismClient {
	const (
		username = "regular-user"
		password = "Regular-user-pwd@1"
	)

	var vars testEnv
	if err := env.Parse(&vars); err != nil {
		t.Fatalf("error parsing env: %v", err)
	}

	auth, err := auth.NewKeycloakAuth("prism-user", auth.KeycloakArgs{
		KeycloakServerUrl:     vars.KeycloakUrl,
		KeycloakAdminUsername: vars.KeycloakAdminUsername,
		KeycloakAdminPassword: vars.KeycloakAdminPassword,
	})
	if err != nil {
		t.Fatalf("error connecting to keycloak: %v", err)
	}

	adminToken, err := auth.AdminLogin(vars.KeycloakAdminUsername, vars.KeycloakAdminPassword)
	if err != nil {
		t.Fatalf("keycloak admin login failed: %v", err)
	}

	if err := auth.CreateUser(adminToken, username, username+"@mail.com", password); err != nil {
		t.Fatalf("error creating user: %v", err)
	}

	user := api.NewUserClient(vars.BackendUrl, vars.KeycloakUrl)
	if err := user.Login(username, password); err != nil {
		t.Fatal(err)
	}

	return user
}
