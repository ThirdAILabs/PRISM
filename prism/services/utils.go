package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"prism/prism/licensing"
	"prism/prism/reports"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ParseRequestBody[T any](r *http.Request) (T, error) {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		slog.Error("error parsing request body", "error", err)
		return data, fmt.Errorf("error parsing request body")
	}
	return data, nil
}

func GetStringFromMap(data map[string]interface{}, key string) (string, error) {
	result, ok := data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in map", key)
	}

	resultString, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("key %s is not a string", key)
	}
	return resultString, nil
}

func GetUUIDFromMap(data map[string]interface{}, key string) (uuid.UUID, error) {
	result, ok := data[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("key %s not found in map", key)
	}

	resultUUID, ok := result.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("key %s is not a uuid", key)
	}
	return resultUUID, nil
}

func GetTimeFromMap(data map[string]interface{}, key string) (time.Time, error) {
	result, ok := data[key]
	if !ok {
		return time.Time{}, fmt.Errorf("key %s not found in map", key)
	}

	resultTime, ok := result.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("key %s is not a time", key)
	}
	return resultTime, nil
}

func GetBoolFromMap(data map[string]interface{}, key string) (bool, error) {
	result, ok := data[key]
	if !ok {
		return false, fmt.Errorf("key %s not found in map", key)
	}

	resultBool, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("key %s is not a bool", key)
	}
	return resultBool, nil
}

type codedError struct {
	err  error
	code int
}

func (e *codedError) Error() string {
	return e.err.Error()
}

func (e *codedError) Unwrap() error {
	return e.err
}

func CodedError(err error, code int) error {
	return &codedError{err: err, code: code}
}

type RestHandler func(r *http.Request) (any, error)

func WrapRestHandler(handler RestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := handler(r)
		if err != nil {
			var cerr *codedError
			if errors.As(err, &cerr) {
				http.Error(w, err.Error(), cerr.code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if res == nil {
			res = struct{}{}
		}

		WriteJsonResponse(w, res)
	}
}

func WriteJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("error serializing response body", "error", err)
		http.Error(w, fmt.Sprintf("error serializing response body: %v", err), http.StatusInternalServerError)
	}
}

func URLParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := chi.URLParam(r, key)

	if len(param) == 0 {
		return uuid.Nil, fmt.Errorf("missing {%v} url parameter", key)
	}

	id, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid '%v' provided: %w", param, err)
	}

	return id, nil
}

func licensingErrorStatus(err error) int {
	switch {
	case errors.Is(err, licensing.ErrLicenseVerificationFailed):
		return http.StatusInternalServerError

	case errors.Is(err, licensing.ErrExpiredLicense):
		return http.StatusForbidden

	case errors.Is(err, licensing.ErrLicenseNotFound):
		return http.StatusNotFound

	case errors.Is(err, licensing.ErrInvalidLicense):
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

func reportErrorStatus(err error) int {
	switch {
	case errors.Is(err, reports.ErrReportNotFound):
		return http.StatusNotFound
	case errors.Is(err, reports.ErrUserCannotAccessReport):
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}
