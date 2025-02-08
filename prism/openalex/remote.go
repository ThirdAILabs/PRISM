package openalex

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

var ErrSearchFailed = errors.New("error performing openalex search")

type RemoteKnowledgeBase struct{}

func NewRemoteKnowledgeBase() KnowledgeBase {
	return &RemoteKnowledgeBase{}
}

func autocompleteHelper(component, query string, dest interface{}) error {
	url := fmt.Sprintf("https://api.openalex.org/autocomplete/%s?q=%s&mailto=kartik@thirdai.com", component, url.QueryEscape(query))

	res, err := http.Get(url)
	if err != nil {
		slog.Error("openalex: autocomplete failed", "query", query, "component", component, "error", err)
		return fmt.Errorf("unable to get autocomplete suggestions")
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(dest); err != nil {
		slog.Error("openalex: error parsing reponse from autocomplete", "query", query, "component", component, "error", err)
		return fmt.Errorf("unable to get autocomplete suggestions")
	}

	return nil
}

type oaResults[T any] struct {
	Results []T `json:"results"`
}

// Response Format: https://docs.openalex.org/how-to-use-the-api/get-lists-of-entities/autocomplete-entities#response-format
type oaAutocompletion struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
	Hint        string `json:"hint"`
}

func (oa *RemoteKnowledgeBase) AutocompleteAuthor(query string) ([]Author, error) {
	var suggestions oaResults[oaAutocompletion]

	if err := autocompleteHelper("authors", query, &suggestions); err != nil {
		return nil, err
	}

	authors := make([]Author, 0, len(suggestions.Results))
	for _, author := range suggestions.Results {
		authors = append(authors, Author{
			AuthorId:     author.Id,
			DisplayName:  author.DisplayName,
			Institutions: []Institution{{InstitutionName: author.Hint}},
		})
	}

	return authors, nil
}

func (oa *RemoteKnowledgeBase) AutocompleteInstitution(query string) ([]Institution, error) {
	var suggestions oaResults[oaAutocompletion]

	if err := autocompleteHelper("institutions", query, &suggestions); err != nil {
		return nil, err
	}

	institutions := make([]Institution, 0, len(suggestions.Results))
	for _, institution := range suggestions.Results {
		institutions = append(institutions, Institution{
			InstitutionName: institution.DisplayName,
			InstitutionId:   institution.Id,
		})
	}

	return institutions, nil
}

// Response Format: https://docs.openalex.org/api-entities/authors/get-lists-of-authors
type oaAuthor struct {
	Id                      string          `json:"id"`
	DisplayName             string          `json:"diplay_name"`
	DisplayNameAlternatives []string        `json:"display_name_alternatives"`
	WorksCount              int             `json:"works_count"`
	Affiliations            []oaAffiliation `json:"affiliations"`
}

type oaInstitution struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
	CountryCode string `json:"country_code"`
}

type oaAffiliation struct {
	Institution oaInstitution `json:"institution"`
}

func (oa *RemoteKnowledgeBase) FindAuthors(author, institution string) ([]Author, error) {
	url := fmt.Sprintf(
		"https://api.openalex.org/authors?filter=display_name.search:%s,affiliations.institution.id:%s&mailto=kartik@thirdai.com",
		url.QueryEscape(author), url.QueryEscape(institution),
	)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("openalex: author search failed", "author", author, "institution", institution, "error", err)
		return nil, ErrSearchFailed
	}
	defer res.Body.Close()

	var results oaResults[oaAuthor]

	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		slog.Error("openalex: error parsing author search response", "author", author, "institution", institution, "error", err)
		return nil, ErrSearchFailed
	}

	authors := make([]Author, 0, len(results.Results))
	for _, result := range results.Results {
		if result.WorksCount > 0 {
			institutions := make([]Institution, 0)
			for i, inst := range result.Affiliations {
				if i < 3 || inst.Institution.CountryCode == "US" {
					institutions = append(institutions, Institution{
						InstitutionName: inst.Institution.DisplayName,
						InstitutionId:   inst.Institution.Id,
					})
				}
			}

			authors = append(authors, Author{
				AuthorId:                result.Id,
				DisplayName:             result.DisplayName,
				DisplayNameAlternatives: result.DisplayNameAlternatives,
				RawAuthorName:           nil,
				Institutions:            institutions,
			})
		}
	}

	return authors, nil
}

type oaMetadata struct {
	NextCursor string `json:"next_cursor"`
}

type oaWorkResults struct {
	Meta    oaMetadata `json:"meta"`
	Results []oaWork   `json:"results"`
}

// Response Format: https://docs.openalex.org/api-entities/works/get-lists-of-works
type oaWork struct {
	Id              string `json:"id"`
	DisplayName     string `json:"display_name"`
	PublicationYear int    `json:"publication_year"`

	Ids oaWorkIds `json:"ids"`

	PrimaryLocation oaLocation   `json:"primary_location"`
	Locations       []oaLocation `json:"locations"`

	BestOaLocation oaLocation `json:"best_oa_location"`

	Authorships []oaAuthorship `json:"authorships"`

	Grants []oaGrant `json:"grants"`
}

func (work *oaWork) getWorkUrl() string {
	if len(work.PrimaryLocation.LandingPageUrl) > 0 {
		return work.PrimaryLocation.LandingPageUrl
	}
	if len(work.Locations) > 0 && len(work.Locations[0].LandingPageUrl) > 0 {
		return work.Locations[0].LandingPageUrl
	}
	return work.Ids.Openalex
}

func (work *oaWork) getOaUrl() string {
	if work.BestOaLocation.IsOA {
		return work.BestOaLocation.LandingPageUrl
	}
	return "none"
}

func (work *oaWork) pdfUrl() string {
	if work.PrimaryLocation.PdfUrl != nil {
		return *work.PrimaryLocation.PdfUrl
	}
	if work.BestOaLocation.PdfUrl != nil {
		return *work.BestOaLocation.PdfUrl
	}
	return ""
}

type oaWorkIds struct {
	Openalex string `json:"openalex"`
}

// This is slightly different from the author above because here we have a subset of the fields
type oaWorkAuthor struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type oaAuthorship struct {
	Author        oaAuthor        `json:"author"`
	Institutions  []oaInstitution `json:"institutions"`
	RawAuthorName string          `json:"raw_author_name"`
}

type oaLocation struct {
	IsOA           bool     `json:"is_oa"`
	LandingPageUrl string   `json:"landing_page_url"`
	Source         oaSource `json:"source"`
	PdfUrl         *string  `json:"pdf_url"`
}

type oaSource struct {
	DisplayName      string `json:"display_name"` // Should we be using "host_organization_name" instead
	HostOrganization string `json:"host_organization"`
}

type oaGrant struct {
	Funder            string `json:"funder"`
	FunderDisplayName string `json:"funder_display_name"`
}

func getYearFilter(startYear, endYear int) string {
	yearFilter := ""
	if startYear >= 0 {
		yearFilter += fmt.Sprintf(",from_publication_date:%d-01-01", startYear)
	}
	if endYear >= 0 {
		yearFilter += fmt.Sprintf(",to_publication_date:%d-12-31", endYear)
	}
	return yearFilter
}

func converOpenalexWork(work oaWork) Work {
	authors := make([]Author, 0)
	for _, author := range work.Authorships {
		institutions := make([]Institution, 0)
		// Here the author affiliations are not provided, so we use the authorship institutions field
		for _, institution := range author.Institutions {
			institutions = append(institutions, Institution{
				InstitutionName: institution.DisplayName,
				InstitutionId:   institution.Id,
			})
		}
		authors = append(authors, Author{
			AuthorId:                author.Author.Id,
			DisplayName:             author.Author.DisplayName,
			DisplayNameAlternatives: author.Author.DisplayNameAlternatives,
			RawAuthorName:           &author.RawAuthorName,
			Institutions:            institutions,
		})
	}

	grants := make([]Grant, 0, len(work.Grants))
	for _, grant := range work.Grants {
		grants = append(grants, Grant{
			FunderId:   grant.Funder,
			FunderName: grant.Funder,
		})
	}

	locations := make([]Location, 0, len(work.Locations))
	for _, loc := range work.Locations {
		locations = append(locations, Location{
			OrganizationId:   loc.Source.HostOrganization,
			OrganizationName: loc.Source.DisplayName,
		})
	}

	return Work{
		WorkId:          work.Id,
		DisplayName:     work.DisplayName,
		WorkUrl:         work.getWorkUrl(),
		OaUrl:           work.getOaUrl(),
		DownloadUrl:     work.pdfUrl(),
		PublicationYear: work.PublicationYear,
		Authors:         authors,
	}
}

func (oa *RemoteKnowledgeBase) StreamWorks(authorId string, startYear, endYear int) chan WorkBatch {
	outputCh := make(chan WorkBatch, 10)

	cursor := "*"

	yearFilter := getYearFilter(startYear, endYear)

	go func() {
		defer close(outputCh)

		for cursor != "" {
			url := fmt.Sprintf("https://api.openalex.org/works?filter=authorships.author.id:%s%s&per-page=200&cursor=%s&mailto=kartik@thirdai.com", authorId, yearFilter, cursor)

			res, err := http.Get(url)
			if err != nil {
				outputCh <- WorkBatch{Works: nil, TargetAuthorIds: nil, Error: fmt.Errorf("openalex: work search failed: %w", err)}
				break
			}

			var results oaResults[oaWork]
			if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
				slog.Error("openalex: error parsing response from work search", "author_id", authorId, "start_year", startYear, "end_year", endYear, "error", err)
				outputCh <- WorkBatch{Works: nil, TargetAuthorIds: nil, Error: fmt.Errorf("error parsing response from open alex: %w", err)}
				break
			}

			works := make([]Work, 0, len(results.Results))
			for _, work := range results.Results {
				works = append(works, converOpenalexWork(work))
			}

			outputCh <- WorkBatch{Works: works, TargetAuthorIds: []string{authorId}, Error: nil}
		}
	}()

	return outputCh
}

func (oa *RemoteKnowledgeBase) FindWorksByTitle(titles []string, startYear, endYear int) ([]Work, error) {
	works := make([]Work, 0, len(titles))

	yearFilter := getYearFilter(startYear, endYear)

	for _, title := range titles {
		url := fmt.Sprintf("https://api.openalex.org/works?filter=title.search:%s%s&per-page=1&mailto=kartik@thirdai.com", url.QueryEscape(title), yearFilter)
		res, err := http.Get(url)
		if err != nil {
			slog.Error("openalex: error searching for work by title", "title", title, "error", err)
			return nil, fmt.Errorf("openalex work search failed: %w", err)
		}

		var results oaResults[oaWork]
		if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
			slog.Error("openalex: error parsing response from work search", "title", title, "start_year", startYear, "end_year", endYear, "error", err)
			return nil, fmt.Errorf("error parsing response from open alex: %w", err)
		}

		if len(results.Results) > 0 {
			works = append(works, converOpenalexWork(results.Results[0]))
		}
	}

	return works, nil
}

func (oa *RemoteKnowledgeBase) GetAuthor(authorId string) (Author, error) {
	url := "https://api.openalex.org/authors?filter=openalex:" + url.QueryEscape(authorId)

	res, err := http.Get(url)
	if err != nil {
		slog.Error("openalex: get author failed", "author_id", authorId, "error", err)
		return Author{}, ErrSearchFailed
	}
	defer res.Body.Close()

	var results oaResults[oaAuthor]

	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		slog.Error("openalex: error parsing get author", "author_id", authorId, "error", err)
		return Author{}, ErrSearchFailed
	}

	if len(results.Results) < 1 {
		slog.Error("openalex: expected 1 author in get author, got 0", "author_id", authorId)
		return Author{}, fmt.Errorf("no authors returned in get author")
	}

	institutions := make([]Institution, 0, len(results.Results[0].Affiliations))
	for _, institution := range results.Results[0].Affiliations {
		institutions = append(institutions, Institution{
			InstitutionName: institution.Institution.DisplayName,
			InstitutionId:   institution.Institution.Id,
		})
	}

	return Author{
		AuthorId:                results.Results[0].Id,
		DisplayName:             results.Results[0].DisplayName,
		DisplayNameAlternatives: results.Results[0].DisplayNameAlternatives,
		RawAuthorName:           nil,
		Institutions:            institutions,
	}, nil
}
