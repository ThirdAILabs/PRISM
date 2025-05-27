package api

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Report struct {
	Id uuid.UUID

	LastAccessedAt time.Time

	AuthorId          string
	AuthorName        string
	Source            string
	Affiliations      string
	ResearchInterests string

	Status string

	Content map[string][]Flag
}

// We have to define a custom Unmarshal method because Flag is an interface so we
// cannot directly deserialize into it.
func (r *Report) UnmarshalJSON(data []byte) error {
	// Define an alias for the Report type to avoid infinite recursion
	type Alias Report

	// Define an anonymous struct with a Content field and an embedded Alias field
	aux := &struct {
		Content map[string][]json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(r), // Embed the Alias type as a pointer
	}

	// Unmarshal the JSON data into the anonymous struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Process the Content field separately
	r.Content = make(map[string][]Flag)
	for flagType, rawFlags := range aux.Content {
		for _, rawFlag := range rawFlags {
			flag, err := ParseFlag(flagType, rawFlag)
			if err != nil {
				return err
			}
			r.Content[flagType] = append(r.Content[flagType], flag)
		}
	}

	return nil
}

type CreateAuthorReportRequest struct {
	AuthorId   string
	AuthorName string
	Source     string
}

type CreateUniversityReportRequest struct {
	UniversityId       string
	UniversityName     string
	UniversityLocation string
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

type CreateHookRequest struct {
	Action   string
	Data     []byte
	Interval int
}

type CreateHookResponse struct {
	Id uuid.UUID
}

type AvailableHookResponse struct {
	Type string
}

type UniversityAuthorFlag struct {
	AuthorId   string
	AuthorName string
	Source     string
	FlagCount  int
}

type UniversityReportFlag struct {
	Total   int
	Authors []UniversityAuthorFlag
}

type UniversityReport struct {
	Id uuid.UUID

	LastAccessedAt time.Time

	UniversityId       string
	UniversityName     string
	UniversityLocation string

	Status string

	Content UniversityReportContent
}

type UniversityReportContent struct {
	TotalAuthors    int
	AuthorsReviewed int
	Flags           map[string][]UniversityAuthorFlag
}

type Autocompletion struct {
	Id   string
	Name string
	Hint string
}

type Author struct {
	AuthorId     string
	AuthorName   string
	Institutions []string
	Source       string
	Interests    []string
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

type MatchedEntity struct {
	Names    string
	Address  string
	Country  string
	Type     string
	Resource string
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

type HookResponse struct {
	Id       uuid.UUID
	Action   string
	Interval int
}
