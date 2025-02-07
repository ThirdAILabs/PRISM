package openalex

type Institution struct {
	InstitutionName string
	InstitutionId   string
}

type Author struct {
	AuthorId                string
	DisplayName             string
	DisplayNameAlternatives []string
	RawAuthorName           *string
	Institutions            []Institution
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
	PublicationYear int
	Authors         []Author
	RawAuthorName   string
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

type KnowledgeBase interface {
	AutocompleteAuthor(query string) ([]Author, error)

	AutocompleteInstitution(query string) ([]Institution, error)

	FindAuthors(author, institution string) ([]Author, error)

	StreamWorks(authorId string, startYear, endYear int) chan WorkBatch

	FindWorksByTitle(titles []string, startYear, endYear int) ([]Work, error)

	GetAuthor(authorId string) (Author, error)
}
