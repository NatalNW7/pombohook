package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	return NewStorage(t.TempDir())
}

func TestStorage_CreateBaseDirectory(t *testing.T) {
	t.Run("should create base directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		basePath := tmpDir + "/nested/pombo"
		s := NewStorage(basePath)

		err := s.SaveConfig(PomboConfig{Server: "wss://x.com", Token: "t"})
		require.NoError(t, err)

		assert.True(t, s.ConfigExists())
	})

	t.Run("should return base path", func(t *testing.T) {
		s := NewStorage("/tmp/test-pombo")
		assert.Equal(t, "/tmp/test-pombo", s.BasePath())
	})
}
