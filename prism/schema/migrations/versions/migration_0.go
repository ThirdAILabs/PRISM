package versions

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Migration0(db *gorm.DB) error {
	type AuthorFlag struct {
		ReportId uuid.UUID `gorm:"type:uuid;primaryKey"`
		FlagHash string    `gorm:"type:char(64);primaryKey"` // This will be the sha256 hash of the flag data (or enough of the flag data to uniquly identify the flag)
		FlagType string    `gorm:"size:40;not null"`
		Date     sql.NullTime
		Data     []byte
	}

	type AuthorReport struct {
		Id uuid.UUID `gorm:"type:uuid;primaryKey"`

		LastUpdatedAt time.Time

		AuthorId   string `gorm:"index"`
		AuthorName string
		Source     string

		QueuedAt     time.Time
		Status       string `gorm:"size:20;not null"`
		QueuedByUser bool

		Flags []AuthorFlag `gorm:"foreignKey:ReportId;constraint:OnDelete:CASCADE"`
	}

	type UserAuthorReport struct {
		Id     uuid.UUID `gorm:"type:uuid;primaryKey"`
		UserId uuid.UUID `gorm:"type:uuid;not null;index"`

		LastAccessedAt time.Time

		ReportId uuid.UUID     `gorm:"type:uuid;not null"`
		Report   *AuthorReport `gorm:"foreignKey:ReportId"`
	}

	type UniversityReport struct {
		Id uuid.UUID `gorm:"type:uuid;primaryKey"`

		LastUpdatedAt time.Time

		UniversityId   string `gorm:"index"`
		UniversityName string

		QueuedAt time.Time
		Status   string `gorm:"size:20;not null"`

		Authors []AuthorReport `gorm:"many2many:university_authors"`
	}

	type UserUniversityReport struct {
		Id     uuid.UUID `gorm:"type:uuid;primaryKey"`
		UserId uuid.UUID `gorm:"type:uuid;not null;index"`

		LastAccessedAt time.Time

		ReportId uuid.UUID         `gorm:"type:uuid;not null"`
		Report   *UniversityReport `gorm:"foreignKey:ReportId"`
	}

	// This uses the structs defined here instead of in schema.go because they need
	// to be consistent with the original schema definition and not reflect any schema
	// changes.
	err := db.AutoMigrate(
		&AuthorFlag{}, &AuthorReport{}, &UserAuthorReport{}, &UniversityReport{}, &UserUniversityReport{},
	)
	if err != nil {
		return fmt.Errorf("initial migration failed: %w", err)
	}
	return nil
}
