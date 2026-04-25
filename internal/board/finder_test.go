package board

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardFinder_FindBySlug(t *testing.T) {
	// Use the existing testdata directory
	testdataDir := filepath.Join("testdata", "board")

	finder := NewCardFinder(testdataDir)
	require.NotNil(t, finder)

	t.Run("exact slug match", func(t *testing.T) {
		result, err := finder.FindBySlug("backlog-task")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "backlog-task", result.Task.Slug)
		assert.Equal(t, "Backlog Task", result.Task.Title)
	})

	t.Run("case-insensitive slug match", func(t *testing.T) {
		result, err := finder.FindBySlug("BACKLOG-TASK")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "backlog-task", result.Task.Slug)
	})

	t.Run("no match", func(t *testing.T) {
		result, err := finder.FindBySlug("nonexistent-task")
		assert.ErrorIs(t, err, ErrNoMatch)
		assert.Nil(t, result)
	})
}

func TestCardFinder_FindByTitleSubstring(t *testing.T) {
	// Use the existing testdata directory
	testdataDir := filepath.Join("testdata", "board")

	finder := NewCardFinder(testdataDir)
	require.NotNil(t, finder)

	t.Run("single match", func(t *testing.T) {
		results, err := finder.FindByTitleSubstring("Backlog")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Backlog Task", results[0].Task.Title)
	})

	t.Run("multiple matches", func(t *testing.T) {
		results, err := finder.FindByTitleSubstring("Task")
		require.NoError(t, err)
		// Should find multiple tasks (backlog-task, todo-task-one, todo-task-two, done-task, unknown-task)
		assert.GreaterOrEqual(t, len(results), 2)
	})

	t.Run("case-insensitive match", func(t *testing.T) {
		results, err := finder.FindByTitleSubstring("BACKLOG")
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Backlog Task", results[0].Task.Title)
	})

	t.Run("no match", func(t *testing.T) {
		results, err := finder.FindByTitleSubstring("xyznonexistent")
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestCardFinder_FindCard(t *testing.T) {
	// Use the existing testdata directory
	testdataDir := filepath.Join("testdata", "board")

	finder := NewCardFinder(testdataDir)
	require.NotNil(t, finder)

	t.Run("slug match takes precedence", func(t *testing.T) {
		result, err := finder.FindCard("backlog-task")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "backlog-task", result.Task.Slug)
	})

	t.Run("title match when no slug match", func(t *testing.T) {
		result, err := finder.FindCard("Backlog")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Backlog Task", result.Task.Title)
	})

	t.Run("no matches", func(t *testing.T) {
		result, err := finder.FindCard("xyznonexistent")
		assert.ErrorIs(t, err, ErrNoMatch)
		assert.Nil(t, result)
	})

	t.Run("multiple title matches returns error", func(t *testing.T) {
		result, err := finder.FindCard("Task")
		assert.ErrorIs(t, err, ErrMultipleMatches)
		assert.Nil(t, result)
		// Error message should contain match details
		assert.Contains(t, err.Error(), "multiple cards match")
	})
}

func TestCardFinder_ArchiveExcluded(t *testing.T) {
	// Use the board-with-archive testdata
	testdataDir := filepath.Join("testdata", "board-with-archive")

	// Verify archive directory exists and has tasks
	archiveDir := filepath.Join(testdataDir, "archive")
	_, err := os.Stat(archiveDir)
	require.NoError(t, err, "archive directory should exist in testdata")

	// Verify there's a task in the archive
	archiveTaskPath := filepath.Join(archiveDir, "archived-task-one.md")
	_, err = os.Stat(archiveTaskPath)
	require.NoError(t, err, "archived task should exist")

	finder := NewCardFinder(testdataDir)

	t.Run("archive slug not found", func(t *testing.T) {
		result, err := finder.FindBySlug("archived-task-one")
		assert.ErrorIs(t, err, ErrNoMatch)
		assert.Nil(t, result)
	})

	t.Run("archive title not found", func(t *testing.T) {
		results, err := finder.FindByTitleSubstring("Archived")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("FindCard skips archive", func(t *testing.T) {
		result, err := finder.FindCard("Archived")
		assert.ErrorIs(t, err, ErrNoMatch)
		assert.Nil(t, result)
	})

	t.Run("non-archive tasks are found", func(t *testing.T) {
		result, err := finder.FindCard("backlog-task")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "1-backlog", result.Column)
	})
}

func TestCardFinder_MultipleColumns(t *testing.T) {
	// Use the existing testdata directory which has multiple columns
	testdataDir := filepath.Join("testdata", "board")

	finder := NewCardFinder(testdataDir)

	t.Run("searches all columns", func(t *testing.T) {
		// All columns have tasks with "Task" in the title
		results, err := finder.FindByTitleSubstring("Task")
		require.NoError(t, err)
		// Should find tasks from multiple columns
		assert.GreaterOrEqual(t, len(results), 2)

		// Verify we have results from different columns
		columns := make(map[string]bool)
		for _, r := range results {
			columns[r.Column] = true
		}
		assert.GreaterOrEqual(t, len(columns), 2, "should find tasks from multiple columns")
	})
}

func TestCardFinder_InvalidDirectory(t *testing.T) {
	t.Run("nonexistent directory", func(t *testing.T) {
		finder := NewCardFinder("/nonexistent/path/that/does/not/exist")
		result, err := finder.FindBySlug("anything")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("file instead of directory", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp(t.TempDir(), "test*.md")
		require.NoError(t, err)
		tmpFile.Close()

		finder := NewCardFinder(tmpFile.Name())
		result, err := finder.FindBySlug("anything")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMatchResult_Struct(t *testing.T) {
	result := MatchResult{
		Task: Task{
			Title: "Test Task",
			Slug:  "test-task",
		},
		Path:   "/path/to/task.md",
		Column: "1-backlog",
	}

	assert.Equal(t, "Test Task", result.Task.Title)
	assert.Equal(t, "test-task", result.Task.Slug)
	assert.Equal(t, "/path/to/task.md", result.Path)
	assert.Equal(t, "1-backlog", result.Column)
}

func TestErrors(t *testing.T) {
	t.Run("ErrNoMatch is correct error", func(t *testing.T) {
		testdataDir := filepath.Join("testdata", "board")
		finder := NewCardFinder(testdataDir)

		_, err := finder.FindBySlug("definitely-does-not-exist")
		assert.True(t, errors.Is(err, ErrNoMatch))
	})

	t.Run("ErrMultipleMatches is correct error", func(t *testing.T) {
		testdataDir := filepath.Join("testdata", "board")
		finder := NewCardFinder(testdataDir)

		_, err := finder.FindCard("Task")
		assert.True(t, errors.Is(err, ErrMultipleMatches))
	})
}
