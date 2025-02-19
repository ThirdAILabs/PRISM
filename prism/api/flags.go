package api

import "fmt"

type Flag interface {
	// This is used to deduplicate flags. Primarily for author flags, it is
	// possible to have the same flag created for multiple works, for instance by
	// finding the author is faculty at an EOC. For work flags, the key is just
	// the flagger type and work id since we can only have 1 flag for a given work.
	Key() string
}

type WorkSummary struct {
	WorkId          string
	DisplayName     string
	WorkUrl         string
	OaUrl           string
	PublicationYear int
}

type AcknowledgementEntity struct {
	Entity  string
	Sources []string
	Aliases []string
}

type TalentContractFlag struct {
	Message            string
	Work               WorkSummary
	Entities           []AcknowledgementEntity
	RawAcknowledements []string
}

func (flag *TalentContractFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("talent-contract-%s", flag.Work.WorkId)
}

type AssociationWithDeniedEntityFlag struct {
	Message            string
	Work               WorkSummary
	Entities           []AcknowledgementEntity
	RawAcknowledements []string
}

func (flag *AssociationWithDeniedEntityFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("association-with-denied-entity-%s", flag.Work.WorkId)
}

type HighRiskFunderFlag struct {
	Message              string
	Work                 WorkSummary
	Funders              []string
	FromAcknowledgements bool
}

func (flag *HighRiskFunderFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("high-risk-funder-%s", flag.Work.WorkId)
}

type AuthorAffiliationFlag struct {
	Message      string
	Work         WorkSummary
	Affiliations []string
}

func (flag *AuthorAffiliationFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("author-affiliation-%s", flag.Work.WorkId)
}

type PotentialAuthorAffiliationFlag struct {
	Message       string
	University    string
	UniversityUrl string
}

func (flag *PotentialAuthorAffiliationFlag) Key() string {
	return fmt.Sprintf("potential-author-affiliation-%s-%s", flag.University, flag.UniversityUrl)
}

type Connection struct {
	DocTitle string
	DocUrl   string
}

type MiscHighRiskAssociationFlag struct {
	Message          string
	DocTitle         string
	DocUrl           string
	DocEntities      []string
	EntityMentioned  string
	Connections      []Connection
	FrequentCoauthor *string
}

func (flag *MiscHighRiskAssociationFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("misc-high-risk-associations-%s-%v-%s", flag.DocTitle, flag.Connections, flag.EntityMentioned)
}

type CoauthorAffiliationFlag struct {
	Message      string
	Work         WorkSummary
	Coauthors    []string
	Affiliations []string
}

func (flag *CoauthorAffiliationFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("coauthor-affiliation-%s", flag.Work.WorkId)
}

type ReportContent struct {
	TalentContracts                []*TalentContractFlag
	AssociationsWithDeniedEntities []*AssociationWithDeniedEntityFlag
	HighRiskFunders                []*HighRiskFunderFlag
	AuthorAffiliations             []*AuthorAffiliationFlag
	PotentialAuthorAffiliations    []*PotentialAuthorAffiliationFlag
	MiscHighRiskAssociations       []*MiscHighRiskAssociationFlag
	CoauthorAffiliations           []*CoauthorAffiliationFlag
}

//The following flags are unused by the frontend, but they are kept in case we
// want to have them in the future.

type MultipleAffiliationFlag struct {
	Message      string
	Work         WorkSummary
	Affiliations []string
}

func (flag *MultipleAffiliationFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("multiple-affiliations-%s", flag.Work.WorkId)
}

type HighRiskPublisherFlag struct {
	Message    string
	Work       WorkSummary
	Publishers []string
}

func (flag *HighRiskPublisherFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("high-risk-publisher-%s", flag.Work.WorkId)
}

type HighRiskCoauthorFlag struct {
	Message   string
	Work      WorkSummary
	Coauthors []string
}

func (flag *HighRiskCoauthorFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("high-risk-coauthor-%s", flag.Work.WorkId)
}
