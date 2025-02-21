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

	LastUpdate time.Time

	AuthorId   string `gorm:"index"`
	AuthorName string
	Source     string

	QueuedAt time.Time
	Status   string `gorm:"size:20;not null"`

	Content *ReportContent `gorm:"foreignKey:ReportId;constraint:OnDelete:CASCADE"`
}

type ReportContent struct {
	ReportId uuid.UUID `gorm:"type:uuid;primaryKey"`

	Content []byte
}

type UserReport struct {
	Id     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId uuid.UUID `gorm:"type:uuid;not null;index"`

	CreatedAt time.Time

	ReportId uuid.UUID `gorm:"type:uuid;not null"`
	Report   *Report   `gorm:"foreignKey:ReportId"`
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
	LicenseId    uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserReportId uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserId       uuid.UUID
	Timestamp    time.Time
}
