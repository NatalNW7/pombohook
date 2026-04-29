package tunnel

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NatalNW7/pombohook/internal/config"
)

// FrameType represents the type of a tunnel protocol frame.
type FrameType string

const (
	FrameTypeRequest  FrameType = "REQUEST"
	FrameTypeRegister FrameType = "REGISTER"
	FrameTypeACK      FrameType = "ACK"
	FrameTypeError    FrameType = "ERROR"
)

// Frame is the protocol unit exchanged between server and CLI over WebSocket.
type Frame struct {
	ID      string      `json:"id"`
	Type    FrameType   `json:"type"`
	Method  string      `json:"method,omitempty"`
	Path    string      `json:"path,omitempty"`
	Headers http.Header `json:"headers,omitempty"`
	Body    []byte      `json:"body,omitempty"`
}

// Encode serializes a Frame to JSON bytes.
func Encode(f Frame) ([]byte, error) {
	data, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("encoding frame: %w", err)
	}
	return data, nil
}

// Decode deserializes JSON bytes into a Frame.
func Decode(data []byte) (Frame, error) {
	var f Frame
	if err := json.Unmarshal(data, &f); err != nil {
		return Frame{}, fmt.Errorf("decoding frame: %w", err)
	}
	return f, nil
}

// NewRequestFrame creates a REQUEST frame with an auto-generated UUID.
func NewRequestFrame(method, path string, headers http.Header, body []byte) Frame {
	return Frame{
		ID:      generateUUID(),
		Type:    FrameTypeRequest,
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}

// NewRegisterFrame creates a REGISTER frame carrying route mappings in the body.
func NewRegisterFrame(routes []config.RouteMapping) Frame {
	body, _ := json.Marshal(routes)
	return Frame{
		ID:   generateUUID(),
		Type: FrameTypeRegister,
		Body: body,
	}
}

// NewACKFrame creates an ACK frame referencing the original frame's ID.
func NewACKFrame(id string) Frame {
	return Frame{
		ID:   id,
		Type: FrameTypeACK,
	}
}

// NewErrorFrame creates an ERROR frame with a message in the body.
func NewErrorFrame(id, message string) Frame {
	return Frame{
		ID:   id,
		Type: FrameTypeError,
		Body: []byte(message),
	}
}
