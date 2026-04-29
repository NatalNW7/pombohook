package cli

import (
	"fmt"
	"io"

	"github.com/NatalNW7/pombohook/internal/config"
	"github.com/NatalNW7/pombohook/internal/storage"
)

// RunRouteAdd adds a route to local storage.
func RunRouteAdd(store *storage.Storage, w io.Writer, path string, port int) error {
	if len(path) == 0 || path[0] != '/' {
		return fmt.Errorf("path must start with /")
	}
	if port <= 0 {
		return fmt.Errorf("port must be a positive number")
	}

	route := config.RouteMapping{Path: path, Port: port}
	if err := store.AddRoute(route); err != nil {
		return err
	}

	fmt.Fprintf(w, "🕊️  Route added: %s → localhost:%d%s\n", path, port, path)
	return nil
}

// RunRouteList lists all routes from local storage.
func RunRouteList(store *storage.Storage, w io.Writer) error {
	routes, err := store.LoadRoutes()
	if err != nil {
		return err
	}

	if len(routes) == 0 {
		fmt.Fprintln(w, "🕊️  No routes registered. Use `pombo route --path=/webhook/test --port=8081`")
		return nil
	}

	fmt.Fprintln(w, "🕊️  Registered routes:")
	for _, r := range routes {
		fmt.Fprintf(w, "    %-20s → localhost:%d%s\n", r.Path, r.Port, r.Path)
	}
	return nil
}

// RunRouteRemove removes a route from local storage.
func RunRouteRemove(store *storage.Storage, w io.Writer, path string) error {
	if err := store.RemoveRoute(path); err != nil {
		return err
	}

	fmt.Fprintf(w, "🕊️  Route removed: %s\n", path)
	return nil
}

// RunRouteClear clears all routes from local storage.
func RunRouteClear(store *storage.Storage, w io.Writer) error {
	if err := store.ClearRoutes(); err != nil {
		return err
	}

	fmt.Fprintln(w, "🕊️  All routes cleared.")
	return nil
}
