package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// TokenMiddleware returns an HTTP middleware that validates
// the Authorization: Bearer <token> header against the expected token.
func TokenMiddleware(expectedToken string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				unauthorized(w)
				logger.Warn("request rejected: missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				unauthorized(w)
				logger.Warn("request rejected: malformed authorization header")
				return
			}

			token := strings.TrimSpace(parts[1])
			if token == "" || token != expectedToken {
				unauthorized(w)
				logger.Warn("request rejected: invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
