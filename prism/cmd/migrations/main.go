package main

import (
	"flag"
	"log"
	"prism/prism/cmd"
	"prism/prism/cmd/migrations/versions"
	"prism/prism/schema"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbUri := flag.String("db", "", "Database URI")
	flag.Parse()

	db, err := gorm.Open(postgres.Open(cmd.UriToDsn(*dbUri)), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	migrator := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID:      "0",
			Migrate: versions.Migration0,
		},
	})

	migrator.InitSchema(func(txn *gorm.DB) error {
		// This is run by the migrator if no previous migration is detected. It
		// allows it to bypass running all the migrations sequentially and just create
		// the latest database state.

		log.Println("clean database detected, running full schema initialization")

		return db.AutoMigrate(
			&schema.AuthorReport{}, &schema.AuthorFlag{}, &schema.UserAuthorReport{},
			&schema.UniversityReport{}, &schema.UserUniversityReport{},
		)
	})

	if err := migrator.Migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
