package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

type KeyValue struct {
	Key   string
	Value string
}

type KeyValueURL struct {
	Key   string
	Value string
	Url   string
}

type Flag interface {
	Type() string

	// This is used to deduplicate flags. Primarily for author flags, it is
	// possible to have the same flag created for multiple works, for instance by
	// finding the author is faculty at an EOC. For work flags, the key is just the
	// hash of the flagger type and work id since we can only have 1 flag for a
	// given work.
	Hash() [sha256.Size]byte

	GetEntities() []string

	MarkDisclosed()

	GetDetailFields() []KeyValue

	GetHeading() string

	// The second arg indicates if the flag can be filtered by date.
	Date() (time.Time, bool)

	GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL

	IsDisclosed() bool
}

const (
	TalentContractType               = "TalentContracts"
	AssociationsWithDeniedEntityType = "AssociationsWithDeniedEntities"
	HighRiskFunderType               = "HighRiskFunders"
	AuthorAffiliationType            = "AuthorAffiliations"
	PotentialAuthorAffiliationType   = "PotentialAuthorAffiliations"
	MiscHighRiskAssociationType      = "MiscHighRiskAssociations"
	CoauthorAffiliationType          = "CoauthorAffiliations"
	NewsArticleType                  = "NewsArticles"
	// Unused flags
	MultipleAffiliationType = "MultipleAffiliations"
	HighRiskPublisherType   = "HighRiskPublishers"
	HighRiskCoauthorType    = "HighRiskCoauthors"
)

func ParseFlag(ftype string, data []byte) (Flag, error) {
	switch ftype {
	case TalentContractType:
		var flag TalentContractFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case AssociationsWithDeniedEntityType:
		var flag AssociationWithDeniedEntityFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case HighRiskFunderType:
		var flag HighRiskFunderFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case AuthorAffiliationType:
		var flag AuthorAffiliationFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case PotentialAuthorAffiliationType:
		var flag PotentialAuthorAffiliationFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case MiscHighRiskAssociationType:
		var flag MiscHighRiskAssociationFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case CoauthorAffiliationType:
		var flag CoauthorAffiliationFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case MultipleAffiliationType:
		var flag MultipleAffiliationFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case HighRiskPublisherType:
		var flag HighRiskPublisherFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case HighRiskCoauthorType:
		var flag HighRiskCoauthorFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	case NewsArticleType:
		var flag NewsArticleFlag
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, fmt.Errorf("error parsing flag of type '%s': %w", ftype, err)
		}
		return &flag, nil

	default:
		return nil, fmt.Errorf("invalid flag type '%s'", ftype)
	}
}

type DisclosableFlag struct {
	Disclosed bool
}

func (flag *DisclosableFlag) MarkDisclosed() {
	flag.Disclosed = true
}

func (flag *DisclosableFlag) IsDisclosed() bool {
	return flag.Disclosed
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
	Message               string
	Work                  WorkSummary
	Entities              []AcknowledgementEntity
	RawAcknowledgements   []string
	FundCodeTriangulation map[string]map[string]bool
}

func (flag *TalentContractFlag) Type() string {
	return TalentContractType
}

func (flag *TalentContractFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
}

func (flag *TalentContractFlag) GetEntities() []string {
	entities := make([]string, 0, len(flag.Entities))
	for _, ack := range flag.Entities {
		entities = append(entities, ack.Entity)
	}
	return entities
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
		{Key: "Acknowledgements", Value: strings.Join(flag.RawAcknowledgements, ", ")},
	}
}

func (flag *TalentContractFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *TalentContractFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Acknowledgements", Value: strings.Join(flag.RawAcknowledgements, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type AssociationWithDeniedEntityFlag struct {
	DisclosableFlag
	Message             string
	Work                WorkSummary
	Entities            []AcknowledgementEntity
	RawAcknowledgements []string
}

func (flag *AssociationWithDeniedEntityFlag) Type() string {
	return AssociationsWithDeniedEntityType
}

func (flag *AssociationWithDeniedEntityFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
}

func (flag *AssociationWithDeniedEntityFlag) GetEntities() []string {
	entities := make([]string, 0, len(flag.Entities))
	for _, ack := range flag.Entities {
		entities = append(entities, ack.Entity)
	}
	return entities
}

func (flag *AssociationWithDeniedEntityFlag) GetHeading() string {
	return "Funding from Denied Entities"
}

func (flag *AssociationWithDeniedEntityFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Disclosed", Value: fmt.Sprintf("%v", flag.Disclosed)},
		{Key: "Title", Value: flag.Work.DisplayName},
		{Key: "URL", Value: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Acknowledgements", Value: strings.Join(flag.RawAcknowledgements, ", ")},
	}
}

func (flag *AssociationWithDeniedEntityFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *AssociationWithDeniedEntityFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Acknowledgements", Value: strings.Join(flag.RawAcknowledgements, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type HighRiskFunderFlag struct {
	DisclosableFlag
	Message               string
	Work                  WorkSummary
	Funders               []string
	RawAcknowledgements   []string
	FundCodeTriangulation map[string]map[string]bool
}

func (flag *HighRiskFunderFlag) Type() string {
	return HighRiskFunderType
}

func (flag *HighRiskFunderFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
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

func (flag *HighRiskFunderFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *HighRiskFunderFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Funders", Value: strings.Join(flag.Funders, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type AuthorAffiliationFlag struct {
	DisclosableFlag
	Message      string
	Work         WorkSummary
	Affiliations []string
}

func (flag *AuthorAffiliationFlag) Type() string {
	return AuthorAffiliationType
}

func (flag *AuthorAffiliationFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
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

func (flag *AuthorAffiliationFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *AuthorAffiliationFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Affiliations", Value: strings.Join(flag.Affiliations, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type PotentialAuthorAffiliationFlag struct {
	DisclosableFlag
	Message       string
	University    string
	UniversityUrl string
}

func (flag *PotentialAuthorAffiliationFlag) Type() string {
	return PotentialAuthorAffiliationType
}

func (flag *PotentialAuthorAffiliationFlag) Hash() [sha256.Size]byte {
	return sha256.Sum256([]byte(flag.Type() + flag.University + flag.UniversityUrl))
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

func (flag *PotentialAuthorAffiliationFlag) Date() (time.Time, bool) {
	// TODO: add date information to this flag and return it here
	return time.Time{}, false
}

func (flag *PotentialAuthorAffiliationFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "University", Value: flag.University, Url: flag.UniversityUrl},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
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

func (flag *MiscHighRiskAssociationFlag) Type() string {
	return MiscHighRiskAssociationType
}

func (flag *MiscHighRiskAssociationFlag) Hash() [sha256.Size]byte {
	data := flag.Type() + flag.DocTitle + flag.EntityMentioned
	for _, conn := range flag.Connections {
		data += conn.DocTitle
	}
	return sha256.Sum256([]byte(data))
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

func (flag *MiscHighRiskAssociationFlag) Date() (time.Time, bool) {
	// TODO: add date information to this flag and return it here
	return time.Time{}, false
}

func (flag *MiscHighRiskAssociationFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.DocTitle, Url: flag.DocUrl},
		{Key: "Entities", Value: strings.Join(flag.DocEntities, ", ")},
		{Key: "Entity Mentioned", Value: flag.EntityMentioned},
	}
	if flag.FrequentCoauthor != nil {
		fields = append(fields, KeyValueURL{Key: "Frequent Coauthor", Value: *flag.FrequentCoauthor})
	}
	for i, conn := range flag.Connections {
		titleKey := fmt.Sprintf("Connection %d", i+1)
		fields = append(fields, KeyValueURL{Key: titleKey, Value: conn.DocTitle, Url: conn.DocUrl})
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type CoauthorAffiliationFlag struct {
	DisclosableFlag
	Message      string
	Work         WorkSummary
	Coauthors    []string
	Affiliations []string
}

func (flag *CoauthorAffiliationFlag) Type() string {
	return CoauthorAffiliationType
}

func (flag *CoauthorAffiliationFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
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

func (flag *CoauthorAffiliationFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *CoauthorAffiliationFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Co-authors", Value: strings.Join(flag.Coauthors, ", ")},
		{Key: "Affiliations", Value: strings.Join(flag.Affiliations, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type NewsArticleFlag struct {
	Message string

	Title       string
	Link        string
	ArticleDate time.Time
}

func (flag *NewsArticleFlag) Type() string {
	return NewsArticleType
}

func (flag *NewsArticleFlag) Hash() [sha256.Size]byte {
	return sha256.Sum256([]byte(flag.Type() + flag.Title))
}

func (flag *NewsArticleFlag) GetEntities() []string {
	return nil
}

func (flag *NewsArticleFlag) GetHeading() string {
	return "Concerning News Articles"
}

func (flag *NewsArticleFlag) GetDetailFields() []KeyValue {
	return []KeyValue{
		{Key: "Article Title", Value: flag.Title},
		{Key: "URL", Value: flag.Link},
		{Key: "Publication Date", Value: flag.ArticleDate.String()},
	}
}

func (flag *NewsArticleFlag) Date() (time.Time, bool) {
	return flag.ArticleDate, true
}

func (flag *NewsArticleFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	return []KeyValueURL{
		{Key: "Article Title", Value: flag.Title, Url: flag.Link},
		{Key: "URL", Value: flag.Link},
		{Key: "Publication Date", Value: flag.ArticleDate.String()},
	}
}

func (flag *NewsArticleFlag) IsDisclosed() bool { return false }

func (flag *NewsArticleFlag) MarkDisclosed() {}

//The following flags are unused by the frontend, but they are kept in case we
// want to have them in the future.

type MultipleAffiliationFlag struct {
	DisclosableFlag
	Message      string
	Work         WorkSummary
	Affiliations []string
}

func (flag *MultipleAffiliationFlag) Type() string {
	return MultipleAffiliationType
}

func (flag *MultipleAffiliationFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
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

func (flag *MultipleAffiliationFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *MultipleAffiliationFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Affiliations", Value: strings.Join(flag.Affiliations, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type HighRiskPublisherFlag struct {
	DisclosableFlag
	Message    string
	Work       WorkSummary
	Publishers []string
}

func (flag *HighRiskPublisherFlag) Type() string {
	return HighRiskPublisherType
}

func (flag *HighRiskPublisherFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
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

func (flag *HighRiskPublisherFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *HighRiskPublisherFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Publishers", Value: strings.Join(flag.Publishers, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}

type HighRiskCoauthorFlag struct {
	DisclosableFlag
	Message   string
	Work      WorkSummary
	Coauthors []string
}

func (flag *HighRiskCoauthorFlag) Type() string {
	return HighRiskCoauthorType
}

func (flag *HighRiskCoauthorFlag) Hash() [sha256.Size]byte {
	// Assumes 1 flag per work
	return sha256.Sum256([]byte(flag.Type() + flag.Work.WorkId))
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

func (flag *HighRiskCoauthorFlag) Date() (time.Time, bool) {
	return flag.Work.PublicationDate, true
}

func (flag *HighRiskCoauthorFlag) GetDetailsFieldsForReport(useDisclosure bool) []KeyValueURL {
	fields := []KeyValueURL{
		{Key: "Title", Value: flag.Work.DisplayName, Url: flag.Work.WorkUrl},
		{Key: "Publication Date", Value: flag.Work.PublicationDate.Format(time.DateOnly)},
		{Key: "Co-authors", Value: strings.Join(flag.Coauthors, ", ")},
	}
	if useDisclosure {
		fields = append([]KeyValueURL{
			{Key: "Disclosed", Value: capitalizeFirstLetter(fmt.Sprintf("%v", flag.Disclosed))},
		}, fields...)
	}
	return fields
}
