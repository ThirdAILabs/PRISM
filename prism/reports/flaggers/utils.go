package flaggers

import (
	"prism/prism/openalex"
	"prism/prism/reports"
	"strings"
	"sync"
)

type CompletedTask[T any] struct {
	Result T
	Error  error
}

func RunInPool[In any, Out any](worker func(In) (Out, error), queue chan In, completed chan CompletedTask[Out], maxWorkers int) {
	workers := min(len(queue), maxWorkers)

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()

				for {
					next, ok := <-queue
					if !ok {
						return
					}

					res, err := worker(next)
					if err != nil {
						completed <- CompletedTask[Out]{Error: err}
					} else {
						completed <- CompletedTask[Out]{Result: res, Error: nil}
					}
				}
			}()
		}

		wg.Wait()

		close(completed)
	}()
}

func parseOpenAlexId(work openalex.Work) string {
	defer reports.LogTiming("parseOpenAlexId")()

	idx := strings.LastIndex(work.WorkId, "/")
	if idx < 0 {
		return ""
	}
	return work.WorkId[idx+1:]
}

func joinParts(parts []string, withPeriod bool) string {
	output := ""
	for _, part := range parts {
		output += part
		if withPeriod && len(part) == 1 {
			output += "."
		}
		output += " "
	}
	return strings.TrimRight(output, " ")
}

func getInitialsCombinations(name string) []string {
	name = strings.Replace(name, ".", "", -1)
	parts := strings.Fields(name)
	if len(parts) < 2 {
		return []string{}
	}

	initials := make([]string, 0, len(parts))
	for _, part := range parts {
		initials = append(initials, part[:1])
	}

	candidates := []string{
		initials[0] + "." + initials[len(initials)-1] + ".",
		initials[0] + initials[len(initials)-1],
		strings.Join(initials, ".") + ".",
		strings.Join(initials, ""),
	}

	nCombinations := (1 << len(initials))
	candidate := make([]string, len(initials))
	for n := 0; n < nCombinations; n++ {
		for i := 0; i < len(initials); i++ {
			if n&(1<<i) > 0 {
				candidate[i] = initials[i]
			} else {
				candidate[i] = parts[i]
			}
		}

		candidates = append(candidates, joinParts(candidate, true), joinParts(candidate, false))
	}

	return candidates
}
