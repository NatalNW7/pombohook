package storage

import "github.com/NatalNW7/pombohook/internal/config"

// PomboConfig holds the CLI connection configuration persisted by `pombo ping`.
type PomboConfig struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

// Re-export RouteMapping for convenience within storage package.
type RouteMapping = config.RouteMapping
