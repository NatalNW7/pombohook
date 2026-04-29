package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *storage.Storage {
	t.Helper()
	return storage.NewStorage(t.TempDir())
}

// --- ping ---

func TestPing(t *testing.T) {
	t.Run("should save config on success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"pong"}`))
		}))
		defer srv.Close()

		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunPing(store, &buf, srv.URL, "valid-token")
		require.NoError(t, err)

		assert.True(t, store.ConfigExists())
		cfg, _ := store.LoadConfig()
		assert.Equal(t, srv.URL, cfg.Server)
		assert.Equal(t, "valid-token", cfg.Token)
		assert.Contains(t, buf.String(), "Connection established")
	})

	t.Run("should error when server unreachable", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunPing(store, &buf, "http://127.0.0.1:1", "token")
		require.Error(t, err)
		assert.Contains(t, buf.String(), "Could not reach server")
	})

	t.Run("should error when token invalid", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
		}))
		defer srv.Close()

		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunPing(store, &buf, srv.URL, "bad-token")
		require.Error(t, err)
		assert.Contains(t, buf.String(), "Authentication failed")
	})

	t.Run("should require server flag", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunPing(store, &buf, "", "token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server")
	})

	t.Run("should require token flag", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunPing(store, &buf, "http://localhost:8080", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token")
	})
}

// --- route ---

func TestRoute(t *testing.T) {
	t.Run("should add route", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunRouteAdd(store, &buf, "/webhook/mp", 8081)
		require.NoError(t, err)

		routes, _ := store.LoadRoutes()
		assert.Len(t, routes, 1)
		assert.Contains(t, buf.String(), "Route added")
	})

	t.Run("should reject duplicate", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		RunRouteAdd(store, &buf, "/webhook/mp", 8081)
		err := RunRouteAdd(store, &buf, "/webhook/mp", 9090)
		require.Error(t, err)
	})

	t.Run("should list routes", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		store.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 8081})
		store.AddRoute(config.RouteMapping{Path: "/webhook/stripe", Port: 3000})

		err := RunRouteList(store, &buf)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "/webhook/mp")
		assert.Contains(t, buf.String(), "/webhook/stripe")
	})

	t.Run("should remove route", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		store.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 8081})

		err := RunRouteRemove(store, &buf, "/webhook/mp")
		require.NoError(t, err)

		routes, _ := store.LoadRoutes()
		assert.Empty(t, routes)
		assert.Contains(t, buf.String(), "Route removed")
	})

	t.Run("should clear routes", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		store.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 8081})

		err := RunRouteClear(store, &buf)
		require.NoError(t, err)

		routes, _ := store.LoadRoutes()
		assert.Empty(t, routes)
		assert.Contains(t, buf.String(), "cleared")
	})

	t.Run("should validate path starts with slash", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunRouteAdd(store, &buf, "webhook/mp", 8081)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "/")
	})

	t.Run("should validate port positive", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunRouteAdd(store, &buf, "/webhook/mp", 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})
}

// --- go/sleep error paths ---

func TestGoErrors(t *testing.T) {
	t.Run("should error when no config", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := ValidateGoPrerequisites(store, &buf)
		require.Error(t, err)
		assert.Contains(t, buf.String(), "pombo ping")
	})

	t.Run("should error when no routes", func(t *testing.T) {
		store := newTestStore(t)
		store.SaveConfig(storage.PomboConfig{Server: "ws://x.com", Token: "t"})
		var buf bytes.Buffer

		err := ValidateGoPrerequisites(store, &buf)
		require.Error(t, err)
		assert.Contains(t, buf.String(), "pombo route")
	})
}

func TestSleepErrors(t *testing.T) {
	t.Run("should error when no pid", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunSleep(store, &buf)
		require.Error(t, err)
		assert.Contains(t, buf.String(), "No pigeon")
	})
}

func TestUnknownCommand(t *testing.T) {
	t.Run("should error on unknown command", func(t *testing.T) {
		var buf bytes.Buffer

		err := Dispatch(&buf, []string{"pombo", "fly"}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unknown command")
	})
}
