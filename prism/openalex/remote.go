package openalex

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"prism/prism/api"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	ErrSearchFailed   = errors.New("error performing openalex search")
	ErrAuthorNotFound = errors.New("author not found")
)

type RemoteKnowledgeBase struct {
	client *resty.Client
}

func NewRemoteKnowledgeBase() KnowledgeBase {
	return &RemoteKnowledgeBase{
		client: resty.New().
			SetBaseURL("https://api.openalex.org").
			AddRetryCondition(func(response *resty.Response, err error) bool {
				if err != nil {
					return true // The err can be non nil for some network errors.
				}
				// There's no reason to retry other 400 requests since the outcome should not change
				return response != nil && (response.StatusCode() > 499 || response.StatusCode() == http.StatusTooManyRequests)
			}).
			SetRetryCount(2).
			// Providing the contact moves requests to a new pool that can have better response times:
			// https://docs.openalex.org/how-to-use-the-api/rate-limits-and-authentication#the-polite-pool
			SetQueryParam("mailto", "contact@thirdai.com"),
	}
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

func (oa *RemoteKnowledgeBase) autocompleteHelper(component, query, filter string) ([]api.Autocompletion, error) {
	res, err := oa.client.R().
		SetResult(&oaResults[oaAutocompletion]{}).
		SetQueryParam("q", query).
		SetQueryParam("filter", filter).
		Get(fmt.Sprintf("/autocomplete/%s", component))

	if err != nil {
		slog.Error("openalex: autocomplete failed", "query", query, "component", component, "error", err)
		return nil, fmt.Errorf("unable to get autocomplete suggestions")
	}

	if !res.IsSuccess() {
		slog.Error("openalex: autocompletion returned error", "status_code", res.StatusCode(), "body", res.String())
		return nil, fmt.Errorf("unable to get autocomplete suggestions")
	}

	// fmt.Println(component, query, res.String())
	results := res.Result().(*oaResults[oaAutocompletion])
	// fmt.Println(results.Results)

	autocompletions := make([]api.Autocompletion, 0, len(results.Results))
	for _, result := range results.Results {
		autocompletions = append(autocompletions, api.Autocompletion{
			Id:   result.Id,
			Name: result.DisplayName,
			Hint: result.Hint,
		})
	}

	return autocompletions, nil
}

func (oa *RemoteKnowledgeBase) AutocompleteAuthor(query string, institutionId string) ([]api.Autocompletion, error) {
	var filterParam = ""
	if institutionId != "" {
		filterParam = fmt.Sprintf("affiliations.institution.id:%s", institutionId)
	}
	return oa.autocompleteHelper("authors", query, filterParam)
}

func (oa *RemoteKnowledgeBase) AutocompleteInstitution(query string) ([]api.Autocompletion, error) {
	return oa.autocompleteHelper("institutions", query, "")
}

func (oa *RemoteKnowledgeBase) AutocompletePaper(query string) ([]api.Autocompletion, error) {
	return oa.autocompleteHelper("works", query, "")
}

// Response Format: https://docs.openalex.org/api-entities/authors/get-lists-of-authors
type oaAuthor struct {
	Id                      string          `json:"id"`
	DisplayName             string          `json:"display_name"`
	DisplayNameAlternatives []string        `json:"display_name_alternatives"`
	WorksCount              int             `json:"works_count"`
	Affiliations            []oaAffiliation `json:"affiliations"`
	Concepts                []oaConcept     `json:"x_concepts"`
}

type oaInstitution struct {
	Id          string `json:"id"`
	DisplayName string `json:"display_name"`
	CountryCode string `json:"country_code"`
}

type oaAffiliation struct {
	Institution oaInstitution `json:"institution"`
}

type oaConcept struct {
	DisplayName string `json:"display_name"`
}

func convertOpenalexAuthor(author oaAuthor) Author {
	institutions := make([]Institution, 0, len(author.Affiliations))
	for i, inst := range author.Affiliations {
		if i < 3 || inst.Institution.CountryCode == "US" {
			institutions = append(institutions, Institution{
				InstitutionName: inst.Institution.DisplayName,
				InstitutionId:   inst.Institution.Id,
				Location:        inst.Institution.CountryCode,
			})
		}
	}
	concepts := make([]string, 0, len(author.Concepts))
	for _, concept := range author.Concepts {
		concepts = append(concepts, concept.DisplayName)
	}

	return Author{
		AuthorId:                author.Id,
		DisplayName:             author.DisplayName,
		DisplayNameAlternatives: author.DisplayNameAlternatives,
		RawAuthorName:           nil,
		Institutions:            institutions,
		Concepts:                concepts,
	}
}

func (oa *RemoteKnowledgeBase) FindAuthors(authorName, institutionId string) ([]Author, error) {
	res, err := oa.client.R().
		SetResult(&oaResults[oaAuthor]{}).
		SetQueryParam("filter", fmt.Sprintf("display_name.search:%s,affiliations.institution.id:%s", authorName, institutionId)).
		Get("/authors")

	if err != nil {
		slog.Error("openalex: author search failed", "author", authorName, "institution", institutionId, "error", err)
		return nil, ErrSearchFailed
	}

	if !res.IsSuccess() {
		slog.Error("openalex: author search returned error", "status_code", res.StatusCode(), "body", res.String())
		return nil, ErrSearchFailed
	}

	results := res.Result().(*oaResults[oaAuthor])

	authors := make([]Author, 0, len(results.Results))
	for _, result := range results.Results {
		if result.WorksCount > 0 {
			authors = append(authors, convertOpenalexAuthor(result))
		}
	}

	return authors, nil
}

func (oa *RemoteKnowledgeBase) FindAuthorByOrcidId(orcidId string) (Author, error) {
	res, err := oa.client.R().
		SetResult(&oaResults[oaAuthor]{}).
		SetQueryParam("filter", fmt.Sprintf("orcid:%s", orcidId)).
		Get("/authors")

	if err != nil {
		slog.Error("openalex: author search failed", "orcid_id", orcidId, "error", err)
		return Author{}, ErrSearchFailed
	}

	if !res.IsSuccess() {
		slog.Error("openalex: author search returned error", "status_code", res.StatusCode(), "body", res.String())
		return Author{}, ErrSearchFailed
	}

	results := res.Result().(*oaResults[oaAuthor])

	if len(results.Results) < 1 {
		return Author{}, ErrAuthorNotFound
	}

	return convertOpenalexAuthor(results.Results[0]), nil
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
	PublicationDate string `json:"publication_date"`

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
	AuthorPosition string          `json:"author_position"`
	Author         oaWorkAuthor    `json:"author"`
	Institutions   []oaInstitution `json:"institutions"`
	RawAuthorName  string          `json:"raw_author_name"`
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

func getYearFilter(startDate, endDate time.Time) string {
	return fmt.Sprintf(",from_publication_date:%s,to_publication_date:%s", startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
}

func convertOpenalexWork(work oaWork) Work {
	authors := make([]Author, 0)
	for _, author := range work.Authorships {
		institutions := make([]Institution, 0)
		// Here the author affiliations are not provided, so we use the authorship institutions field
		for _, institution := range author.Institutions {
			institutions = append(institutions, Institution{
				InstitutionName: institution.DisplayName,
				InstitutionId:   institution.Id,
				Location:        institution.CountryCode,
			})
		}
		authors = append(authors, Author{
			AuthorId:      author.Author.Id,
			DisplayName:   author.Author.DisplayName,
			RawAuthorName: &author.RawAuthorName,
			Institutions:  institutions,
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

	publicationDate, err := time.Parse(time.DateOnly, work.PublicationDate)
	if err != nil {
		slog.Error("error parsing openalex publication date", "error", err)
	}

	return Work{
		WorkId:          work.Id,
		DisplayName:     work.DisplayName,
		WorkUrl:         work.getWorkUrl(),
		OaUrl:           work.getOaUrl(),
		DownloadUrl:     work.pdfUrl(),
		PublicationDate: publicationDate,
		Authors:         authors,
		Grants:          grants,
		Locations:       locations,
	}
}

func (oa *RemoteKnowledgeBase) StreamWorks(authorId string, startDate, endDate time.Time) chan WorkBatch {
	outputCh := make(chan WorkBatch, 10)

	cursor := "*"

	yearFilter := getYearFilter(startDate, endDate)

	go func() {
		defer close(outputCh)

		for cursor != "" {
			res, err := oa.client.R().
				SetResult(&oaWorkResults{}).
				SetQueryParam("filter", fmt.Sprintf("authorships.author.id:%s%s", authorId, yearFilter)).
				SetQueryParam("per-page", "200").
				SetQueryParam("cursor", cursor).
				Get("/works")

			if err != nil {
				outputCh <- WorkBatch{Works: nil, TargetAuthorIds: nil, Error: fmt.Errorf("openalex: work search failed: %w", err)}
				break
			}

			if !res.IsSuccess() {
				outputCh <- WorkBatch{Works: nil, TargetAuthorIds: nil, Error: fmt.Errorf("openalex: work search failed: openalex returned status_code=%d body=%s", res.StatusCode(), res.String())}
				break
			}

			results := res.Result().(*oaWorkResults)

			works := make([]Work, 0, len(results.Results))
			for _, work := range results.Results {
				works = append(works, convertOpenalexWork(work))
			}

			outputCh <- WorkBatch{Works: works, TargetAuthorIds: []string{authorId}, Error: nil}

			cursor = results.Meta.NextCursor
		}
	}()

	return outputCh
}

func (oa *RemoteKnowledgeBase) FindWorksByTitle(titles []string, startDate, endDate time.Time) ([]Work, error) {
	works := make([]Work, 0, len(titles))

	yearFilter := getYearFilter(startDate, endDate)

	for _, title := range titles {
		res, err := oa.client.R().
			SetResult(&oaResults[oaWork]{}).
			SetQueryParam("filter", fmt.Sprintf("title.search:%s%s", title, yearFilter)).
			SetQueryParam("per-page", "1").
			Get("/works")

		if err != nil {
			slog.Error("openalex: error searching for work by title", "title", title, "error", err)
			return nil, fmt.Errorf("openalex work search failed: %w", err)
		}

		if !res.IsSuccess() {
			return nil, fmt.Errorf("openalex work search failed: openalex returned status_code=%d body=%s", res.StatusCode(), res.String())
		}

		results := res.Result().(*oaResults[oaWork])

		if len(results.Results) > 0 {
			works = append(works, convertOpenalexWork(results.Results[0]))
		}
	}

	return works, nil
}

func (oa *RemoteKnowledgeBase) GetAuthor(authorId string) (Author, error) {
	res, err := oa.client.R().
		SetResult(&oaResults[oaAuthor]{}).
		SetQueryParam("filter", "openalex:"+authorId).
		Get("authors")

	if err != nil {
		slog.Error("openalex: get author failed", "author_id", authorId, "error", err)
		return Author{}, ErrSearchFailed
	}

	if !res.IsSuccess() {
		slog.Error("openalex: author search returned error", "status_code", res.StatusCode(), "body", res.String())
		return Author{}, ErrSearchFailed
	}

	results := res.Result().(*oaResults[oaAuthor])

	if len(results.Results) < 1 {
		slog.Error("openalex: expected 1 author in get author, got 0", "author_id", authorId)
		return Author{}, fmt.Errorf("no authors returned in get author")
	}

	institutions := make([]Institution, 0, len(results.Results[0].Affiliations))
	for _, institution := range results.Results[0].Affiliations {
		institutions = append(institutions, Institution{
			InstitutionName: institution.Institution.DisplayName,
			InstitutionId:   institution.Institution.Id,
			Location:        institution.Institution.CountryCode,
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

func (oa *RemoteKnowledgeBase) GetInstitutionAuthors(institutionId string, startDate, endDate time.Time) ([]InstitutionAuthor, error) {
	filter := fmt.Sprintf("institutions.id:%s%s", institutionId, getYearFilter(startDate, endDate))
	cursor := "*"

	seen := make(map[string]bool)
	authors := make([]InstitutionAuthor, 0)

	for cursor != "" {
		res, err := oa.client.R().
			SetResult(&oaWorkResults{}).
			SetQueryParam("filter", filter).
			SetQueryParam("cursor", cursor).
			SetQueryParam("per-page", "200").
			Get("/works")

		if err != nil {
			slog.Error("openalex: get institution authors failed", "institution_id", institutionId, "error", err)
			return nil, ErrSearchFailed
		}

		if !res.IsSuccess() {
			slog.Error("openalex: get institution authors returned error", "status_code", res.StatusCode(), "body", res.String())
			return nil, ErrSearchFailed
		}

		result := res.Result().(*oaWorkResults)

		for _, work := range result.Results {
			for _, author := range work.Authorships {
				if author.AuthorPosition == "last" {
					for _, inst := range author.Institutions {
						if inst.Id == institutionId && !seen[author.Author.Id] {
							seen[author.Author.Id] = true
							authors = append(authors, InstitutionAuthor{
								AuthorId:   author.Author.Id,
								AuthorName: author.Author.DisplayName,
							})
							break
						}
					}
				}
			}
		}

		cursor = result.Meta.NextCursor
	}

	return authors, nil
}
