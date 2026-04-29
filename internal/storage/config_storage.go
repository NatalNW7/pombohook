package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFile = "config.json"

// SaveConfig writes the PomboConfig to config.json.
func (s *Storage) SaveConfig(cfg PomboConfig) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(s.filePath(configFile), data, 0600)
}

// LoadConfig reads the PomboConfig from config.json.
func (s *Storage) LoadConfig() (PomboConfig, error) {
	data, err := os.ReadFile(s.filePath(configFile))
	if err != nil {
		return PomboConfig{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg PomboConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return PomboConfig{}, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// ConfigExists returns true if config.json exists.
func (s *Storage) ConfigExists() bool {
	_, err := os.Stat(s.filePath(configFile))
	return err == nil
}
