package board

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindGitRoot_FindsGitRoot(t *testing.T) {
	// Start from deep within the test git repo
	startDir := "testdata/git-repo/project-board"
	root, err := FindGitRoot(startDir)
	require.NoError(t, err)

	// Check that the directory name ends with "git-repo"
	assert.Equal(t, "git-repo", filepath.Base(root))
}

func TestFindGitRoot_NotInGitRepo(t *testing.T) {
	// Use a temp directory which is not a git repo
	tempDir := t.TempDir()
	_, err := FindGitRoot(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no git repository found")
}

func TestFindGitRoot_FromCurrentDir(t *testing.T) {
	// The actual project root should be a git repo
	root, err := FindGitRoot(".")
	require.NoError(t, err)

	// Verify it contains .git
	gitDir := filepath.Join(root, ".git")
	info, err := os.Stat(gitDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFindBoardDirectory_WithConfig(t *testing.T) {
	repoRoot := "testdata/git-repo"
	boardDir, err := FindBoardDirectory(repoRoot)
	require.NoError(t, err)

	// Should prefer the directory with board.toml
	assert.Equal(t, "project-board", filepath.Base(boardDir))
}

func TestFindBoardDirectory_ByNameOnly(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	myBoardDir := filepath.Join(tempDir, "my-board")
	err := os.MkdirAll(myBoardDir, 0755)
	require.NoError(t, err)

	// Create a dummy task file (no board.toml)
	err = os.WriteFile(filepath.Join(myBoardDir, "task.md"), []byte("# Task\n"), 0644)
	require.NoError(t, err)

	boardDir, err := FindBoardDirectory(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "my-board", filepath.Base(boardDir))
}

func TestFindBoardDirectory_NoBoardFound(t *testing.T) {
	// Create a temporary directory with no board directories
	tempDir := t.TempDir()
	err := os.MkdirAll(filepath.Join(tempDir, "some-dir"), 0755)
	require.NoError(t, err)

	_, err = FindBoardDirectory(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no board directory found")
}

func TestFindBoardDirectory_MultipleConfigs(t *testing.T) {
	// Create a temporary directory with multiple board.toml files
	tempDir := t.TempDir()

	board1 := filepath.Join(tempDir, "board1")
	board2 := filepath.Join(tempDir, "board2")
	err := os.MkdirAll(board1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(board2, 0755)
	require.NoError(t, err)

	// Create board.toml in both
	err = os.WriteFile(filepath.Join(board1, "board.toml"), []byte("name = \"Board1\""), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(board2, "board.toml"), []byte("name = \"Board2\""), 0644)
	require.NoError(t, err)

	_, err = FindBoardDirectory(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple directories with board.toml found")
}

func TestFindBoardDirectory_MultipleByName(t *testing.T) {
	// Create a temporary directory with multiple "board" directories (no configs)
	tempDir := t.TempDir()

	board1 := filepath.Join(tempDir, "my-board")
	board2 := filepath.Join(tempDir, "project-board")
	err := os.MkdirAll(board1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(board2, 0755)
	require.NoError(t, err)

	_, err = FindBoardDirectory(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple directories with 'board' in name found")
}

func TestFindBoardDirectory_SkipsHiddenDirs(t *testing.T) {
	// Create a temporary directory with a hidden board directory
	tempDir := t.TempDir()

	hiddenBoard := filepath.Join(tempDir, ".board")
	visibleBoard := filepath.Join(tempDir, "visible-board")
	err := os.MkdirAll(hiddenBoard, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(visibleBoard, 0755)
	require.NoError(t, err)

	boardDir, err := FindBoardDirectory(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "visible-board", filepath.Base(boardDir))
}

func TestFindBoardDirectory_ConfigPreferredOverName(t *testing.T) {
	// Create a temp directory with both a config directory and a name-match directory
	tempDir := t.TempDir()

	configBoard := filepath.Join(tempDir, "weird-name") // No "board" in name
	nameBoard := filepath.Join(tempDir, "my-board")     // Has "board" in name
	err := os.MkdirAll(configBoard, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(nameBoard, 0755)
	require.NoError(t, err)

	// Create board.toml in the weird-name directory
	err = os.WriteFile(filepath.Join(configBoard, "board.toml"), []byte("name = \"Weird\""), 0644)
	require.NoError(t, err)

	// Should prefer the one with config
	boardDir, err := FindBoardDirectory(tempDir)
	require.NoError(t, err)
	assert.Equal(t, "weird-name", filepath.Base(boardDir))
}

func TestFindProjectBoard_Integration(t *testing.T) {
	// Start from within a project board directory
	startDir := "testdata/git-repo/project-board"
	boardDir, err := FindProjectBoard(startDir)
	require.NoError(t, err)

	// Should find the project-board directory
	assert.Equal(t, "project-board", filepath.Base(boardDir))
}

func TestFindProjectBoard_NotInGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	_, err := FindProjectBoard(tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no git repository found")
}
