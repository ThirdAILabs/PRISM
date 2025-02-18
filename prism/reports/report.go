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

type TypeToFlags struct {
	AuthorAssociationsEOC  []*flaggers.AuthorIsAssociatedWithEOCFlag `json:"doj_press_release_eoc"`
	CoauthorAffiliationEOC []*flaggers.EOCCoauthorAffiliationsFlag   `json:"oa_coauthor_affiliation_eoc"`
	AuthorFacultyAtEOC     []*flaggers.AuthorIsFacultyAtEOCFlag      `json:"uni_faculty_eoc"`
	AcknowledgementEOC     []*flaggers.EOCAcknowledgemntsFlag        `json:"oa_acknowledgement_eoc"`
	AuthorAffiliationEOC   []*flaggers.EOCAuthorAffiliationsFlag     `json:"oa_author_affiliation_eoc"`
	FunderEOC              []*flaggers.EOCFundersFlag                `json:"oa_funder_eoc"`
}

// TODO(Anyone): This format is should be simplified and cleaned, doing it like this now for compatability
type ReportContent struct {
	AuthorName  string            `json:"name"`
	RiskScore   int               `json:"risk_score"`
	Connections []ConnectionField `json:"connections"`

	TypeToFlags TypeToFlags `json:"type_to_flag"`
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

	typeToFlags := TypeToFlags{}

	for _, flag := range flags {
		switch flag := flag.(type) {
		case *flaggers.AuthorIsAssociatedWithEOCFlag:
			miscPotentialHighRiskAssoc = append(miscPotentialHighRiskAssoc, flag.Connection())
			miscPotentialHighRiskAssocDetails = append(miscPotentialHighRiskAssocDetails, flag.Details())
			typeToFlags.AuthorAssociationsEOC = append(typeToFlags.AuthorAssociationsEOC, flag)

		case *flaggers.EOCCoauthorAffiliationsFlag:
			papersWithHighRiskInstitutions = append(papersWithHighRiskInstitutions, flag.Connection())
			papersWithHighRiskInstitutionsDetails = append(papersWithHighRiskInstitutionsDetails, flag.Details())
			typeToFlags.CoauthorAffiliationEOC = append(typeToFlags.CoauthorAffiliationEOC, flag)

		case *flaggers.AuthorIsFacultyAtEOCFlag:
			potentialHighRiskApptsAtInstitutions = append(potentialHighRiskApptsAtInstitutions, flag.Connection())
			potentialHighRiskApptsAtInstitutionsDetails = append(potentialHighRiskApptsAtInstitutionsDetails, flag.Details())
			typeToFlags.AuthorFacultyAtEOC = append(typeToFlags.AuthorFacultyAtEOC, flag)

		case *flaggers.EOCAcknowledgemntsFlag:
			typeToFlags.AcknowledgementEOC = append(typeToFlags.AcknowledgementEOC, flag)
			if hasForeignTalentProgram(flag) {
				papersWithForeignTalentPrograms = append(papersWithForeignTalentPrograms, flag.Connection())
				papersWithForeignTalentProgramsDetails = append(papersWithForeignTalentProgramsDetails, flag.Details())
			} else if hasDeniedEntity(flag) {
				papersWithDeniedEntities = append(papersWithDeniedEntities, flag.Connection())
				papersWithDeniedEntitiesDetails = append(papersWithDeniedEntitiesDetails, flag.Details())
			} else {
				papersWithHighRiskFunding = append(papersWithHighRiskFunding, flag.Connection())
				papersWithHighRiskFundingDetails = append(papersWithHighRiskFundingDetails, flag.Details())
			}

		case *flaggers.EOCAuthorAffiliationsFlag:
			typeToFlags.AuthorAffiliationEOC = append(typeToFlags.AuthorAffiliationEOC, flag)
			highRiskApptsAtInstitutions = append(highRiskApptsAtInstitutions, flag.Connection())
			highRiskApptsAtInstitutionsDetails = append(highRiskApptsAtInstitutionsDetails, flag.Details())

		case *flaggers.EOCFundersFlag:
			papersWithHighRiskFunding = append(papersWithHighRiskFunding, flag.Connection())
			papersWithHighRiskFundingDetails = append(papersWithHighRiskFundingDetails, flag.Details())
			typeToFlags.FunderEOC = append(typeToFlags.FunderEOC, flag)
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
