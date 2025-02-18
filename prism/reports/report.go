package reports

import (
	"prism/prism/reports/flaggers"
	"strings"
)

type ConnectionField struct {
	Title       string                `json:"title"`
	Count       int                   `json:"count"`
	Connections []flaggers.Connection `json:"connections"`
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
	papersWithDeniedEntities := make([]flaggers.Connection, 0)
	papersWithHighRiskFunding := make([]flaggers.Connection, 0)
	papersWithHighRiskInstitutions := make([]flaggers.Connection, 0)
	highRiskApptsAtInstitutions := make([]flaggers.Connection, 0)
	potentialHighRiskApptsAtInstitutions := make([]flaggers.Connection, 0)
	miscPotentialHighRiskAssoc := make([]flaggers.Connection, 0)

	typeToFlags := TypeToFlags{}

	for _, flag := range flags {
		switch flag := flag.(type) {
		case *flaggers.AuthorIsAssociatedWithEOCFlag:
			typeToFlags.AuthorAssociationsEOC = append(typeToFlags.AuthorAssociationsEOC, flag)
			miscPotentialHighRiskAssoc = append(miscPotentialHighRiskAssoc, flag.Connection())

		case *flaggers.EOCCoauthorAffiliationsFlag:
			typeToFlags.CoauthorAffiliationEOC = append(typeToFlags.CoauthorAffiliationEOC, flag)
			papersWithHighRiskInstitutions = append(papersWithHighRiskInstitutions, flag.Connection())

		case *flaggers.AuthorIsFacultyAtEOCFlag:
			typeToFlags.AuthorFacultyAtEOC = append(typeToFlags.AuthorFacultyAtEOC, flag)
			potentialHighRiskApptsAtInstitutions = append(potentialHighRiskApptsAtInstitutions, flag.Connection())

		case *flaggers.EOCAcknowledgemntsFlag:
			typeToFlags.AcknowledgementEOC = append(typeToFlags.AcknowledgementEOC, flag)
			if hasForeignTalentProgram(flag) {
				papersWithForeignTalentPrograms = append(papersWithForeignTalentPrograms, flag.Connection())
			} else if hasDeniedEntity(flag) {
				papersWithDeniedEntities = append(papersWithDeniedEntities, flag.Connection())
			} else {
				papersWithHighRiskFunding = append(papersWithHighRiskFunding, flag.Connection())
			}
		case *flaggers.EOCAuthorAffiliationsFlag:
			typeToFlags.AuthorAffiliationEOC = append(typeToFlags.AuthorAffiliationEOC, flag)
			highRiskApptsAtInstitutions = append(highRiskApptsAtInstitutions, flag.Connection())

		case *flaggers.EOCFundersFlag:
			typeToFlags.FunderEOC = append(typeToFlags.FunderEOC, flag)
			papersWithHighRiskFunding = append(papersWithHighRiskFunding, flag.Connection())
		}
	}

	connections := []ConnectionField{
		{
			Title:       "Papers with foreign talent programs",
			Connections: papersWithForeignTalentPrograms,
		},
		{
			Title:       "Papers with denied entities",
			Connections: papersWithDeniedEntities,
		},
		{
			Title:       "Papers with high-risk funding sources",
			Connections: papersWithHighRiskFunding,
		},
		{
			Title:       "Papers with high-risk foreign institutions",
			Connections: papersWithHighRiskInstitutions,
		},
		{
			Title:       "High-risk appointments at foreign institutions",
			Connections: highRiskApptsAtInstitutions,
		},
		{
			Title:       "Potential high-risk appointments at foreign institutions",
			Connections: potentialHighRiskApptsAtInstitutions,
		},
		{
			Title:       "Miscellaneous potential high-risk associations",
			Connections: miscPotentialHighRiskAssoc,
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
