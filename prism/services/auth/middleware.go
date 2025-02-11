package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type TokenVerifier interface {
	VerifyToken(token string) (uuid.UUID, error)
}

func Middleware(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			token, err := getToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			userId, err := verifier.VerifyToken(token)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			reqCtx := r.Context()
			reqCtx = context.WithValue(reqCtx, userIdContextKey, userId)
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
