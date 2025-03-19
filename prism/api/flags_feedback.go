package api

type FlagFeedback interface {
	Type() string
}

type AuthorIssue struct {
	IncorrectAuthor bool
	AuthorNotFound  bool
}

type TalentContractFeedback struct {
	AuthorIssue
}

func (feedback *TalentContractFeedback) Type() string {
	return TalentContractType
}

type AssociationWithDeniedEntityFeedback struct {
	AuthorIssue
	ForeignEntityNotFound bool // glagged foreign entity is not found in the url
}

func (feedback *AssociationWithDeniedEntityFeedback) Type() string {
	return AssociationsWithDeniedEntityType
}

type HighRiskFunderFeedback struct {
	AuthorIssue
	FundingNotEoc bool
}

func (feedback *HighRiskFunderFeedback) Type() string {
	return HighRiskFunderType
}

type AuthorAffiliationFeedback struct {
	AuthorIssue
	Alliliates []string // the affiliation is incorrect
}

func (feedback *AuthorAffiliationFeedback) Type() string {
	return AuthorAffiliationType
}

type PotentialAuthorAffiliationFeedback struct {
	AuthorIssue
	IncorrectInstitutionAffiliation bool
}

func (feedback *PotentialAuthorAffiliationFeedback) Type() string {
	return PotentialAuthorAffiliationType
}

type MiscHighRiskAssociationFeedback struct {
	AuthorIssue
	EntityNotMentioned bool     // the entity is not mentioned in the Press release
	IncorrectDocUrl    string   // the document url for this document title is incorrect
	Affiliates         []string // the affiliated authors are incorrect
}

func (feedback *MiscHighRiskAssociationFeedback) Type() string {
	return MiscHighRiskAssociationType
}

type CoauthorAffiliationFeedback struct {
	AuthorIssue
	Affiliates []string // the affiliated authors are incorrect
}

func (feedback *CoauthorAffiliationFeedback) Type() string {
	return CoauthorAffiliationType
}
