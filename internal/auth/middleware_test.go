package auth

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testToken = "secret-pombo-token"

func authLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func protectedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
}

func TestTokenMiddleware(t *testing.T) {
	t.Run("should pass when valid token", func(t *testing.T) {
		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(protectedHandler())

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("should return 401 when no auth header", func(t *testing.T) {
		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(protectedHandler())

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "unauthorized")
	})

	t.Run("should return 401 when malformed header", func(t *testing.T) {
		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(protectedHandler())

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "NotBearer "+testToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 401 when wrong token", func(t *testing.T) {
		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(protectedHandler())

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 401 when empty bearer", func(t *testing.T) {
		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(protectedHandler())

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should not leak expected token in response", func(t *testing.T) {
		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(protectedHandler())

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "Bearer wrong")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.NotContains(t, rec.Body.String(), testToken)
	})

	t.Run("should pass request to next handler", func(t *testing.T) {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(next)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.True(t, called)
	})

	t.Run("should preserve request body for next handler", func(t *testing.T) {
		var receivedBody string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.WriteHeader(http.StatusOK)
		})

		middleware := TokenMiddleware(testToken, authLogger())
		handler := middleware(next)

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"data":"test"}`))
		req.Header.Set("Authorization", "Bearer "+testToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"data":"test"}`, receivedBody)
	})
}
