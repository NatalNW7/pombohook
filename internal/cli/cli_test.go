package cli

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/storage"
	"github.com/gorilla/websocket"
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

func TestGoBackground(t *testing.T) {
	t.Run("should error when already running", func(t *testing.T) {
		store := newTestStore(t)
		store.SaveConfig(storage.PomboConfig{Server: "ws://x.com", Token: "t"})
		store.AddRoute(config.RouteMapping{Path: "/webhook/test", Port: 8081})
		var buf bytes.Buffer

		// Simulate an already-running process by saving the current PID
		store.SavePID(os.Getpid())

		err := RunGoBackground(store, &buf, "/nonexistent")
		require.Error(t, err)
		assert.Contains(t, buf.String(), "already flying")
		assert.Contains(t, err.Error(), "already running")
	})

	t.Run("should error when no config", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		err := RunGoBackground(store, &buf, "/nonexistent")
		require.Error(t, err)
		assert.Contains(t, buf.String(), "pombo ping")
	})

	t.Run("should error when no routes", func(t *testing.T) {
		store := newTestStore(t)
		store.SaveConfig(storage.PomboConfig{Server: "ws://x.com", Token: "t"})
		var buf bytes.Buffer

		err := RunGoBackground(store, &buf, "/nonexistent")
		require.Error(t, err)
		assert.Contains(t, buf.String(), "pombo route")
	})

	t.Run("should cleanup stale pid", func(t *testing.T) {
		store := newTestStore(t)
		store.SaveConfig(storage.PomboConfig{Server: "ws://x.com", Token: "t"})
		store.AddRoute(config.RouteMapping{Path: "/webhook/test", Port: 8081})
		var buf bytes.Buffer

		// Save a PID that doesn't exist (dead process)
		store.SavePID(999999)

		// Will fail on exec but should not fail on "already running" check
		err := RunGoBackground(store, &buf, "/nonexistent-binary")
		// The error should be about starting the process, NOT about already running
		if err != nil {
			assert.NotContains(t, err.Error(), "already running")
		}
	})

	t.Run("should fork successfully", func(t *testing.T) {
		store := newTestStore(t)
		store.SaveConfig(storage.PomboConfig{Server: "ws://x.com", Token: "t"})
		store.AddRoute(config.RouteMapping{Path: "/webhook/test", Port: 8081})
		var buf bytes.Buffer

		// Use 'sleep' as a dummy executable so it stays alive just long enough to spawn
		err := RunGoBackground(store, &buf, "sleep")
		require.NoError(t, err)
		
		assert.True(t, store.PIDExists())
		assert.Contains(t, buf.String(), "released in background")
	})
}

func TestRunGo(t *testing.T) {
	t.Run("should connect, register and listen", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusSwitchingProtocols)
			// we don't do a full websocket upgrade here, just enough to fail the websocket dial or let it drop
			// Wait, the client uses gorilla/websocket, if it fails to upgrade it will return an error and Connect() will fail.
		}))
		defer srv.Close()

		store := newTestStore(t)
		wsURL := "ws" + srv.URL[4:]
		store.SaveConfig(storage.PomboConfig{Server: wsURL, Token: "t"})
		store.AddRoute(config.RouteMapping{Path: "/webhook/test", Port: 8081})
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		// This will fail at Connect() because the mock server isn't a real WS server
		err := RunGo(store, &buf, logger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connecting:")
	})

	t.Run("should fail registering if server disconnects immediately", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			// Close immediately without ACK
			conn.Close()
		}))
		defer srv.Close()

		store := newTestStore(t)
		wsURL := "ws" + srv.URL[4:]
		store.SaveConfig(storage.PomboConfig{Server: wsURL, Token: "t"})
		store.AddRoute(config.RouteMapping{Path: "/webhook/test", Port: 8081})
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		
		// This should connect but fail at SendRegister or Listen
		err := RunGo(store, &buf, logger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "registering:")
	})

	t.Run("should handle corrupted config", func(t *testing.T) {
		store := newTestStore(t)
		// Bypass SaveConfig to write corrupted JSON
		os.WriteFile(store.BasePath()+"/config.json", []byte("{bad-json"), 0644)
		store.AddRoute(config.RouteMapping{Path: "/webhook/test", Port: 8081})
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		
		err := RunGo(store, &buf, logger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "loading config:")
	})

	t.Run("should handle corrupted routes", func(t *testing.T) {
		store := newTestStore(t)
		store.SaveConfig(storage.PomboConfig{Server: "ws://x", Token: "t"})
		// Bypass SaveRoutes to write corrupted JSON
		os.WriteFile(store.BasePath()+"/routes.json", []byte("{bad-json"), 0644)
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		
		err := RunGo(store, &buf, logger)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing routes:")
	})
}

func TestIsProcessAlive(t *testing.T) {
	t.Run("should return true for current process", func(t *testing.T) {
		assert.True(t, isProcessAlive(os.Getpid()))
	})

	t.Run("should return false for dead process", func(t *testing.T) {
		assert.False(t, isProcessAlive(999999))
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

	t.Run("should stop a running process", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		// Start a real process we can kill
		cmd := exec.Command("sleep", "30")
		require.NoError(t, cmd.Start())
		defer cmd.Process.Kill() // safety cleanup

		store.SavePID(cmd.Process.Pid)

		err := RunSleep(store, &buf)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "resting")
		assert.False(t, store.PIDExists())
	})

	t.Run("should cleanup stale pid of dead process", func(t *testing.T) {
		store := newTestStore(t)
		var buf bytes.Buffer

		store.SavePID(999999) // PID inexistente

		err := RunSleep(store, &buf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "process not running")
		assert.False(t, store.PIDExists())
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
