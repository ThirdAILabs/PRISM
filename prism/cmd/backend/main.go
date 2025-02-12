package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"prism/openalex"
	"prism/schema"
	"prism/search"
	"prism/services"
	"prism/services/auth"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func loadConfig() Config {
	var configPath string

	flag.StringVar(&configPath, "config", "", "path to load config from")
	flag.Parse()

	if configPath == "" {
		log.Fatal("must specify config path")
	}

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("unable to open config file: %v", err)
	}

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	return config
}

func uriToDsn(uri string) string {
	parts, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("error parsing db uri: %v", err)
	}
	pwd, _ := parts.User.Password()
	dbname := strings.TrimPrefix(parts.Path, "/")
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", parts.Hostname(), parts.User.Username(), pwd, dbname, parts.Port())
}

func initDb(uri string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(uriToDsn(uri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	err = db.AutoMigrate(
		&schema.Report{}, &schema.ReportContent{}, &schema.License{},
		&schema.LicenseUser{}, &schema.LicenseUsage{},
	)
	if err != nil {
		log.Fatalf("error migrating db schema: %v", err)
	}

	return db
}

func initLogging(logFile *os.File) {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	slog.Info("logging initialized", "log_file", logFile.Name())
}

func main() {
	config := loadConfig()

	logFile, err := os.OpenFile(config.logfile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	initLogging(logFile)

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

	db := initDb(config.PostgresUri)

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
