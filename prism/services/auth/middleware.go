package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIdContextKey contextKey = "user_id"
	emailContextKey  contextKey = "email_id"
)

type TokenVerifier interface {
	VerifyToken(token string) (uuid.UUID, string, error)
}

func Middleware(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			token, err := getToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			userId, email, err := verifier.VerifyToken(token)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			reqCtx := r.Context()
			reqCtx = context.WithValue(reqCtx, userIdContextKey, userId)
			reqCtx = context.WithValue(reqCtx, emailContextKey, email)
			next.ServeHTTP(w, r.WithContext(reqCtx))
		}

		return http.HandlerFunc(handler)
	}
}

func GetUserId(r *http.Request) (uuid.UUID, error) {
	userUntyped := r.Context().Value(userIdContextKey)
	if userUntyped == nil {
		return uuid.Nil, fmt.Errorf("user_id field not found in request context")
	}
	userId, ok := userUntyped.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid value for user_id field")
	}
	return userId, nil
}

func GetUserEmail(r *http.Request) (string, error) {
	emailUntyped := r.Context().Value(emailContextKey)
	if emailUntyped == nil {
		return "", fmt.Errorf("email_id field not found in request context")
	}
	email, ok := emailUntyped.(string)
	if !ok {
		return "", fmt.Errorf("invalid value for email_id field")
	}
	return email, nil
}
