package services

import (
	"encoding/json"
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

func WriteJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("error serializing response body", "error", err)
		http.Error(w, fmt.Sprintf("error serializing response body: %v", err), http.StatusInternalServerError)
	}
}

func WriteSuccess(w http.ResponseWriter) {
	WriteJsonResponse(w, struct{}{})
}
