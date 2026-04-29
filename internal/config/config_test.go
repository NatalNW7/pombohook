package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearEnv(t *testing.T) {
	t.Helper()
	os.Unsetenv("PORT")
	os.Unsetenv("POMBOHOOK_TOKEN")
	os.Unsetenv("LOG_LEVEL")
}

func TestLoadServerConfig(t *testing.T) {
	t.Run("should return config when all env vars set", func(t *testing.T) {
		clearEnv(t)
		os.Setenv("PORT", "8080")
		os.Setenv("POMBOHOOK_TOKEN", "my-secret-token")
		defer clearEnv(t)

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.Port)
		assert.Equal(t, "my-secret-token", cfg.AuthToken)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("should return error when PORT missing", func(t *testing.T) {
		clearEnv(t)
		os.Setenv("POMBOHOOK_TOKEN", "my-secret-token")
		defer clearEnv(t)

		_, err := LoadServerConfig()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "PORT")
	})

	t.Run("should return error when TOKEN missing", func(t *testing.T) {
		clearEnv(t)
		os.Setenv("PORT", "8080")
		defer clearEnv(t)

		_, err := LoadServerConfig()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "POMBOHOOK_TOKEN")
	})

	t.Run("should use default log level when not set", func(t *testing.T) {
		clearEnv(t)
		os.Setenv("PORT", "8080")
		os.Setenv("POMBOHOOK_TOKEN", "my-secret-token")
		defer clearEnv(t)

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("should override log level when set", func(t *testing.T) {
		clearEnv(t)
		os.Setenv("PORT", "8080")
		os.Setenv("POMBOHOOK_TOKEN", "my-secret-token")
		os.Setenv("LOG_LEVEL", "debug")
		defer clearEnv(t)

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.LogLevel)
	})
}
