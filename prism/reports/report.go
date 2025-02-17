package reports

import (
	"prism/prism/reports/flaggers"
	"strings"
)

type ConnectionField struct {
	Title       string                `json:"title"`
	Count       int                   `json:"count"`
	Connections []flaggers.Connection `json:"connections"`
	Details     []interface{}         `json:"details"`
	Disclosed   []bool                `json:"disclosed"`
}

// TODO(Anyone): This format is should be simplified and cleaned, doing it like this now for compatability
type ReportContent struct {
	AuthorName  string            `json:"name"`
	RiskScore   int               `json:"risk_score"`
	Connections []ConnectionField `json:"connections"`

	TypeToFlags map[string][]flaggers.Flag `json:"type_to_flag"`
}

func hasForeignTalentProgram(flag *flaggers.EOCAcknowledgemntsFlag) bool {
	if strings.Contains(flag.FlagMessage, "talent") {
		return true
	}

	for _, entity := range flag.Entities {
		for _, source := range entity.Sources {
			if strings.Contains(strings.ToLower(source), "foreign talent") {
				return true
			}
		}
	}
	return false
}

func hasDeniedEntity(flag *flaggers.EOCAcknowledgemntsFlag) bool {
	for _, entity := range flag.Entities {
		for _, source := range entity.Sources {
			if strings.Contains(strings.ToLower(source), "chinese military") {
				return true
			}
		}
	}
	return false
}

func FormatReport(authorname string, flags []flaggers.Flag) ReportContent {
	papersWithForeignTalentPrograms := make([]flaggers.Connection, 0)
	papersWithForeignTalentProgramsDetails := make([]interface{}, 0)

	papersWithDeniedEntities := make([]flaggers.Connection, 0)
	papersWithDeniedEntitiesDetails := make([]interface{}, 0)

	papersWithHighRiskFunding := make([]flaggers.Connection, 0)
	papersWithHighRiskFundingDetails := make([]interface{}, 0)

	papersWithHighRiskInstitutions := make([]flaggers.Connection, 0)
	papersWithHighRiskInstitutionsDetails := make([]interface{}, 0)

	highRiskApptsAtInstitutions := make([]flaggers.Connection, 0)
	highRiskApptsAtInstitutionsDetails := make([]interface{}, 0)

	potentialHighRiskApptsAtInstitutions := make([]flaggers.Connection, 0)
	potentialHighRiskApptsAtInstitutionsDetails := make([]interface{}, 0)

	miscPotentialHighRiskAssoc := make([]flaggers.Connection, 0)
	miscPotentialHighRiskAssocDetails := make([]interface{}, 0)

	typeToFlags := make(map[string][]flaggers.Flag)

	for _, flag := range flags {
		typeToFlags[string(flag.Type())] = append(typeToFlags[string(flag.Type())], flag)
		switch f := flag.(type) {
		case *flaggers.AuthorIsAssociatedWithEOCFlag:
			miscPotentialHighRiskAssoc = append(miscPotentialHighRiskAssoc, f.Connection())
			miscPotentialHighRiskAssocDetails = append(miscPotentialHighRiskAssocDetails, f.Details())

		case *flaggers.EOCCoauthorAffiliationsFlag:
			papersWithHighRiskInstitutions = append(papersWithHighRiskInstitutions, f.Connection())
			papersWithHighRiskInstitutionsDetails = append(papersWithHighRiskInstitutionsDetails, f.Details())

		case *flaggers.AuthorIsFacultyAtEOCFlag:
			potentialHighRiskApptsAtInstitutions = append(potentialHighRiskApptsAtInstitutions, f.Connection())
			potentialHighRiskApptsAtInstitutionsDetails = append(potentialHighRiskApptsAtInstitutionsDetails, f.Details())

		case *flaggers.EOCAcknowledgemntsFlag:
			if hasForeignTalentProgram(f) {
				papersWithForeignTalentPrograms = append(papersWithForeignTalentPrograms, f.Connection())
				papersWithForeignTalentProgramsDetails = append(papersWithForeignTalentProgramsDetails, f.Details())
			} else if hasDeniedEntity(f) {
				papersWithDeniedEntities = append(papersWithDeniedEntities, f.Connection())
				papersWithDeniedEntitiesDetails = append(papersWithDeniedEntitiesDetails, f.Details())
			} else {
				papersWithHighRiskFunding = append(papersWithHighRiskFunding, f.Connection())
				papersWithHighRiskFundingDetails = append(papersWithHighRiskFundingDetails, f.Details())
			}

		case *flaggers.EOCAuthorAffiliationsFlag:
			highRiskApptsAtInstitutions = append(highRiskApptsAtInstitutions, f.Connection())
			highRiskApptsAtInstitutionsDetails = append(highRiskApptsAtInstitutionsDetails, f.Details())

		case *flaggers.EOCFundersFlag:
			papersWithHighRiskFunding = append(papersWithHighRiskFunding, f.Connection())
			papersWithHighRiskFundingDetails = append(papersWithHighRiskFundingDetails, f.Details())
		}
	}

	connections := []ConnectionField{
		{
			Title:       "Papers with foreign talent programs",
			Connections: papersWithForeignTalentPrograms,
			Details:     papersWithForeignTalentProgramsDetails,
		},
		{
			Title:       "Papers with denied entities",
			Connections: papersWithDeniedEntities,
			Details:     papersWithDeniedEntitiesDetails,
		},
		{
			Title:       "Papers with high-risk funding sources",
			Connections: papersWithHighRiskFunding,
			Details:     papersWithHighRiskFundingDetails,
		},
		{
			Title:       "Papers with high-risk foreign institutions",
			Connections: papersWithHighRiskInstitutions,
			Details:     papersWithHighRiskInstitutionsDetails,
		},
		{
			Title:       "High-risk appointments at foreign institutions",
			Connections: highRiskApptsAtInstitutions,
			Details:     highRiskApptsAtInstitutionsDetails,
		},
		{
			Title:       "Potential high-risk appointments at foreign institutions",
			Connections: potentialHighRiskApptsAtInstitutions,
			Details:     potentialHighRiskApptsAtInstitutionsDetails,
		},
		{
			Title:       "Miscellaneous potential high-risk associations",
			Connections: miscPotentialHighRiskAssoc,
			Details:     miscPotentialHighRiskAssocDetails,
		},
	}

	totalScore := 0
	for i, conn := range connections {
		connections[i].Count = len(conn.Connections)
		totalScore += len(conn.Connections)
	}

	return ReportContent{
		AuthorName:  authorname,
		RiskScore:   totalScore,
		Connections: connections,
		TypeToFlags: typeToFlags,
	}
}
