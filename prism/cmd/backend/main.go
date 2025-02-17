package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"prism/prism/cmd"
	"prism/prism/openalex"
	"prism/prism/search"
	"prism/prism/services"
	"prism/prism/services/auth"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Config struct {
	PostgresUri   string `yaml:"postgres_uri"`
	EntityNdbPath string `yaml:"entity_ndb"`
	Keycloak      auth.KeycloakArgs
	Port          int
	Logfile       string
	NdbLicense    string `yaml:"ndb_license"`
}

func (c *Config) logfile() string {
	if c.Logfile == "" {
		return "prism_backend.log"
	}
	return c.Logfile
}

func (c *Config) port() int {
	if c.Port == 0 {
		return 3000
	}
	return c.Port
}

func main() {
	var config Config
	cmd.LoadConfig(&config)

	logFile, err := os.OpenFile(config.logfile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	cmd.InitLogging(logFile)

	openalex := openalex.NewRemoteKnowledgeBase()

	if strings.HasPrefix(config.NdbLicense, "file ") {
		err := search.SetLicensePath(strings.TrimPrefix(config.NdbLicense, "file "))
		if err != nil {
			log.Fatalf("error activating license at path '%s': %v", config.NdbLicense, err)
		}
	} else {
		err := search.SetLicenseKey(config.NdbLicense)
		if err != nil {
			log.Fatalf("error activating license: %v", err)
		}
	}

	entityNdb, err := search.NewNeuralDB(config.EntityNdbPath)
	if err != nil {
		log.Fatalf("unable to load entity ndb: %v", err)
	}

	db := cmd.InitDb(config.PostgresUri)

	userAuth, err := auth.NewKeycloakAuth("prism-user", config.Keycloak)
	if err != nil {
		log.Fatalf("error initializing keycloak user auth: %v", err)
	}

	adminAuth, err := auth.NewKeycloakAuth("prism-admin", config.Keycloak)
	if err != nil {
		log.Fatalf("error initializing keycloak admin auth: %v", err)
	}

	backend := services.NewBackend(db, openalex, entityNdb, userAuth, adminAuth)

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
