package config

import (
	"fmt"
	"os"
	"strings"
)

// ServerConfig holds the configuration for the PomboHook API server.
type ServerConfig struct {
	Port           string
	AuthToken      string
	LogLevel       string
	AllowedOrigins []string
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

	allowedOrigins := parseAllowedOrigins(os.Getenv("POMBOHOOK_ALLOWED_ORIGINS"))

	return ServerConfig{
		Port:           port,
		AuthToken:      token,
		LogLevel:       logLevel,
		AllowedOrigins: allowedOrigins,
	}, nil
}

// parseAllowedOrigins splits a comma-separated list of origins.
// Returns ["*"] when raw is empty or "*", preserving the default permissive behavior.
func parseAllowedOrigins(raw string) []string {
	if raw == "" || raw == "*" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
