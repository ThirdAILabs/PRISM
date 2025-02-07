package flaggers

import (
	"fmt"
	"log/slog"
	"prism/llms"
	"prism/search"
	"slices"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ndbScoreThreshold = 6.5
	indelSimThreshold = 0.95
	jaroSimThreshold  = 0.93
	maxLLMWorkers     = 30
)

type AliasRecord struct {
	Id    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Alias string    `gorm:"not null"`

	EntityId uuid.UUID
	Entity   *EntityRecord `gorm:"foreignKey:EntityId;constraint:OnDelete:CASCADE"`
}

type EntityRecord struct {
	Id   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"not null"`

	SourceId uuid.UUID
	Source   *SourceRecord `gorm:"foreignKey:SourceId;constraint:OnDelete:CASCADE"`
}

type SourceRecord struct {
	Id   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"not null"`
	Link string
}

type EntityStore struct {
	db    *gorm.DB
	ndb   search.NeuralDB
	flash search.Flash
}

func (store *EntityStore) allAliases() ([]string, error) {
	var records []AliasRecord

	if err := store.db.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("db alias lookup failed: %w", err)
	}

	data := make([]string, 0, len(records))
	for _, record := range records {
		data = append(data, record.Alias)
	}

	return data, nil
}

type SourceToAliases = map[string][]string

func (store *EntityStore) exactLookup(queries []string) (SourceToAliases, error) {
	var records []AliasRecord

	if err := store.db.Preload("Entity").Preload("Entity.Source").Find(&records, "alias IN ?", queries).Error; err != nil {
		return nil, fmt.Errorf("db entity lookup failed: %w", err)
	}

	sourceToAliases := make(map[string][]string)
	for _, record := range records {
		source := record.Entity.Source.Name
		aliases, ok := sourceToAliases[source]
		if !ok {
			aliases = make([]string, 1)
		}
		sourceToAliases[source] = append(aliases, record.Alias)
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
		slog.Error("llm match verification error", "query", task.query, "matches", task.matches, "error", err)
		return verificationTask{}, err
	}
	return verificationTask{query: task.query, matches: filtered}, nil
}

func (store *EntityStore) SearchEntities(queries []string) (map[string]SourceToAliases, error) {
	slog.Info("searching for entities queries", "queries", queries)

	queue := make(chan verificationTask, len(queries))
	for _, query := range queries {
		ndbMatches, err := store.ndbLookup(query)
		if err != nil {
			slog.Error("ndb entity lookup failed", "error", err)
		}
		flashMatches, err := store.flashLookup(query)
		if err != nil {
			slog.Error("flash lookup failed", "error", err)
		}

		matches := slices.Concat(ndbMatches, flashMatches)
		slog.Info("matches found for queries", "matches", matches)
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
			continue
		}

		sourceToAliases, err := store.exactLookup(task.Result.matches)
		if err != nil {
			slog.Error("error completing exact entity lookup", "error", err)
		} else {
			results[task.Result.query] = sourceToAliases
		}
	}

	return results, nil
}
