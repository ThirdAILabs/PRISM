package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	userIdContextKey = "user_id"
)

func isConflict(err error) bool {
	apiErr, ok := err.(*gocloak.APIError)
	// Keycloak returns 409 if user/realm etc already exists when creating it.
	return ok && apiErr.Code == http.StatusConflict
}

func pArg[T any](value T) *T {
	p := new(T)
	*p = value
	return p
}

var boolArg = pArg[bool]
var intArg = pArg[int]
var strArg = pArg[string]

func adminLogin(client *gocloak.GoCloak, adminUsername, adminPassword string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// The "master" realm is the default admin realm in Keycloak.
	adminToken, err := client.LoginAdmin(ctx, adminUsername, adminPassword, "master")
	if err != nil {
		return "", fmt.Errorf("error during keycloak admin login: %w", err)
	}
	return adminToken.AccessToken, nil
}

func getUserID(ctx context.Context, client *gocloak.GoCloak, adminToken, username string) (*string, error) {
	users, err := client.GetUsers(ctx, adminToken, "master", gocloak.GetUsersParams{
		Username: &username,
		Max:      intArg(1),
		Exact:    boolArg(true),
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving user id: %w", err)
	}
	if len(users) == 1 {
		return users[0].ID, nil
	}
	return nil, nil
}

func createAdminIfNotExists(client *gocloak.GoCloak, adminToken, username, email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	existingUserId, err := getUserID(ctx, client, adminToken, username)
	if err != nil {
		return "", fmt.Errorf("error checking for existing admin : %w", err)
	}
	if existingUserId != nil {
		slog.Info("KEYCLOAK: admin user has already been created")
		return *existingUserId, nil
	}

	userId, err := client.CreateUser(ctx, adminToken, "master", gocloak.User{
		Username:      &username,
		Email:         &email,
		Enabled:       boolArg(true),
		EmailVerified: boolArg(true),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      strArg("password"),
				Value:     &password,
				Temporary: boolArg(false),
			},
		},
	})

	if err != nil {
		if isConflict(err) {
			userId, err := getUserID(ctx, client, adminToken, username)
			slog.Info("KEYCLOAK: admin user has already been created")
			if err != nil {
				return "", fmt.Errorf("error retrieving existing admin after conflict creating admin: %w", err)
			}
			if userId == nil {
				return "", fmt.Errorf("no user found after conflict creating admin")
			}
			return *userId, nil
		}
		return "", fmt.Errorf("error creating new admin: %w", err)
	}

	return userId, nil
}

func assignAdminRole(client *gocloak.GoCloak, adminToken, userId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	roles, err := client.GetRealmRoles(ctx, adminToken, "master", gocloak.GetRoleParams{})
	if err != nil {
		return fmt.Errorf("error getting keycloak roles: %w", err)
	}
	for _, role := range roles {
		if *role.Name == "admin" {
			err := client.AddRealmRoleToUser(ctx, adminToken, "master", userId, []gocloak.Role{*role})
			if err != nil {
				return fmt.Errorf("error assigning admin role: %w", err)
			}
		}
	}
	return nil
}

func createRealm(client *gocloak.GoCloak, adminToken, realmName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serverInfo, err := client.GetServerInfo(ctx, adminToken)
	if err != nil {
		return fmt.Errorf("error getting keycloak server info: %w", err)
	}

	args := gocloak.RealmRepresentation{
		Realm:                &realmName,
		Enabled:              boolArg(true),
		IdentityProviders:    &[]interface{}{},
		DefaultRoles:         &[]string{"user"},
		RegistrationAllowed:  boolArg(true),
		ResetPasswordAllowed: boolArg(true),
		AccessCodeLifespan:   intArg(1500),
	}

	if serverInfo.Themes != nil {
		for _, theme := range serverInfo.Themes.Login {
			if theme.Name == "custom-theme" {
				args.LoginTheme = strArg("custom-theme")
				args.AccountTheme = strArg("custom-theme")
				args.AdminTheme = strArg("custom-theme")
				args.EmailTheme = strArg("custom-theme")
				args.DisplayName = &realmName
				args.DisplayNameHTML = strArg("<div class='kc-logo-text'><span>Keycloak</span></div>")
			}
		}
	}

	_, err = client.CreateRealm(ctx, adminToken, args)
	if err != nil {
		if isConflict(err) {
			slog.Info(fmt.Sprintf("KEYCLOAK: realm '%v' has already been created", realmName))
			return nil // Ok if realm already exists
		}
		return fmt.Errorf("error creating realm: %w", err)
	}
	return nil
}

func createClient(client *gocloak.GoCloak, adminToken, realm string, redirectUrls []string, rootUrl string) error {
	clientName := realm + "-login"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clients, err := client.GetClients(ctx, adminToken, realm, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		return fmt.Errorf("error listing existing clients for realm: %w", err)
	}
	if len(clients) == 1 {
		slog.Info(fmt.Sprintf("KEYCLOAK: client '%v' already exists for realm '%v'", clientName, realm))
		return nil
	}

	_, err = client.CreateClient(ctx, adminToken, realm, gocloak.Client{
		ClientID:                  &clientName,
		Enabled:                   boolArg(true),
		PublicClient:              boolArg(true),    // Public client that doesn't require a secret for authentication.
		RedirectURIs:              &redirectUrls,    // URIs where the client will redirect after authentication.
		RootURL:                   &rootUrl,         // Root URL for the client application.
		BaseURL:                   strArg("/login"), // Base URL for the client application.
		DirectAccessGrantsEnabled: boolArg(false),   // Direct grants like password flow are disabled.
		ServiceAccountsEnabled:    boolArg(false),   // Service accounts are disabled.
		StandardFlowEnabled:       boolArg(true),    // Standard authorization code flow is enabled.
		ImplicitFlowEnabled:       boolArg(false),   // Implicit flow is disabled.
		FullScopeAllowed:          boolArg(false),   // Limit access to only allowed scopes.
		DefaultClientScopes:       &[]string{"profile", "email", "openid", "roles"},
		OptionalClientScopes:      &[]string{"offline_access", "microprofile-jwt"},
		ProtocolMappers: &[]gocloak.ProtocolMapperRepresentation{
			{
				Name:            strArg("auidience resolve"),            // Protocol mappers adjust tokens for clients.
				Protocol:        strArg("openid-connect"),               // The OIDC protocol used for authentication.
				ProtocolMapper:  strArg("oidc-audience-resolve-mapper"), // Mapper to add audience claim in tokens.
				ConsentRequired: boolArg(false),
				Config:          &map[string]string{},
			},
		},
		WebOrigins: &redirectUrls,
	})
	if err != nil {
		if isConflict(err) {
			slog.Info(fmt.Sprintf("KEYCLOAK: client '%v' has already been created for realm '%v'", clientName, realm))
			return nil
		}
		return fmt.Errorf("error creating realm client: %w", err)
	}
	return nil
}

type KeycloakArgs struct {
	KeycloakServerUrl string

	KeycloakAdminUsername string
	KeycloakAdminPassword string

	AdminUsername string
	AdminEmail    string
	AdminPassword string

	PublicHostname  string
	PrivateHostname string

	SslLogin bool

	Verbose bool
}

type KeycloakAuth struct {
	keycloak *gocloak.GoCloak
	realm    string
}

func New(db *gorm.DB, realm string, args KeycloakArgs) (*KeycloakAuth, error) {
	client := gocloak.NewClient(args.KeycloakServerUrl)
	restyClient := client.RestyClient()
	restyClient.SetDebug(args.Verbose) // Adds logging for every request

	if args.SslLogin {
		cert, err := tls.LoadX509KeyPair("/model_bazaar/certs/traefik.crt", "/model_bazaar/certs/traefik.key")
		if err != nil {
			return nil, fmt.Errorf("error loading cert: %w", err)
		}
		restyClient.SetCertificates(cert)
	} else {
		restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	adminToken, err := adminLogin(client, args.KeycloakAdminUsername, args.KeycloakAdminPassword)
	if err != nil {
		slog.Error("KEYCLOAK: admin login failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: admin login successful")

	userId, err := createAdminIfNotExists(client, adminToken, args.AdminUsername, args.AdminEmail, args.AdminPassword)
	if err != nil {
		slog.Error("KEYCLOAK: new admin creation failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: new admin creation successful")

	err = assignAdminRole(client, adminToken, userId)
	if err != nil {
		slog.Error("KEYCLOAK: admin role assignment failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: admin role assignment successful")

	err = createRealm(client, adminToken, realm)
	if err != nil {
		slog.Error("KEYCLOAK: realm creation failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: realm creation successful")

	redirectUrls := []string{
		fmt.Sprintf("http://%v/*", args.PublicHostname),
		fmt.Sprintf("https://%v/*", args.PublicHostname),
		fmt.Sprintf("http://%v:80/*", args.PublicHostname),
		fmt.Sprintf("https://%v:80/*", args.PublicHostname),
		fmt.Sprintf("http://%v/*", args.PrivateHostname),
		fmt.Sprintf("https://%v/*", args.PrivateHostname),
		fmt.Sprintf("http://%v:80/*", args.PrivateHostname),
		fmt.Sprintf("https://%v:80/*", args.PrivateHostname),
		"http://localhost/*",
		"https://localhost/*",
		"http://localhost:80/*",
		"https://localhost:80/*",
		"http://127.0.0.1/*",
		"https://127.0.0.1/*",
		"*",
	}
	err = createClient(client, adminToken, realm, redirectUrls, args.KeycloakServerUrl)
	if err != nil {
		slog.Error("KEYCLOAK: client creation failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: client creation successful")

	return &KeycloakAuth{keycloak: client, realm: realm}, nil
}

func getToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")

	if len(header) > 7 && strings.ToLower(header[:7]) == "bearer " {
		return header[7:], nil
	}

	return "", fmt.Errorf("missing or invalid authorization header")
}

func (auth *KeycloakAuth) VerifyToken(token string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userInfo, err := auth.keycloak.GetUserInfo(ctx, token, auth.realm)
	if err != nil {
		slog.Error("unable to verify token with keycloak", "error", err)
		return uuid.Nil, fmt.Errorf("unable to verify access token: %w", err)
	}

	if userInfo.Sub == nil {
		slog.Error("missing user identifier in keycloak response")
		return uuid.Nil, fmt.Errorf("missing user identifier in keycloak response")
	}

	userId, err := uuid.Parse(*userInfo.Sub)
	if err != nil {
		slog.Error("unable to parse user id from keycloak", "id", *userInfo.Sub, "error", err)
		return uuid.Nil, fmt.Errorf("invalid uuid '%v' returned from keycloak: %v", *userInfo.Sub, err)
	}

	return userId, nil
}
