package versions

import "gorm.io/gorm"

func Migration2(db *gorm.DB) error {
	type AuthorReport struct{}

	if err := db.Migrator().RenameColumn(&AuthorReport{}, "queued_by_user", "for_university_report"); err != nil {
		return err
	}

	if err := db.Exec("UPDATE author_reports SET for_university_report = NOT for_university_report").Error; err != nil {
		return err
	}

	return nil
}

func Rollback2(db *gorm.DB) error {
	type AuthorReport struct{}

	if err := db.Exec("UPDATE author_reports SET for_university_report = NOT for_university_report").Error; err != nil {
		return err
	}

	if err := db.Migrator().RenameColumn(&AuthorReport{}, "for_university_report", "queued_by_user"); err != nil {
		return err
	}

	return nil
}
