package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NatalNW7/pombohook/internal/auth"
	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/queue"
	"github.com/NatalNW7/pombohook/internal/router"
	"github.com/NatalNW7/pombohook/internal/tunnel"
	"github.com/stretchr/testify/assert"
)

const testServerToken = "test-server-token"

func serverTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestServer() *Server {
	logger := serverTestLogger()
	cfg := &config.ServerConfig{
		Port:      "8080",
		AuthToken: testServerToken,
		LogLevel:  "info",
	}
	registry := router.NewRouteRegistry()
	tm := tunnel.NewTunnelManager(logger)
	q := queue.NewWebhookQueue(20)
	authMiddleware := auth.TokenMiddleware(cfg.AuthToken, logger)

	srv := NewServer(cfg, registry, tm, q, authMiddleware, logger)
	srv.SetupRoutes()
	return srv
}

func TestHandlers_Ping(t *testing.T) {
	t.Run("ping should return 200 with pong", func(t *testing.T) {
		srv := newTestServer()

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "Bearer "+testServerToken)
		rec := httptest.NewRecorder()

		srv.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "pong")
	})

	t.Run("ping should reject without auth", func(t *testing.T) {
		srv := newTestServer()

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()

		srv.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("ping should only accept GET", func(t *testing.T) {
		srv := newTestServer()

		req := httptest.NewRequest(http.MethodPost, "/ping", nil)
		req.Header.Set("Authorization", "Bearer "+testServerToken)
		rec := httptest.NewRecorder()

		srv.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}
