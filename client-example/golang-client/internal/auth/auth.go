package auth

import (
	"fmt"
	"os"
	"strings"
)

// Config holds authentication configuration
type Config struct {
	Username string
	Password string
	Server   string
}

// LoadFromEnv loads authentication configuration from environment variables
func LoadFromEnv() (*Config, error) {
	username := os.Getenv("USER")
	if username == "" {
		return nil, fmt.Errorf("USER environment variable is required")
	}

	password := os.Getenv("PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("PASSWORD environment variable is required")
	}

	server := os.Getenv("SERVER")
	if server == "" {
		return nil, fmt.Errorf("SERVER environment variable is required")
	}

	// Ensure server URL doesn't have trailing slash
	server = strings.TrimSuffix(server, "/")

	return &Config{
		Username: username,
		Password: password,
		Server:   server,
	}, nil
}


