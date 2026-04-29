package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/NatalNW7/pombohook/internal/config"
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
	httpSrv  *http.Server
}

// NewServer creates a new Server with all dependencies injected.
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
	// /ws and catch-all will be added in later phases.
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.Port),
		Handler: s.mux,
	}

	s.logger.Info("starting server", "port", s.config.Port)
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}
