package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLevel(t *testing.T) {
	t.Run("should return debug level", func(t *testing.T) {
		assert.Equal(t, slog.LevelDebug, parseLevel("debug"))
	})

	t.Run("should return warn level", func(t *testing.T) {
		assert.Equal(t, slog.LevelWarn, parseLevel("warn"))
	})

	t.Run("should return error level", func(t *testing.T) {
		assert.Equal(t, slog.LevelError, parseLevel("error"))
	})

	t.Run("should default to info for info string", func(t *testing.T) {
		assert.Equal(t, slog.LevelInfo, parseLevel("info"))
	})

	t.Run("should default to info for unknown string", func(t *testing.T) {
		assert.Equal(t, slog.LevelInfo, parseLevel("unknown"))
	})

	t.Run("should default to info for empty string", func(t *testing.T) {
		assert.Equal(t, slog.LevelInfo, parseLevel(""))
	})

	t.Run("should handle case insensitive input", func(t *testing.T) {
		assert.Equal(t, slog.LevelDebug, parseLevel("DEBUG"))
		assert.Equal(t, slog.LevelWarn, parseLevel("WARN"))
	})
}
