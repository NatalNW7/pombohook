package server

import (
	"encoding/json"
	"net/http"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/tunnel"
	"github.com/gorilla/websocket"
)

// handlePing responds with {"message":"pong"} for GET /ping.
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}

// handleWS upgrades the HTTP connection to a WebSocket and registers it with the tunnel manager.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("failed to upgrade websocket", "error", err)
		return
	}

	// Set connection immediately so we can use tunnel methods
	s.tunnel.SetConnection(conn)

	// Read the first frame which must be a RegisterFrame
	frame, err := s.tunnel.ReadFrame()
	if err != nil {
		s.logger.Error("failed to read register frame", "error", err)
		s.tunnel.RemoveConnection()
		conn.Close()
		return
	}

	if frame.Type != tunnel.FrameTypeRegister {
		s.logger.Error("expected register frame", "type", frame.Type)
		s.tunnel.RemoveConnection()
		conn.Close()
		return
	}

	tunnelID := r.RemoteAddr
	var req []config.RouteMapping
	if err := json.Unmarshal(frame.Body, &req); err != nil {
		s.logger.Error("invalid register frame body", "error", err)
		s.tunnel.Send(tunnel.NewErrorFrame("", "invalid register body"))
		return
	}

	// Clear previous routes for this tunnel and add new ones
	s.registry.UnregisterAll(tunnelID)
	for _, r := range req {
		s.registry.Register(r.Path, r.Port, tunnelID)
	}

	// Send ACK
	ackFrame := tunnel.NewACKFrame(frame.ID)
	s.tunnel.Send(ackFrame)

	s.logger.Info("tunnel client connected and registered", "tunnel_id", tunnelID, "routes", len(req))

	// Flush queued webhooks
	go func() {
		frames := s.queue.DrainAll()
		for _, queuedFrame := range frames {
			if sendErr := s.tunnel.Send(queuedFrame); sendErr != nil {
				s.logger.Error("failed to deliver queued frame", "error", sendErr)
				// Put it back in queue (simple retry logic)
				s.queue.Enqueue(queuedFrame)
				// Break to maintain ordering of the rest
				break
			}
			s.logger.Info("delivered queued webhook", "path", queuedFrame.Path)
		}
	}()

	// Listen for frames (acks, errors) and handle disconnect
	go func() {
		defer func() {
			s.registry.UnregisterAll(tunnelID)
			s.tunnel.RemoveConnection()
			conn.Close()
			s.logger.Info("tunnel client disconnected", "tunnel_id", tunnelID)
		}()
		
		for {
			_, err := s.tunnel.ReadFrame()
			if err != nil {
				break
			}
			// In Phase 9, we don't process acks/errors strictly, just drop them or log
		}
	}()
}
