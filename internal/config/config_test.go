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

	t.Run("should default allowed origins to wildcard", func(t *testing.T) {
		t.Setenv("PORT", "8080")
		t.Setenv("POMBOHOOK_TOKEN", "my-secret-token")

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, []string{"*"}, cfg.AllowedOrigins)
	})

	t.Run("should parse allowed origins from env var", func(t *testing.T) {
		t.Setenv("PORT", "8080")
		t.Setenv("POMBOHOOK_TOKEN", "my-secret-token")
		t.Setenv("POMBOHOOK_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")

		cfg, err := LoadServerConfig()

		require.NoError(t, err)
		assert.Equal(t, []string{"https://example.com", "https://api.example.com"}, cfg.AllowedOrigins)
	})
}

func TestParseAllowedOrigins(t *testing.T) {
	t.Run("should default to wildcard when empty", func(t *testing.T) {
		result := parseAllowedOrigins("")
		assert.Equal(t, []string{"*"}, result)
	})

	t.Run("should default to wildcard when star", func(t *testing.T) {
		result := parseAllowedOrigins("*")
		assert.Equal(t, []string{"*"}, result)
	})

	t.Run("should parse comma-separated origins", func(t *testing.T) {
		result := parseAllowedOrigins("https://a.com,https://b.com")
		assert.Equal(t, []string{"https://a.com", "https://b.com"}, result)
	})

	t.Run("should handle single origin", func(t *testing.T) {
		result := parseAllowedOrigins("https://only.com")
		assert.Equal(t, []string{"https://only.com"}, result)
	})

	t.Run("should trim whitespace from origins", func(t *testing.T) {
		result := parseAllowedOrigins(" https://a.com , https://b.com ")
		assert.Equal(t, []string{"https://a.com", "https://b.com"}, result)
	})

	t.Run("should skip empty entries", func(t *testing.T) {
		result := parseAllowedOrigins("https://a.com,,https://b.com,")
		assert.Equal(t, []string{"https://a.com", "https://b.com"}, result)
	})
}
