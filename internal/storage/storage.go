package storage

import (
	"os"
	"path/filepath"
)

// Storage manages local file persistence in ~/.pombo/.
type Storage struct {
	basePath string
}

// NewStorage creates a new Storage instance with the given base directory path.
func NewStorage(basePath string) *Storage {
	return &Storage{basePath: basePath}
}

// BasePath returns the base directory path used by this storage instance.
func (s *Storage) BasePath() string {
	return s.basePath
}

func (s *Storage) ensureDir() error {
	return os.MkdirAll(s.basePath, 0755) // #nosec G301 -- basePath derived from os.UserHomeDir(), not user input; 0755 is standard for CLI config directories
}

func (s *Storage) filePath(name string) string {
	return filepath.Join(s.basePath, name)
}
