package cmd

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
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

func OpenDB(uri string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(uriToDsn(uri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	return db
}

func InitLogging(logFile *os.File) {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	slog.Info("logging initialized", "log_file", logFile.Name())
}

func LoadEnvFile() {
	var configPath string

	flag.StringVar(&configPath, "env", "", "path to load env from")
	flag.Parse()

	if configPath == "" {
		log.Printf("no env file specified, using os.Environ only")
		return
	}

	log.Printf("loading env from file %s", configPath)
	err := godotenv.Load(configPath)
	if err != nil {
		log.Fatalf("error loading .env file '%s': %v", configPath, err)
	}
}
