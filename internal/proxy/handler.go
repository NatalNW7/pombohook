package proxy

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/NatalNW7/pombohook/internal/queue"
	"github.com/NatalNW7/pombohook/internal/router"
	"github.com/NatalNW7/pombohook/internal/tunnel"
)

// TunnelSender defines the interface for sending frames to the connected CLI.
type TunnelSender interface {
	IsOnline() bool
	Send(frame tunnel.Frame) error
}

// Queuer defines the interface for queueing webhooks when CLI is offline.
type Queuer interface {
	Enqueue(frame tunnel.Frame) error
}

// ProxyHandler is the catch-all HTTP handler that forwards incoming webhooks to the CLI.
type ProxyHandler struct {
	registry *router.RouteRegistry
	tunnel   TunnelSender
	queue    Queuer
	logger   *slog.Logger
}

// NewProxyHandler creates a new ProxyHandler.
func NewProxyHandler(registry *router.RouteRegistry, tunnel TunnelSender, queue Queuer, logger *slog.Logger) *ProxyHandler {
	return &ProxyHandler{
		registry: registry,
		tunnel:   tunnel,
		queue:    queue,
		logger:   logger,
	}
}

// ServeHTTP handles incoming webhooks.
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log all incoming proxy requests for debugging
	h.logger.Debug("proxy handler received request", "path", r.URL.Path, "method", r.Method)

	// 1. Lookup r.URL.Path no registry
	_, found := h.registry.Lookup(r.URL.Path)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "route not found"})
		h.logger.Warn("webhook dropped: route not found", "path", r.URL.Path)
		return
	}

	// 2. Serializa request em Frame
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
	}

	frame := tunnel.NewRequestFrame(r.Method, r.URL.Path, r.Header, body)

	w.Header().Set("Content-Type", "application/json")

	// 3. Se tunnel online
	if h.tunnel.IsOnline() {
		// go tunnel.Send(frame) ← background
		go func() {
			if err := h.tunnel.Send(frame); err != nil {
				h.logger.Error("failed to send frame", "error", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "delivered"})
		h.logger.Info("webhook delivered to tunnel", "path", r.URL.Path)
		return
	}

	// 4. Se tunnel offline
	err := h.queue.Enqueue(frame)
	if err == queue.ErrQueueFull {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "queue full, try again later"})
		h.logger.Warn("webhook dropped: queue full", "path", r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
	h.logger.Info("webhook queued", "path", r.URL.Path)
}
