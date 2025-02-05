package flaggers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"prism/openalex"
	"strings"

	"go.etcd.io/bbolt"
)

var cacheBucket = []byte("ack-cache")

type AcknowledgementsExtractor struct {
	cache      *bbolt.DB
	maxWorkers int
}

type Entity struct {
	Entity   string
	Type     string
	Position int
}

type Acknowledgement struct {
	RawText            string
	SearchableEntities []Entity
	MiscEntities       []Entity
}

type Acknowledgements struct {
	OpenAlexId       string
	Acknowledgements []Acknowledgement
}

func (extractor *AcknowledgementsExtractor) GetAcknowledgements(works []openalex.Work) (chan Acknowledgements, chan error) {
	outputCh := make(chan Acknowledgements, len(works))
	errorCh := make(chan error, 10)

	queue := make(chan openalex.Work, len(works))

	for _, work := range works {
		id := parseOpenAlexId(work)
		if id == "" {
			continue
		}

		cachedAck, err := extractor.checkCache(id)
		if err != nil {
			slog.Error("error checking cache for acknowledgement", "id", id, "error", err)
			errorCh <- fmt.Errorf("error checking acknowledgment cache: %w", err)
		} else if cachedAck != nil {
			outputCh <- *cachedAck
			continue
		}

		queue <- work
	}
	close(queue)

	nWorkers := min(len(queue), extractor.maxWorkers)
	for i := 0; i < nWorkers; i++ {
		go extractor.worker(queue, outputCh, errorCh)
	}

	return outputCh, errorCh
}

func (extractor *AcknowledgementsExtractor) worker(queue chan openalex.Work, outputCh chan Acknowledgements, errorCh chan error) {
	for {
		next, done := <-queue
		if done {
			return
		}

		id := parseOpenAlexId(next)

		acks, err := extractor.extractAcknowledgments(next)
		if err != nil {
			slog.Error("error extracting acknowledgements for work", "id", id, "name", next.DisplayName, "error", err)
			errorCh <- fmt.Errorf("error extracting acknowledgments: %w", err)
		} else {
			outputCh <- acks
		}

		extractor.updateCache(id, acks)
	}
}

func parseOpenAlexId(work openalex.Work) string {
	idx := strings.LastIndex(work.WorkId, "/")
	if idx < 0 {
		return ""
	}
	return work.WorkId[idx+1:]
}

func (extractor *AcknowledgementsExtractor) checkCache(id string) (*Acknowledgements, error) {
	slog.Info("checking acknowledgement cache", "id", id)
	var data []byte
	err := extractor.cache.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(cacheBucket)

		data = bucket.Get([]byte(id))

		return nil
	})

	if err != nil {
		slog.Error("cache access failed", "id", id, "error", err)
		return nil, fmt.Errorf("cache access failed: %w", err)
	}

	if data == nil {
		slog.Info("no cached acknowledgements found", "id", id)
		return nil, nil
	}

	var acks Acknowledgements
	if err := json.Unmarshal(data, &acks); err != nil {
		slog.Info("error parsing cache data", "id", id, "error", err)
		return nil, fmt.Errorf("error parsing cache data: %w", err)
	}

	slog.Info("found cached acknowledgements", "id", id)

	return &acks, nil
}

func (extractor *AcknowledgementsExtractor) updateCache(id string, acks Acknowledgements) {
	slog.Info("updating acknowledgement cache", "id", id)

	data, err := json.Marshal(acks)
	if err != nil {
		slog.Error("error updating acknowledgements cache: error serializing data", "id", id, "error", err)
		return // No error since cache update isn't critical
	}

	if err := extractor.cache.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(cacheBucket).Put([]byte(id), data)
	}); err != nil {
		slog.Error("error updating acknowledgements cache: bolt db error", "id", id, "error", err)
	}

	slog.Info("successfully updated acknowledgements cache", "id", id)
}

func (extractor *AcknowledgementsExtractor) extractAcknowledgments(work openalex.Work) (Acknowledgements, error) {

}
