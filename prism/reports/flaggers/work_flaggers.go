package flaggers

import (
	"fmt"
	"prism/openalex"
	"slices"
	"strings"
)

type OpenAlexMultipleAffiliationsFlagger struct{}

func (flagger *OpenAlexMultipleAffiliationsFlagger) Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error) {
	flags := make([]WorkFlag, 0)

	for _, work := range works {
		for _, author := range work.Authors {
			if len(author.Institutions) > 1 && slices.Contains(targetAuthorIds, author.AuthorId) {
				institutionNames := author.InstitutionNames()
				flags = append(flags, WorkFlag{
					FlaggerType: OAMultipleAffiliations,
					Title:       "Multiple Affiliations",
					Message:     fmt.Sprintf("%s has multiple affilitions in work '%s'\n%s", author.DisplayName, work.GetDisplayName(), strings.Join(institutionNames, "\n")),
					AuthorIds:   targetAuthorIds,
					Work:        work,
					MultipleAssociations: &MultipleAssociationsFlag{
						AuthorName:   author.DisplayName,
						Affiliations: institutionNames,
					},
				})
			}
			break
		}
	}

	return flags, nil
}

type eocSet map[string]struct{}

func (s *eocSet) contains(entity string) bool {
	_, exists := (*s)[entity]
	return exists
}

type OpenAlexFunderIsEOC struct {
	concerningFunders  eocSet
	concerningEntities eocSet
}

func (flagger *OpenAlexFunderIsEOC) Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error) {
	flags := make([]WorkFlag, 0)

	for _, work := range works {
		concerningFunders := make([]string, 0)
		for _, grant := range work.Grants {
			if flagger.concerningEntities.contains(grant.FunderId) || flagger.concerningFunders.contains(grant.FunderId) {
				concerningFunders = append(concerningFunders, grant.FunderName)
			}
		}

		if len(concerningFunders) > 0 {
			flags = append(flags, WorkFlag{
				FlaggerType: OAFunderIsEOC,
				Title:       "Funder is Entity of Concern",
				Message:     fmt.Sprintf("The following funders of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningFunders, "\n")),
				AuthorIds:   targetAuthorIds,
				Work:        work,
				EOCFunders: &EOCFundersFlag{
					Funders: concerningFunders,
				},
			})
		}
	}

	return flags, nil
}

type OpenAlexPublisherIsEOC struct {
	concerningPublishers eocSet
}

func (flagger *OpenAlexPublisherIsEOC) Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error) {
	flags := make([]WorkFlag, 0)

	for _, work := range works {
		concerningPublishers := make([]string, 0)
		for _, loc := range work.Locations {
			if flagger.concerningPublishers.contains(loc.OrganizationId) {
				concerningPublishers = append(concerningPublishers, loc.OrganizationName)
			}
		}

		if len(concerningPublishers) > 0 {
			flags = append(flags, WorkFlag{
				FlaggerType: OAPublisherIsEOC,
				Title:       "Publisher is Entity of Concern",
				Message:     fmt.Sprintf("The following publishers of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningPublishers, "\n")),
				AuthorIds:   targetAuthorIds,
				Work:        work,
				EOCPublishers: &EOCPublishersFlag{
					Publishers: concerningPublishers,
				},
			})
		}
	}

	return flags, nil
}

type OpenAlexCoauthorIsEOC struct {
	concerningEntities eocSet
}

func (flagger *OpenAlexCoauthorIsEOC) Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error) {
	flags := make([]WorkFlag, 0)

	for _, work := range works {
		concerningAuthors := make([]string, 0)
		for _, author := range work.Authors {
			if flagger.concerningEntities.contains(author.AuthorId) {
				concerningAuthors = append(concerningAuthors, author.DisplayName)
			}
		}

		if len(concerningAuthors) > 0 {
			flags = append(flags, WorkFlag{
				FlaggerType: OACoathorIsEOC,
				Title:       "Co-author is Entity of Concern",
				Message:     fmt.Sprintf("The following co-authors of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningAuthors, "\n")),
				AuthorIds:   targetAuthorIds,
				Work:        work,
				EOCCoauthors: &EOCCoauthorsFlag{
					Coauthors: concerningAuthors,
				},
			})
		}
	}

	return flags, nil
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

type OpenAlexAuthorAffiliationIsEOC struct {
	concerningEntities     eocSet
	concerningInstitutions eocSet
}

func (flagger *OpenAlexAuthorAffiliationIsEOC) Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error) {
	flags := make([]WorkFlag, 0)

	for _, work := range works {
		concerningInstitutions := make(map[string]bool)
		for _, author := range work.Authors {
			if !slices.Contains(targetAuthorIds, author.AuthorId) {
				continue
			}
			for _, institution := range author.Institutions {
				if flagger.concerningEntities.contains(institution.InstitutionId) ||
					flagger.concerningInstitutions.contains(institution.InstitutionId) {
					concerningInstitutions[institution.InstitutionName] = true
				}
			}
		}

		if len(concerningInstitutions) > 0 {
			concerningInstitutions := getKeys(concerningInstitutions)
			flags = append(flags, WorkFlag{
				FlaggerType: OAAuthorAffiliationIsEOC,
				Title:       "This author is affiliated with entities of concern",
				Message:     fmt.Sprintf("In '%s', this author is affiliated with entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningInstitutions, "\n")),
				AuthorIds:   targetAuthorIds,
				Work:        work,
				EOCAuthorAffiliations: &EOCAuthorAffiliationsFlag{
					Institutions: concerningInstitutions,
				},
			})
		}
	}

	return flags, nil
}

type OpenAlexCoauthorAffiliationIsEOC struct {
	concerningEntities     eocSet
	concerningInstitutions eocSet
}

func (flagger *OpenAlexCoauthorAffiliationIsEOC) Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error) {
	flags := make([]WorkFlag, 0)

	for _, work := range works {
		concerningInstitutions := make(map[string]bool)
		concerningCoauthors := make(map[string]bool)
		for _, author := range work.Authors {
			if slices.Contains(targetAuthorIds, author.AuthorId) {
				continue
			}
			for _, institution := range author.Institutions {
				if flagger.concerningEntities.contains(institution.InstitutionId) ||
					flagger.concerningInstitutions.contains(institution.InstitutionId) {
					concerningInstitutions[institution.InstitutionName] = true
					concerningCoauthors[author.DisplayName] = true
				}
			}
		}

		if len(concerningInstitutions) > 0 {
			concerningCoauthors := getKeys(concerningCoauthors)
			concerningInstitutions := getKeys(concerningInstitutions)
			flags = append(flags, WorkFlag{
				FlaggerType: OACoauthorAffiliationIsEOC,
				Title:       "This co-authors of work are affiliated with entities of concern",
				Message:     fmt.Sprintf("In '%s', some of the co-authors are affiliated with entities of concern:\n%s\n\nAffiliated authors:\n%s", work.GetDisplayName(), strings.Join(concerningInstitutions, "\n"), strings.Join(concerningCoauthors, "\n")),
				AuthorIds:   targetAuthorIds,
				Work:        work,
				EOCCoauthorAffiliations: &EOCCoauthorAffiliationsFlag{
					Institutions: concerningInstitutions,
					Authors:      concerningCoauthors,
				},
			})
		}
	}

	return flags, nil
}
