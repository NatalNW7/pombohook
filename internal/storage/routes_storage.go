package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

const routesFile = "routes.json"

// SaveRoutes writes the route mappings to routes.json.
func (s *Storage) SaveRoutes(routes []RouteMapping) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	data, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling routes: %w", err)
	}

	return os.WriteFile(s.filePath(routesFile), data, 0600)
}

// LoadRoutes reads route mappings from routes.json. Returns empty slice if file doesn't exist.
func (s *Storage) LoadRoutes() ([]RouteMapping, error) {
	data, err := os.ReadFile(s.filePath(routesFile))
	if err != nil {
		if os.IsNotExist(err) {
			return []RouteMapping{}, nil
		}
		return nil, fmt.Errorf("reading routes: %w", err)
	}

	var routes []RouteMapping
	if err := json.Unmarshal(data, &routes); err != nil {
		return nil, fmt.Errorf("parsing routes: %w", err)
	}

	return routes, nil
}

// AddRoute appends a route mapping. Returns error if path already exists.
func (s *Storage) AddRoute(route RouteMapping) error {
	routes, err := s.LoadRoutes()
	if err != nil {
		return err
	}

	for _, r := range routes {
		if r.Path == route.Path {
			return fmt.Errorf("route %q already exists", route.Path)
		}
	}

	routes = append(routes, route)
	return s.SaveRoutes(routes)
}

// RemoveRoute removes a route by path. Returns error if not found.
func (s *Storage) RemoveRoute(path string) error {
	routes, err := s.LoadRoutes()
	if err != nil {
		return err
	}

	found := false
	filtered := make([]RouteMapping, 0, len(routes))
	for _, r := range routes {
		if r.Path == path {
			found = true
			continue
		}
		filtered = append(filtered, r)
	}

	if !found {
		return fmt.Errorf("route %q not found", path)
	}

	return s.SaveRoutes(filtered)
}

// ClearRoutes removes all routes.
func (s *Storage) ClearRoutes() error {
	return s.SaveRoutes([]RouteMapping{})
}
