package api

import (
	"encoding/json"
	"fmt"
)

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
	ForeignEntityNotFound bool // flagged foreign entity is not found in the url
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

func ParseFlagFeedback(flagType string, data []byte) (FlagFeedback, error) {
	switch flagType {
	case TalentContractType:
		var flag TalentContractFeedback
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, err
		}
		return &flag, nil
	case AssociationsWithDeniedEntityType:
		var flag AssociationWithDeniedEntityFeedback
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, err
		}
		return &flag, nil
	case HighRiskFunderType:
		var flag HighRiskFunderFeedback
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, err
		}
		return &flag, nil
	case AuthorAffiliationType:
		var flag AuthorAffiliationFeedback
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, err
		}
		return &flag, nil
	case PotentialAuthorAffiliationType:
		var flag PotentialAuthorAffiliationFeedback
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, err
		}
		return &flag, nil
	case MiscHighRiskAssociationType:
		var flag MiscHighRiskAssociationFeedback
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, err
		}
		return &flag, nil
	case CoauthorAffiliationType:
		var flag CoauthorAffiliationFeedback
		if err := json.Unmarshal(data, &flag); err != nil {
			return nil, err
		}
		return &flag, nil
	default:
		return nil, fmt.Errorf("invalid flag type: %s", flagType)
	}
}
