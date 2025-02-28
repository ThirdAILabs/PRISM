package flaggers

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"prism/prism/llms"
	"prism/prism/search"
	"slices"
	"strings"
)

const (
	ndbScoreThreshold = 6.5
	indelSimThreshold = 0.95
	jaroSimThreshold  = 0.93
	maxLLMWorkers     = 30
)

type EntityStore struct {
	aliasToSource map[string]string
	ndb           search.NeuralDB
	flash         search.Flash
}

func createNdb(ndbPath string, aliases []string) (search.NeuralDB, error) {
	ndb, err := search.NewNeuralDB(ndbPath)
	if err != nil {
		return search.NeuralDB{}, fmt.Errorf("error creating ndb: %w", err)
	}

	if err := ndb.Insert("aliases", "aliases", aliases, nil, nil); err != nil {
		ndb.Free()
		return search.NeuralDB{}, fmt.Errorf("ndb insertion failed: %w", err)
	}

	return ndb, nil
}

func createCsv(aliases []string) (string, error) {
	filename := "entities.csv"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return "", fmt.Errorf("error creating csv file for training: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	rows := make([][]string, 0, len(aliases)+1)
	rows = append(rows, []string{"entity"})
	for _, alias := range aliases {
		rows = append(rows, []string{alias})
	}

	if err := writer.WriteAll(rows); err != nil {
		return "", fmt.Errorf("error writing rows: %w", err)
	}

	return filename, nil
}

func createFlash(aliases []string) (search.Flash, error) {
	csv, err := createCsv(aliases)
	if err != nil {
		return search.Flash{}, fmt.Errorf("error creating csv: %w", err)
	}

	flash, err := search.NewFlash()
	if err != nil {
		return search.Flash{}, fmt.Errorf("error creating flash")
	}

	if err := flash.Train(csv); err != nil {
		flash.Free()
		return search.Flash{}, fmt.Errorf("error training flash: %w", err)
	}

	return flash, nil
}

func NewEntityStore(ndbPath string, aliasToSource map[string]string) (*EntityStore, error) {
	store := &EntityStore{aliasToSource: aliasToSource}

	aliases, err := store.allAliases()
	if err != nil {
		return nil, err
	}

	ndb, err := createNdb(ndbPath, aliases)
	if err != nil {
		return nil, fmt.Errorf("error creating entity ndb: %w", err)
	}

	flash, err := createFlash(aliases)
	if err != nil {
		return nil, fmt.Errorf("error creating entity flash: %w", err)
	}

	store.ndb = ndb
	store.flash = flash

	return store, nil
}

func (store *EntityStore) Free() {
	store.ndb.Free()
	store.flash.Free()
}

func (store *EntityStore) allAliases() ([]string, error) {
	data := make([]string, 0, len(store.aliasToSource))
	for k := range store.aliasToSource {
		data = append(data, k)
	}

	return data, nil
}

type SourceToAliases = map[string][]string

func (store *EntityStore) exactLookup(queries []string) (SourceToAliases, error) {
	sourceToAliases := make(map[string][]string)
	for _, query := range queries {
		source, ok := store.aliasToSource[query]
		if ok {
			sourceToAliases[source] = append(sourceToAliases[source], query)
		}
	}

	return sourceToAliases, nil
}

func (store *EntityStore) ndbLookup(query string) ([]string, error) {
	results, err := store.ndb.Query(query, 5, nil)
	if err != nil {
		return nil, fmt.Errorf("ndb entity lookup failed: %w", err)
	}

	if len(results) == 0 {
		return []string{}, nil
	}

	matches := make([]string, 0, len(results))

	highScore := results[0].Score
	for _, result := range results {
		if result.Score >= ndbScoreThreshold && (highScore-result.Score) < 0.5 {
			matches = append(matches, result.Text)
		}
	}

	return matches, nil
}

func (store *EntityStore) flashLookup(query string) ([]string, error) {
	candidates, err := store.flash.Predict(query, 3)
	if err != nil {
		return nil, fmt.Errorf("flash lookup failed: %w", err)
	}

	filtered := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if IndelSimilarity(query, candidate) >= indelSimThreshold ||
			JaroWinklerSimilarity(query, candidate) >= jaroSimThreshold {
			filtered = append(filtered, candidate)
		}
	}

	return filtered, nil
}

const llmValidationPromptTemplate = `Return a python list containing True or False based on whether this entity: %s, is an equivalent entity to the following entities.
Here are the entities: ["%s"]
DO NOT ENTER 'FALSE' JUST BECAUSE THEY HAVE A DIFFERENT NAME, they can still be the same entity with a different name (e.g. same person, organization, school, etc. with different aliases).
Answer with a python list of True or False, and nothing else. Do not provide any explanation or extra words/characters. Ensure that the python syntax is absolutely correct and runnable. Also ensure that the resulting list is the same length as the list of entities.
Answer:
`

func filterMatchesWithLLM(query string, matches []string) ([]string, error) {
	llm := llms.New()

	res, err := llm.Generate(fmt.Sprintf(llmValidationPromptTemplate, query, strings.Join(matches, `", "`)), &llms.Options{
		Model:        llms.GPT4oMini,
		ZeroTemp:     true,
		SystemPrompt: "You are a helpful python assistant who responds in python lists only.",
	})
	if err != nil {
		return nil, fmt.Errorf("llm match verification failed: %w", err)
	}

	flags := strings.Split(res, ",")

	if len(flags) != len(matches) {
		return nil, fmt.Errorf("llm verification returned incorrect number of flags query=%v, matches=%v, flags=%v", query, matches, flags)
	}

	filtered := make([]string, 0, len(matches))
	for i, match := range matches {
		if strings.Contains(flags[i], "True") {
			filtered = append(filtered, match)
		}
	}

	return filtered, nil
}

type verificationTask struct {
	query   string
	matches []string
}

func filterMatchesWorker(task verificationTask) (verificationTask, error) {
	filtered, err := filterMatchesWithLLM(task.query, task.matches)
	if err != nil {
		return verificationTask{}, err
	}
	return verificationTask{query: task.query, matches: filtered}, nil
}

func (store *EntityStore) SearchEntities(logger *slog.Logger, queries []string) (map[string]SourceToAliases, error) {
	queue := make(chan verificationTask, len(queries))
	for _, query := range queries {
		ndbMatches, err := store.ndbLookup(query)
		if err != nil {
			logger.Error("ndb entity lookup failed", "error", err)
		}
		flashMatches, err := store.flashLookup(query)
		if err != nil {
			logger.Error("flash lookup failed", "error", err)
		}

		matches := slices.Concat(ndbMatches, flashMatches)
		if len(matches) > 0 {
			queue <- verificationTask{query: query, matches: matches}
		}
	}
	close(queue)

	nWorkers := min(maxLLMWorkers, len(queue))
	completed := make(chan CompletedTask[verificationTask], len(queue))
	RunInPool(filterMatchesWorker, queue, completed, nWorkers)

	results := make(map[string]SourceToAliases)
	for task := range completed {
		if task.Error != nil {
			logger.Error("llm match verification error", "error", task.Error)
			continue
		}

		sourceToAliases, err := store.exactLookup(task.Result.matches)
		if err != nil {
			logger.Error("error completing exact entity lookup", "error", err)
		} else {
			results[task.Result.query] = sourceToAliases
		}
	}

	return results, nil
}
