package flaggers

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"prism/prism/api"
	"prism/prism/llms"
	"prism/prism/openalex"
	"prism/prism/reports/flaggers/eoc"
	"prism/prism/search"
	"prism/prism/triangulation"
)

type WorkFlagger interface {
	Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error)

	Name() string

	DisableForUniversityReport() bool
}

func getWorkSummary(w openalex.Work) api.WorkSummary {
	return api.WorkSummary{
		WorkId:          w.WorkId,
		DisplayName:     w.DisplayName,
		WorkUrl:         w.WorkUrl,
		OaUrl:           w.DownloadUrl,
		PublicationDate: w.PublicationDate,
	}
}

type OpenAlexMultipleAffiliationsFlagger struct{}

func (flagger *OpenAlexMultipleAffiliationsFlagger) Name() string {
	return "MultipleAffiliations"
}

func (flagger *OpenAlexMultipleAffiliationsFlagger) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	for _, work := range works {
		for _, author := range work.Authors {
			if len(author.Institutions) > 1 && slices.Contains(targetAuthorIds, author.AuthorId) {
				affiliations := author.InstitutionNames()
				flags = append(flags, &api.MultipleAffiliationFlag{
					Message:      fmt.Sprintf("%s has multiple affilitions in work '%s'\n%s", author.DisplayName, work.GetDisplayName(), strings.Join(affiliations, "\n")),
					Work:         getWorkSummary(work),
					Affiliations: affiliations,
				})
				break
			}
		}
	}

	return flags, nil
}

func (flagger *OpenAlexMultipleAffiliationsFlagger) DisableForUniversityReport() bool {
	return false
}

type OpenAlexFunderIsEOC struct {
	concerningFunders  eoc.EocSet
	concerningEntities eoc.EocSet
}

func (flagger *OpenAlexFunderIsEOC) Name() string {
	return "FunderEOC"
}

func (flagger *OpenAlexFunderIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	for _, work := range works {
		concerningFunders := make([]string, 0)
		for _, grant := range work.Grants {
			if flagger.concerningEntities.Contains(grant.FunderId) || flagger.concerningFunders.Contains(grant.FunderId) {
				concerningFunders = append(concerningFunders, grant.FunderName)
			}
		}

		if len(concerningFunders) > 0 {
			flags = append(flags, &api.HighRiskFunderFlag{
				Message: fmt.Sprintf("The following funders of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningFunders, "\n")),
				Work:    getWorkSummary(work),
				Funders: concerningFunders,
			})
		}
	}

	return flags, nil
}

func (flagger *OpenAlexFunderIsEOC) DisableForUniversityReport() bool {
	return false
}

type OpenAlexPublisherIsEOC struct {
	concerningPublishers eoc.EocSet
}

func (flagger *OpenAlexPublisherIsEOC) Name() string {
	return "PublisherEOC"
}

func (flagger *OpenAlexPublisherIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	for _, work := range works {
		concerningPublishers := make([]string, 0)
		for _, loc := range work.Locations {
			if flagger.concerningPublishers.Contains(loc.OrganizationId) {
				concerningPublishers = append(concerningPublishers, loc.OrganizationName)
			}
		}

		if len(concerningPublishers) > 0 {
			flags = append(flags, &api.HighRiskPublisherFlag{
				Message:    fmt.Sprintf("The following publishers of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningPublishers, "\n")),
				Work:       getWorkSummary(work),
				Publishers: concerningPublishers,
			})
		}
	}

	return flags, nil
}

func (flagger *OpenAlexPublisherIsEOC) DisableForUniversityReport() bool {
	return false
}

type OpenAlexCoauthorIsEOC struct {
	concerningEntities eoc.EocSet
}

func (flagger *OpenAlexCoauthorIsEOC) Name() string {
	return "CoauthorEOC"
}

func (flagger *OpenAlexCoauthorIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	for _, work := range works {
		concerningAuthors := make([]string, 0)
		for _, author := range work.Authors {
			if flagger.concerningEntities.Contains(author.AuthorId) {
				concerningAuthors = append(concerningAuthors, author.DisplayName)
			}
		}

		if len(concerningAuthors) > 0 {
			flags = append(flags, &api.HighRiskCoauthorFlag{
				Message:   fmt.Sprintf("The following co-authors of work '%s' are entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningAuthors, "\n")),
				Work:      getWorkSummary(work),
				Coauthors: concerningAuthors,
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

func (flagger *OpenAlexCoauthorIsEOC) DisableForUniversityReport() bool {
	return false
}

type OpenAlexAuthorAffiliationIsEOC struct {
	concerningEntities     eoc.EocSet
	concerningInstitutions eoc.EocSet
}

func (flagger *OpenAlexAuthorAffiliationIsEOC) Name() string {
	return "AuthorAffiliationEOC"
}

func (flagger *OpenAlexAuthorAffiliationIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	for _, work := range works {
		concerningAffiliations := make(map[string]bool)
		for _, author := range work.Authors {
			if !slices.Contains(targetAuthorIds, author.AuthorId) {
				continue
			}
			for _, institution := range author.Institutions {
				if flagger.concerningEntities.Contains(institution.InstitutionId) ||
					flagger.concerningInstitutions.Contains(institution.InstitutionId) {
					concerningAffiliations[institution.InstitutionName] = true
				}
			}
		}

		if len(concerningAffiliations) > 0 {
			concerningAffiliations := getKeys(concerningAffiliations)
			flags = append(flags, &api.AuthorAffiliationFlag{
				Message:      fmt.Sprintf("In '%s', this author is affiliated with entities of concern:\n%s", work.GetDisplayName(), strings.Join(concerningAffiliations, "\n")),
				Work:         getWorkSummary(work),
				Affiliations: concerningAffiliations,
			})
		}
	}

	return flags, nil
}

func (flagger *OpenAlexAuthorAffiliationIsEOC) DisableForUniversityReport() bool {
	return false
}

type OpenAlexCoauthorAffiliationIsEOC struct {
	concerningEntities     eoc.EocSet
	concerningInstitutions eoc.EocSet
}

func (flagger *OpenAlexCoauthorAffiliationIsEOC) Name() string {
	return "CoauthorAffiliationEOC"
}

func (flagger *OpenAlexCoauthorAffiliationIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	for _, work := range works {
		concerningAffiliations := make(map[string]bool)
		concerningCoauthors := make(map[string]bool)
		for _, author := range work.Authors {
			if slices.Contains(targetAuthorIds, author.AuthorId) {
				continue
			}
			for _, institution := range author.Institutions {
				if flagger.concerningEntities.Contains(institution.InstitutionId) ||
					flagger.concerningInstitutions.Contains(institution.InstitutionId) {
					concerningAffiliations[institution.InstitutionName] = true
					concerningCoauthors[author.DisplayName] = true
				}
			}
		}

		if len(concerningAffiliations) > 0 {
			concerningCoauthors := getKeys(concerningCoauthors)
			concerningAffiliations := getKeys(concerningAffiliations)
			flags = append(flags, &api.CoauthorAffiliationFlag{
				Message:      fmt.Sprintf("In '%s', some of the co-authors are affiliated with entities of concern:\n%s\n\nAffiliated authors:\n%s", work.GetDisplayName(), strings.Join(concerningAffiliations, "\n"), strings.Join(concerningCoauthors, "\n")),
				Work:         getWorkSummary(work),
				Coauthors:    concerningCoauthors,
				Affiliations: concerningAffiliations,
			})
		}
	}

	return flags, nil
}

func (flagger *OpenAlexCoauthorAffiliationIsEOC) DisableForUniversityReport() bool {
	return false
}

func BuildWatchlistEntityIndex(aliasToSource map[string]string) *search.EntityIndex[string] {
	records := make([]search.Record[string], 0, len(aliasToSource))
	for alias, source := range aliasToSource {
		records = append(records, search.Record[string]{Entity: alias, Metadata: source})
	}
	return search.NewIndex(records)
}

type OpenAlexAcknowledgementIsEOC struct {
	openalex        openalex.KnowledgeBase
	entityLookup    *search.EntityIndex[string]
	authorCache     DataCache[openalex.Author]
	extractor       AcknowledgementsExtractor
	sussyBakas      []string
	triangulationDB *triangulation.TriangulationDB
}

func (flagger *OpenAlexAcknowledgementIsEOC) Name() string {
	return "AcknowledgementEOC"
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

func (flagger *OpenAlexAcknowledgementIsEOC) containsSussyBakas(text string) bool {
	text = fmt.Sprintf(" %s ", strings.ToLower(strings.TrimSpace(text)))

	for _, sussyBaka := range flagger.sussyBakas {
		if strings.Contains(text, fmt.Sprintf(" %s ", sussyBaka)) {
			return true
		}
	}
	return false
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

	newText = punctCleaningRe.ReplaceAllString(newText, " ")

	return flagger.containsSussyBakas(newText)
}

type SourceToAliases map[string][]string

func (flagger *OpenAlexAcknowledgementIsEOC) searchWatchlistEntities(entities []string) map[string]SourceToAliases {
	matches := make(map[string]SourceToAliases)

	for _, entity := range entities {
		results := flagger.entityLookup.Query(entity, 10)

		sourceToAliases := make(SourceToAliases)
		for _, result := range results {
			sim := IndelSimilarity(entity, result.Entity)
			if sim > 0.9 {
				sourceToAliases[result.Metadata] = append(sourceToAliases[result.Entity], result.Entity)
			}
		}
		if len(sourceToAliases) > 0 {
			matches[entity] = sourceToAliases
		}
	}

	return matches
}

func (flagger *OpenAlexAcknowledgementIsEOC) checkAcknowledgementEntities(
	acknowledgements []Acknowledgement, allAuthorNames []string,
) (bool, map[string]SourceToAliases, string, error) {
	message := ""
	flagged := false

	flaggedEntities := make(map[string]SourceToAliases)

	for _, ack := range acknowledgements {
		nameInAck := false
		for _, name := range allAuthorNames {
			if strings.Contains(ack.RawText, name) {
				nameInAck = true
				break
			}
		}

		sussyBakaFlag := flagger.checkForSussyBaka(ack)
		if sussyBakaFlag {
			flagged = true
		}

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
			matches := flagger.searchWatchlistEntities(entityQueries)

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

func (flagger *OpenAlexAcknowledgementIsEOC) verifyGrantRecipientWithLLM(authorName string, grantNumber string, acknowledgementText string) (bool, error) {
	llm := llms.New()

	prompt := `Analyze this paper acknowledgment and determine if author %s might be the primary recipient/investigator of grant code %s.

	Important instructions:
	1. The author might be referred to by their initials (e.g., "J.S." or "JS" for "John Smith") in the text.
	2. For grant codes that are listed as supporting general work without specifying who the primary recipient is, assume that the given author could be a primary recipient.
	3. Only return "false" if there is EXPLICIT evidence that someone else is a primary recipient of the grant.
	4. If multiple grants are listed together, and it's not clear who is the primary recipient of which grant, return "true".

	Acknowledgment:
	%s
	`

	res, err := llm.Generate(fmt.Sprintf(prompt, authorName, grantNumber, acknowledgementText), &llms.Options{
		Model:        llms.GPT4oMini,
		ZeroTemp:     true,
		SystemPrompt: "You are a scientific paper analysis assistant who responds with only 'true' or 'false'.",
	})
	if err != nil {
		return true, fmt.Errorf("llm match verification failed: %w", err)
	}

	if strings.Contains(strings.ToLower(res), "true") {
		return true, nil
	}

	return false, nil
}

func (flagger *OpenAlexAcknowledgementIsEOC) checkForGrantRecipient(
	fundCodes map[string]bool, acknowledgements []Acknowledgement, allAuthorNames []string,
) (map[string]map[string]bool, error) {
	triangulationResults := make(map[string]map[string]bool)

	for _, ack := range acknowledgements {
		for _, entity := range ack.SearchableEntities {
			if len(entity.FundCodes) > 0 {
				if flagger.containsSussyBakas(entity.EntityText) {
					if _, ok := triangulationResults[entity.EntityText]; !ok {
						triangulationResults[entity.EntityText] = make(map[string]bool)
					}

					for _, grantNumber := range entity.FundCodes {
						if res, ok := fundCodes[grantNumber]; ok && !res {
							triangulationResults[entity.EntityText][grantNumber] = false
							continue
						}

						for _, authorName := range allAuthorNames {
							result, err := flagger.triangulationDB.IsAuthorGrantRecipient(authorName, grantNumber)
							if err != nil {
								continue
							}

							if result {
								result, err = flagger.verifyGrantRecipientWithLLM(authorName, grantNumber, ack.RawText)
								if err != nil {
									continue
								}
							}
							fundCodes[grantNumber] = result
							triangulationResults[entity.EntityText][grantNumber] = result
							break
						}
					}
				}
			}
		}
	}

	return triangulationResults, nil
}

var talentPrograms = []string{
	"Department of Defense - Foreign Talent Programs that Pose a Threat to National Security Interests of the United States",
	"Foreign Talent Recruitment Programs",
}

var deniedEntities = []string{
	"Chinese Military Companies Operating in the United States",
}

func containsSource(entities []api.AcknowledgementEntity, sourcesOfInterest []string) bool {
	for _, entity := range entities {
		for _, source := range entity.Sources {
			if slices.Contains(sourcesOfInterest, source) {
				return true
			}
		}
	}
	return false
}

func createAcknowledgementFlag(work openalex.Work, message string, entities []api.AcknowledgementEntity, rawAcks []string, triangulationResults map[string]map[string]bool) api.Flag {
	if strings.Contains(message, "talent") || strings.Contains(message, "Talent") || containsSource(entities, talentPrograms) {
		return &api.TalentContractFlag{
			Message:               message,
			Work:                  getWorkSummary(work),
			Entities:              entities,
			RawAcknowledgement:    rawAcks,
			FundCodeTriangulation: triangulationResults,
		}
	} else if containsSource(entities, deniedEntities) {
		return &api.AssociationWithDeniedEntityFlag{
			Message:            message,
			Work:               getWorkSummary(work),
			Entities:           entities,
			RawAcknowledgement: rawAcks,
		}
	} else {
		entityNames := make([]string, 0, len(entities))
		for _, entity := range entities {
			entityNames = append(entityNames, entity.Entity)
		}
		return &api.HighRiskFunderFlag{
			Message:               message,
			Work:                  getWorkSummary(work),
			Funders:               entityNames,
			RawAcknowledgement:    rawAcks,
			FundCodeTriangulation: triangulationResults,
		}
	}
}

func (flagger *OpenAlexAcknowledgementIsEOC) DisableForUniversityReport() bool {
	return true
}

func (flagger *OpenAlexAcknowledgementIsEOC) Flag(logger *slog.Logger, works []openalex.Work, targetAuthorIds []string) ([]api.Flag, error) {
	flags := make([]api.Flag, 0)

	remaining := make([]openalex.Work, 0)

	workIdToWork := make(map[string]openalex.Work)

	for _, work := range works {
		workId := parseOpenAlexId(work)
		if workId == "" {
			logger.Warn("unable to parse work id", "work_name", work.DisplayName, "work_id", work.WorkId)
			continue
		}

		workIdToWork[workId] = work

		if work.DownloadUrl == "" {
			// This is fairly common so we just ignore it and continue
			continue
		}

		remaining = append(remaining, work)
	}

	allAuthorNames, err := flagger.getAuthorNames(targetAuthorIds)
	if err != nil {
		logger.Error("error getting author names", "target_authors", targetAuthorIds, "error", err)
		return nil, fmt.Errorf("error getting author infos: %w", err)
	}

	acknowledgementsStream := flagger.extractor.GetAcknowledgements(logger, remaining)

	fundCodes := make(map[string]bool)

	for acks := range acknowledgementsStream {
		if acks.Error != nil {
			logger.Warn("error retreiving acknowledgments for work", "error", acks.Error)
			continue
		}

		workLogger := logger.With("work_id", acks.Result.WorkId)
		if len(acks.Result.Acknowledgements) == 0 {
			continue
		}

		flagged, flaggedEntities, message, err := flagger.checkAcknowledgementEntities(
			acks.Result.Acknowledgements, allAuthorNames,
		)
		if err != nil {
			workLogger.Error("error checking acknowledgements: skipping work", "error", err)
			continue
		}

		var triangulationResults map[string]map[string]bool

		if flagged {
			var err error
			triangulationResults, err = flagger.checkForGrantRecipient(
				fundCodes, acks.Result.Acknowledgements, allAuthorNames,
			)
			if err != nil {
				workLogger.Error("error checking for grant recipient", "error", err)
				continue
			}
		}

		if flagged {
			ackTexts := make([]string, 0, len(acks.Result.Acknowledgements))
			for _, ack := range acks.Result.Acknowledgements {
				ackTexts = append(ackTexts, ack.RawText)
			}

			entities := make([]api.AcknowledgementEntity, 0, len(flaggedEntities))
			for entity, sourceToAliases := range flaggedEntities {
				sources, allAliases := getAllSourcesAndAliases(sourceToAliases)
				entities = append(entities, api.AcknowledgementEntity{
					Entity:  entity,
					Sources: sources,
					Aliases: allAliases,
				})
			}

			flags = append(flags, createAcknowledgementFlag(
				workIdToWork[acks.Result.WorkId],
				fmt.Sprintf("%s\n%s", message, strings.Join(ackTexts, "\n")),
				entities,
				ackTexts,
				triangulationResults))
		}
	}

	updateFundCodeTriangulation := func(flagFundCodeTriangulation map[string]map[string]bool) {
		for funder, innerMap := range flagFundCodeTriangulation {
			for grantNumber := range innerMap {
				if val, ok := fundCodes[grantNumber]; ok {
					flagFundCodeTriangulation[funder][grantNumber] = val
				}
			}
		}
	}

	for _, flag := range flags {
		switch f := flag.(type) {
		case *api.TalentContractFlag:
			updateFundCodeTriangulation(f.FundCodeTriangulation)
		case *api.HighRiskFunderFlag:
			updateFundCodeTriangulation(f.FundCodeTriangulation)
		default:
			continue
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
