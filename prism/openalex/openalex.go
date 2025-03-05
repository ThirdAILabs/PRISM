package openalex

import (
	"prism/prism/api"
	"time"
)

type Institution struct {
	InstitutionName string
	InstitutionId   string
	Location        string
}

type Author struct {
	AuthorId                string
	DisplayName             string
	DisplayNameAlternatives []string // Only set when calling GetAuthor, not returned via works
	RawAuthorName           *string  // Only set when getting author via works
	Institutions            []Institution
	Concepts                []string // Only set when calling FindAuthors
}

func (a *Author) InstitutionNames() []string {
	names := make([]string, 0, len(a.Institutions))
	for _, institution := range a.Institutions {
		names = append(names, institution.InstitutionName)
	}
	return names
}

type Grant struct {
	FunderId   string
	FunderName string
}

type Location struct {
	OrganizationId   string
	OrganizationName string
}

type Work struct {
	WorkId          string
	DisplayName     string
	WorkUrl         string
	OaUrl           string
	DownloadUrl     string
	PublicationDate time.Time
	Authors         []Author
	Grants          []Grant
	Locations       []Location
}

func (w *Work) GetDisplayName() string {
	if len(w.DisplayName) > 0 {
		return w.DisplayName
	}
	return "unknown"
}

type WorkBatch struct {
	Works           []Work
	TargetAuthorIds []string
	Error           error
}

type InstitutionAuthor struct {
	AuthorId   string
	AuthorName string
}

type KnowledgeBase interface {
	AutocompleteAuthor(query string) ([]api.Autocompletion, error)

	AutocompleteInstitution(query string) ([]api.Autocompletion, error)

	AutocompletePaper(query string) ([]api.Autocompletion, error)

	FindAuthors(authorName, institutionId string) ([]Author, error)

	FindAuthorByOrcidId(orcidId string) (Author, error)

	StreamWorks(authorId string, startDate, endDate time.Time) chan WorkBatch

	FindWorksByTitle(titles []string, startDate, endDate time.Time) ([]Work, error)

	GetAuthor(authorId string) (Author, error)

	GetInstitutionAuthors(institutionId string, startDate, endDate time.Time) ([]InstitutionAuthor, error)
}
