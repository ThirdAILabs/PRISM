package api

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	Id uuid.UUID

	CreatedAt time.Time

	AuthorId   string
	AuthorName string
	Source     string

	Status string

	Content ReportContent
}

type CreateReportRequest struct {
	AuthorId   string
	AuthorName string
	Source     string
}

type CreateReportResponse struct {
	Id uuid.UUID
}

const (
	OpenAlexSource      = "openalex"
	GoogleScholarSource = "google-scholar"
	UnstructuredSource  = "unstructured"
	ScopusSource        = "scopus"
)

type Author struct {
	AuthorId          string
	AuthorName        string
	Institutions      []string
	Source            string
	Interests []string
}

type GScholarSearchResults struct {
	Authors []Author
	Cursor  string
}

type Institution struct {
	InstitutionId   string
	InstitutionName string
	Location        string
}

type FormalRelationResponse struct {
	HasFormalRelation bool
}

type MatchEntitiesResponse struct {
	Entities []string
}

type CreateLicenseRequest struct {
	Name       string
	Expiration time.Time
}

type CreateLicenseResponse struct {
	Id      uuid.UUID
	License string
}

type ActivateLicenseRequest struct {
	License string
}

type License struct {
	Id          uuid.UUID
	Name        string
	Expiration  time.Time
	Deactivated bool
}
