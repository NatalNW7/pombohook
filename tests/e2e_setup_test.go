package tests

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/NatalNW7/pombohook/internal/auth"
	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/forward"
	"github.com/NatalNW7/pombohook/internal/queue"
	"github.com/NatalNW7/pombohook/internal/router"
	"github.com/NatalNW7/pombohook/internal/server"
	"github.com/NatalNW7/pombohook/internal/tunnel"
)

func e2eLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// setupE2EEnvironment sets up an in-memory server, a tunnel client, and a target localhost mock.
func setupE2EEnvironment(t *testing.T, queueCapacity int) (*httptest.Server, *httptest.Server, *tunnel.TunnelClient, func() string, func() string, func(string), func(string), int) {
	logger := e2eLogger()
	token := "e2e-token"

	// 1. Setup Server
	cfg := &config.ServerConfig{
		Port:      "8080",
		AuthToken: token,
		LogLevel:  "debug",
	}

	registry := router.NewRouteRegistry()
	tunnelMgr := tunnel.NewTunnelManager(logger)
	webhookQueue := queue.NewWebhookQueue(queueCapacity)
	authMW := auth.TokenMiddleware(token, logger)

	srv := server.NewServer(cfg, registry, tunnelMgr, webhookQueue, authMW, logger)
	srv.SetupRoutes()
	tsServer := httptest.NewServer(srv.Handler())

	// 2. Setup "localhost" target mock
	var mu sync.Mutex
	var receivedBodyStr string
	var receivedMethodStr string

	// Create mock target
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		receivedBodyStr = string(body)
		receivedMethodStr = r.Method
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))

	getReceivedBody := func() string {
		mu.Lock()
		defer mu.Unlock()
		return receivedBodyStr
	}

	getReceivedMethod := func() string {
		mu.Lock()
		defer mu.Unlock()
		return receivedMethodStr
	}

	setReceivedBody := func(val string) {
		mu.Lock()
		defer mu.Unlock()
		receivedBodyStr = val
	}

	setReceivedMethod := func(val string) {
		mu.Lock()
		defer mu.Unlock()
		receivedMethodStr = val
	}

	// Extract target port
	parts := strings.Split(targetServer.URL, ":")
	var targetPort int
	for _, c := range parts[len(parts)-1] {
		targetPort = targetPort*10 + int(c-'0')
	}

	// 3. Setup CLI Client & Forwarder
	wsURL := strings.Replace(tsServer.URL, "http://", "ws://", 1) + "/ws"
	routes := map[string]int{"/webhook/test": targetPort}
	fwd := forward.NewForwarder(routes, logger)
	client := tunnel.NewTunnelClient(
		wsURL,
		token,
		func(f tunnel.Frame) {
			if f.Type == tunnel.FrameTypeRequest {
				fwd.Forward(f)
			}
		},
		logger,
	)

	return tsServer, targetServer, client, getReceivedBody, getReceivedMethod, setReceivedBody, setReceivedMethod, targetPort
}
