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

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// wsTestServer creates a test WebSocket server, returns the server and a channel
// that receives the server-side websocket.Conn once a client connects.
func wsTestServer(t *testing.T) (*httptest.Server, <-chan *websocket.Conn) {
	t.Helper()
	connCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		connCh <- conn
	}))
	t.Cleanup(srv.Close)
	return srv, connCh
}

// wsConnect dials the test server and returns the client-side websocket.Conn.
func wsConnect(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return conn
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestTunnelManager_Status(t *testing.T) {
	t.Run("should report offline when no connection", func(t *testing.T) {
		tm := NewTunnelManager(testLogger())
		assert.False(t, tm.IsOnline())
	})

	t.Run("should report online after set connection", func(t *testing.T) {
		tm := NewTunnelManager(testLogger())
		srv, connCh := wsTestServer(t)
		_ = wsConnect(t, srv)
		serverConn := <-connCh

		tm.SetConnection(serverConn)
		assert.True(t, tm.IsOnline())
	})

	t.Run("should report offline after remove connection", func(t *testing.T) {
		tm := NewTunnelManager(testLogger())
		srv, connCh := wsTestServer(t)
		_ = wsConnect(t, srv)
		serverConn := <-connCh

		tm.SetConnection(serverConn)
		tm.RemoveConnection()
		assert.False(t, tm.IsOnline())
	})
}

func TestTunnelManager_Send(t *testing.T) {
	t.Run("should send frame when online", func(t *testing.T) {
		tm := NewTunnelManager(testLogger())
		srv, connCh := wsTestServer(t)
		clientConn := wsConnect(t, srv)
		serverConn := <-connCh

		tm.SetConnection(serverConn)

		frame := NewRequestFrame("POST", "/webhook/mp", nil, []byte(`{"test":true}`))
		err := tm.Send(frame)
		require.NoError(t, err)

		// Read from client side
		_, msg, err := clientConn.ReadMessage()
		require.NoError(t, err)

		received, err := Decode(msg)
		require.NoError(t, err)
		assert.Equal(t, frame.ID, received.ID)
		assert.Equal(t, "/webhook/mp", received.Path)
	})

	t.Run("should return error when sending to offline", func(t *testing.T) {
		tm := NewTunnelManager(testLogger())

		frame := NewRequestFrame("POST", "/test", nil, nil)
		err := tm.Send(frame)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("should handle concurrent send calls", func(t *testing.T) {
		tm := NewTunnelManager(testLogger())
		srv, connCh := wsTestServer(t)
		clientConn := wsConnect(t, srv)
		serverConn := <-connCh
		tm.SetConnection(serverConn)

		const numSends = 20
		var wg sync.WaitGroup

		// Send concurrently
		errs := make([]error, numSends)
		for i := 0; i < numSends; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				frame := NewRequestFrame("POST", "/concurrent", nil, []byte(`{}`))
				errs[idx] = tm.Send(frame)
			}(i)
		}

		// Read all messages
		received := 0
		done := make(chan struct{})
		go func() {
			for i := 0; i < numSends; i++ {
				clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, _, err := clientConn.ReadMessage()
				if err != nil {
					break
				}
				received++
			}
			close(done)
		}()

		wg.Wait()
		<-done

		for _, err := range errs {
			assert.NoError(t, err)
		}
		assert.Equal(t, numSends, received)
	})
}

func TestTunnelManager_ReadFrame(t *testing.T) {
	t.Run("should read register frame from client", func(t *testing.T) {
		tm := NewTunnelManager(testLogger())
		srv, connCh := wsTestServer(t)
		clientConn := wsConnect(t, srv)
		serverConn := <-connCh
		tm.SetConnection(serverConn)

		// Client sends a REGISTER frame
		regFrame := NewRegisterFrame(nil)
		data, err := Encode(regFrame)
		require.NoError(t, err)
		err = clientConn.WriteMessage(websocket.TextMessage, data)
		require.NoError(t, err)

		// Server reads it
		received, err := tm.ReadFrame()
		require.NoError(t, err)
		assert.Equal(t, FrameTypeRegister, received.Type)
		assert.Equal(t, regFrame.ID, received.ID)
	})
}
