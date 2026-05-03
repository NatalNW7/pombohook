package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NatalNW7/pombohook/internal/auth"
	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/queue"
	"github.com/NatalNW7/pombohook/internal/router"
	"github.com/NatalNW7/pombohook/internal/tunnel"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// newTestServerWithDeps creates a server returning all deps for direct inspection.
func newTestServerWithDeps() (*Server, *router.RouteRegistry, *tunnel.TunnelManager, *queue.WebhookQueue) {
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
	return srv, registry, tm, q
}

func wsAuthHeader() http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+testServerToken)
	return h
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

func TestHandlers_WS(t *testing.T) {
	t.Run("ws should accept connection and register routes", func(t *testing.T) {
		srv := newTestServer()
		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, wsAuthHeader())
		require.NoError(t, err)
		defer conn.Close()

		// Send REGISTER
		routes := []config.RouteMapping{{Path: "/webhook/mp", Port: 8081}}
		regFrame := tunnel.NewRegisterFrame(routes)
		data, _ := tunnel.Encode(regFrame)
		conn.WriteMessage(websocket.TextMessage, data)

		// Read ACK
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn.ReadMessage()
		require.NoError(t, err)

		ack, err := tunnel.Decode(msg)
		require.NoError(t, err)
		assert.Equal(t, tunnel.FrameTypeACK, ack.Type)
		assert.Equal(t, regFrame.ID, ack.ID)
	})

	t.Run("ws should register routes in registry", func(t *testing.T) {
		srv, registry, _, _ := newTestServerWithDeps()
		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, wsAuthHeader())
		require.NoError(t, err)
		defer conn.Close()

		routes := []config.RouteMapping{{Path: "/webhook/mp", Port: 8081}}
		regFrame := tunnel.NewRegisterFrame(routes)
		data, _ := tunnel.Encode(regFrame)
		conn.WriteMessage(websocket.TextMessage, data)

		// Read ACK
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		conn.ReadMessage()
		time.Sleep(50 * time.Millisecond)

		// Verify route is registered
		route, found := registry.Lookup("/webhook/mp")
		assert.True(t, found)
		assert.Equal(t, 8081, route.Port)
	})

	t.Run("ws should flush queued webhooks on connect", func(t *testing.T) {
		srv, _, _, q := newTestServerWithDeps()

		// Pre-queue a frame
		queuedFrame := tunnel.NewRequestFrame("POST", "/webhook/mp", nil, []byte(`{"q":1}`))
		q.Enqueue(queuedFrame)

		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, wsAuthHeader())
		require.NoError(t, err)
		defer conn.Close()

		routes := []config.RouteMapping{{Path: "/webhook/mp", Port: 8081}}
		regFrame := tunnel.NewRegisterFrame(routes)
		data, _ := tunnel.Encode(regFrame)
		conn.WriteMessage(websocket.TextMessage, data)

		// Read ACK first
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		conn.ReadMessage()

		// Then read the flushed queued frame
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn.ReadMessage()
		require.NoError(t, err)

		flushed, err := tunnel.Decode(msg)
		require.NoError(t, err)
		assert.Equal(t, tunnel.FrameTypeRequest, flushed.Type)
		assert.Equal(t, "/webhook/mp", flushed.Path)

		// Queue should be empty now
		assert.Equal(t, 0, q.Len())
	})

	t.Run("ws should cleanup routes on disconnect", func(t *testing.T) {
		srv, registry, _, _ := newTestServerWithDeps()
		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, wsAuthHeader())
		require.NoError(t, err)

		routes := []config.RouteMapping{{Path: "/webhook/mp", Port: 8081}}
		regFrame := tunnel.NewRegisterFrame(routes)
		data, _ := tunnel.Encode(regFrame)
		conn.WriteMessage(websocket.TextMessage, data)

		// Read ACK
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		conn.ReadMessage()
		time.Sleep(50 * time.Millisecond)

		// Verify route exists before disconnect
		_, found := registry.Lookup("/webhook/mp")
		assert.True(t, found)

		// Close connection
		conn.Close()
		time.Sleep(200 * time.Millisecond)

		// Routes should be cleaned up
		_, found = registry.Lookup("/webhook/mp")
		assert.False(t, found)
	})

	t.Run("ws should require auth", func(t *testing.T) {
		srv := newTestServer()
		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
		// No auth header
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestServer_StartAndShutdown(t *testing.T) {
	t.Run("should start listening and shutdown gracefully", func(t *testing.T) {
		logger := serverTestLogger()
		cfg := &config.ServerConfig{
			Port:      "0",
			AuthToken: testServerToken,
			LogLevel:  "info",
		}
		registry := router.NewRouteRegistry()
		tm := tunnel.NewTunnelManager(logger)
		q := queue.NewWebhookQueue(20)
		authMW := auth.TokenMiddleware(cfg.AuthToken, logger)

		srv := NewServer(cfg, registry, tm, q, authMW, logger)
		srv.SetupRoutes()

		errCh := make(chan error, 1)
		go func() { errCh <- srv.Start() }()

		// Give server time to bind
		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		require.NoError(t, err)

		startErr := <-errCh
		assert.ErrorIs(t, startErr, http.ErrServerClosed)
	})

	t.Run("shutdown before start should be no-op", func(t *testing.T) {
		srv := newTestServer()
		err := srv.Shutdown(context.Background())
		assert.NoError(t, err)
	})
}
