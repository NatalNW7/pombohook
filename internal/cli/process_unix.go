//go:build !windows

package cli

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true, // Detach from parent session
	}
}

// isProcessAlive checks if a process with the given PID is running.
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

// stopProcess attempts to gracefully stop the process, falling back to SIGKILL.
func stopProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found")
	}

	// Check if process exists
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return fmt.Errorf("process not running")
	}

	// Send SIGTERM
	process.Signal(syscall.SIGTERM)

	// Wait up to 5 seconds
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return nil // Process exited
		}
		time.Sleep(100 * time.Millisecond)
	}

	// If still alive, SIGKILL
	if err := process.Signal(syscall.Signal(0)); err == nil {
		process.Signal(syscall.SIGKILL)
	}

	return nil
}
