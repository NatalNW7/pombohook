package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadServerConfig(t *testing.T) {
	t.Run("should return config when all env vars set", func(t *testing.T) {
		t.Setenv("PORT", "8080")
		t.Setenv("POMBOHOOK_TOKEN", "my-secret-token")

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.Port)
		assert.Equal(t, "my-secret-token", cfg.AuthToken)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("should return error when PORT missing", func(t *testing.T) {
		t.Setenv("POMBOHOOK_TOKEN", "my-secret-token")

		_, err := LoadServerConfig()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "PORT")
	})

	t.Run("should return error when TOKEN missing", func(t *testing.T) {
		t.Setenv("PORT", "8080")

		_, err := LoadServerConfig()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "POMBOHOOK_TOKEN")
	})

	t.Run("should use default log level when not set", func(t *testing.T) {
		t.Setenv("PORT", "8080")
		t.Setenv("POMBOHOOK_TOKEN", "my-secret-token")

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("should override log level when set", func(t *testing.T) {
		t.Setenv("PORT", "8080")
		t.Setenv("POMBOHOOK_TOKEN", "my-secret-token")
		t.Setenv("LOG_LEVEL", "debug")

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.LogLevel)
	})
}
