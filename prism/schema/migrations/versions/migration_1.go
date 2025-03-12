package versions

import "gorm.io/gorm"

func Migration1(db *gorm.DB) error {
	type AuthorReport struct{}
	type UniversityReport struct{}

	if err := db.Migrator().RenameColumn(&AuthorReport{}, "queued_at", "status_updated_at"); err != nil {
		return err
	}

	if err := db.Migrator().RenameColumn(&UniversityReport{}, "queued_at", "status_updated_at"); err != nil {
		return err
	}

	return nil
}

func Rollback1(db *gorm.DB) error {
	type AuthorReport struct{}
	type UniversityReport struct{}

	if err := db.Migrator().RenameColumn(&AuthorReport{}, "status_updated_at", "queued_at"); err != nil {
		return err
	}

	if err := db.Migrator().RenameColumn(&UniversityReport{}, "status_updated_at", "queued_at"); err != nil {
		return err
	}

	return nil
}
