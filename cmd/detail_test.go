package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEpicSummary(t *testing.T) {
	dir := t.TempDir()
	for _, column := range []string{"1-todo", "2-done"} {
		require.NoError(t, os.Mkdir(filepath.Join(dir, column), 0o755))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "1-todo", "epic.md"), []byte("---\ntype: epic\n---\n# Epic\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "1-todo", "child.md"), []byte("---\nepic: epic\n---\n# Child\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "2-done", "finished.md"), []byte("---\nepic: epic\n---\n# Finished\n"), 0o644))

	summary, err := loadEpicSummary(dir, "epic")
	require.NoError(t, err)
	assert.Contains(t, summary, "Child")
	assert.Contains(t, summary, "todo")
	assert.Contains(t, summary, "Finished")
	assert.Contains(t, summary, "done")
}
