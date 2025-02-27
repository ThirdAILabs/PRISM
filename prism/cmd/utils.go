package cmd

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/url"
	"os"
	"prism/prism/schema"
	"strings"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func uriToDsn(uri string) string {
	parts, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("error parsing db uri: %v", err)
	}
	pwd, _ := parts.User.Password()
	dbname := strings.TrimPrefix(parts.Path, "/")
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", parts.Hostname(), parts.User.Username(), pwd, dbname, parts.Port())
}

func InitDb(uri string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(uriToDsn(uri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	err = db.AutoMigrate(
		&schema.AuthorReport{}, &schema.AuthorFlag{}, &schema.UserAuthorReport{},
		&schema.UniversityReport{}, &schema.UserUniversityReport{},
		&schema.License{}, &schema.LicenseUser{}, &schema.LicenseUsage{},
	)
	if err != nil {
		log.Fatalf("error migrating db schema: %v", err)
	}

	return db
}

func InitLogging(logFile *os.File) {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	slog.Info("logging initialized", "log_file", logFile.Name())
}

func LoadConfig(config any) {
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

	if err := yaml.NewDecoder(file).Decode(config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}
}
