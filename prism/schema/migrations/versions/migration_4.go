package versions

import (
	"gorm.io/gorm"
)

func Migration4(db *gorm.DB) error {
	type AuthorReport struct {
		Affiliations      string
		ResearchInterests string
	}

	if err := db.Migrator().AddColumn(&AuthorReport{}, "Affiliations"); err != nil {
		return err
	}

	if err := db.Migrator().AddColumn(&AuthorReport{}, "ResearchInterests"); err != nil {
		return err
	}

	return nil
}

func Rollback4(db *gorm.DB) error {
	type AuthorReport struct {
		Affiliations      string
		ResearchInterests string
	}

	if err := db.Migrator().DropColumn(&AuthorReport{}, "Affiliations"); err != nil {
		return err
	}

	if err := db.Migrator().DropColumn(&AuthorReport{}, "ResearchInterests"); err != nil {
		return err
	}
	return nil
}
