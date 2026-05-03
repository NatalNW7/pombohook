package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigStorage_SaveAndLoad(t *testing.T) {
	t.Run("should save and load config", func(t *testing.T) {
		s := newTestStorage(t)
		cfg := PomboConfig{Server: "wss://example.com", Token: "my-token"}

		err := s.SaveConfig(cfg)
		require.NoError(t, err)

		loaded, err := s.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, cfg, loaded)
	})

	t.Run("should return error when config not found", func(t *testing.T) {
		s := newTestStorage(t)

		_, err := s.LoadConfig()
		require.Error(t, err)
	})

	t.Run("should report config exists correctly", func(t *testing.T) {
		s := newTestStorage(t)

		assert.False(t, s.ConfigExists())

		err := s.SaveConfig(PomboConfig{Server: "wss://x.com", Token: "t"})
		require.NoError(t, err)

		assert.True(t, s.ConfigExists())
	})

	t.Run("should return error when unable to save config", func(t *testing.T) {
		s := NewStorage("/dev/null/foo")
		err := s.SaveConfig(PomboConfig{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating storage directory")
	})

	t.Run("should return error when config contains bad JSON", func(t *testing.T) {
		s := newTestStorage(t)
		err := s.ensureDir()
		require.NoError(t, err)
		
		err = os.WriteFile(s.filePath(configFile), []byte("invalid json"), 0600)
		require.NoError(t, err)

		_, err = s.LoadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing config")
	})
}
