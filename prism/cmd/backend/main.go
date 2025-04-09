package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"prism/prism/api"
	"prism/prism/cmd"
	"prism/prism/licensing"
	"prism/prism/openalex"
	"prism/prism/reports"
	"prism/prism/schema/migrations"
	"prism/prism/search"
	"prism/prism/services"
	"prism/prism/services/auth"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Config struct {
	PostgresUri  string `env:"DB_URI,notEmpty,required"`
	Logfile      string `env:"LOGFILE,notEmpty" envDefault:"prism_backend.log"`
	PrismLicense string `env:"PRISM_LICENSE,notEmpty,required"`

	Port int `env:"PORT" envDefault:"8000"`

	SearchableEntitiesData string `env:"SEARCHABLE_ENTITIES_DATA,notEmpty,required"`

	Keycloak struct {
		ServerUrl string `env:"SERVER_URL,notEmpty,required"`

		KeycloakAdminUsername string `env:"ADMIN_USERNAME,notEmpty,required"`
		KeycloakAdminPassword string `env:"ADMIN_PASSWORD,notEmpty,required"`

		PublicHostname  string `env:"PUBLIC_HOSTNAME,notEmpty,required"`
		PrivateHostname string `env:"PRIVATE_HOSTNAME,notEmpty,required"`

		SslInLogin bool `env:"SSL_IN_LOGIN" envDefault:"false"`
		Verbose    bool `env:"VERBOSE" envDefault:"false"`
	} `envPrefix:"KEYCLOAK_"`

	// This variable is directly loaded by the openai client library, it is just
	// listed here so that and error is raised if it's missing.
	OpenaiKey string `env:"OPENAI_API_KEY,notEmpty,required"`

	ResourceFolder string `env:"RESOURCE_FOLDER,notEmpty,required"`
}

func (c *Config) logfile() string {
	if c.Logfile == "" {
		return "prism_backend.log"
	}
	return c.Logfile
}

func (c *Config) port() int {
	if c.Port == 0 {
		return 8000
	}
	return c.Port
}

func loadSearchableEntities(entityPath string) []api.MatchedEntity {
	const entityNdbPath = "searchable_entities.ndb"
	if err := os.RemoveAll(entityNdbPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("error deleting existing ndb: %v", err)
	}

	time.Sleep(2 * time.Second)

	file, err := os.Open(entityPath)
	if err != nil {
		log.Fatalf("error opening searchable entities: %v", err)
	}

	var entities []api.MatchedEntity
	if err := json.NewDecoder(file).Decode(&entities); err != nil {
		log.Fatalf("error parsing searchable entities: %v", err)
	}

	log.Printf("loaded %d searchable entities", len(entities))

	return entities
}

func verifyResourceFolder(resourceFolder string) {
	info, err := os.Stat(resourceFolder)
	if os.IsNotExist(err) {
		log.Fatalf("resource folder '%s' does not exist", resourceFolder)
	}

	if !info.IsDir() {
		log.Fatalf("resource folder '%s' is not a directory", resourceFolder)
	}

	logoPath := filepath.Join(resourceFolder, "prism-logo.png")
	logo, err := os.ReadFile(logoPath)
	if err != nil {
		log.Fatalf("error reading logo: %v", err)
	}

	if len(logo) == 0 {
		log.Fatalf("logo is empty")
	}

	headerLogoPath := filepath.Join(resourceFolder, "prism-header-logo.png")
	headerLogo, err := os.ReadFile(headerLogoPath)
	if err != nil {
		log.Fatalf("error reading header logo: %v", err)
	}
	if len(headerLogo) == 0 {
		log.Fatalf("header logo is empty")
	}
}

func main() {
	cmd.LoadEnvFile()

	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	logFile, err := os.OpenFile(config.logfile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	cmd.InitLogging(logFile)

	openalex := openalex.NewRemoteKnowledgeBase()

	licensing, err := licensing.NewLicenseVerifier(config.PrismLicense)
	if err != nil {
		log.Fatalf("error initializing licensing: %v", err)
	}

	if err := search.SetLicenseKey(config.PrismLicense); err != nil {
		log.Fatalf("error activating license key: %v", err)
	}

	db := cmd.OpenDB(config.PostgresUri)
	migrations.RunMigrations(db)

	keycloakArgs := auth.KeycloakArgs{
		KeycloakServerUrl:     config.Keycloak.ServerUrl,
		KeycloakAdminUsername: config.Keycloak.KeycloakAdminUsername,
		KeycloakAdminPassword: config.Keycloak.KeycloakAdminPassword,
		PublicHostname:        config.Keycloak.PublicHostname,
		PrivateHostname:       config.Keycloak.PrivateHostname,
		SslLogin:              config.Keycloak.SslInLogin,
		Verbose:               config.Keycloak.Verbose,
	}

	userAuth, err := auth.NewKeycloakAuth("prism-user", keycloakArgs)
	if err != nil {
		log.Fatalf("error initializing keycloak user auth: %v", err)
	}

	verifyResourceFolder(config.ResourceFolder)

	reportManager := reports.NewManager(db)

	reportManager.StartReportUpdateCheck()
	defer reportManager.StopReportUpdateCheck()

	hooks := services.NewHookService(db, map[string]services.Hook{})

	hooks.RunHooks(30 * time.Minute)
	defer hooks.Stop()

	backend := services.NewBackend(
		services.NewReportService(reportManager, licensing, openalex, config.ResourceFolder),
		services.NewSearchService(openalex, loadSearchableEntities(config.SearchableEntitiesData)),
		services.NewAutoCompleteService(openalex),
		hooks,
		userAuth,
	)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},                                       // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Allow all HTTP methods
		AllowedHeaders:   []string{"*"},                                       // Allow all headers
		ExposedHeaders:   []string{"*"},                                       // Expose all headers
		AllowCredentials: true,                                                // Allow cookies/auth headers
		MaxAge:           300,                                                 // Cache preflight response for 5 minutes
	}))

	r.Mount("/api/v1", backend.Routes())

	slog.Info("starting server", "port", config.port())
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.port()), r)
	if err != nil {
		log.Fatalf("listen and serve returned error: %v", err.Error())
	}
}
