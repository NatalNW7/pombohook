package tunnel

import (
	"net/http"
	"strings"
	"testing"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtocol_Encode(t *testing.T) {
	t.Run("should encode frame to valid json", func(t *testing.T) {
		frame := Frame{
			ID:     "test-id",
			Type:   FrameTypeRequest,
			Method: "POST",
			Path:   "/webhook/mp",
			Body:   []byte(`{"event":"payment"}`),
		}

		data, err := Encode(frame)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"id":"test-id"`)
		assert.Contains(t, string(data), `"type":"REQUEST"`)
	})
}

func TestProtocol_Decode(t *testing.T) {
	t.Run("should decode valid json to frame", func(t *testing.T) {
		raw := `{"id":"abc","type":"REQUEST","method":"POST","path":"/hook"}`

		frame, err := Decode([]byte(raw))
		require.NoError(t, err)
		assert.Equal(t, "abc", frame.ID)
		assert.Equal(t, FrameTypeRequest, frame.Type)
		assert.Equal(t, "POST", frame.Method)
		assert.Equal(t, "/hook", frame.Path)
	})

	t.Run("should return error when decoding invalid json", func(t *testing.T) {
		_, err := Decode([]byte(`{invalid`))
		require.Error(t, err)
	})
}

func TestProtocol_Roundtrip(t *testing.T) {
	t.Run("should roundtrip encode decode", func(t *testing.T) {
		original := Frame{
			ID:      "roundtrip-id",
			Type:    FrameTypeRequest,
			Method:  "PUT",
			Path:    "/webhook/stripe",
			Headers: http.Header{"Content-Type": []string{"application/json"}},
			Body:    []byte(`{"key":"value"}`),
		}

		data, err := Encode(original)
		require.NoError(t, err)

		decoded, err := Decode(data)
		require.NoError(t, err)

		assert.Equal(t, original.ID, decoded.ID)
		assert.Equal(t, original.Type, decoded.Type)
		assert.Equal(t, original.Method, decoded.Method)
		assert.Equal(t, original.Path, decoded.Path)
		assert.Equal(t, original.Headers.Get("Content-Type"), decoded.Headers.Get("Content-Type"))
		assert.Equal(t, original.Body, decoded.Body)
	})
}

func TestProtocol_EdgeCases(t *testing.T) {
	t.Run("should handle empty body", func(t *testing.T) {
		frame := Frame{ID: "e1", Type: FrameTypeRequest, Method: "GET", Path: "/test"}

		data, err := Encode(frame)
		require.NoError(t, err)

		decoded, err := Decode(data)
		require.NoError(t, err)
		assert.Empty(t, decoded.Body)
	})

	t.Run("should handle nil headers", func(t *testing.T) {
		frame := Frame{ID: "e2", Type: FrameTypeRequest, Method: "GET", Path: "/test"}

		data, err := Encode(frame)
		require.NoError(t, err)

		decoded, err := Decode(data)
		require.NoError(t, err)
		assert.Empty(t, decoded.Headers)
	})

	t.Run("should handle large body", func(t *testing.T) {
		largeBody := []byte(strings.Repeat("x", 1024*1024)) // 1MB

		frame := Frame{ID: "e3", Type: FrameTypeRequest, Method: "POST", Path: "/big", Body: largeBody}

		data, err := Encode(frame)
		require.NoError(t, err)

		decoded, err := Decode(data)
		require.NoError(t, err)
		assert.Equal(t, len(largeBody), len(decoded.Body))
	})
}

func TestProtocol_Quality(t *testing.T) {
	t.Run("should preserve body bytes exactly", func(t *testing.T) {
		binaryBody := []byte{0x00, 0x01, 0xFF, 0xFE, 0x80}
		frame := Frame{ID: "q1", Type: FrameTypeRequest, Method: "POST", Path: "/bin", Body: binaryBody}

		data, err := Encode(frame)
		require.NoError(t, err)

		decoded, err := Decode(data)
		require.NoError(t, err)
		assert.Equal(t, binaryBody, decoded.Body)
	})

	t.Run("should generate unique ids for request frames", func(t *testing.T) {
		f1 := NewRequestFrame("GET", "/a", nil, nil)
		f2 := NewRequestFrame("GET", "/b", nil, nil)

		assert.NotEmpty(t, f1.ID)
		assert.NotEmpty(t, f2.ID)
		assert.NotEqual(t, f1.ID, f2.ID)
	})
}

func TestProtocol_Factories(t *testing.T) {
	t.Run("should create register frame with routes", func(t *testing.T) {
		routes := []config.RouteMapping{
			{Path: "/webhook/mp", Port: 8081},
			{Path: "/webhook/stripe", Port: 3000},
		}

		frame := NewRegisterFrame(routes)

		assert.Equal(t, FrameTypeRegister, frame.Type)
		assert.NotEmpty(t, frame.ID)
		// Body should contain serialized routes
		assert.NotEmpty(t, frame.Body)
		assert.Contains(t, string(frame.Body), "/webhook/mp")
	})

	t.Run("should create ack frame with matching id", func(t *testing.T) {
		frame := NewACKFrame("original-id")

		assert.Equal(t, FrameTypeACK, frame.Type)
		assert.Equal(t, "original-id", frame.ID)
	})

	t.Run("should create error frame with message", func(t *testing.T) {
		frame := NewErrorFrame("err-id", "something went wrong")

		assert.Equal(t, FrameTypeError, frame.Type)
		assert.Equal(t, "err-id", frame.ID)
		assert.Contains(t, string(frame.Body), "something went wrong")
	})
}
