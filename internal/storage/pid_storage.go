package storage

import (
	"fmt"
	"os"
	"strconv"
)

const pidFile = "pombo.pid"

// SavePID writes the process ID to pombo.pid.
func (s *Storage) SavePID(pid int) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	return os.WriteFile(s.filePath(pidFile), []byte(strconv.Itoa(pid)), 0600)
}

// LoadPID reads the process ID from pombo.pid.
func (s *Storage) LoadPID() (int, error) {
	data, err := os.ReadFile(s.filePath(pidFile))
	if err != nil {
		return 0, fmt.Errorf("reading PID: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("parsing PID: %w", err)
	}

	return pid, nil
}

// RemovePID deletes the pombo.pid file.
func (s *Storage) RemovePID() error {
	return os.Remove(s.filePath(pidFile))
}

// PIDExists returns true if pombo.pid exists.
func (s *Storage) PIDExists() bool {
	_, err := os.Stat(s.filePath(pidFile))
	return err == nil
}
