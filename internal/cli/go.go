package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/NatalNW7/pombohook/internal/forward"
	"github.com/NatalNW7/pombohook/internal/storage"
	"github.com/NatalNW7/pombohook/internal/tunnel"
)

// RunGo starts the tunnel client in foreground mode.
// This is the core logic shared by both foreground and daemon modes.
func RunGo(store *storage.Storage, w io.Writer, logger *slog.Logger) error {
	if err := ValidateGoPrerequisites(store, w); err != nil {
		return err
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	routesList, err := store.LoadRoutes()
	if err != nil {
		return fmt.Errorf("loading routes: %w", err)
	}

	routes := make(map[string]int)
	for _, r := range routesList {
		routes[r.Path] = r.Port
	}

	fwd := forward.NewForwarder(routes, logger)
	client := tunnel.NewTunnelClient(
		cfg.Server,
		cfg.Token,
		func(frame tunnel.Frame) {
			if frame.Type == tunnel.FrameTypeRequest {
				fwd.Forward(frame)
			}
		},
		logger,
	)

	fmt.Fprintln(w, "🕊️  Pigeon is flying! Listening for webhooks...")
	for path, port := range routes {
		fmt.Fprintf(w, "    %-20s → localhost:%d%s\n", path, port, path)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("connecting: %w", err)
	}

	if err := client.SendRegister(routesList); err != nil {
		return fmt.Errorf("registering: %w", err)
	}

	client.Listen() // blocks until disconnect
	return nil
}

// RunGoBackground forks the current process in daemon mode.
func RunGoBackground(store *storage.Storage, w io.Writer, executablePath string) error {
	if err := ValidateGoPrerequisites(store, w); err != nil {
		return err
	}

	// Check if already running
	if store.PIDExists() {
		pid, err := store.LoadPID()
		if err == nil {
			if isProcessAlive(pid) {
				fmt.Fprintf(w, "✗ Pigeon is already flying in background (PID: %d)\n", pid)
				return fmt.Errorf("already running with PID %d", pid)
			}
			// Stale PID file, clean up
			store.RemovePID()
		}
	}

	// Open log file
	logPath := filepath.Join(store.BasePath(), "pombo.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}

	// Fork: re-execute `pombo go --daemon`
	cmd := exec.Command(executablePath, "go", "--daemon")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Detach from parent session
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("starting background process: %w", err)
	}

	// Save PID
	if err := store.SavePID(cmd.Process.Pid); err != nil {
		return fmt.Errorf("saving PID: %w", err)
	}

	logFile.Close() // Parent closes its handle; child keeps its own

	fmt.Fprintf(w, "🕊️  Pigeon released in background (PID: %d)\n", cmd.Process.Pid)
	fmt.Fprintf(w, "    Logs: %s\n", logPath)
	fmt.Fprintln(w, "    Stop: pombo sleep")

	return nil
}

// isProcessAlive checks if a process with the given PID is running.
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
