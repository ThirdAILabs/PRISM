package flaggers

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

func TestRunInpool(t *testing.T) {
	worker := func(i int) (string, error) {
		if i%4 == 3 {
			time.Sleep(time.Duration(10-i) * time.Millisecond)
			return "", fmt.Errorf("error")
		}
		return fmt.Sprintf("%d-%d", i, i), nil
	}

	queue := make(chan int, 10)

	for i := 0; i < 10; i++ {
		queue <- i
	}

	close(queue)

	output := make(chan CompletedTask[string], 10)

	RunInPool(worker, queue, output, 5, nil)

	success, errors := 0, 0
	for result := range output {
		if result.Error != nil {
			errors++
		} else {
			success++
		}
	}

	if success != 8 || errors != 2 {
		t.Fatal("invalid results")
	}
}

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
