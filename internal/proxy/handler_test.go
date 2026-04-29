package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/NatalNW7/pombohook/internal/queue"
	"github.com/NatalNW7/pombohook/internal/router"
	"github.com/NatalNW7/pombohook/internal/tunnel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTunnelSender struct {
	online    bool
	frames    []tunnel.Frame
	mu        sync.Mutex
	waitGroup sync.WaitGroup
}

func (m *mockTunnelSender) IsOnline() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.online
}

func (m *mockTunnelSender) Send(frame tunnel.Frame) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.frames = append(m.frames, frame)
	m.waitGroup.Done()
	return nil
}

func proxyTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestProxyHandler(online bool, qCap int) (*ProxyHandler, *mockTunnelSender, *queue.WebhookQueue, *router.RouteRegistry) {
	reg := router.NewRouteRegistry()
	ts := &mockTunnelSender{online: online}
	q := queue.NewWebhookQueue(qCap)
	logger := proxyTestLogger()

	handler := NewProxyHandler(reg, ts, q, logger)
	return handler, ts, q, reg
}

func TestProxyHandler_ServeHTTP(t *testing.T) {
	t.Run("should return 404 for unregistered path", func(t *testing.T) {
		handler, _, _, _ := newTestProxyHandler(true, 20)

		req := httptest.NewRequest(http.MethodPost, "/unknown", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "route not found")
	})

	t.Run("should return 200 when cli online", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/webhook/mp", 8080, "t1")

		req := httptest.NewRequest(http.MethodPost, "/webhook/mp", nil)
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "delivered")
	})

	t.Run("should send frame in background when online", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/webhook/mp", 8080, "t1")

		req := httptest.NewRequest(http.MethodPost, "/webhook/mp", bytes.NewReader([]byte(`{"a":1}`)))
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)

		ts.waitGroup.Wait() // wait for background goroutine to call Send
		ts.mu.Lock()
		defer ts.mu.Unlock()

		require.Len(t, ts.frames, 1)
		frame := ts.frames[0]
		assert.Equal(t, "POST", frame.Method)
		assert.Equal(t, "/webhook/mp", frame.Path)
		assert.Equal(t, `{"a":1}`, string(frame.Body))
	})

	t.Run("should return 202 when cli offline and queue ok", func(t *testing.T) {
		handler, _, q, reg := newTestProxyHandler(false, 20)
		reg.Register("/webhook/mp", 8080, "t1")

		req := httptest.NewRequest(http.MethodPost, "/webhook/mp", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)
		assert.Contains(t, rec.Body.String(), "queued")
		assert.Equal(t, 1, q.Len())
	})

	t.Run("should return 503 when cli offline and queue full", func(t *testing.T) {
		handler, _, _, reg := newTestProxyHandler(false, 1)
		reg.Register("/webhook/mp", 8080, "t1")

		// Fill queue
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/webhook/mp", nil))

		// Try again
		req := httptest.NewRequest(http.MethodPost, "/webhook/mp", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Contains(t, rec.Body.String(), "queue full")
	})
}

func TestProxyHandler_Quality(t *testing.T) {
	t.Run("should preserve POST method in frame", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/hook", 8080, "t1")

		req := httptest.NewRequest(http.MethodPost, "/hook", nil)
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)
		ts.waitGroup.Wait()

		assert.Equal(t, "POST", ts.frames[0].Method)
	})

	t.Run("should preserve GET method in frame", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/hook", 8080, "t1")

		req := httptest.NewRequest(http.MethodGet, "/hook", nil)
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)
		ts.waitGroup.Wait()

		assert.Equal(t, "GET", ts.frames[0].Method)
	})

	t.Run("should preserve PUT method in frame", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/hook", 8080, "t1")

		req := httptest.NewRequest(http.MethodPut, "/hook", nil)
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)
		ts.waitGroup.Wait()

		assert.Equal(t, "PUT", ts.frames[0].Method)
	})

	t.Run("should preserve path in frame", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/hook", 8080, "t1")

		req := httptest.NewRequest(http.MethodPost, "/hook", nil)
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)
		ts.waitGroup.Wait()

		assert.Equal(t, "/hook", ts.frames[0].Path)
	})

	t.Run("should preserve headers in frame", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/hook", 8080, "t1")

		req := httptest.NewRequest(http.MethodPost, "/hook", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Custom", "pombo")
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)
		ts.waitGroup.Wait()

		assert.Equal(t, "application/json", ts.frames[0].Headers.Get("Content-Type"))
		assert.Equal(t, "pombo", ts.frames[0].Headers.Get("X-Custom"))
	})

	t.Run("should preserve body in frame", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/hook", 8080, "t1")

		payload := `{"test": 123}`
		req := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader([]byte(payload)))
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)
		ts.waitGroup.Wait()

		assert.Equal(t, payload, string(ts.frames[0].Body))
	})

	t.Run("should handle empty body", func(t *testing.T) {
		handler, ts, _, reg := newTestProxyHandler(true, 20)
		reg.Register("/hook", 8080, "t1")

		req := httptest.NewRequest(http.MethodPost, "/hook", nil)
		rec := httptest.NewRecorder()

		ts.waitGroup.Add(1)
		handler.ServeHTTP(rec, req)
		ts.waitGroup.Wait()

		assert.Empty(t, ts.frames[0].Body)
	})
}
