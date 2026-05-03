package tunnel

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// TunnelManager manages a single WebSocket connection to a CLI client (V1: single-developer).
type TunnelManager struct {
	mu       sync.RWMutex
	conn     *websocket.Conn
	isOnline bool
	logger   *slog.Logger
}

// NewTunnelManager creates a new TunnelManager.
func NewTunnelManager(logger *slog.Logger) *TunnelManager {
	return &TunnelManager{logger: logger}
}

// SetConnection registers a WebSocket connection as the active tunnel.
func (tm *TunnelManager) SetConnection(conn *websocket.Conn) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.conn = conn
	tm.isOnline = true
	tm.logger.Info("tunnel connected")
}

// RemoveConnection marks the tunnel as offline and clears the connection.
func (tm *TunnelManager) RemoveConnection() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.conn = nil
	tm.isOnline = false
	tm.logger.Info("tunnel disconnected")
}

// IsOnline returns true if a CLI client is currently connected.
func (tm *TunnelManager) IsOnline() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.isOnline
}

// Send serializes and sends a Frame to the connected CLI via WebSocket.
// Returns an error if no client is connected.
func (tm *TunnelManager) Send(frame Frame) error {
	tm.mu.RLock()
	if !tm.isOnline || tm.conn == nil {
		tm.mu.RUnlock()
		return fmt.Errorf("tunnel not connected")
	}
	conn := tm.conn
	tm.mu.RUnlock()

	data, err := Encode(frame)
	if err != nil {
		return fmt.Errorf("encoding frame: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("writing to websocket: %w", err)
	}

	return nil
}

// ReadFrame reads and decodes the next Frame from the connected CLI.
func (tm *TunnelManager) ReadFrame() (Frame, error) {
	tm.mu.RLock()
	conn := tm.conn
	tm.mu.RUnlock()

	if conn == nil {
		return Frame{}, fmt.Errorf("tunnel not connected")
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		return Frame{}, fmt.Errorf("reading from websocket: %w", err)
	}

	return Decode(msg)
}
