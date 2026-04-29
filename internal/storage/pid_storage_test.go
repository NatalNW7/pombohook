package storage

import (
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
}
