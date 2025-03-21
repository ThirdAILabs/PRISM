package flaggers

import (
	"slices"
	"testing"
)

func TestInitialsCombinations(t *testing.T) {
	combos := getInitialsCombinations("aa bb cc")

	expected := []string{
		"ac", "a.c.", "abc", "a.b.c.",
		"a b c", "a. b. c.", "aa b c", "aa b. c.", "a bb c", "a. bb c.", "a b cc", "a. b. cc",
		"aa bb c", "aa bb c.", "aa b cc", "aa b. cc", "a bb cc", "a. bb cc", "aa bb cc", "aa bb cc",
	}
	slices.Sort(combos)
	slices.Sort(expected)

	if !slices.Equal(combos, expected) {
		t.Fatal("invalid combinations")
	}
}
