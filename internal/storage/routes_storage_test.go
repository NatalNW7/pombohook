package storage

import (
	"os"
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

	t.Run("should return error when unable to save routes", func(t *testing.T) {
		s := NewStorage("/dev/null/foo")
		err := s.SaveRoutes([]RouteMapping{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating storage directory")
	})

	t.Run("should return error when routes file has invalid content", func(t *testing.T) {
		s := newTestStorage(t)
		err := s.ensureDir()
		require.NoError(t, err)
		
		err = os.WriteFile(s.filePath(routesFile), []byte("invalid-json"), 0600)
		require.NoError(t, err)

		_, err = s.LoadRoutes()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing routes")
	})
}
