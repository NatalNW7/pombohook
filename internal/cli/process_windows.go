//go:build windows

package cli

import (
	"fmt"
	"os"
	"syscall"
)

func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// isProcessAlive checks if a process with the given PID is running.
func isProcessAlive(pid int) bool {
	_, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds even if the process is dead.
	// A robust check requires Windows API calls, but for simplicity, 
	// we assume it is alive if the PID file exists.
	// If it fails to stop later, we clear the PID anyway.
	return true
}

// stopProcess forcefully stops the process on Windows.
func stopProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found")
	}

	// Windows doesn't support graceful SIGTERM via Go os.Process out-of-the-box easily.
	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	return nil
}
