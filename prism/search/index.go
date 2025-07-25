package search

import (
	"fmt"
	"math"
	"prism/prism/llms"
	"sort"
	"strings"
	"unicode"
)

type Record[T any] struct {
	Entity   string
	Metadata T
}

type tokenFreq struct {
	RecordId uint32
	Freq     uint32
}

type invertedIndex map[string][]tokenFreq

type EntityIndex[T any] struct {
	records []Record[T]

	index    invertedIndex
	recLens  []int
	totalLen int

	ngram int
	k1    float32
	b     float32

	llm llms.LLM
}

func tokenize(str string, ngram int) []string {
	lowercaseStr := strings.ToLower(str)

	var result strings.Builder
	for _, char := range lowercaseStr {
		if unicode.IsPunct(char) {
			result.WriteRune(' ')
		} else {
			result.WriteRune(char)
		}
	}

	tokens := make([]string, 0)
	for _, word := range strings.Fields(result.String()) {
		tokens = append(tokens, word)

		for i := 1; i < len(word); i++ {
			tokens = append(tokens, word[max(0, i-ngram):i])
		}
	}

	return tokens
}

func buildIndex[T any](records []Record[T], ngram int) (int, []int, invertedIndex) {
	totalLen := 0
	recordLens := make([]int, 0, len(records))

	index := make(invertedIndex)

	for recId, record := range records {
		tokens := tokenize(record.Entity, ngram)
		recordLens = append(recordLens, len(tokens))
		totalLen += len(tokens)

		recordTokens := make(map[string]uint32)
		for _, token := range tokens {
			recordTokens[token]++
		}

		for token, freq := range recordTokens {
			index[token] = append(index[token], tokenFreq{RecordId: uint32(recId), Freq: freq})
		}
	}

	return totalLen, recordLens, index
}

func idf(tf, n float32) float32 {
	tf_ := float64(tf)
	n_ := float64(n)

	return float32(math.Log(1.0 + (n_-tf_+0.5)/(tf_+0.5)))
}

func bm25(idf, recTf, len, avglen, k1, b float32) float32 {
	num := recTf * (k1 + 1.0)
	denom := recTf + k1*(1.0-b+b*len/avglen)

	return idf * num / denom
}

const (
	default_k1    = 1.2
	default_b     = 0.75
	default_ngram = 4
)

type candidatePair struct {
	recordId uint32
	score    float32
}

func NewIndex[T any](records []Record[T]) *EntityIndex[T] {
	totalLen, recLens, index := buildIndex(records, default_ngram)

	return &EntityIndex[T]{
		records:  records,
		index:    index,
		recLens:  recLens,
		totalLen: totalLen,
		ngram:    default_ngram,
		k1:       default_k1,
		b:        default_b,
		llm:      llms.New(),
	}
}

func (index *EntityIndex[T]) Query(query string, k int) []Record[T] {
	candidates := make(map[uint32]float32)

	avgLen := float32(index.totalLen) / float32(len(index.records))
	for _, token := range tokenize(query, index.ngram) {
		records, ok := index.index[token]
		if !ok {
			continue
		}

		tf := float32(len(records))
		idf := idf(tf, float32(len(index.records)))

		for _, record := range records {
			score := bm25(idf, float32(record.Freq), float32(index.recLens[record.RecordId]), avgLen, index.k1, index.b)
			candidates[record.RecordId] += score
		}
	}

	pairs := make([]candidatePair, 0, len(candidates))
	for recordId, score := range candidates {
		pairs = append(pairs, candidatePair{recordId, score})
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score == pairs[j].score {
			return pairs[i].recordId < pairs[j].recordId
		}
		return pairs[i].score > pairs[j].score
	})

	topk := min(k, len(pairs))
	results := make([]Record[T], 0, topk)
	for i := range topk {
		results = append(results, index.records[pairs[i].recordId])
	}

	return results
}

func (index *EntityIndex[T]) QueryWithLLMValidation(query string, k int) ([]Record[T], error) {
	results := index.Query(query, k)
	return index.llmValidate(query, results)
}

const llmValidationPromptTemplate = `Return a python list containing True or False based on whether this entity: %s, is an equivalent entity to the following entities.
Here are the entities: ["%s"]
DO NOT ENTER 'FALSE' JUST BECAUSE THEY HAVE A DIFFERENT NAME, they can still be the same entity with a different name (e.g. same person, organization, school, etc. with different aliases).
Answer with a python list of True or False, and nothing else. Do not provide any explanation or extra words/characters. Ensure that the python syntax is absolutely correct and runnable. Also ensure that the resulting list is the same length as the list of entities.
Answer:
`

func (index *EntityIndex[T]) llmValidate(query string, results []Record[T]) ([]Record[T], error) {
	if len(results) == 0 {
		return results, nil
	}

	matchStrings := make([]string, 0, len(results))
	for _, result := range results {
		matchStrings = append(matchStrings, result.Entity)
	}

	res, err := index.llm.Generate(fmt.Sprintf(llmValidationPromptTemplate, query, strings.Join(matchStrings, `", "`)), &llms.Options{
		Model:        llms.GPT4oMini,
		ZeroTemp:     true,
		SystemPrompt: "You are a helpful python assistant who responds in python lists only.",
	})
	if err != nil {
		return nil, fmt.Errorf("llm match verification failed: %w", err)
	}

	flags := strings.Split(res, ",")

	if len(flags) != len(matchStrings) {
		return nil, fmt.Errorf("llm verification returned incorrect number of flags query=%v, matches=%v, flags=%v", query, matchStrings, flags)
	}

	filtered := make([]Record[T], 0, len(results))
	for i, result := range results {
		if strings.Contains(flags[i], "True") {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

type ManyToOneIndex[T any] struct {
	metadata []T
	index    EntityIndex[int]
}

func NewManyToOneIndex[T any](entities [][]string, metadata []T) *ManyToOneIndex[T] {
	records := make([]Record[int], 0, len(entities))
	for i, entitySet := range entities {
		for _, entity := range entitySet {
			records = append(records, Record[int]{Entity: entity, Metadata: i})
		}
	}

	return &ManyToOneIndex[T]{
		metadata: metadata,
		index:    *NewIndex(records),
	}
}

func (index *ManyToOneIndex[T]) Query(query string, k int) []Record[T] {
	indexResults := index.index.Query(query, k)

	results := make([]Record[T], 0, len(indexResults))
	for _, result := range indexResults {
		results = append(results, Record[T]{Entity: result.Entity, Metadata: index.metadata[result.Metadata]})
	}

	return results
}
