package tests

import (
	"testing"

	"prism/prism/api"
	"prism/prism/services/auth"
)

const (
	backendUrl = "http://localhost"

	keycloakUrl           = "http://localhost/keycloak"
	keycloakAdminUsername = "kc-admin"
	keycloakAdminPassword = "KC-admin-pwd@1"

	username = "regular-user"
	password = "Regular-user-pwd@1"
)

func setupKeycloakUsers(t *testing.T) {
	auth, err := auth.NewKeycloakAuth("prism-user", auth.KeycloakArgs{
		KeycloakServerUrl:     keycloakUrl,
		KeycloakAdminUsername: keycloakAdminUsername,
		KeycloakAdminPassword: keycloakAdminPassword,
		PublicHostname:        "",
		PrivateHostname:       "",
	})
	if err != nil {
		t.Fatalf("error connecting to keycloak: %v", err)
	}

	adminToken, err := auth.AdminLogin(keycloakAdminUsername, keycloakAdminPassword)
	if err != nil {
		t.Fatalf("keycloak admin login failed: %v", err)
	}

	if err := auth.CreateUser(adminToken, username, username+"@mail.com", password); err != nil {
		t.Fatalf("error creating user: %v", err)
	}

}

func setupTestEnv(t *testing.T) *api.PrismClient {
	setupKeycloakUsers(t)

	user := api.NewUserClient(backendUrl, keycloakUrl)
	if err := user.Login(username, password); err != nil {
		t.Fatal(err)
	}

	return user
}
