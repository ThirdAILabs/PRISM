package api

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	Id uuid.UUID

	CreatedAt time.Time

	AuthorId    string
	DisplayName string
	Source      string
	StartYear   int
	EndYear     int

	Status string

	Content any `json:"omitempty"`
}

type CreateReportRequest struct {
	AuthorId    string
	DisplayName string
	Source      string
	StartYear   int
	EndYear     int
}

type CreateReportResponse struct {
	Id uuid.UUID
}

const (
	OpenAlexSource      = "openalex"
	GoogleScholarSource = "google-scholar"
	ScopusSource        = "scopus"
)

type Author struct {
	AuthorId     string
	DisplayName  string
	Institutions []string
	Source       string
	WorksCount   int
}

type Institution struct {
	DisplayName string
}
