package schema

import (
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

	QueuedAt time.Time
	Status   string `gorm:"size:20;not null"`

	Flags []AuthorFlag `gorm:"foreignKey:ReportId;constraint:OnDelete:CASCADE"`
}

type AuthorFlag struct {
	Id       uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReportId uuid.UUID `gorm:"type:uuid"`
	FlagType string    `gorm:"size:40;not null"`
	FlagKey  string
	Data     []byte
}

type UserAuthorReport struct {
	Id     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId uuid.UUID `gorm:"type:uuid;not null;index"`

	CreatedAt time.Time

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

	CreatedAt time.Time

	ReportId uuid.UUID         `gorm:"type:uuid;not null"`
	Report   *UniversityReport `gorm:"foreignKey:ReportId"`
}

type License struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Secret      []byte
	Name        string
	Expiration  time.Time
	Deactivated bool
}

type LicenseUser struct {
	UserId    uuid.UUID `gorm:"type:uuid;primaryKey"`
	LicenseId uuid.UUID `gorm:"type:uuid"`

	License *License `gorm:"foreignKey:LicenseId"`
}

const (
	UniversityReportType = "university"
	AuthorReportType     = "author"
)

type LicenseUsage struct {
	LicenseId  uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReportId   uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReportType string    `gorm:"size:40"`
	UserId     uuid.UUID
	Timestamp  time.Time
}
