package migrations

import (
	"log"
	"prism/prism/schema"
	"prism/prism/schema/migrations/versions"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func GetMigrator(db *gorm.DB) *gormigrate.Gormigrate {
	migrator := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID:      "0",
			Migrate: versions.Migration0,
		},
		{
			ID:       "1",
			Migrate:  versions.Migration1,
			Rollback: versions.Rollback1,
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

	return migrator
}

func RunMigrations(db *gorm.DB) {
	log.Println("running db migrations")

	migrator := GetMigrator(db)

	if err := migrator.Migrate(); err != nil {
		log.Fatalf("db migration failed: %v", err)
	}

	log.Println("db migrations complete")
}
