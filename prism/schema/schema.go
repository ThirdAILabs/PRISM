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
	AuthorId    string
	DisplayName string
	Source      string
	StartYear   int
	EndYear     int

	Status string `gorm:"size:20;not null"`

	Content *ReportContent `gorm:"foreignKey:ReportId;constraint:OnDelete:CASCADE"`
}

type ReportContent struct {
	ReportId uuid.UUID `gorm:"type:uuid;primaryKey"`

	Content []byte
}
