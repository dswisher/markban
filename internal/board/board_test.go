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
	assert.Equal(t, "in-progress", name)
	assert.Equal(t, 3, order)
}

func TestColumnName_WithoutPrefix(t *testing.T) {
	name, order := columnName("backlog")
	assert.Equal(t, "backlog", name)
	assert.Equal(t, -1, order)
}
