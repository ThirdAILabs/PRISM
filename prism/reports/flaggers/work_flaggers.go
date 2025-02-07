package flaggers

import (
	"fmt"
	"log/slog"
	"prism/openalex"
	"regexp"
	"slices"
	"strings"
)

type WorkFlagger interface {
	Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error)

	Name() string
}

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

type cachedAckFlag struct {
	Message  string
	FlagData *EOCAcknowledgemntsFlag
}

type OpenAlexAcknowledgementIsEOC struct {
	openalex     openalex.KnowledgeBase
	entityLookup EntityStore
	flagCache    DataCache[cachedAckFlag]
	authorCache  DataCache[openalex.Author]
	extractor    AcknowledgementsExtractor
	sussyBakas   []string
}

func (flagger *OpenAlexAcknowledgementIsEOC) getAuthorNames(authorIds []string) ([]string, error) {
	authorNames := make([]string, 0, len(authorIds))

	for _, authorId := range authorIds {

		if cachedAuthor := flagger.authorCache.Lookup(authorId); cachedAuthor != nil {
			authorNames = append(authorNames, cachedAuthor.DisplayName)
			authorNames = append(authorNames, cachedAuthor.DisplayNameAlternatives...)
			authorNames = append(authorNames, getInitialsCombinations(cachedAuthor.DisplayName)...)
			continue
		}

		authorInfo, err := flagger.openalex.GetAuthor(authorId)
		if err != nil {
			slog.Error("error getting author info from open alex", "author_id", authorId, "error", err)
			return nil, fmt.Errorf("error retrieving author info: %w", err)
		}

		authorNames = append(authorNames, authorInfo.DisplayName)
		authorNames = append(authorNames, authorInfo.DisplayNameAlternatives...)
		authorNames = append(authorNames, getInitialsCombinations(authorInfo.DisplayName)...)

		flagger.authorCache.Update(authorId, authorInfo)
	}

	return authorNames, nil
}

var punctCleaningRe = regexp.MustCompile(`[.,()!?:"']`)

func (flagger *OpenAlexAcknowledgementIsEOC) checkForSussyBaka(ack Acknowledgement) bool {
	slices.SortFunc(ack.MiscEntities, func(a, b Entity) int {
		if a.StartPosition < b.StartPosition {
			return -1
		}
		if b.StartPosition > a.StartPosition {
			return 1
		}
		return 0
	})

	prevEndPos := 0
	newText := ""
	for _, entity := range ack.MiscEntities {
		if entity.EntityType == "person" {
			if entity.StartPosition >= prevEndPos {
				newText += ack.RawText[prevEndPos:entity.StartPosition]
				prevEndPos += entity.StartPosition + len(entity.EntityText)
			}
		}
	}
	newText += ack.RawText[prevEndPos:]

	newText = strings.ToLower(strings.TrimSpace(newText))
	newText = punctCleaningRe.ReplaceAllString(newText, " ")
	newText = fmt.Sprintf(" %s ", newText)

	for _, sussyBaka := range flagger.sussyBakas {
		if strings.Contains(newText, fmt.Sprintf(" %s ", sussyBaka)) {
			return true
		}
	}
	return false
}

func flagCacheKey(workId string, targetAuthorIds []string) string {
	return fmt.Sprintf("%s;%v", workId, targetAuthorIds)
}

func (flagger *OpenAlexAcknowledgementIsEOC) Flag(works []openalex.Work, targetAuthorIds []string) ([]WorkFlag, error) {
	flags := make([]WorkFlag, 0)

	remaining := make([]openalex.Work, 0)

	workIdToWork := make(map[string]openalex.Work)

	for _, work := range works {
		workId := parseOpenAlexId(work)
		if workId == "" {
			continue
		}

		workIdToWork[workId] = work

		if cacheEntry := flagger.flagCache.Lookup(flagCacheKey(workId, targetAuthorIds)); cacheEntry != nil {
			if cacheEntry.FlagData != nil {
				flags = append(flags, WorkFlag{
					FlaggerType:        OAAcknowledgementIsEOC,
					Title:              "Acknowledgements are entities of concern",
					Message:            cacheEntry.Message,
					AuthorIds:          targetAuthorIds,
					Work:               work,
					EOCAcknowledgemnts: cacheEntry.FlagData,
				})
			}
			continue
		}

		remaining = append(remaining, work)
	}

	allAuthorNames, err := flagger.getAuthorNames(targetAuthorIds)
	if err != nil {
		return nil, fmt.Errorf("error getting author infos: %w", err)
	}

	acknowledgementsStream := flagger.extractor.GetAcknowledgements(remaining)

	for acks := range acknowledgementsStream {
		if acks.Error != nil {
			slog.Error("error retreiving acknowledgments for work", "error", acks.Error)
			continue
		}
		if len(acks.Result.Acknowledgements) == 0 {
			continue
		}

		var message string
		flagged := false

		flaggedEntities := make(map[string]SourceToAliases)

		for _, ack := range acks.Result.Acknowledgements {
			nameInAck := false
			for _, name := range allAuthorNames {
				if strings.Contains(ack.RawText, name) {
					nameInAck = true
					break
				}
			}

			sussyBakaFlag := flagger.checkForSussyBaka(ack)
			flagged = flagged || sussyBakaFlag

			entityQueries := make([]string, 0)
			if sussyBakaFlag {
				for _, entity := range ack.SearchableEntities {
					entityQueries = append(entityQueries, entity.EntityText)
				}
			}

			if nameInAck && !flagged {
				for _, entity := range ack.SearchableEntities {
					entityQueries = append(entityQueries, entity.EntityText)
				}
				for _, entity := range ack.MiscEntities {
					entityQueries = append(entityQueries, entity.EntityText)
				}
			}

			if len(entityQueries) > 0 {
				matches, err := flagger.entityLookup.SearchEntities(entityQueries)
				if err != nil {
					return nil, fmt.Errorf("error looking up entity matches: %w", err)
				}

				for _, entity := range entityQueries {
					if sources, ok := matches[entity]; ok {
						message += messageFromAcknowledgmentMatches(entity, sources)
						flagged = true
						flaggedEntities[entity] = sources
					}
				}
			}
		}

		if flagged {
			ackTexts := make([]string, 0, len(acks.Result.Acknowledgements))
			for _, ack := range acks.Result.Acknowledgements {
				ackTexts = append(ackTexts, ack.RawText)
			}

			entities := make([]EOCAcknowledgementEntity, 0, len(flaggedEntities))
			for entity, sourceToAliases := range flaggedEntities {
				sources, allAliases := getAllSourcesAndAliases(sourceToAliases)
				entities = append(entities, EOCAcknowledgementEntity{
					Entity:  entity,
					Sources: sources,
					Aliases: allAliases,
				})
			}

			msg := fmt.Sprintf("%s\n%s", message, strings.Join(ackTexts, "\n"))
			flag := WorkFlag{
				FlaggerType: OAAcknowledgementIsEOC,
				Title:       "Acknowledgements are entities of concern",
				Message:     fmt.Sprintf("%s\n%s", message, strings.Join(ackTexts, "\n")),
				AuthorIds:   targetAuthorIds,
				Work:        workIdToWork[acks.Result.WorkId],
				EOCAcknowledgemnts: &EOCAcknowledgemntsFlag{
					Entities:           entities,
					RawAcknowledements: ackTexts,
				},
			}

			flagger.flagCache.Update(flagCacheKey(acks.Result.WorkId, targetAuthorIds), cachedAckFlag{
				Message:  msg,
				FlagData: flag.EOCAcknowledgemnts,
			})

			flags = append(flags, flag)
		} else {
			flagger.flagCache.Update(flagCacheKey(acks.Result.WorkId, targetAuthorIds), cachedAckFlag{
				FlagData: nil,
			})
		}
	}

	return flags, nil
}

func getAllSourcesAndAliases(matches SourceToAliases) ([]string, []string) {
	sources := make([]string, 0, len(matches))
	aliases := make([]string, 0, len(matches))

	for k, v := range matches {
		sources = append(sources, k)
		aliases = append(aliases, v...)
	}

	return sources, aliases
}

func messageFromAcknowledgmentMatches(entity string, matches SourceToAliases) string {
	sources, aliases := getAllSourcesAndAliases(matches)

	return fmt.Sprintf("'%s' was in %s as %s\n", entity, strings.Join(sources, ", "), strings.Join(aliases, ", "))
}
