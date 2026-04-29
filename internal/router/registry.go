package router

import (
	"errors"
	"sync"
)

// ErrRouteAlreadyExists is returned when registering a path that is already active.
var ErrRouteAlreadyExists = errors.New("route already registered")

// Route represents an active webhook path mapped to a local port via a tunnel.
type Route struct {
	Path     string
	Port     int
	TunnelID string
}

// RouteRegistry is a thread-safe registry of active webhook routes.
type RouteRegistry struct {
	mu     sync.RWMutex
	routes map[string]*Route
}

// NewRouteRegistry creates an empty RouteRegistry.
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make(map[string]*Route),
	}
}

// Register adds a route. Returns ErrRouteAlreadyExists if path is already registered.
func (r *RouteRegistry) Register(path string, port int, tunnelID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.routes[path]; exists {
		return ErrRouteAlreadyExists
	}

	r.routes[path] = &Route{
		Path:     path,
		Port:     port,
		TunnelID: tunnelID,
	}
	return nil
}

// Lookup returns the route for a given path, or false if not found.
func (r *RouteRegistry) Lookup(path string) (*Route, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	route, found := r.routes[path]
	return route, found
}

// Unregister removes a route by path. Returns true if the route existed.
func (r *RouteRegistry) Unregister(path string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.routes[path]; !exists {
		return false
	}

	delete(r.routes, path)
	return true
}

// UnregisterAll removes all routes belonging to a given tunnelID. Returns the count removed.
func (r *RouteRegistry) UnregisterAll(tunnelID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for path, route := range r.routes {
		if route.TunnelID == tunnelID {
			delete(r.routes, path)
			count++
		}
	}
	return count
}

// ListRoutes returns a snapshot of all registered routes.
func (r *RouteRegistry) ListRoutes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]Route, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, *route)
	}
	return routes
}
