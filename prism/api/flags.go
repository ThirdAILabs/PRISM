package api

import (
	"fmt"
	"slices"
)

type Flag interface {
	// This is used to deduplicate flags. Primarily for author flags, it is
	// possible to have the same flag created for multiple works, for instance by
	// finding the author is faculty at an EOC. For work flags, the key is just
	// the flagger type and work id since we can only have 1 flag for a given work.
	Key() string

	GetEntities() []string

	MarkDisclosed()
}

type DisclosableFlag struct {
	Disclosed bool
}

func (flag *DisclosableFlag) MarkDisclosed() {
	flag.Disclosed = true
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
	DisclosableFlag
	Message            string
	Work               WorkSummary
	Entities           []AcknowledgementEntity
	RawAcknowledements []string
}

func (flag *TalentContractFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("talent-contract-%s", flag.Work.WorkId)
}

func (flag *TalentContractFlag) GetEntities() []string {
	return flag.RawAcknowledements
}

type AssociationWithDeniedEntityFlag struct {
	DisclosableFlag
	Message            string
	Work               WorkSummary
	Entities           []AcknowledgementEntity
	RawAcknowledements []string
}

func (flag *AssociationWithDeniedEntityFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("association-with-denied-entity-%s", flag.Work.WorkId)
}

func (flag *AssociationWithDeniedEntityFlag) GetEntities() []string {
	return flag.RawAcknowledements
}

type HighRiskFunderFlag struct {
	DisclosableFlag
	Message            string
	Work               WorkSummary
	Funders            []string
	RawAcknowledements []string
}

func (flag *HighRiskFunderFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("high-risk-funder-%s", flag.Work.WorkId)
}

func (flag *HighRiskFunderFlag) GetEntities() []string {
	return flag.Funders
}

type AuthorAffiliationFlag struct {
	DisclosableFlag
	Message      string
	Work         WorkSummary
	Affiliations []string
}

func (flag *AuthorAffiliationFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("author-affiliation-%s", flag.Work.WorkId)
}

func (flag *AuthorAffiliationFlag) GetEntities() []string {
	return flag.Affiliations
}

type PotentialAuthorAffiliationFlag struct {
	DisclosableFlag
	Message       string
	University    string
	UniversityUrl string
}

func (flag *PotentialAuthorAffiliationFlag) Key() string {
	return fmt.Sprintf("potential-author-affiliation-%s-%s", flag.University, flag.UniversityUrl)
}

func (flag *PotentialAuthorAffiliationFlag) GetEntities() []string {
	return []string{flag.University}
}

type Connection struct {
	DocTitle string
	DocUrl   string
}

type MiscHighRiskAssociationFlag struct {
	DisclosableFlag
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

func (flag *MiscHighRiskAssociationFlag) GetEntities() []string {
	entities := make([]string, 0)
	if flag.FrequentCoauthor != nil {
		entities = append(entities, *flag.FrequentCoauthor)
	}
	if len(flag.Connections) > 0 { // Not primary connection, means EntityMentioned is not the author
		entities = append(entities, flag.EntityMentioned)
	}
	return entities
}

type CoauthorAffiliationFlag struct {
	DisclosableFlag
	Message      string
	Work         WorkSummary
	Coauthors    []string
	Affiliations []string
}

func (flag *CoauthorAffiliationFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("coauthor-affiliation-%s", flag.Work.WorkId)
}

func (flag *CoauthorAffiliationFlag) GetEntities() []string {
	return slices.Concat(flag.Coauthors, flag.Affiliations)
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
	DisclosableFlag
	Message      string
	Work         WorkSummary
	Affiliations []string
}

func (flag *MultipleAffiliationFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("multiple-affiliations-%s", flag.Work.WorkId)
}

func (flag *MultipleAffiliationFlag) GetEntities() []string {
	return flag.Affiliations
}

type HighRiskPublisherFlag struct {
	DisclosableFlag
	Message    string
	Work       WorkSummary
	Publishers []string
}

func (flag *HighRiskPublisherFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("high-risk-publisher-%s", flag.Work.WorkId)
}

func (flag *HighRiskPublisherFlag) GetEntities() []string {
	return flag.Publishers
}

type HighRiskCoauthorFlag struct {
	DisclosableFlag
	Message   string
	Work      WorkSummary
	Coauthors []string
}

func (flag *HighRiskCoauthorFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("high-risk-coauthor-%s", flag.Work.WorkId)
}

func (flag *HighRiskCoauthorFlag) GetEntities() []string {
	return flag.Coauthors
}
