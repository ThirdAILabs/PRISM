package flaggers

import (
	"fmt"
	"prism/prism/openalex"
	"strings"
)

type flagType string

// Author Flagger Types
const (
	AuthorIsFacultyAtEOC      flagType = "uni_faculty_eoc"
	AuthorIsAssociatedWithEOC flagType = "doj_press_release_eoc"
)

// Work Flagger Types
const (
	OAMultipleAffiliations     flagType = "oa_multi_affil" // Not used right now
	OAFunderIsEOC              flagType = "oa_funder_eoc"
	OAPublisherIsEOC           flagType = "oa_publisher_eoc" // Not used right now
	OACoauthorIsEOC            flagType = "oa_coauther_eoc"  // Not used right now
	OAAuthorAffiliationIsEOC   flagType = "oa_author_affiliation_eoc"
	OACoauthorAffiliationIsEOC flagType = "oa_coauthor_affiliation_eoc"
	OAAcknowledgementIsEOC     flagType = "oa_acknowledgement_eoc"
)

type Connection struct {
	Title       string       `json:"title"`
	Url         string       `json:"url"`
	Connections []Connection `json:"connections"`
}

type Flag interface {
	Type() flagType

	Connection() Connection

	Details() interface{}

	// This is used to deduplicate flags. Primarily for author flags, it is
	// possible to have the same flag created for multiple works, for instance by
	// finding the author is faculty at an EOC. For work flags, the key is just
	// the flagger type and work id since we can only have 1 flag for a given work.
	Key() string
}

/*
 * Author Flags
 */

type AuthorIsFacultyAtEOCFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	University    string
	UniversityUrl string
}

func (flag *AuthorIsFacultyAtEOCFlag) Type() flagType {
	return AuthorIsFacultyAtEOC
}

func (flag *AuthorIsFacultyAtEOCFlag) Connection() Connection {
	return Connection{
		Title: flag.University,
		Url:   flag.UniversityUrl,
	}
}

func (flag *AuthorIsFacultyAtEOCFlag) Details() interface{} {
	return struct {
		University    string `json:"university"`
		UniversityUrl string `json:"university_url"`
	}{
		University:    flag.University,
		UniversityUrl: flag.UniversityUrl,
	}
}

func (flag *AuthorIsFacultyAtEOCFlag) Key() string {
	return fmt.Sprintf("%v-%s-%s", flag.FlagType, flag.University, flag.UniversityUrl)
}

type Node struct {
	DocTitle string
	DocUrl   string
}

type AuthorIsAssociatedWithEOCFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	DocTitle         string
	DocUrl           string
	DocEntities      []string
	EntityMentioned  string
	ConnectionLevel  string
	Nodes            []Node
	FrequentCoauthor *string
}

func (flag *AuthorIsAssociatedWithEOCFlag) Type() flagType {
	return AuthorIsAssociatedWithEOC
}

func (flag *AuthorIsAssociatedWithEOCFlag) Connection() Connection {
	connection := Connection{
		Title: flag.DocTitle,
		Url:   flag.DocUrl,
	}

	for i := len(flag.Nodes) - 1; i >= 0; i-- {
		connection = Connection{
			Title:       flag.Nodes[i].DocTitle,
			Url:         flag.Nodes[i].DocUrl,
			Connections: []Connection{connection},
		}
	}

	return connection
}

func (flag *AuthorIsAssociatedWithEOCFlag) Details() interface{} {
	return struct {
		DocTitle         string   `json:"doc_title"`
		DocUrl           string   `json:"doc_url"`
		DocEntities      []string `json:"doc_entities"`
		EntityMentioned  string   `json:"entity_mentioned"`
		ConnectionLevel  string   `json:"connection_level"`
		FrequentCoauthor *string  `json:"frequent_coauthor,omitempty"`
	}{
		DocTitle:         flag.DocTitle,
		DocUrl:           flag.DocUrl,
		DocEntities:      flag.DocEntities,
		EntityMentioned:  flag.EntityMentioned,
		ConnectionLevel:  flag.ConnectionLevel,
		FrequentCoauthor: flag.FrequentCoauthor,
	}
}

func (flag *AuthorIsAssociatedWithEOCFlag) Key() string {
	return fmt.Sprintf("%v-%s-%s-%v", flag.FlagType, flag.DocTitle, flag.EntityMentioned, flag.Nodes)
}

/*
 * Work Flags
 */

func workFlagKey(flagType flagType, workId string) string {
	// Note: this assumes that only 1 flag of a given type is created per work.
	// If that changes this should be updated.
	return fmt.Sprintf("%v-%s", flagType, workId)
}

type MultipleAssociationsFlag struct { // This flag is deprecated
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	Work openalex.Work

	AuthorName   string
	Affiliations []string
}

func (flag *MultipleAssociationsFlag) Type() flagType {
	return OAMultipleAffiliations
}

func (flag *MultipleAssociationsFlag) Connection() Connection {
	return Connection{
		Title: flag.FlagMessage,
	}
}

func (flag *MultipleAssociationsFlag) Details() interface{} {
	return struct {
		AuthorName   string   `json:"author_name"`
		Affiliations []string `json:"affiliations"`
	}{
		AuthorName:   flag.AuthorName,
		Affiliations: flag.Affiliations,
	}
}

func (flag *MultipleAssociationsFlag) Key() string {
	return workFlagKey(flag.FlagType, flag.Work.WorkId)
}

type EOCFundersFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	Work openalex.Work

	Funders []string
}

func (flag *EOCFundersFlag) Type() flagType {
	return OAFunderIsEOC
}

func (flag *EOCFundersFlag) Connection() Connection {
	return Connection{
		Title: flag.Work.DisplayName,
		Url:   flag.Work.WorkUrl,
	}
}

func (flag *EOCFundersFlag) Details() interface{} {
	return struct {
		Work    openalex.Work `json:"work"`
		Funders []string      `json:"funders"`
	}{
		Work:    flag.Work,
		Funders: flag.Funders,
	}
}

func (flag *EOCFundersFlag) Key() string {
	return workFlagKey(flag.FlagType, flag.Work.WorkId)
}

type EOCPublishersFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	Work openalex.Work

	Publishers []string
}

func (flag *EOCPublishersFlag) Type() flagType {
	return OAPublisherIsEOC
}

func (flag *EOCPublishersFlag) Connection() Connection {
	return Connection{
		Title: flag.Work.DisplayName,
		Url:   flag.Work.WorkUrl,
	}
}

func (flag *EOCPublishersFlag) Details() interface{} {
	return struct {
		Work       openalex.Work `json:"work"`
		Publishers []string      `json:"publishers"`
	}{
		Work:       flag.Work,
		Publishers: flag.Publishers,
	}
}

func (flag *EOCPublishersFlag) Key() string {
	return workFlagKey(flag.FlagType, flag.Work.WorkId)
}

type EOCCoauthorsFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	Work openalex.Work

	Coauthors []string
}

func (flag *EOCCoauthorsFlag) Type() flagType {
	return OACoauthorIsEOC
}

func (flag *EOCCoauthorsFlag) Connection() Connection {
	return Connection{
		Title: flag.Work.DisplayName,
		Url:   flag.Work.WorkUrl,
	}
}

func (flag *EOCCoauthorsFlag) Details() interface{} {
	return struct {
		Work      openalex.Work `json:"work"`
		Coauthors []string      `json:"coauthors"`
	}{
		Work:      flag.Work,
		Coauthors: flag.Coauthors,
	}
}

func (flag *EOCCoauthorsFlag) Key() string {
	return workFlagKey(flag.FlagType, flag.Work.WorkId)
}

type EOCAuthorAffiliationsFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	Work openalex.Work

	Institutions []string
}

func (flag *EOCAuthorAffiliationsFlag) Type() flagType {
	return OAAuthorAffiliationIsEOC
}

func (flag *EOCAuthorAffiliationsFlag) Connection() Connection {
	return Connection{
		Title: strings.Join(flag.Institutions, " "),
		Url:   flag.Work.WorkUrl,
	}
}

func (flag *EOCAuthorAffiliationsFlag) Details() interface{} {
	return struct {
		Work         openalex.Work `json:"work"`
		Institutions []string      `json:"institutions"`
	}{
		Work:         flag.Work,
		Institutions: flag.Institutions,
	}
}

func (flag *EOCAuthorAffiliationsFlag) Key() string {
	return workFlagKey(flag.FlagType, flag.Work.WorkId)
}

type EOCCoauthorAffiliationsFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	Work openalex.Work

	Institutions []string
	Authors      []string
}

func (flag *EOCCoauthorAffiliationsFlag) Type() flagType {
	return OACoauthorAffiliationIsEOC
}

func (flag *EOCCoauthorAffiliationsFlag) Connection() Connection {
	return Connection{
		Title: flag.Work.DisplayName,
		Url:   flag.Work.WorkUrl,
	}
}

func (flag *EOCCoauthorAffiliationsFlag) Details() interface{} {
	return struct {
		Work         openalex.Work `json:"work"`
		Institutions []string      `json:"institutions"`
		Authors      []string      `json:"authors"`
	}{
		Work:         flag.Work,
		Institutions: flag.Institutions,
		Authors:      flag.Authors,
	}
}

func (flag *EOCCoauthorAffiliationsFlag) Key() string {
	return workFlagKey(flag.FlagType, flag.Work.WorkId)
}

type EOCAcknowledgementEntity struct {
	Entity  string
	Sources []string
	Aliases []string
}

type EOCAcknowledgemntsFlag struct {
	FlagType    flagType
	FlagTitle   string // Do we still need this?
	FlagMessage string // Do we still need this?

	Work openalex.Work

	Entities           []EOCAcknowledgementEntity
	RawAcknowledements []string
}

func (flag *EOCAcknowledgemntsFlag) Type() flagType {
	return OAAcknowledgementIsEOC
}

func (flag *EOCAcknowledgemntsFlag) Connection() Connection {
	return Connection{
		Title: flag.Work.DisplayName,
		Url:   flag.Work.WorkUrl,
	}
}

func (flag *EOCAcknowledgemntsFlag) Details() interface{} {
	return struct {
		Work               openalex.Work              `json:"work"`
		Entities           []EOCAcknowledgementEntity `json:"entities"`
		RawAcknowledements []string                   `json:"raw_acknowledgements"`
	}{
		Work:               flag.Work,
		Entities:           flag.Entities,
		RawAcknowledements: flag.RawAcknowledements,
	}
}

func (flag *EOCAcknowledgemntsFlag) Key() string {
	return workFlagKey(flag.FlagType, flag.Work.WorkId)
}
