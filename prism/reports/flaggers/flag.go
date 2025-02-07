package flaggers

import "prism/openalex"

// Author Flagger Types
const (
	AuthorIsFacultyAtEOC      = "uni_faculty_eoc"
	AuthorIsAssociatedWithEOC = "doj_press_release_eoc"
)

// Work Flagger Types
const (
	OAMultipleAffiliations     = "oa_multi_affil"
	OAFunderIsEOC              = "oa_funder_eoc"
	OAPublisherIsEOC           = "oa_publisher_eoc"
	OACoathorIsEOC             = "oa_coather_eoc"
	OAAuthorAffiliationIsEOC   = "oa_author_affiliation_eoc"
	OACoauthorAffiliationIsEOC = "oa_coauthor_affiliation_eoc"
	OAAcknowledgementIsEOC     = "oa_acknowledgement_eoc"
)

type Flag struct {
	FlaggerType   string
	Title         string
	Message       string
	UniversityUrl string
	Affiliations  []string
	Metadata      map[string]any
}

type WorkFlag struct {
	FlaggerType string
	Title       string
	Message     string
	AuthorIds   []string
	Work        openalex.Work

	// For OpenAlexMultipleAffiliationsFlagger
	MultipleAssociations *MultipleAssociationsFlag

	// For OpenAlexFunderIsEOC
	EOCFunders *EOCFundersFlag

	// For OpenAlexPublisherIsEOC
	EOCPublishers *EOCPublishersFlag

	// For OpenAlexCoauthorIsEOC
	EOCCoauthors *EOCCoauthorsFlag

	// For OpenAlexAuthorAffiliationIsEOC
	EOCAuthorAffiliations *EOCAuthorAffiliationsFlag

	// For OpenAlexCoauthorAffiliationIsEOC
	EOCCoauthorAffiliations *EOCCoauthorAffiliationsFlag

	// For OpenAlexAcknowledgementIsEOC
	EOCAcknowledgemnts *EOCAcknowledgemntsFlag
}

type MultipleAssociationsFlag struct {
	AuthorName   string
	Affiliations []string
}

type EOCFundersFlag struct {
	Funders []string
}

type EOCPublishersFlag struct {
	Publishers []string
}

type EOCCoauthorsFlag struct {
	Coauthors []string
}

type EOCAuthorAffiliationsFlag struct {
	Institutions []string
}

type EOCCoauthorAffiliationsFlag struct {
	Institutions []string
	Authors      []string
}

type EOCAcknowledgementEntity struct {
	Entity  string
	Sources []string
	Aliases []string
}

type EOCAcknowledgemntsFlag struct {
	Entities           []EOCAcknowledgementEntity
	RawAcknowledements []string
}
