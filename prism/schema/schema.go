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

type Report struct {
	Id uuid.UUID `gorm:"type:uuid;primaryKey"`

	UserId uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time

	// Report params
	AuthorId   string
	AuthorName string
	Source     string
	StartYear  int
	EndYear    int

	Status string `gorm:"size:20;not null"`

	Content *ReportContent `gorm:"foreignKey:ReportId;constraint:OnDelete:CASCADE"`
}

type ReportContent struct {
	ReportId uuid.UUID `gorm:"type:uuid;primaryKey"`

	Content []byte
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

type LicenseUsage struct {
	LicenseId uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReportId  uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId    uuid.UUID
	Timestamp time.Time
}
