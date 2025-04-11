package versions

import (
	"gorm.io/gorm"
)

func Migration5(db *gorm.DB) error {
	type UniversityReport struct {
		UniversityLocation string
	}

	if err := db.Migrator().AddColumn(&UniversityReport{}, "UniversityLocation"); err != nil {
		return err
	}

	return nil
}

func Rollback5(db *gorm.DB) error {
	type UniversityReport struct {
		UniversityLocation string
	}

	if err := db.Migrator().DropColumn(&UniversityReport{}, "UniversityLocation"); err != nil {
		return err
	}

	return nil
}
