package board

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadBoard_ColumnOrder(t *testing.T) {
	board, err := LoadBoard("testdata/board")
	require.NoError(t, err)

	require.Len(t, board.Columns, 4)

	// Numeric-prefixed columns come first in prefix order.
	assert.Equal(t, "backlog", board.Columns[0].Name)
	assert.Equal(t, 1, board.Columns[0].Order)

	assert.Equal(t, "todo", board.Columns[1].Name)
	assert.Equal(t, 2, board.Columns[1].Order)

	// Conventional name "unknown" has no special order (falls back to 50),
	// which sorts before "done" (100).
	assert.Equal(t, "unknown", board.Columns[2].Name)

	// Conventional name "done" sorts last.
	assert.Equal(t, "done", board.Columns[3].Name)
	assert.Equal(t, 100, board.Columns[3].Order)
}

func TestLoadBoard_TasksLoaded(t *testing.T) {
	board, err := LoadBoard("testdata/board")
	require.NoError(t, err)

	// backlog has one task
	require.Len(t, board.Columns[0].Tasks, 1)
	assert.Equal(t, "Backlog Task", board.Columns[0].Tasks[0].Title)

	// todo has two tasks
	require.Len(t, board.Columns[1].Tasks, 2)
}

func TestLoadBoard_SkipsHiddenDirs(t *testing.T) {
	board, err := LoadBoard("testdata/board")
	require.NoError(t, err)

	for _, col := range board.Columns {
		assert.False(t, len(col.Name) > 0 && col.Name[0] == '.')
	}
}

func TestLoadBoard_InvalidDir(t *testing.T) {
	_, err := LoadBoard("testdata/nonexistent")
	assert.Error(t, err)
}

func TestColumnName_WithPrefix(t *testing.T) {
	name, order := columnName("3-in-progress")
	assert.Equal(t, "in progress", name)
	assert.Equal(t, 3, order)
}

func TestColumnName_WithoutPrefix(t *testing.T) {
	name, order := columnName("backlog")
	assert.Equal(t, "backlog", name)
	assert.Equal(t, -1, order)
}

func TestColumnName_HyphenToSpace(t *testing.T) {
	name, order := columnName("in-progress")
	assert.Equal(t, "in progress", name)
	assert.Equal(t, -1, order)
}

func TestLoadBoard_WithConfig(t *testing.T) {
	board, err := LoadBoard("testdata/board-with-config")
	require.NoError(t, err)
	assert.Equal(t, "Test Project", board.Name)
}

func TestLoadBoard_WithoutConfig(t *testing.T) {
	board, err := LoadBoard("testdata/board")
	require.NoError(t, err)
	assert.Equal(t, "", board.Name)
}

func TestLoadBoard_ExcludesArchiveDir(t *testing.T) {
	board, err := LoadBoard("testdata/board-with-archive")
	require.NoError(t, err)

	// Should have 2 columns (backlog, todo), not 3 (archive excluded)
	require.Len(t, board.Columns, 2)
	assert.Equal(t, "backlog", board.Columns[0].Name)
	assert.Equal(t, "todo", board.Columns[1].Name)
}

func TestLoadArchive(t *testing.T) {
	tasks, err := LoadArchive("testdata/board-with-archive")
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	// Tasks should be sorted by title
	assert.Equal(t, "Archived Task One", tasks[0].Title)
	assert.Equal(t, "Archived Task Two", tasks[1].Title)
}

func TestLoadArchive_NoArchiveDir(t *testing.T) {
	tasks, err := LoadArchive("testdata/board")
	require.NoError(t, err)
	assert.Nil(t, tasks)
}

func TestIsArchiveDir(t *testing.T) {
	tests := []struct {
		dirName string
		want    bool
	}{
		{"archive", true},
		{"99-archive", true},
		{"5-archive", true},
		{"Archive", true},
		{"ARCHIVE", true},
		{"backlog", false},
		{"done", false},
		{"archived", false},
	}

	for _, tt := range tests {
		t.Run(tt.dirName, func(t *testing.T) {
			got := isArchiveDir(tt.dirName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsDoneColumn(t *testing.T) {
	// Note: isDoneColumn receives the name AFTER columnName strips the numeric prefix
	// So "4-done" becomes "done" before isDoneColumn is called
	tests := []struct {
		name string
		want bool
	}{
		{"done", true},
		{"Done", true},
		{"DONE", true},
		{"Done Tasks", false}, // normalizes to "done-tasks", not "done"
		{"backlog", false},
		{"todo", false},
		{"in progress", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDoneColumn(tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadBoard_DoneColumnSortByDate(t *testing.T) {
	board, err := LoadBoard("testdata/board-with-done")
	require.NoError(t, err)

	// Find the done column
	var doneCol *Column
	for i := range board.Columns {
		if board.Columns[i].Name == "done" {
			doneCol = &board.Columns[i]
			break
		}
	}
	require.NotNil(t, doneCol, "should have a done column")
	require.Len(t, doneCol.Tasks, 4, "done column should have 4 tasks")

	// Expected order (most recent first):
	// 1. recent-task (2026-04-20, medium)
	// 2. same-date-no-priority (2026-04-20, no priority - comes after medium)
	// 3. older-task (2026-04-15, high)
	// 4. no-date-task (no date - sorted to end)
	assert.Equal(t, "recent-task", doneCol.Tasks[0].Slug)
	assert.Equal(t, "same-date-no-priority", doneCol.Tasks[1].Slug)
	assert.Equal(t, "older-task", doneCol.Tasks[2].Slug)
	assert.Equal(t, "no-date-task", doneCol.Tasks[3].Slug)

	// Verify dates
	assert.Equal(t, "2026-04-20", doneCol.Tasks[0].Done.Format("2006-01-02"))
	assert.Equal(t, "2026-04-20", doneCol.Tasks[1].Done.Format("2006-01-02"))
	assert.Equal(t, "2026-04-15", doneCol.Tasks[2].Done.Format("2006-01-02"))
	assert.True(t, doneCol.Tasks[3].Done.IsZero())
}
