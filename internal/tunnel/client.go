package tunnel

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/gorilla/websocket"
)

// FrameHandler is a callback function invoked for each received Frame.
type FrameHandler func(Frame)

// TunnelClient manages a WebSocket connection to the PomboHook server.
type TunnelClient struct {
	serverURL string
	token     string
	conn      *websocket.Conn
	handler   FrameHandler
	logger    *slog.Logger
	done      chan struct{}
}

// NewTunnelClient creates a new TunnelClient.
func NewTunnelClient(serverURL, token string, handler FrameHandler, logger *slog.Logger) *TunnelClient {
	return &TunnelClient{
		serverURL: serverURL,
		token:     token,
		handler:   handler,
		logger:    logger,
		done:      make(chan struct{}),
	}
}

// Connect dials the server WebSocket endpoint with the auth token in the header.
func (c *TunnelClient) Connect() error {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(c.serverURL, header)
	if err != nil {
		return fmt.Errorf("connecting to server: %w", err)
	}

	c.conn = conn
	c.logger.Info("connected to server", "url", c.serverURL)
	return nil
}

// SendRegister sends a REGISTER frame with routes and waits for an ACK response.
func (c *TunnelClient) SendRegister(routes []config.RouteMapping) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	frame := NewRegisterFrame(routes)
	data, err := Encode(frame)
	if err != nil {
		return fmt.Errorf("encoding register frame: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("sending register frame: %w", err)
	}

	c.logger.Info("register frame sent, waiting for ACK")

	// Wait for ACK
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("waiting for ACK: %w", err)
	}
	// Reset deadline
	c.conn.SetReadDeadline(time.Time{})

	ack, err := Decode(msg)
	if err != nil {
		return fmt.Errorf("decoding ACK: %w", err)
	}

	if ack.Type == FrameTypeError {
		return fmt.Errorf("server error: %s", string(ack.Body))
	}

	if ack.Type != FrameTypeACK || ack.ID != frame.ID {
		return fmt.Errorf("unexpected response: type=%s id=%s", ack.Type, ack.ID)
	}

	c.logger.Info("routes registered successfully")
	return nil
}

// Listen reads frames from the server and dispatches them to the handler.
// This method blocks until the connection is closed or an error occurs.
func (c *TunnelClient) Listen() {
	if c.conn == nil {
		return
	}

	defer close(c.done)

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.logger.Info("connection closed by server")
			} else {
				c.logger.Error("read error", "error", err)
			}
			return
		}

		frame, err := Decode(msg)
		if err != nil {
			c.logger.Error("failed to decode frame", "error", err)
			continue
		}

		c.handler(frame)
	}
}

// Close gracefully closes the WebSocket connection.
func (c *TunnelClient) Close() {
	if c.conn != nil {
		c.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		c.conn.Close()
		c.logger.Info("connection closed")
	}
}
