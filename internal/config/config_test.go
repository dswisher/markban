package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUserConfig_FileNotExists(t *testing.T) {
	// Create a temporary directory for our test config
	tmpDir := t.TempDir()

	// Override the config path function
	originalGetConfigPath := getConfigPathFunc
	getConfigPathFunc = func() string {
		return filepath.Join(tmpDir, "nonexistent", "markban.toml")
	}
	defer func() { getConfigPathFunc = originalGetConfigPath }()

	cfg, err := LoadUserConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.MarkdownViewer)
}

func TestLoadUserConfig_ValidConfig(t *testing.T) {
	// Create a temporary directory for our test config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "markban")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "markban.toml")
	configContent := `markdown_viewer = "bat"
`
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Override the config path function
	originalGetConfigPath := getConfigPathFunc
	getConfigPathFunc = func() string {
		return configPath
	}
	defer func() { getConfigPathFunc = originalGetConfigPath }()

	cfg, err := LoadUserConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "bat", cfg.MarkdownViewer)
}

func TestLoadUserConfig_InvalidTOML(t *testing.T) {
	// Create a temporary directory for our test config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "markban")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "markban.toml")
	configContent := `this is not valid toml {{{
`
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Override the config path function
	originalGetConfigPath := getConfigPathFunc
	getConfigPathFunc = func() string {
		return configPath
	}
	defer func() { getConfigPathFunc = originalGetConfigPath }()

	cfg, err := LoadUserConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "parsing config file")
}

func TestLoadUserConfig_EmptyViewer(t *testing.T) {
	// Create a temporary directory for our test config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "markban")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "markban.toml")
	configContent := ``
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Override the config path function
	originalGetConfigPath := getConfigPathFunc
	getConfigPathFunc = func() string {
		return configPath
	}
	defer func() { getConfigPathFunc = originalGetConfigPath }()

	cfg, err := LoadUserConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.MarkdownViewer)
}

func TestGetConfigPath(t *testing.T) {
	// This test verifies that getConfigPath returns a non-empty path
	// The actual path depends on the environment, so we just check it's not empty
	path := getConfigPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "markban.toml")

	// Check that it contains the expected directory structure
	if runtime.GOOS == "windows" {
		assert.Contains(t, path, "markban")
	} else {
		assert.Contains(t, path, ".config")
		assert.Contains(t, path, "markban")
	}
}
