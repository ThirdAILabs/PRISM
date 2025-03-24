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

	StatusUpdatedAt     time.Time
	Status              string `gorm:"size:20;not null"`
	ForUniversityReport bool

	Flags []AuthorFlag `gorm:"foreignKey:ReportId;constraint:OnDelete:CASCADE"`
}

type AuthorFlag struct {
	ReportId  uuid.UUID `gorm:"type:uuid;primaryKey"`
	FlagHash  string    `gorm:"type:char(64);primaryKey"` // This will be the sha256 hash of the flag data (or enough of the flag data to uniquly identify the flag)
	FlagType  string    `gorm:"size:40;not null"`
	Date      sql.NullTime
	Data      []byte
	Feedbacks []FlagFeedback `gorm:"foreignKey:ReportId,FlagHash;references:ReportId,FlagHash;constraint:OnDelete:CASCADE"`
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

	StatusUpdatedAt time.Time
	Status          string `gorm:"size:20;not null"`

	Authors []AuthorReport `gorm:"many2many:university_authors"`
}

type UserUniversityReport struct {
	Id     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId uuid.UUID `gorm:"type:uuid;not null;index"`

	LastAccessedAt time.Time

	ReportId uuid.UUID         `gorm:"type:uuid;not null"`
	Report   *UniversityReport `gorm:"foreignKey:ReportId"`
}

type FlagFeedback struct {
	Id         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId     uuid.UUID `gorm:"type:uuid;not null;index"`
	ReportId   uuid.UUID `gorm:"type:uuid;not null"`
	FlagHash   string    `gorm:"type:char(64);not null"`
	Timestamp  time.Time
	Data       []byte
	AuthorFlag AuthorFlag `gorm:"foreignKey:ReportId,FlagHash;references:ReportId,FlagHash"`
}
