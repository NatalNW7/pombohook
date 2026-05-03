package cli

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

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

	process, err := os.FindProcess(pid)
	if err != nil {
		store.RemovePID()
		fmt.Fprintln(w, "✗ No pigeon is flying. Nothing to stop.")
		return fmt.Errorf("process not found")
	}

	// Check if process exists
	if err := process.Signal(syscall.Signal(0)); err != nil {
		store.RemovePID()
		fmt.Fprintln(w, "✗ No pigeon is flying. Nothing to stop.")
		return fmt.Errorf("process not running")
	}

	// Send SIGTERM
	process.Signal(syscall.SIGTERM)

	// Wait up to 5 seconds
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			break // Process exited
		}
		time.Sleep(100 * time.Millisecond)
	}

	// If still alive, SIGKILL
	if err := process.Signal(syscall.Signal(0)); err == nil {
		process.Signal(syscall.SIGKILL)
	}

	store.RemovePID()
	fmt.Fprintln(w, "🕊️  Pigeon is resting. Background session stopped.")
	return nil
}

