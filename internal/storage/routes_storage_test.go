package storage

import (
	"testing"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutesStorage_SaveAndLoad(t *testing.T) {
	t.Run("should save and load routes", func(t *testing.T) {
		s := newTestStorage(t)
		routes := []RouteMapping{
			{Path: "/webhook/mp", Port: 8081},
			{Path: "/webhook/stripe", Port: 3000},
		}

		err := s.SaveRoutes(routes)
		require.NoError(t, err)

		loaded, err := s.LoadRoutes()
		require.NoError(t, err)
		assert.Equal(t, routes, loaded)
	})

	t.Run("should add route without duplicate", func(t *testing.T) {
		s := newTestStorage(t)

		err := s.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 8081})
		require.NoError(t, err)

		err = s.AddRoute(config.RouteMapping{Path: "/webhook/stripe", Port: 3000})
		require.NoError(t, err)

		loaded, err := s.LoadRoutes()
		require.NoError(t, err)
		assert.Len(t, loaded, 2)
	})

	t.Run("should reject duplicate route path", func(t *testing.T) {
		s := newTestStorage(t)

		err := s.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 8081})
		require.NoError(t, err)

		err = s.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 9090})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("should remove existing route", func(t *testing.T) {
		s := newTestStorage(t)
		s.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 8081})
		s.AddRoute(config.RouteMapping{Path: "/webhook/stripe", Port: 3000})

		err := s.RemoveRoute("/webhook/mp")
		require.NoError(t, err)

		loaded, err := s.LoadRoutes()
		require.NoError(t, err)
		assert.Len(t, loaded, 1)
		assert.Equal(t, "/webhook/stripe", loaded[0].Path)
	})

	t.Run("should return error removing unknown route", func(t *testing.T) {
		s := newTestStorage(t)

		err := s.RemoveRoute("/webhook/nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("should clear all routes", func(t *testing.T) {
		s := newTestStorage(t)
		s.AddRoute(config.RouteMapping{Path: "/webhook/mp", Port: 8081})
		s.AddRoute(config.RouteMapping{Path: "/webhook/stripe", Port: 3000})

		err := s.ClearRoutes()
		require.NoError(t, err)

		loaded, err := s.LoadRoutes()
		require.NoError(t, err)
		assert.Empty(t, loaded)
	})

	t.Run("should return empty when no routes", func(t *testing.T) {
		s := newTestStorage(t)

		loaded, err := s.LoadRoutes()
		require.NoError(t, err)
		assert.Empty(t, loaded)
	})
}
