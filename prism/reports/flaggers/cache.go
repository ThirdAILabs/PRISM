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
}

func NewCache[T any](bucket, path string) (DataCache[T], error) {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		slog.Error("error opening cache db", "bucket", bucket, "error", err)
		return DataCache[T]{}, fmt.Errorf("error creating cache: %w", err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		slog.Error("error creating cache bucket", "bucket", bucket, "error", err)
		return DataCache[T]{}, fmt.Errorf("error creating cache: %w", err)
	}

	return DataCache[T]{db: db, bucket: []byte(bucket)}, nil
}

func (cache *DataCache[T]) Close() error {
	return cache.db.Close()
}

func (cache *DataCache[T]) Lookup(key string) *T {
	slog.Info("checking cache", "bucket", string(cache.bucket), "key", key)

	var data []byte
	err := cache.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(cache.bucket)

		data = bucket.Get([]byte(key))

		return nil
	})
	if err != nil {
		slog.Error("cache access failed", "bucket", string(cache.bucket), "key", key, "error", err)
		return nil
	}

	if data == nil {
		slog.Info("no cached entry found", "bucket", string(cache.bucket), "key", key)
		return nil
	}

	entry := new(T)
	if err := json.Unmarshal(data, entry); err != nil {
		slog.Info("error parsing cache data", "bucket", string(cache.bucket), "key", key, "error", err)
		return nil
	}

	slog.Info("found cached entry", "bucket", string(cache.bucket), "key", key)

	return entry
}

func (cache *DataCache[T]) Update(key string, entry T) {
	slog.Info("updating cache", "bucket", string(cache.bucket), "key", key)

	data, err := json.Marshal(entry)
	if err != nil {
		slog.Error("error updating cache: error serializing data", "bucket", string(cache.bucket), "key", key, "error", err)
		return // No error since cache update isn't critical
	}

	if err := cache.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(cache.bucket).Put([]byte(key), data)
	}); err != nil {
		slog.Error("cache update failed", "bucket", string(cache.bucket), "key", key, "error", err)
	}

	slog.Info("successfully updated cache", "bucket", string(cache.bucket), "key", key)
}
