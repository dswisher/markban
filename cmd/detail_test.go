package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dswisher/markban/internal/board"
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

func TestSortEpicSubtasksByDependency(t *testing.T) {
	subtasks := []board.Task{
		{Slug: "third", DependsOn: []string{"second"}},
		{Slug: "second", DependsOn: []string{"first"}},
		{Slug: "first"},
		{Slug: "external-dependent", DependsOn: []string{"outside-epic"}},
	}

	ordered := sortEpicSubtasks(subtasks)
	assert.Equal(t, []string{"first", "second", "third", "external-dependent"}, []string{
		ordered[0].Slug, ordered[1].Slug, ordered[2].Slug, ordered[3].Slug,
	})
}
