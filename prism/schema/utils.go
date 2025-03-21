package schema

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	tables := []any{&AuthorReport{}, &AuthorFlag{}, &UserAuthorReport{},
		&UniversityReport{}, &UserUniversityReport{}}

	testUri := os.Getenv("TEST_DB_URI")
	if testUri == "" {
		t.Fatal("TEST_DB_URI env not set")
	}

	db, err := gorm.Open(postgres.Open(UriToDsn(testUri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			t.Fatalf("error dropping table %T: %v", table, err)
		}
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
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", parts.Hostname(), parts.User.Username(), pwd, dbname, parts.Port())
}
