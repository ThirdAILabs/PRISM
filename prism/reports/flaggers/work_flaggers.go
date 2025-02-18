package flaggers

import (
	"fmt"
	"log/slog"
	"prism/prism/openalex"
	"prism/prism/reports/flaggers/eoc"
	"regexp"
	"slices"
	"strings"
)

type WorkFlagger interface {
	Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error)

	Name() flagType
}

type OpenAlexMultipleAffiliationsFlagger struct{}

func (flagger *OpenAlexMultipleAffiliationsFlagger) Name() flagType {
	return OAMultipleAffiliations
}

func (flagger *OpenAlexMultipleAffiliationsFlagger) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error) {
	flags := make([]Flag, 0)

	for _, work := range works {
		logger.Info("processing work", "work_id", work.WorkId, "target_authors", targetAuthorIds)

		for _, author := range work.Authors {
			if len(author.Institutions) > 1 && slices.Contains(targetAuthorIds, author.AuthorId) {
				institutionNames := author.InstitutionNames()
				flags = append(flags, &MultipleAssociationsFlag{
					FlagType:     OAMultipleAffiliations,
					FlagTitle:    "Multiple Affiliations",
					FlagMessage:  fmt.Sprintf("%s has multiple affilitions in work '%s'\n%s", author.DisplayName, work.GetDisplayName(), strings.Join(institutionNames, "\n")),
					Work:         work,
					AuthorName:   author.DisplayName,
					Affiliations: institutionNames,
				})
				logger.Info("found multiple affiliations", "author_id", author.AuthorId, "author_name", author.DisplayName, "institutions", institutionNames)
				break
			}
		}
		logger.Info("processed work", "work_id", work.WorkId, "target_authors", targetAuthorIds)
	}

	logger.Info("flagger completed", "n_flags", len(flags))

	return flags, nil
}

type OpenAlexFunderIsEOC struct {
	concerningFunders  eoc.EocSet
	concerningEntities eoc.EocSet
}

func (flagger *OpenAlexFunderIsEOC) Name() flagType {
	return OAFunderIsEOC
}

func (flagger *OpenAlexFunderIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error) {
	flags := make([]Flag, 0)

	for _, work := range works {
		logger.Info("processing work", "work_id", work.WorkId, "target_authors", targetAuthorIds)

		concerningFunders := make([]string, 0)
		for _, grant := range work.Grants {
			if flagger.concerningEntities.Contains(grant.FunderId) || flagger.concerningFunders.Contains(grant.FunderId) {
				concerningFunders = append(concerningFunders, grant.FunderName)
			}
		}

		if len(concerningFunders) > 0 {
			flags = append(flags, &EOCFundersFlag{
				FlagType:    OAFunderIsEOC,
				FlagTitle:   "Funder is Entity of Concern",
				FlagMessage: fmt.Sprintf("The following funders of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningFunders, "\n")),
				Work:        work,
				Funders:     concerningFunders,
			})
			logger.Info("found concerning funders", "funders", concerningFunders)
		}
		logger.Info("processed work", "work_id", work.WorkId, "target_authors", targetAuthorIds)
	}

	logger.Info("flagger completed", "n_flags", len(flags))

	return flags, nil
}

type OpenAlexPublisherIsEOC struct {
	concerningPublishers eoc.EocSet
}

func (flagger *OpenAlexPublisherIsEOC) Name() flagType {
	return OAPublisherIsEOC
}

func (flagger *OpenAlexPublisherIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error) {
	flags := make([]Flag, 0)

	for _, work := range works {
		logger.Info("processing work", "work_id", work.WorkId, "target_authors", targetAuthorIds)

		concerningPublishers := make([]string, 0)
		for _, loc := range work.Locations {
			if flagger.concerningPublishers.Contains(loc.OrganizationId) {
				concerningPublishers = append(concerningPublishers, loc.OrganizationName)
			}
		}

		if len(concerningPublishers) > 0 {
			flags = append(flags, &EOCPublishersFlag{
				FlagType:    OAPublisherIsEOC,
				FlagTitle:   "Publisher is Entity of Concern",
				FlagMessage: fmt.Sprintf("The following publishers of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningPublishers, "\n")),
				Work:        work,
				Publishers:  concerningPublishers,
			})
			logger.Info("found concerning publishers", "publishers", concerningPublishers)
		}
		logger.Info("processed work", "work_id", work.WorkId, "target_authors", targetAuthorIds)
	}

	logger.Info("flagger completed", "n_flags", len(flags))

	return flags, nil
}

type OpenAlexCoauthorIsEOC struct {
	concerningEntities eoc.EocSet
}

func (flagger *OpenAlexCoauthorIsEOC) Name() flagType {
	return OACoauthorIsEOC
}

func (flagger *OpenAlexCoauthorIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error) {
	flags := make([]Flag, 0)

	for _, work := range works {
		logger.Info("processing work", "work_id", work.WorkId, "target_authors", targetAuthorIds)

		concerningAuthors := make([]string, 0)
		for _, author := range work.Authors {
			if flagger.concerningEntities.Contains(author.AuthorId) {
				concerningAuthors = append(concerningAuthors, author.DisplayName)
			}
		}

		if len(concerningAuthors) > 0 {
			flags = append(flags, &EOCCoauthorsFlag{
				FlagType:    OACoauthorIsEOC,
				FlagTitle:   "Co-author is Entity of Concern",
				FlagMessage: fmt.Sprintf("The following co-authors of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningAuthors, "\n")),
				Work:        work,
				Coauthors:   concerningAuthors,
			})
			logger.Info("found concerning coauthors", "coauthors", concerningAuthors)
		}
		logger.Info("processed work", "work_id", work.WorkId, "target_authors", targetAuthorIds)
	}

	logger.Info("flagger completed", "n_flags", len(flags))

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
	concerningEntities     eoc.EocSet
	concerningInstitutions eoc.EocSet
}

func (flagger *OpenAlexAuthorAffiliationIsEOC) Name() flagType {
	return OAAuthorAffiliationIsEOC
}

func (flagger *OpenAlexAuthorAffiliationIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error) {
	flags := make([]Flag, 0)

	for _, work := range works {
		logger.Info("processing work", "work_id", work.WorkId, "target_authors", targetAuthorIds)

		concerningInstitutions := make(map[string]bool)
		for _, author := range work.Authors {
			if !slices.Contains(targetAuthorIds, author.AuthorId) {
				continue
			}
			for _, institution := range author.Institutions {
				if flagger.concerningEntities.Contains(institution.InstitutionId) ||
					flagger.concerningInstitutions.Contains(institution.InstitutionId) {
					concerningInstitutions[institution.InstitutionName] = true
				}
			}
		}

		if len(concerningInstitutions) > 0 {
			concerningInstitutions := getKeys(concerningInstitutions)
			flags = append(flags, &EOCAuthorAffiliationsFlag{
				FlagType:     OAAuthorAffiliationIsEOC,
				FlagTitle:    "This author is affiliated with entities of concern",
				FlagMessage:  fmt.Sprintf("In '%s', this author is affiliated with entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningInstitutions, "\n")),
				Work:         work,
				Institutions: concerningInstitutions,
			})
			logger.Info("found concerning affiliations", "institutions", concerningInstitutions)
		}
		logger.Info("processed work", "work_id", work.WorkId, "target_authors", targetAuthorIds)
	}

	logger.Info("flagger completed", "n_flags", len(flags))

	return flags, nil
}

type OpenAlexCoauthorAffiliationIsEOC struct {
	concerningEntities     eoc.EocSet
	concerningInstitutions eoc.EocSet
}

func (flagger *OpenAlexCoauthorAffiliationIsEOC) Name() flagType {
	return OACoauthorAffiliationIsEOC
}

func (flagger *OpenAlexCoauthorAffiliationIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error) {
	flags := make([]Flag, 0)

	for _, work := range works {
		logger.Info("processing work", "work_id", work.WorkId, "target_authors", targetAuthorIds)

		concerningInstitutions := make(map[string]bool)
		concerningCoauthors := make(map[string]bool)
		for _, author := range work.Authors {
			if slices.Contains(targetAuthorIds, author.AuthorId) {
				continue
			}
			for _, institution := range author.Institutions {
				if flagger.concerningEntities.Contains(institution.InstitutionId) ||
					flagger.concerningInstitutions.Contains(institution.InstitutionId) {
					concerningInstitutions[institution.InstitutionName] = true
					concerningCoauthors[author.DisplayName] = true
				}
			}
		}

		if len(concerningInstitutions) > 0 {
			concerningCoauthors := getKeys(concerningCoauthors)
			concerningInstitutions := getKeys(concerningInstitutions)
			flags = append(flags, &EOCCoauthorAffiliationsFlag{
				FlagType:     OACoauthorAffiliationIsEOC,
				FlagTitle:    "This co-authors of work are affiliated with entities of concern",
				FlagMessage:  fmt.Sprintf("In '%s', some of the co-authors are affiliated with entities of concern:\n%s\n\nAffiliated authors:\n%s", work.GetDisplayName(), strings.Join(concerningInstitutions, "\n"), strings.Join(concerningCoauthors, "\n")),
				Work:         work,
				Institutions: concerningInstitutions,
				Authors:      concerningCoauthors,
			})
			logger.Info("found concerning coauthor affiliations", "institutions", concerningInstitutions, "coauthors", concerningCoauthors)
		}
		logger.Info("processed work", "work_id", work.WorkId, "target_authors", targetAuthorIds)
	}

	logger.Info("flagger completed", "n_flags", len(flags))

	return flags, nil
}

type cachedAckFlag struct {
	Message  string
	FlagData *EOCAcknowledgemntsFlag
}

type OpenAlexAcknowledgementIsEOC struct {
	openalex     openalex.KnowledgeBase
	entityLookup *EntityStore
	flagCache    DataCache[cachedAckFlag]
	authorCache  DataCache[openalex.Author]
	extractor    AcknowledgementsExtractor
	sussyBakas   []string
}

func (flagger *OpenAlexAcknowledgementIsEOC) Name() flagType {
	return OAAcknowledgementIsEOC
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
				prevEndPos = entity.StartPosition + len(entity.EntityText)
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

func (flagger *OpenAlexAcknowledgementIsEOC) checkAcknowledgementEntities(
	logger *slog.Logger, acknowledgements []Acknowledgement, allAuthorNames []string,
) (bool, map[string]SourceToAliases, string, error) {
	message := ""
	flagged := false

	flaggedEntities := make(map[string]SourceToAliases)

	for _, ack := range acknowledgements {
		nameInAck := false
		for _, name := range allAuthorNames {
			if strings.Contains(ack.RawText, name) {
				logger.Info("author name detected in acknowledgements", "name", name)
				nameInAck = true
				break
			}
		}

		sussyBakaFlag := flagger.checkForSussyBaka(ack)
		if sussyBakaFlag {
			logger.Info("sussy baka detected")
			flagged = true
		}

		logger.Info("ACKNOWLEDGEMENT TEXT", "text", ack.RawText)

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
			matches, err := flagger.entityLookup.SearchEntities(logger, entityQueries)
			if err != nil {
				return false, nil, "", fmt.Errorf("error looking up entity matches: %w", err)
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

	return flagged, flaggedEntities, message, nil
}

func flagCacheKey(workId string, targetAuthorIds []string) string {
	return fmt.Sprintf("%s;%v", workId, targetAuthorIds)
}

func (flagger *OpenAlexAcknowledgementIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]Flag, error) {
	flags := make([]Flag, 0)

	remaining := make([]openalex.Work, 0)

	workIdToWork := make(map[string]openalex.Work)

	for _, work := range works {
		logger.Info("processing work", "work_id", work.WorkId, "target_authors", targetAuthorIds)

		workId := parseOpenAlexId(work)
		if workId == "" {
			logger.Error("unable to parse work id", "work_name", work.DisplayName, "work_id", work.WorkId)
			continue
		}

		workIdToWork[workId] = work

		if cacheEntry := flagger.flagCache.Lookup(flagCacheKey(workId, targetAuthorIds)); cacheEntry != nil {
			logger.Info("found cached entry for work", "work_id", work.WorkId)
			if cacheEntry.FlagData != nil {
				flags = append(flags, &EOCAcknowledgemntsFlag{
					FlagType:           OAAcknowledgementIsEOC,
					FlagTitle:          "Acknowledgements are entities of concern",
					FlagMessage:        cacheEntry.Message,
					Work:               work,
					Entities:           cacheEntry.FlagData.Entities,
					RawAcknowledements: cacheEntry.FlagData.RawAcknowledements,
				})
				logger.Info("cached entry contains flag", "work_id", workId)
			}
			continue
		}

		if work.DownloadUrl == "" {
			logger.Info("work has no download url", "work_id", workId)
			continue
		}

		logger.Info("queueing work for further processing", "work_id", workId)
		remaining = append(remaining, work)
	}

	allAuthorNames, err := flagger.getAuthorNames(targetAuthorIds)
	if err != nil {
		logger.Error("error getting author names", "target_authors", targetAuthorIds, "error", err)
		return nil, fmt.Errorf("error getting author infos: %w", err)
	}

	acknowledgementsStream := flagger.extractor.GetAcknowledgements(logger, remaining)

	for acks := range acknowledgementsStream {
		if acks.Error != nil {
			logger.Error("error retreiving acknowledgments for work", "error", acks.Error)
			continue
		}

		workLogger := logger.With("work_id", acks.Result.WorkId)
		workLogger.Info("processing acknowledgments for work")
		if len(acks.Result.Acknowledgements) == 0 {
			workLogger.Info("work contains no acknowledgments: skipping work")
			continue
		}

		workLogger.Info("found acknowledgements", "n_acks", len(acks.Result.Acknowledgements))

		flagged, flaggedEntities, message, err := flagger.checkAcknowledgementEntities(
			workLogger, acks.Result.Acknowledgements, allAuthorNames,
		)
		if err != nil {
			workLogger.Error("error checking acknowledgements: skipping work", "error", err)
			continue
		}

		workLogger.Info("found flagged entities in acknowledgements", "n_entities", len(flaggedEntities))

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
			flag := &EOCAcknowledgemntsFlag{
				FlagType:           OAAcknowledgementIsEOC,
				FlagTitle:          "Acknowledgements are entities of concern",
				FlagMessage:        fmt.Sprintf("%s\n%s", message, strings.Join(ackTexts, "\n")),
				Work:               workIdToWork[acks.Result.WorkId],
				Entities:           entities,
				RawAcknowledements: ackTexts,
			}

			flagger.flagCache.Update(flagCacheKey(acks.Result.WorkId, targetAuthorIds), cachedAckFlag{
				Message:  msg,
				FlagData: flag,
			})

			flags = append(flags, flag)
		} else {
			flagger.flagCache.Update(flagCacheKey(acks.Result.WorkId, targetAuthorIds), cachedAckFlag{
				FlagData: nil,
			})
		}

		workLogger.Info("processed acknowledgements for work")
	}

	logger.Info("flagger completed", "n_flags", len(flags))

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
