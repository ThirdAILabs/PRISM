package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func ParseRequestBody[T any](r *http.Request) (T, error) {
	var data T
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		slog.Error("error parsing request body", "error", err)
		return data, fmt.Errorf("error parsing request body")
	}
	return data, nil
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
