package tests

import (
	"testing"
	"time"

	"prism/prism/api"
	"prism/prism/services/auth"
)

const (
	backendUrl = "http://localhost"

	keycloakUrl           = "http://localhost/keycloak"
	keycloakAdminUsername = "kc-admin"
	keycloakAdminPassword = "KC-admin-pwd@1"

	regularUserUsername = "regular-user"
	regularUserPassword = "Regular-user-pwd@1"

	adminUserUsername = "admin-user"
	adminUserPassword = "Admin-user-pwd@1"
)

func setupKeycloakUsers(t *testing.T) {
	users := []struct {
		realm    string
		username string
		password string
	}{
		{"prism-user", regularUserUsername, regularUserPassword},
		{"prism-admin", adminUserUsername, adminUserPassword},
	}

	for _, user := range users {
		auth, err := auth.NewKeycloakAuth(user.realm, auth.KeycloakArgs{
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

		if err := auth.CreateUser(adminToken, user.username, user.username+"@mail.com", user.password); err != nil {
			t.Fatalf("error creating user: %v", err)
		}
	}
}

func setupTestEnv(t *testing.T) (*api.UserClient, *api.AdminClient) {
	setupKeycloakUsers(t)

	user := api.NewUserClient(backendUrl, keycloakUrl)
	if err := user.Login(regularUserUsername, regularUserPassword); err != nil {
		t.Fatal(err)
	}

	admin := api.NewAdminClient(backendUrl, keycloakUrl)
	if err := admin.Login(adminUserUsername, adminUserPassword); err != nil {
		t.Fatal(err)
	}

	license, err := admin.CreateLicense("test license", time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	if err := user.ActivateLicense(license); err != nil {
		t.Fatal(err)
	}

	return user, admin
}
