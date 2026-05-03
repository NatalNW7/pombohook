package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/proxy"
	"github.com/NatalNW7/pombohook/internal/queue"
	"github.com/NatalNW7/pombohook/internal/router"
	"github.com/NatalNW7/pombohook/internal/tunnel"
)

// Server is the PomboHook API server.
type Server struct {
	config   *config.ServerConfig
	registry *router.RouteRegistry
	tunnel   *tunnel.TunnelManager
	queue    *queue.WebhookQueue
	auth     func(http.Handler) http.Handler
	logger   *slog.Logger
	mux      *http.ServeMux
	mu       sync.Mutex
	httpSrv  *http.Server
}

// NewServer creates a new Server with all dependencies injected.
// Design note: authMiddleware is injected as a function parameter (not stored as a field)
// for better dependency injection and testability. This diverges from the original spec
// which showed auth as a struct field.
func NewServer(
	cfg *config.ServerConfig,
	registry *router.RouteRegistry,
	tm *tunnel.TunnelManager,
	q *queue.WebhookQueue,
	authMiddleware func(http.Handler) http.Handler,
	logger *slog.Logger,
) *Server {
	return &Server{
		config:   cfg,
		registry: registry,
		tunnel:   tm,
		queue:    q,
		auth:     authMiddleware,
		logger:   logger,
		mux:      http.NewServeMux(),
	}
}

// SetupRoutes configures the HTTP mux with all endpoints.
func (s *Server) SetupRoutes() {
	s.mux.Handle("/ping", s.auth(http.HandlerFunc(s.handlePing)))

	// Tunnel connection endpoint
	s.mux.Handle("/ws", s.auth(http.HandlerFunc(s.handleWS)))

	// Proxy handler (catch-all)
	// We only wrap proxy in auth if we want webhooks to be authenticated,
	// but generally webhooks from external providers (Stripe, MP) won't have our internal token.
	// We'll leave the catch-all without our TokenMiddleware, relying on the target to validate signatures.
	proxyHandler := proxy.NewProxyHandler(s.registry, s.tunnel, s.queue, s.logger)
	s.mux.Handle("/", proxyHandler)
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	s.mu.Lock()
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.Port),
		Handler: s.mux,
	}
	s.mu.Unlock()

	s.logger.Info("starting server", "port", s.config.Port)
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

// Handler returns the server's HTTP handler (mux) for testing.
func (s *Server) Handler() http.Handler {
	return s.mux
}
