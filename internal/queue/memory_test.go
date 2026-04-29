package queue

import (
	"sync"
	"testing"

	"github.com/NatalNW7/pombohook/internal/tunnel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookQueue_Enqueue(t *testing.T) {
	t.Run("should enqueue frame successfully", func(t *testing.T) {
		q := NewWebhookQueue(20)
		frame := tunnel.NewRequestFrame("POST", "/hook", nil, nil)

		err := q.Enqueue(frame)
		require.NoError(t, err)
		assert.Equal(t, 1, q.Len())
	})

	t.Run("should enqueue up to max capacity", func(t *testing.T) {
		q := NewWebhookQueue(3)

		for i := 0; i < 3; i++ {
			err := q.Enqueue(tunnel.NewRequestFrame("POST", "/hook", nil, nil))
			require.NoError(t, err)
		}

		assert.Equal(t, 3, q.Len())
	})

	t.Run("should return error when queue full", func(t *testing.T) {
		q := NewWebhookQueue(2)

		q.Enqueue(tunnel.NewRequestFrame("POST", "/a", nil, nil))
		q.Enqueue(tunnel.NewRequestFrame("POST", "/b", nil, nil))

		err := q.Enqueue(tunnel.NewRequestFrame("POST", "/c", nil, nil))
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrQueueFull)
	})
}

func TestWebhookQueue_Drain(t *testing.T) {
	t.Run("should drain all in fifo order", func(t *testing.T) {
		q := NewWebhookQueue(20)

		f1 := tunnel.NewRequestFrame("POST", "/first", nil, nil)
		f2 := tunnel.NewRequestFrame("POST", "/second", nil, nil)
		f3 := tunnel.NewRequestFrame("POST", "/third", nil, nil)

		q.Enqueue(f1)
		q.Enqueue(f2)
		q.Enqueue(f3)

		drained := q.DrainAll()
		require.Len(t, drained, 3)
		assert.Equal(t, f1.ID, drained[0].ID)
		assert.Equal(t, f2.ID, drained[1].ID)
		assert.Equal(t, f3.ID, drained[2].ID)
	})

	t.Run("should return empty slice when drain empty queue", func(t *testing.T) {
		q := NewWebhookQueue(20)

		drained := q.DrainAll()
		assert.Empty(t, drained)
	})

	t.Run("should clear queue after drain", func(t *testing.T) {
		q := NewWebhookQueue(20)
		q.Enqueue(tunnel.NewRequestFrame("POST", "/x", nil, nil))
		q.Enqueue(tunnel.NewRequestFrame("POST", "/y", nil, nil))

		q.DrainAll()
		assert.Equal(t, 0, q.Len())
	})

	t.Run("should allow enqueue after drain", func(t *testing.T) {
		q := NewWebhookQueue(2)
		q.Enqueue(tunnel.NewRequestFrame("POST", "/a", nil, nil))
		q.Enqueue(tunnel.NewRequestFrame("POST", "/b", nil, nil))

		q.DrainAll()

		err := q.Enqueue(tunnel.NewRequestFrame("POST", "/c", nil, nil))
		require.NoError(t, err)
		assert.Equal(t, 1, q.Len())
	})
}

func TestWebhookQueue_Len(t *testing.T) {
	t.Run("should return correct length", func(t *testing.T) {
		q := NewWebhookQueue(20)
		assert.Equal(t, 0, q.Len())

		q.Enqueue(tunnel.NewRequestFrame("POST", "/a", nil, nil))
		assert.Equal(t, 1, q.Len())

		q.Enqueue(tunnel.NewRequestFrame("POST", "/b", nil, nil))
		assert.Equal(t, 2, q.Len())
	})
}

func TestWebhookQueue_Concurrent(t *testing.T) {
	t.Run("should handle concurrent enqueue safely", func(t *testing.T) {
		q := NewWebhookQueue(200)
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				q.Enqueue(tunnel.NewRequestFrame("POST", "/concurrent", nil, nil))
			}()
		}

		wg.Wait()
		assert.Equal(t, 100, q.Len())
	})

	t.Run("should handle concurrent enqueue and drain", func(t *testing.T) {
		q := NewWebhookQueue(200)
		var wg sync.WaitGroup

		// Enqueue 50
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				q.Enqueue(tunnel.NewRequestFrame("POST", "/mixed", nil, nil))
			}()
		}

		wg.Wait()

		// Drain while more are being added
		var drainWg sync.WaitGroup
		var drained []tunnel.Frame

		drainWg.Add(1)
		go func() {
			defer drainWg.Done()
			drained = q.DrainAll()
		}()

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				q.Enqueue(tunnel.NewRequestFrame("POST", "/after", nil, nil))
			}()
		}

		drainWg.Wait()
		wg.Wait()

		// Total across drained + remaining should be 100
		total := len(drained) + q.Len()
		assert.Equal(t, 100, total)
	})
}
