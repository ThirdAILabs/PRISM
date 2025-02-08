package flaggers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"go.etcd.io/bbolt"
)

type DataCache[T any] struct {
	db     *bbolt.DB
	bucket []byte
	logger *slog.Logger
}

func NewCache[T any](bucket, path string) (DataCache[T], error) {
	logger := slog.With("bucket", bucket)

	logger.Info("creating new cache", "path", path)

	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		logger.Error("error opening cache db", "error", err)
		return DataCache[T]{}, fmt.Errorf("error creating cache: %w", err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		logger.Error("error creating cache bucket", "error", err)
		return DataCache[T]{}, fmt.Errorf("error creating cache: %w", err)
	}

	logger.Info("cache initialized")

	return DataCache[T]{db: db, bucket: []byte(bucket), logger: logger}, nil
}

func (cache *DataCache[T]) Close() error {
	return cache.db.Close()
}

func (cache *DataCache[T]) Lookup(key string) *T {
	cache.logger.Info("checking cache", "key", key)

	var entry *T
	err := cache.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(cache.bucket)

		data := bucket.Get([]byte(key))
		if data != nil {
			entry = new(T)
			if err := json.Unmarshal(data, entry); err != nil {
				return fmt.Errorf("error parsing cache data: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		cache.logger.Error("cache access failed", "key", key, "error", err)
		return nil // No error since cache update isn't critical
	}

	if entry != nil {
		cache.logger.Info("found cached entry", "key", key)
	} else {
		cache.logger.Info("no cached entry found", "key", key)
	}

	return entry
}

func (cache *DataCache[T]) Update(key string, entry T) {
	cache.logger.Info("updating cache", "key", key)

	data, err := json.Marshal(entry)
	if err != nil {
		cache.logger.Error("error updating cache: error serializing data", "key", key, "error", err)
		return // No error since cache update isn't critical
	}

	if err := cache.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(cache.bucket).Put([]byte(key), data)
	}); err != nil {
		cache.logger.Error("cache update failed", "key", key, "error", err)
		return // No error since cache update isn't critical
	}

	cache.logger.Info("successfully updated cache", "key", key)
}
