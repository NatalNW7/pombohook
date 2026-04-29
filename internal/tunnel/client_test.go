package tunnel

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// echoWSServer creates a test WS server that upgrades and optionally sends frames.
func echoWSServer(t *testing.T, onConnect func(conn *websocket.Conn)) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		if onConnect != nil {
			onConnect(conn)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func clientTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestTunnelClient_Connect(t *testing.T) {
	t.Run("should connect to server successfully", func(t *testing.T) {
		done := make(chan struct{})
		srv := echoWSServer(t, func(conn *websocket.Conn) {
			close(done)
			// Keep connection alive briefly
			time.Sleep(100 * time.Millisecond)
		})

		wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
		client := NewTunnelClient(wsURL, "test-token", func(f Frame) {}, clientTestLogger())

		err := client.Connect()
		require.NoError(t, err)
		defer client.Close()

		select {
		case <-done:
			// Server received connection
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for connection")
		}
	})

	t.Run("should fail connect with invalid url", func(t *testing.T) {
		client := NewTunnelClient("ws://127.0.0.1:1", "token", func(f Frame) {}, clientTestLogger())

		err := client.Connect()
		require.Error(t, err)
	})
}

func TestTunnelClient_SendRegister(t *testing.T) {
	t.Run("should send register frame", func(t *testing.T) {
		receivedCh := make(chan Frame, 1)
		srv := echoWSServer(t, func(conn *websocket.Conn) {
			// Read REGISTER frame
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			f, _ := Decode(msg)
			receivedCh <- f

			// Send ACK back
			ack := NewACKFrame(f.ID)
			data, _ := Encode(ack)
			conn.WriteMessage(websocket.TextMessage, data)

			time.Sleep(100 * time.Millisecond)
		})

		wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
		client := NewTunnelClient(wsURL, "test-token", func(f Frame) {}, clientTestLogger())
		err := client.Connect()
		require.NoError(t, err)
		defer client.Close()

		routes := []config.RouteMapping{{Path: "/webhook/mp", Port: 8081}}
		err = client.SendRegister(routes)
		require.NoError(t, err)

		select {
		case f := <-receivedCh:
			assert.Equal(t, FrameTypeRegister, f.Type)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for register frame")
		}
	})
}

func TestTunnelClient_Listen(t *testing.T) {
	t.Run("should receive and dispatch request frame", func(t *testing.T) {
		handledCh := make(chan Frame, 1)
		handler := func(f Frame) {
			handledCh <- f
		}

		srv := echoWSServer(t, func(conn *websocket.Conn) {
			// Send a REQUEST frame to client
			frame := NewRequestFrame("POST", "/webhook/mp", nil, []byte(`{"event":"pay"}`))
			data, _ := Encode(frame)
			conn.WriteMessage(websocket.TextMessage, data)

			time.Sleep(200 * time.Millisecond)
		})

		wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
		client := NewTunnelClient(wsURL, "test-token", handler, clientTestLogger())
		err := client.Connect()
		require.NoError(t, err)
		defer client.Close()

		go client.Listen()

		select {
		case f := <-handledCh:
			assert.Equal(t, FrameTypeRequest, f.Type)
			assert.Equal(t, "/webhook/mp", f.Path)
			assert.Equal(t, "POST", f.Method)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for frame dispatch")
		}
	})

	t.Run("should call handler for each received frame", func(t *testing.T) {
		var mu sync.Mutex
		var handled []Frame
		handler := func(f Frame) {
			mu.Lock()
			handled = append(handled, f)
			mu.Unlock()
		}

		srv := echoWSServer(t, func(conn *websocket.Conn) {
			for i := 0; i < 3; i++ {
				frame := NewRequestFrame("POST", "/multi", nil, nil)
				data, _ := Encode(frame)
				conn.WriteMessage(websocket.TextMessage, data)
				time.Sleep(10 * time.Millisecond)
			}
			time.Sleep(200 * time.Millisecond)
		})

		wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
		client := NewTunnelClient(wsURL, "test-token", handler, clientTestLogger())
		err := client.Connect()
		require.NoError(t, err)
		defer client.Close()

		go client.Listen()
		time.Sleep(500 * time.Millisecond)

		mu.Lock()
		assert.Len(t, handled, 3)
		mu.Unlock()
	})

	t.Run("should handle connection close gracefully", func(t *testing.T) {
		srv := echoWSServer(t, func(conn *websocket.Conn) {
			// Close immediately
			conn.Close()
		})

		wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
		client := NewTunnelClient(wsURL, "test-token", func(f Frame) {}, clientTestLogger())
		err := client.Connect()
		require.NoError(t, err)

		// Listen should return without panic when server closes
		client.Listen()
		// If we reach here, graceful shutdown worked
	})
}
