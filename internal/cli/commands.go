package cli

import (
	"fmt"
	"io"

	"github.com/NatalNW7/pombohook/internal/storage"
)

// ValidateGoPrerequisites checks that config and routes exist before starting.
func ValidateGoPrerequisites(store *storage.Storage, w io.Writer) error {
	if !store.ConfigExists() {
		fmt.Fprintln(w, "✗ No config found. Run `pombo ping` first to connect to a server.")
		return fmt.Errorf("no config found")
	}

	routes, err := store.LoadRoutes()
	if err != nil {
		return err
	}
	if len(routes) == 0 {
		fmt.Fprintln(w, "✗ No routes registered. Run `pombo route --path=/webhook/test --port=8081` first.")
		return fmt.Errorf("no routes found")
	}

	return nil
}

// RunSleep stops a background pombo process.
func RunSleep(store *storage.Storage, w io.Writer) error {
	if !store.PIDExists() {
		fmt.Fprintln(w, "✗ No pigeon is flying. Nothing to stop.")
		return fmt.Errorf("no pid found")
	}

	pid, err := store.LoadPID()
	if err != nil {
		return err
	}

	if err := stopProcess(pid); err != nil {
		store.RemovePID()
		fmt.Fprintln(w, "✗ Pigeon was not flying or could not be stopped.")
		return err
	}

	store.RemovePID()
	fmt.Fprintln(w, "🕊️  Pigeon is resting. Background session stopped.")
	return nil
}

