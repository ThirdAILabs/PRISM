package api

const (
	IncorrectAcknowledgementType = "IncorrectAcknowledgement"
	IncorrectAuthorType          = "IncorrectAuthor"
	IncorrectPaperTrailType      = "IncorrectPaperTrail"
	IncorrectEntityDetectionType = "IncorrectEntityDetection"
	EntityNotEocType             = "EntityNotEoc"
	IncorrectAffiliationType     = "IncorrectAffiliation"
)

type Feedbacks interface {
	Type() string
}

type IncorrectAcknowledgement struct {
	Text   string
	DocUrl string
}

func (feedback *IncorrectAcknowledgement) Type() string {
	return IncorrectAcknowledgementType
}

type IncorrectAuthor struct {
	AuthorId string
}

func (feedback *IncorrectAuthor) Type() string {
	return IncorrectAuthorType
}

type IncorrectPaperTrail struct {
	DocTitle string
	DocUrl   string
}

func (feedback *IncorrectPaperTrail) Type() string {
	return IncorrectPaperTrailType
}

type IncorrectEntityDetection struct {
	DocUrl string
	Entity string
}

func (feedback *IncorrectEntityDetection) Type() string {
	return IncorrectEntityDetectionType
}

type EntityNotEoc struct {
	Entity string
}

func (feedback *EntityNotEoc) Type() string {
	return EntityNotEocType
}

type IncorrectAffiliation struct {
	Affiliate string
	DocUrl    string
}

func (feedback *IncorrectAffiliation) Type() string {
	return IncorrectAffiliationType
}
