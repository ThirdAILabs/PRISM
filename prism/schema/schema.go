package schema

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const (
	ReportQueued     = "queued"
	ReportInProgress = "in-progress"
	ReportFailed     = "failed"
	ReportCompleted  = "complete"
)

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

type AuthorFlag struct {
	Id       uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReportId uuid.UUID `gorm:"type:uuid"`
	FlagType string    `gorm:"size:40;not null"`
	FlagKey  string
	Date     sql.NullTime
	Data     []byte
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
