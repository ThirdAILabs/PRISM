package entity_search

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

type Record[T any] struct {
	Entity   string
	Metadata T
}

type recordCnt struct {
	recordId uint32
	cnt      uint32
}

type invertedIndex map[string][]recordCnt

type EntityIndex[T any] struct {
	records []Record[T]

	index    invertedIndex
	recLens  []int
	totalLen int

	ngram int
	k1    float32
	b     float32
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

		for token, cnt := range recordTokens {
			index[token] = append(index[token], recordCnt{uint32(recId), cnt})
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
			score := bm25(idf, float32(record.cnt), float32(index.recLens[record.recordId]), avgLen, index.k1, index.b)
			candidates[record.recordId] += score
		}
	}

	pairs := make([]candidatePair, 0, len(candidates))
	for recordId, score := range candidates {
		pairs = append(pairs, candidatePair{recordId, score})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})

	topk := min(k, len(pairs))
	results := make([]Record[T], 0, topk)
	for i := range topk {
		results = append(results, index.records[pairs[i].recordId])
	}

	return results
}
