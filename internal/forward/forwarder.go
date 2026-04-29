package forward

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/NatalNW7/pombohook/internal/tunnel"
)

// Forwarder receives tunnel Frames and forwards them as HTTP requests to localhost.
type Forwarder struct {
	client *http.Client
	routes map[string]int // path → port
	logger *slog.Logger
}

// NewForwarder creates a new Forwarder with the given route map and logger.
func NewForwarder(routes map[string]int, logger *slog.Logger) *Forwarder {
	return &Forwarder{
		client: &http.Client{Timeout: 30 * time.Second},
		routes: routes,
		logger: logger,
	}
}

// Forward takes a REQUEST Frame and sends it to the matching localhost target.
func (f *Forwarder) Forward(frame tunnel.Frame) {
	port, found := f.routes[frame.Path]
	if !found {
		f.logger.Warn("no route for path", "path", frame.Path)
		return
	}

	url := f.buildURL(frame.Path, port)

	var body *bytes.Reader
	if frame.Body != nil {
		body = bytes.NewReader(frame.Body)
	} else {
		body = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(frame.Method, url, body)
	if err != nil {
		f.logger.Error("failed to create request", "path", frame.Path, "error", err)
		return
	}

	// Copy headers from frame
	for key, values := range frame.Headers {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	resp, err := f.client.Do(req)
	if err != nil {
		f.logger.Error("delivery failed", "path", frame.Path, "error", err)
		return
	}
	defer resp.Body.Close()

	f.logger.Info("webhook delivered",
		"method", frame.Method,
		"path", frame.Path,
		"target", url,
		"status", resp.StatusCode,
	)
}

// buildURL constructs the localhost URL from path and port.
func (f *Forwarder) buildURL(path string, port int) string {
	return fmt.Sprintf("http://localhost:%d%s", port, path)
}
