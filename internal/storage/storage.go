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

func (s *Storage) ensureDir() error {
	return os.MkdirAll(s.basePath, 0755)
}

func (s *Storage) filePath(name string) string {
	return filepath.Join(s.basePath, name)
}
