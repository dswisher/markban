package board

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindGitRoot searches upward from startDir for a directory containing a .git folder.
// Returns the path to the git root directory, or an error if not found.
func FindGitRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path %q: %w", startDir, err)
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		info, err := os.Stat(gitDir)
		if err == nil && info.IsDir() {
			return dir, nil
		}

		// Move up to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no git repository found starting from %q", startDir)
}

// FindBoardDirectory searches for a subdirectory within rootDir using these rules:
// 1. First, look for a subdirectory containing a board.toml file (best match)
// 2. If not found, look for a subdirectory with "board" in its name (case-insensitive)
// Returns the path to the board directory, or an error if not found or multiple found.
func FindBoardDirectory(rootDir string) (string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return "", fmt.Errorf("cannot read directory %q: %w", rootDir, err)
	}

	var dirsWithConfig []string
	var dirsWithBoardInName []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		entryPath := filepath.Join(rootDir, entry.Name())

		// Check if it contains board.toml
		boardTomlPath := filepath.Join(entryPath, "board.toml")
		if _, err := os.Stat(boardTomlPath); err == nil {
			dirsWithConfig = append(dirsWithConfig, entry.Name())
		}

		// Check if name contains "board" (case-insensitive)
		if strings.Contains(strings.ToLower(entry.Name()), "board") {
			dirsWithBoardInName = append(dirsWithBoardInName, entry.Name())
		}
	}

	// Prefer directories with board.toml
	if len(dirsWithConfig) == 1 {
		return filepath.Join(rootDir, dirsWithConfig[0]), nil
	}
	if len(dirsWithConfig) > 1 {
		return "", fmt.Errorf("multiple directories with board.toml found in %q: %v", rootDir, dirsWithConfig)
	}

	// Fall back to directories with "board" in name
	if len(dirsWithBoardInName) == 1 {
		return filepath.Join(rootDir, dirsWithBoardInName[0]), nil
	}
	if len(dirsWithBoardInName) > 1 {
		return "", fmt.Errorf("multiple directories with 'board' in name found in %q: %v", rootDir, dirsWithBoardInName)
	}

	return "", fmt.Errorf("no board directory found in %q (looked for board.toml or 'board' in name)", rootDir)
}

// FindProjectBoard attempts to auto-discover a board directory.
// It first finds the git root from startDir, then looks for a board subdirectory.
// Returns the absolute path to the board directory, or an error if not found.
func FindProjectBoard(startDir string) (string, error) {
	gitRoot, err := FindGitRoot(startDir)
	if err != nil {
		return "", err
	}

	boardDir, err := FindBoardDirectory(gitRoot)
	if err != nil {
		return "", err
	}

	return boardDir, nil
}
