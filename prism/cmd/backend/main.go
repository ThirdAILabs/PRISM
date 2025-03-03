package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"prism/prism/api"
	"prism/prism/cmd"
	"prism/prism/licensing"
	"prism/prism/openalex"
	"prism/prism/search"
	"prism/prism/services"
	"prism/prism/services/auth"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Config struct {
	PostgresUri            string `yaml:"postgres_uri"`
	SearchableEntitiesPath string `yaml:"searchable_entities"`
	Keycloak               auth.KeycloakArgs
	Port                   int
	Logfile                string
	PrismLicense           string `yaml:"prism_license"`
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

func buildEntityNdb(entityPath string) services.EntitySearch {
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

	s := time.Now()
	es, err := services.NewEntitySearch(entityNdbPath)
	if err != nil {
		log.Fatalf("error openning ndb: %v", err)
	}

	if err := es.Insert(entities); err != nil {
		log.Fatalf("error inserting into ndb: %v", err)
	}

	e := time.Now()

	log.Printf("searchable entity ndb construction time=%.3f", e.Sub(s).Seconds())

	return es
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

	licensing, err := licensing.NewLicenseVerifier(config.PrismLicense)
	if err != nil {
		log.Fatalf("error initializing licensing: %v", err)
	}

	if err := search.SetLicenseKey(config.PrismLicense); err != nil {
		log.Fatalf("error activating license key: %v", err)
	}

	entitySearch := buildEntityNdb(config.SearchableEntitiesPath)

	db := cmd.InitDb(config.PostgresUri)

	userAuth, err := auth.NewKeycloakAuth("prism-user", config.Keycloak)
	if err != nil {
		log.Fatalf("error initializing keycloak user auth: %v", err)
	}

	backend := services.NewBackend(db, openalex, entitySearch, userAuth, licensing)

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
