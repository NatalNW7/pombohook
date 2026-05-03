package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPIDStorage_SaveAndLoad(t *testing.T) {
	t.Run("should save and load pid", func(t *testing.T) {
		s := newTestStorage(t)

		err := s.SavePID(12345)
		require.NoError(t, err)

		pid, err := s.LoadPID()
		require.NoError(t, err)
		assert.Equal(t, 12345, pid)
	})

	t.Run("should remove pid file", func(t *testing.T) {
		s := newTestStorage(t)
		s.SavePID(12345)

		err := s.RemovePID()
		require.NoError(t, err)

		assert.False(t, s.PIDExists())
	})

	t.Run("should report pid exists correctly", func(t *testing.T) {
		s := newTestStorage(t)

		assert.False(t, s.PIDExists())

		s.SavePID(99)
		assert.True(t, s.PIDExists())
	})

	t.Run("should return error when unable to save pid", func(t *testing.T) {
		s := NewStorage("/dev/null/foo")
		err := s.SavePID(123)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating storage directory")
	})

	t.Run("should return error when pid file has invalid content", func(t *testing.T) {
		s := newTestStorage(t)
		err := s.ensureDir()
		require.NoError(t, err)
		
		err = os.WriteFile(s.filePath(pidFile), []byte("not-an-int"), 0600)
		require.NoError(t, err)

		_, err = s.LoadPID()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing PID")
	})
}
