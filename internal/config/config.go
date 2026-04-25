// Package config provides functionality for loading and managing user configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Config represents the user configuration from ~/.config/markban/markban.toml.
type Config struct {
	MarkdownViewer string `toml:"markdown_viewer"`
}

// getConfigPathFunc is a variable to allow overriding in tests.
var getConfigPathFunc = getConfigPath

// LoadUserConfig loads the user configuration from the default config path.
// If the config file does not exist, it returns a default (empty) config.
// If the config file exists but cannot be parsed, it returns an error.
func LoadUserConfig() (*Config, error) {
	configPath := getConfigPathFunc()

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, return default config
		return &Config{}, nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", configPath, err)
	}

	// Parse the TOML
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", configPath, err)
	}

	return &cfg, nil
}

// getConfigPath returns the path to the user configuration file.
// On Unix systems: ~/.config/markban/markban.toml
// On Windows: %LOCALAPPDATA%\markban\markban.toml
func getConfigPath() string {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Use LOCALAPPDATA on Windows
		configDir = os.Getenv("LOCALAPPDATA")
		if configDir == "" {
			// Fallback to USERPROFILE
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
	default:
		// Unix-like systems (Linux, macOS, etc.)
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			// Default to ~/.config
			home, err := os.UserHomeDir()
			if err == nil {
				configDir = filepath.Join(home, ".config")
			}
		}
	}

	return filepath.Join(configDir, "markban", "markban.toml")
}
