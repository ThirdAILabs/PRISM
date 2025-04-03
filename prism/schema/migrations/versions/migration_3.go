package versions

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Migration3(db *gorm.DB) error {

	type UserAuthorReport struct {
		Id uuid.UUID `gorm:"type:uuid;primaryKey"`
	}

	type AuthorReportHook struct {
		Id uuid.UUID `gorm:"type:uuid;primaryKey"`

		UserReportId uuid.UUID         `gorm:"type:uuid;not null"`
		UserReport   *UserAuthorReport `gorm:"foreignKey:UserReportId"`

		Action string
		Data   []byte

		LastRanAt time.Time
		Interval  int
	}

	if err := db.AutoMigrate(&AuthorReportHook{}); err != nil {
		return err
	}

	return nil
}

func Rollback3(db *gorm.DB) error {

	if err := db.Migrator().DropTable("author_report_hooks"); err != nil {
		return err
	}

	return nil
}
