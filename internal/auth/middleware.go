// Package auth also contains middleware utilities for securing API endpoints using JWT.
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// WithAuth is middleware that validates JWT tokens from the Authorization header
// before allowing access to the next HTTP handler.
func WithAuth(next http.HandlerFunc, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			http.Error(w, " is null unauth", http.StatusUnauthorized)
			return
		}
		logrus.Infof("header: %v", h)
		parts := strings.SplitN(h, " ", 2)
		logrus.Infof("parts: %v", parts)
		if len(parts) != 2 {
			http.Error(w, "length unauth", http.StatusUnauthorized)
			return
		}
		tok := parts[1]
		parsed, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) { return []byte(secret), nil })
		if err != nil || !parsed.Valid {
			http.Error(w, "parse unauth", http.StatusUnauthorized)
			return
		}
		// attach claims
		ctx := context.WithValue(r.Context(), "claims", parsed.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// JSON writes the provided value as a JSON response with the appropriate headers.
func JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
