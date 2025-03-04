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
)

type KeycloakAuth struct {
	keycloak *gocloak.GoCloak
	realm    string
	logger   *slog.Logger
}

type KeycloakArgs struct {
	KeycloakServerUrl string `yaml:"keycloak_server_url"`

	KeycloakAdminUsername string `yaml:"keycloak_admin_username"`
	KeycloakAdminPassword string `yaml:"keycloak_admin_password"`

	PublicHostname  string `yaml:"public_hostname"`
	PrivateHostname string `yaml:"private_hostname"`

	SslLogin bool `yaml:"ssl_login"`

	Verbose bool
}

func NewKeycloakAuth(realm string, args KeycloakArgs) (*KeycloakAuth, error) {
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

	auth := &KeycloakAuth{keycloak: client, realm: realm, logger: slog.With("logger", "keycloak", "realm", realm)}

	adminToken, err := auth.AdminLogin(args.KeycloakAdminUsername, args.KeycloakAdminPassword)
	if err != nil {
		auth.logger.Error("admin login failed", "error", err)
		return nil, err
	}
	auth.logger.Info("admin login successful")

	if err := auth.createRealm(adminToken); err != nil {
		auth.logger.Error("realm creation failed", "error", err)
		return nil, err
	}
	auth.logger.Info("realm creation successful")

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

	if err := auth.createClient(adminToken, redirectUrls, args.KeycloakServerUrl); err != nil {
		auth.logger.Error("client creation failed", "error", err)
		return nil, err
	}
	auth.logger.Info("client creation successful")

	return auth, nil
}

func isConflict(err error) bool {
	apiErr, ok := err.(*gocloak.APIError)
	// Keycloak returns 409 if user/realm etc already exists when creating it.
	return ok && apiErr.Code == http.StatusConflict
}

func (auth *KeycloakAuth) AdminLogin(adminUsername, adminPassword string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// The "master" realm is the default admin realm in Keycloak.
	adminToken, err := auth.keycloak.LoginAdmin(ctx, adminUsername, adminPassword, "master")
	if err != nil {
		return "", fmt.Errorf("error during keycloak admin login: %w", err)
	}
	return adminToken.AccessToken, nil
}

func (auth *KeycloakAuth) createRealm(adminToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverInfo, err := auth.keycloak.GetServerInfo(ctx, adminToken)
	if err != nil {
		return fmt.Errorf("error getting keycloak server info: %w", err)
	}

	args := gocloak.RealmRepresentation{
		Realm:                        &auth.realm,
		Enabled:                      gocloak.BoolP(true),
		IdentityProviders:            &[]interface{}{},
		DefaultRoles:                 &[]string{"user"},
		RegistrationAllowed:          gocloak.BoolP(true),
		ResetPasswordAllowed:         gocloak.BoolP(true),
		AccessCodeLifespan:           gocloak.IntP(1500),
		PasswordPolicy:               gocloak.StringP("length(8) and digits(1) and lowerCase(1) and upperCase(1) and specialChars(1)"),
		BruteForceProtected:          gocloak.BoolP(true),
		MaxFailureWaitSeconds:        gocloak.IntP(900),
		MinimumQuickLoginWaitSeconds: gocloak.IntP(60),
		WaitIncrementSeconds:         gocloak.IntP(60),
		QuickLoginCheckMilliSeconds:  gocloak.Int64P(int64(1000)),
		MaxDeltaTimeSeconds:          gocloak.IntP(43200),
		FailureFactor:                gocloak.IntP(30),
		SMTPServer: &map[string]string{
			"host":     "smtp.sendgrid.net",
			"port":     "465",
			"from":     "platform@thirdai.com",
			"replyTo":  "platform@thirdai.com",
			"ssl":      "true",
			"starttls": "true",
			"auth":     "true",
			"user":     "apikey",
			"password": "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",
		},
	}

	if serverInfo.Themes != nil {
		for _, theme := range serverInfo.Themes.Login {
			if theme.Name == "custom-theme" {
				args.LoginTheme = gocloak.StringP("custom-theme")
				args.AccountTheme = gocloak.StringP("custom-theme")
				args.AdminTheme = gocloak.StringP("custom-theme")
				args.EmailTheme = gocloak.StringP("custom-theme")
				args.DisplayName = &auth.realm
				args.DisplayNameHTML = gocloak.StringP("<div class='kc-logo-text'><span>Keycloak</span></div>")
			}
		}
	}

	_, err = auth.keycloak.CreateRealm(ctx, adminToken, args)
	if err != nil {
		if isConflict(err) {
			auth.logger.Info("realm has already been created")
			return nil // Ok if realm already exists
		}
		return fmt.Errorf("error creating realm: %w", err)
	}
	return nil
}

func (auth *KeycloakAuth) createClient(adminToken string, redirectUrls []string, rootUrl string) error {
	clientName := auth.realm + "-login-client"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clients, err := auth.keycloak.GetClients(ctx, adminToken, auth.realm, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		return fmt.Errorf("error listing existing clients for realm: %w", err)
	}
	if len(clients) == 1 {
		auth.logger.Info("client already exists for realm", "client", clientName)
		return nil
	}

	_, err = auth.keycloak.CreateClient(ctx, adminToken, auth.realm, gocloak.Client{
		ClientID:                  &clientName,
		Enabled:                   gocloak.BoolP(true),
		PublicClient:              gocloak.BoolP(true),  // Public client that doesn't require a secret for authentication.
		RedirectURIs:              &redirectUrls,        // URIs where the client will redirect after authentication.
		RootURL:                   &rootUrl,             // Root URL for the client application.
		BaseURL:                   gocloak.StringP("/"), // Base URL for the client application.
		DirectAccessGrantsEnabled: gocloak.BoolP(true),  // Direct grants like password flow are enabled, this allows for login with username and password via keycloak.
		ServiceAccountsEnabled:    gocloak.BoolP(false), // Service accounts are disabled.
		StandardFlowEnabled:       gocloak.BoolP(true),  // Standard authorization code flow is enabled.
		ImplicitFlowEnabled:       gocloak.BoolP(false), // Implicit flow is disabled.
		FullScopeAllowed:          gocloak.BoolP(false), // Limit access to only allowed scopes.
		DefaultClientScopes:       &[]string{"profile", "email", "openid", "roles"},
		OptionalClientScopes:      &[]string{"offline_access", "microprofile-jwt"},
		ProtocolMappers: &[]gocloak.ProtocolMapperRepresentation{
			{
				Name:            gocloak.StringP("auidience resolve"),            // Protocol mappers adjust tokens for clients.
				Protocol:        gocloak.StringP("openid-connect"),               // The OIDC protocol used for authentication.
				ProtocolMapper:  gocloak.StringP("oidc-audience-resolve-mapper"), // Mapper to add audience claim in tokens.
				ConsentRequired: gocloak.BoolP(false),
				Config:          &map[string]string{},
			},
		},
		WebOrigins: &redirectUrls,
	})
	if err != nil {
		if isConflict(err) {
			auth.logger.Info("client has already been created for realm", "client", clientName)
			return nil
		}
		return fmt.Errorf("error creating realm client: %w", err)
	}
	return nil
}

func getToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")

	if len(header) > 7 && strings.ToLower(header[:7]) == "bearer " {
		return header[7:], nil
	}

	return "", fmt.Errorf("missing or invalid authorization header")
}

func (auth *KeycloakAuth) VerifyToken(token string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userInfo, err := auth.keycloak.GetUserInfo(ctx, token, auth.realm)
	if err != nil {
		auth.logger.Error("unable to verify token with keycloak", "error", err)
		return uuid.Nil, fmt.Errorf("unable to verify access token: %w", err)
	}

	if userInfo.Sub == nil {
		auth.logger.Error("missing user identifier in keycloak response")
		return uuid.Nil, fmt.Errorf("missing user identifier in keycloak response")
	}

	userId, err := uuid.Parse(*userInfo.Sub)
	if err != nil {
		auth.logger.Error("unable to parse user id from keycloak", "id", *userInfo.Sub, "error", err)
		return uuid.Nil, fmt.Errorf("invalid uuid '%v' returned from keycloak: %v", *userInfo.Sub, err)
	}

	return userId, nil
}

// This is just for the purpose of integration tests. It is used to create users
// that can be used for the tests. It is not used in the backend
func (auth *KeycloakAuth) CreateUser(adminToken, username, email, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	auth.logger.Info("creating user", "user", username)

	userId, err := auth.keycloak.CreateUser(ctx, adminToken, auth.realm, gocloak.User{
		Username:      gocloak.StringP(username),
		Email:         gocloak.StringP(email),
		FirstName:     gocloak.StringP(username),
		LastName:      gocloak.StringP(username),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
	})
	if err != nil {
		if isConflict(err) {
			auth.logger.Info("user already exists", "user", username)
			return nil
		}
		auth.logger.Error("error creating user", "user", username, "error", err)
		return fmt.Errorf("error creating user '%s' in realm '%s': %w", username, auth.realm, err)
	}

	if err := auth.keycloak.SetPassword(ctx, adminToken, userId, auth.realm, password, false); err != nil {
		auth.logger.Error("error creating user", "user", username, "error", err)
		return fmt.Errorf("error creating user '%s' in realm '%s': %w", username, auth.realm, err)
	}

	auth.logger.Info("user created successfully", "user", username)

	return nil
}
