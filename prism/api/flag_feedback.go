package api

import (
	"encoding/json"
	"fmt"
)

type FlagFeedback struct {
	IncorrectAuthor       bool
	AuthorNotFound        bool
	EntityNotFound        bool
	EntityNotEoc          bool
	IncorrectDoc          KeyValue // DocTitle, DocUrl
	IncorrectAffiliations []string
}

func ParseFlagFeedback(data []byte) (FlagFeedback, error) {
	var feedback FlagFeedback
	if err := json.Unmarshal(data, &feedback); err != nil {
		return FlagFeedback{}, fmt.Errorf("error parsing flag feedback: %w", err)
	}
	return feedback, nil
}
