package api

type FlagFeedback interface {
	Type() string
}

type TalentContractFeedback struct {
	Feedbacks []Feedbacks
}

func (feedback *TalentContractFeedback) Type() string {
	return TalentContractType
}

type AssociationWithDeniedEntityFeedback struct {
	Feedbacks []Feedbacks
}

func (feedback *AssociationWithDeniedEntityFeedback) Type() string {
	return AssociationsWithDeniedEntityType
}

type HighRiskFunderFeedback struct {
	Feedbacks []Feedbacks
}

func (feedback *HighRiskFunderFeedback) Type() string {
	return HighRiskFunderType
}

type AuthorAffiliationFeedback struct {
	Feedbacks []Feedbacks
}

func (feedback *AuthorAffiliationFeedback) Type() string {
	return AuthorAffiliationType
}

type PotentialAuthorAffiliationFeedback struct {
	Feedbacks []Feedbacks
}

func (feedback *PotentialAuthorAffiliationFeedback) Type() string {
	return PotentialAuthorAffiliationType
}

type MiscHighRiskAssociationFeedback struct {
	Feedbacks []Feedbacks
}

func (feedback *MiscHighRiskAssociationFeedback) Type() string {
	return MiscHighRiskAssociationType
}

type CoauthorAffiliationFeedback struct {
	Feedbacks []Feedbacks
}

func (feedback *CoauthorAffiliationFeedback) Type() string {
	return CoauthorAffiliationType
}
