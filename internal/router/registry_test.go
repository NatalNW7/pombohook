package router

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteRegistry_Register(t *testing.T) {
	t.Run("should register route successfully", func(t *testing.T) {
		r := NewRouteRegistry()

		err := r.Register("/webhook/mp", 8081, "tunnel-1")
		require.NoError(t, err)
	})

	t.Run("should return error for duplicate path", func(t *testing.T) {
		r := NewRouteRegistry()

		err := r.Register("/webhook/mp", 8081, "tunnel-1")
		require.NoError(t, err)

		err = r.Register("/webhook/mp", 9090, "tunnel-2")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrRouteAlreadyExists)
	})
}

func TestRouteRegistry_Lookup(t *testing.T) {
	t.Run("should lookup registered route", func(t *testing.T) {
		r := NewRouteRegistry()
		r.Register("/webhook/mp", 8081, "tunnel-1")

		route, found := r.Lookup("/webhook/mp")
		assert.True(t, found)
		assert.Equal(t, "/webhook/mp", route.Path)
		assert.Equal(t, 8081, route.Port)
		assert.Equal(t, "tunnel-1", route.TunnelID)
	})

	t.Run("should return false for unknown path", func(t *testing.T) {
		r := NewRouteRegistry()

		_, found := r.Lookup("/webhook/unknown")
		assert.False(t, found)
	})
}

func TestRouteRegistry_Unregister(t *testing.T) {
	t.Run("should unregister existing route", func(t *testing.T) {
		r := NewRouteRegistry()
		r.Register("/webhook/mp", 8081, "tunnel-1")

		removed := r.Unregister("/webhook/mp")
		assert.True(t, removed)

		_, found := r.Lookup("/webhook/mp")
		assert.False(t, found)
	})

	t.Run("should return false when unregister unknown", func(t *testing.T) {
		r := NewRouteRegistry()

		removed := r.Unregister("/webhook/unknown")
		assert.False(t, removed)
	})

	t.Run("should allow reregister after unregister", func(t *testing.T) {
		r := NewRouteRegistry()

		r.Register("/webhook/mp", 8081, "tunnel-1")
		r.Unregister("/webhook/mp")

		err := r.Register("/webhook/mp", 9090, "tunnel-2")
		require.NoError(t, err)

		route, found := r.Lookup("/webhook/mp")
		assert.True(t, found)
		assert.Equal(t, 9090, route.Port)
	})
}

func TestRouteRegistry_UnregisterAll(t *testing.T) {
	t.Run("should unregister all by tunnel id", func(t *testing.T) {
		r := NewRouteRegistry()

		r.Register("/webhook/mp", 8081, "tunnel-1")
		r.Register("/webhook/stripe", 3000, "tunnel-1")
		r.Register("/webhook/other", 4000, "tunnel-2")

		count := r.UnregisterAll("tunnel-1")
		assert.Equal(t, 2, count)

		_, found := r.Lookup("/webhook/mp")
		assert.False(t, found)
		_, found = r.Lookup("/webhook/stripe")
		assert.False(t, found)
		// tunnel-2 route should still exist
		_, found = r.Lookup("/webhook/other")
		assert.True(t, found)
	})

	t.Run("should return zero when unregister all unknown tunnel", func(t *testing.T) {
		r := NewRouteRegistry()

		count := r.UnregisterAll("nonexistent")
		assert.Equal(t, 0, count)
	})
}

func TestRouteRegistry_List(t *testing.T) {
	t.Run("should list all routes", func(t *testing.T) {
		r := NewRouteRegistry()
		r.Register("/webhook/mp", 8081, "tunnel-1")
		r.Register("/webhook/stripe", 3000, "tunnel-1")

		routes := r.ListRoutes()
		assert.Len(t, routes, 2)
	})

	t.Run("should return empty list when no routes", func(t *testing.T) {
		r := NewRouteRegistry()

		routes := r.ListRoutes()
		assert.Empty(t, routes)
	})
}

func TestRouteRegistry_Concurrent(t *testing.T) {
	t.Run("should handle concurrent register and lookup", func(t *testing.T) {
		r := NewRouteRegistry()
		var wg sync.WaitGroup

		// Register 50 routes concurrently
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				path := "/webhook/" + string(rune('A'+idx))
				r.Register(path, 8000+idx, "tunnel-1")
			}(i)
		}

		// Lookup concurrently while registering
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				path := "/webhook/" + string(rune('A'+idx))
				r.Lookup(path)
			}(i)
		}

		wg.Wait()

		routes := r.ListRoutes()
		assert.Len(t, routes, 50)
	})
}
