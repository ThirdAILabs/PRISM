package api

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type KeyValue struct {
	Key   string
	Value string
}

type Flag interface {
	// This is used to deduplicate flags. Primarily for author flags, it is
	// possible to have the same flag created for multiple works, for instance by
	// finding the author is faculty at an EOC. For work flags, the key is just
	// the flagger type and work id since we can only have 1 flag for a given work.
	Key() string

	GetEntities() []string

	MarkDisclosed()

	GetDetailFields() []KeyValue

	GetHeading() string
	Before(time.Time) bool

	After(time.Time) bool
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
	PublicationDate time.Time
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
	LikelyGrantRecipient bool
}

func (flag *TalentContractFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("talent-contract-%s", flag.Work.WorkId)
}

func (flag *TalentContractFlag) GetEntities() []string {
	return flag.RawAcknowledements
}

func (flag *TalentContractFlag) GetHeading() string {
	return "Talent Contracts"
}

func (flag *TalentContractFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Acknowledgements", Value: strings.Join(flag.RawAcknowledements, ", ")},
	}
}

func (flag *TalentContractFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *TalentContractFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
}

type AssociationWithDeniedEntityFlag struct {
	DisclosableFlag
	Message            string
	Work               WorkSummary
	Entities           []AcknowledgementEntity
	RawAcknowledements []string
	LikelyGrantRecipient bool

}

func (flag *AssociationWithDeniedEntityFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("association-with-denied-entity-%s", flag.Work.WorkId)
}

func (flag *AssociationWithDeniedEntityFlag) GetEntities() []string {
	return flag.RawAcknowledements
}

func (flag *AssociationWithDeniedEntityFlag) GetHeading() string {
	return "Funding from Denied Entities"
}

func (flag *AssociationWithDeniedEntityFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Paper Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Acknowledgements", Value: strings.Join(flag.RawAcknowledements, ", ")},
	}
}

func (flag *AssociationWithDeniedEntityFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *AssociationWithDeniedEntityFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
}

type HighRiskFunderFlag struct {
	DisclosableFlag
	Message              string
	Work                 WorkSummary
	Funders              []string
	FromAcknowledgements bool
	LikelyGrantRecipient bool
}

func (flag *HighRiskFunderFlag) Key() string {
	// Assumes 1 flag per work
	return fmt.Sprintf("high-risk-funder-%s", flag.Work.WorkId)
}

func (flag *HighRiskFunderFlag) GetEntities() []string {
	return flag.Funders
}

func (flag *HighRiskFunderFlag) GetHeading() string {
	return "High Risk Funding Sources"
}

func (flag *HighRiskFunderFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Paper Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Funders", Value: strings.Join(flag.Funders, ", ")},
	}
}

func (flag *HighRiskFunderFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *HighRiskFunderFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
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

func (flag *AuthorAffiliationFlag) GetHeading() string {
	return "Affiliations with High Risk Foreign Institutes"
}

func (flag *AuthorAffiliationFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Paper Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Affiliations", Value: strings.Join(flag.Affiliations, ", ")},
	}
}

func (flag *AuthorAffiliationFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *AuthorAffiliationFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
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

func (flag *PotentialAuthorAffiliationFlag) GetHeading() string {
	return "Appointments at High Risk Foreign Institutes"
}

func (flag *PotentialAuthorAffiliationFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "University", Value: flag.University},
		{Key: "University URL", Value: flag.UniversityUrl},
	}
}

func (flag *PotentialAuthorAffiliationFlag) Before(t time.Time) bool {
	// TODO: add date information to this flag and check it here
	return true
}

func (flag *PotentialAuthorAffiliationFlag) After(t time.Time) bool {
	// TODO: add date information to this flag and check it here
	return true
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

func (flag *MiscHighRiskAssociationFlag) GetHeading() string {
	return "Miscellaneous High Risk Connections"
}

func (flag *MiscHighRiskAssociationFlag) GetDetailFields() []KeyValue {
	fields := []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Doc Title", Value: flag.DocTitle},
		{Key: "Doc URL", Value: flag.DocUrl},
		{Key: "Doc Entities", Value: strings.Join(flag.DocEntities, ", ")},
		{Key: "Entity Mentioned", Value: flag.EntityMentioned},
	}
	if flag.FrequentCoauthor != nil {
		fields = append(fields, KeyValue{Key: "Frequent Coauthor", Value: *flag.FrequentCoauthor})
	}
	for i, conn := range flag.Connections {
		titleKey := fmt.Sprintf("Connection %d Title", i+1)
		urlKey := fmt.Sprintf("Connection %d URL", i+1)
		fields = append(fields, KeyValue{Key: titleKey, Value: conn.DocTitle})
		fields = append(fields, KeyValue{Key: urlKey, Value: conn.DocUrl})
	}
	return fields
}

func (flag *MiscHighRiskAssociationFlag) Before(t time.Time) bool {
	// TODO: add date information to this flag and check it here
	return true
}

func (flag *MiscHighRiskAssociationFlag) After(t time.Time) bool {
	// TODO: add date information to this flag and check it here
	return true
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

func (flag *CoauthorAffiliationFlag) GetHeading() string {
	return "Co-authors' affiliations with High Risk Foreign Institutes"
}

func (flag *CoauthorAffiliationFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Paper Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Co-authors", Value: strings.Join(flag.Coauthors, ", ")},
		{Key: "Affiliations", Value: strings.Join(flag.Affiliations, ", ")},
	}
}

func (flag *CoauthorAffiliationFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *CoauthorAffiliationFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
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

func addFlags[T Flag](groups map[string][]Flag, flags []T) {
	for _, flag := range flags {
		key := flag.GetHeading()
		groups[key] = append(groups[key], flag)
	}
}

func (rc *ReportContent) GroupFlags() map[string][]Flag {
	groups := make(map[string][]Flag)

	addFlags(groups, rc.TalentContracts)
	addFlags(groups, rc.AssociationsWithDeniedEntities)
	addFlags(groups, rc.HighRiskFunders)
	addFlags(groups, rc.AuthorAffiliations)
	addFlags(groups, rc.PotentialAuthorAffiliations)
	addFlags(groups, rc.MiscHighRiskAssociations)
	addFlags(groups, rc.CoauthorAffiliations)

	return groups
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

func (flag *MultipleAffiliationFlag) GetHeading() string {
	return flag.Message
}

func (flag *MultipleAffiliationFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Paper Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Affiliations", Value: strings.Join(flag.Affiliations, ", ")},
	}
}

func (flag *MultipleAffiliationFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *MultipleAffiliationFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
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

func (flag *HighRiskPublisherFlag) GetHeading() string {
	return flag.Message
}

func (flag *HighRiskPublisherFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Paper Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Publishers", Value: strings.Join(flag.Publishers, ", ")},
	}
}

func (flag *HighRiskPublisherFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *HighRiskPublisherFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
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

func (flag *HighRiskCoauthorFlag) GetHeading() string {
	return flag.Message
}

func (flag *HighRiskCoauthorFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Paper Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.String()},
		{Key: "Co-authors", Value: strings.Join(flag.Coauthors, ", ")},
	}
}

func (flag *HighRiskCoauthorFlag) Before(t time.Time) bool {
	return flag.Work.PublicationDate.Before(t)
}

func (flag *HighRiskCoauthorFlag) After(t time.Time) bool {
	return flag.Work.PublicationDate.After(t)
}
