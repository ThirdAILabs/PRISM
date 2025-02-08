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
	OACoauthorIsEOC            = "oa_coauther_eoc"
	OAAuthorAffiliationIsEOC   = "oa_author_affiliation_eoc"
	OACoauthorAffiliationIsEOC = "oa_coauthor_affiliation_eoc"
	OAAcknowledgementIsEOC     = "oa_acknowledgement_eoc"
)

type Flag interface {
	GetType() string

	GetTitle() string

	GetMessage() string
}

/*
 * Author Flags
 */

type AuthorFlag struct {
	FlaggerType string
	Title       string
	Message     string

	// For AuthorIsFacultyAtEOCFlagger
	AuthorIsFacultyAtEOC *AuthorIsFacultyAtEOCFlag

	// For AuthorIsAssociatedWithEOCFlagger
	AuthorIsAssociatedWithEOC *AuthorIsAssociatedWithEOCFlag
}

type AuthorIsFacultyAtEOCFlag struct {
	University    string
	UniversityUrl string
}

type Node struct {
	DocTitle string
	DocUrl   string
}

type AuthorIsAssociatedWithEOCFlag struct {
	DocTitle         string
	DocUrl           string
	DocEntities      []string
	EntityMentioned  string
	Connection       string
	Nodes            []Node
	FrequentCoauthor *string
}

func (flag *AuthorFlag) GetType() string {
	return flag.FlaggerType
}

func (flag *AuthorFlag) GetTitle() string {
	return flag.Title
}

func (flag *AuthorFlag) GetMessage() string {
	return flag.Message
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

/*
 * Work Flags
 */
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

func (flag *WorkFlag) GetType() string {
	return flag.FlaggerType
}

func (flag *WorkFlag) GetTitle() string {
	return flag.Title
}

func (flag *WorkFlag) GetMessage() string {
	return flag.Message
}
