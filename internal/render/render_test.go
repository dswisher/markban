package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dswisher/markban/internal/board"
)

func makeTestBoard() *board.Board {
	return &board.Board{
		Columns: []board.Column{
			{
				Name:  "Backlog",
				Order: 1,
				Tasks: []board.Task{
					{Title: "First Task", Blurb: "A short blurb."},
					{Title: "Second Task"},
				},
			},
			{
				Name:  "Done",
				Order: 100,
				Tasks: []board.Task{},
			},
		},
	}
}

func assertBoardHTML(t *testing.T, html string) {
	t.Helper()
	assert.Contains(t, html, "Backlog")
	assert.Contains(t, html, "Done")
	assert.Contains(t, html, "First Task")
	assert.Contains(t, html, "A short blurb.")
	assert.Contains(t, html, "Second Task")
	// Empty column should render the placeholder.
	assert.Contains(t, html, "No tasks")
	// Should be valid enough HTML to have the doctype and a body.
	assert.True(t, strings.HasPrefix(strings.TrimSpace(html), "<!DOCTYPE html>"))
	assert.Contains(t, html, "</html>")
}

func TestRenderHTML(t *testing.T) {
	b := makeTestBoard()

	var buf strings.Builder
	err := renderHTML(b, &buf)
	require.NoError(t, err)

	assertBoardHTML(t, buf.String())
}

func TestRenderHTML_ContainsSSEScript(t *testing.T) {
	b := makeTestBoard()

	var buf strings.Builder
	err := renderHTML(b, &buf)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `new EventSource("/events")`)
}

func TestRenderToDir(t *testing.T) {
	b := makeTestBoard()

	dir := t.TempDir()
	buildDir := filepath.Join(dir, ".build")

	err := RenderToDir(b, buildDir)
	require.NoError(t, err)

	indexPath := filepath.Join(buildDir, "index.html")
	data, err := os.ReadFile(indexPath)
	require.NoError(t, err)

	assertBoardHTML(t, string(data))
}
