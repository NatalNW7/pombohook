package config

import (
	"fmt"
	"os"
)

// ServerConfig holds the configuration for the PomboHook API server.
type ServerConfig struct {
	Port      string
	AuthToken string
	LogLevel  string
}

// RouteMapping represents a path-to-port mapping for webhook forwarding.
type RouteMapping struct {
	Path string `json:"path"`
	Port int    `json:"port"`
}

// LoadServerConfig loads server configuration from environment variables.
// Required: PORT, POMBOHOOK_TOKEN.
// Optional: LOG_LEVEL (default: "info").
func LoadServerConfig() (ServerConfig, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return ServerConfig{}, fmt.Errorf("required environment variable PORT is not set")
	}

	token := os.Getenv("POMBOHOOK_TOKEN")
	if token == "" {
		return ServerConfig{}, fmt.Errorf("required environment variable POMBOHOOK_TOKEN is not set")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return ServerConfig{
		Port:      port,
		AuthToken: token,
		LogLevel:  logLevel,
	}, nil
}
