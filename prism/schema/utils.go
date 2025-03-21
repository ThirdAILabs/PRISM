package schema

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/url"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func closeGormDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("error getting sql.DB: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		t.Fatalf("error closing database connection: %v", err)
	}
}

func SetupTestDB(t *testing.T) *gorm.DB {
	tables := []any{&AuthorReport{}, &AuthorFlag{}, &UserAuthorReport{},
		&UniversityReport{}, &UserUniversityReport{}}

	testUri := os.Getenv("TEST_DB_URI")
	if testUri == "" {
		t.Fatal("TEST_DB_URI env not set")
	}

	rootDB, err := gorm.Open(postgres.Open(UriToDsn(testUri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	dbName := fmt.Sprintf("prism_test_%d_%d", os.Getpid(), rand.Int())

	if err := rootDB.Exec("CREATE DATABASE " + dbName).Error; err != nil {
		t.Fatalf("error creating database: %v", err)
	}

	t.Cleanup(func() {
		if err := rootDB.Exec("DROP DATABASE IF EXISTS " + dbName).Error; err != nil {
			t.Fatalf("error dropping database: %v", err)
		}
		closeGormDB(t, rootDB)
	})

	db, err := gorm.Open(postgres.Open(UriToDsn(testUri+"/"+dbName)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	t.Cleanup(func() {
		closeGormDB(t, db)
	})

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			t.Fatalf("error dropping table %T: %v", table, err)
		}
	}

	if err := db.Migrator().DropTable("university_authors"); err != nil {
		t.Fatalf("error dropping table university_authors: %v", err)
	}

	if err := db.AutoMigrate(tables...); err != nil {
		t.Fatalf("error migrating tables: %v", err)
	}

	return db
}

func UriToDsn(uri string) string {
	parts, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("error parsing db uri: %v", err)
	}
	pwd, _ := parts.User.Password()
	dbname := strings.TrimPrefix(parts.Path, "/")

	if dbname != "" {
		dbname = "dbname=" + dbname
	}

	return fmt.Sprintf("host=%v user=%v password=%v %v port=%v", parts.Hostname(), parts.User.Username(), pwd, dbname, parts.Port())
}
