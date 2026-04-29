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
	"github.com/stretchr/testify/require"
)

func TestServer_Create(t *testing.T) {
	t.Run("should create server with valid config", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		cfg := &config.ServerConfig{
			Port:      "9090",
			AuthToken: "tok",
			LogLevel:  "info",
		}
		registry := router.NewRouteRegistry()
		tm := tunnel.NewTunnelManager(logger)
		q := queue.NewWebhookQueue(20)
		authMW := auth.TokenMiddleware(cfg.AuthToken, logger)

		srv := NewServer(cfg, registry, tm, q, authMW, logger)
		require.NotNil(t, srv)
	})

	t.Run("should route ping to handler", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		cfg := &config.ServerConfig{
			Port:      "9090",
			AuthToken: "tok",
			LogLevel:  "info",
		}
		registry := router.NewRouteRegistry()
		tm := tunnel.NewTunnelManager(logger)
		q := queue.NewWebhookQueue(20)
		authMW := auth.TokenMiddleware(cfg.AuthToken, logger)

		srv := NewServer(cfg, registry, tm, q, authMW, logger)
		srv.SetupRoutes()

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rec := httptest.NewRecorder()

		srv.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "pong")
	})
}
