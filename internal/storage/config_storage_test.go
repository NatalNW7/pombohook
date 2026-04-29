package storage

import (
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
}
