package forward

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NatalNW7/pombohook/internal/tunnel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func forwarderLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestForwarder_Forward(t *testing.T) {
	t.Run("should forward POST preserving path", func(t *testing.T) {
		var receivedMethod, receivedPath string
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			receivedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		port := extractPort(t, target.URL)
		routes := map[string]int{"/webhook/mp": port}
		fwd := NewForwarder(routes, forwarderLogger())

		frame := tunnel.NewRequestFrame("POST", "/webhook/mp", nil, []byte(`{"event":"pay"}`))
		fwd.Forward(frame)

		assert.Equal(t, "POST", receivedMethod)
		assert.Equal(t, "/webhook/mp", receivedPath)
	})

	t.Run("should forward GET preserving path", func(t *testing.T) {
		var receivedMethod string
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		port := extractPort(t, target.URL)
		routes := map[string]int{"/webhook/verify": port}
		fwd := NewForwarder(routes, forwarderLogger())

		frame := tunnel.NewRequestFrame("GET", "/webhook/verify", nil, nil)
		fwd.Forward(frame)

		assert.Equal(t, "GET", receivedMethod)
	})

	t.Run("should forward PUT preserving path", func(t *testing.T) {
		var receivedMethod string
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		port := extractPort(t, target.URL)
		routes := map[string]int{"/webhook/update": port}
		fwd := NewForwarder(routes, forwarderLogger())

		frame := tunnel.NewRequestFrame("PUT", "/webhook/update", nil, []byte(`{}`))
		fwd.Forward(frame)

		assert.Equal(t, "PUT", receivedMethod)
	})
}

func TestForwarder_Quality(t *testing.T) {
	t.Run("should preserve all headers", func(t *testing.T) {
		var receivedContentType, receivedCustom string
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedContentType = r.Header.Get("Content-Type")
			receivedCustom = r.Header.Get("X-Custom")
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		port := extractPort(t, target.URL)
		routes := map[string]int{"/hook": port}
		fwd := NewForwarder(routes, forwarderLogger())

		headers := http.Header{}
		headers.Set("Content-Type", "application/json")
		headers.Set("X-Custom", "pombo-value")

		frame := tunnel.NewRequestFrame("POST", "/hook", headers, []byte(`{}`))
		fwd.Forward(frame)

		assert.Equal(t, "application/json", receivedContentType)
		assert.Equal(t, "pombo-value", receivedCustom)
	})

	t.Run("should preserve body exactly", func(t *testing.T) {
		var receivedBody string
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		port := extractPort(t, target.URL)
		routes := map[string]int{"/hook": port}
		fwd := NewForwarder(routes, forwarderLogger())

		payload := `{"id":123,"event":"payment.created","data":{"amount":99.99}}`
		frame := tunnel.NewRequestFrame("POST", "/hook", nil, []byte(payload))
		fwd.Forward(frame)

		assert.Equal(t, payload, receivedBody)
	})
}

func TestForwarder_BuildURL(t *testing.T) {
	t.Run("should build url with path and port", func(t *testing.T) {
		fwd := NewForwarder(nil, forwarderLogger())

		url := fwd.buildURL("/webhook/mp", 8081)
		assert.Equal(t, "http://localhost:8081/webhook/mp", url)
	})
}

func TestForwarder_Errors(t *testing.T) {
	t.Run("should log error when target offline", func(t *testing.T) {
		routes := map[string]int{"/hook": 1} // port 1 = closed
		fwd := NewForwarder(routes, forwarderLogger())

		frame := tunnel.NewRequestFrame("POST", "/hook", nil, []byte(`{}`))
		// Should not panic — just logs the error
		fwd.Forward(frame)
	})

	t.Run("should handle unknown path", func(t *testing.T) {
		routes := map[string]int{"/known": 8080}
		fwd := NewForwarder(routes, forwarderLogger())

		frame := tunnel.NewRequestFrame("POST", "/unknown/path", nil, nil)
		// Should not panic — just logs warning
		fwd.Forward(frame)
	})
}

func TestForwarder_EdgeCases(t *testing.T) {
	t.Run("should handle empty body frame", func(t *testing.T) {
		var receivedBody []byte
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		port := extractPort(t, target.URL)
		routes := map[string]int{"/hook": port}
		fwd := NewForwarder(routes, forwarderLogger())

		frame := tunnel.NewRequestFrame("POST", "/hook", nil, nil)
		fwd.Forward(frame)

		assert.Empty(t, receivedBody)
	})
}

// extractPort parses the port number from an httptest URL like "http://127.0.0.1:PORT"
func extractPort(t *testing.T, rawURL string) int {
	t.Helper()
	parts := strings.Split(rawURL, ":")
	require.Len(t, parts, 3)
	var port int
	_, err := io.ReadAll(strings.NewReader(parts[2]))
	require.NoError(t, err)
	// Parse the port
	for _, c := range parts[2] {
		port = port*10 + int(c-'0')
	}
	return port
}
