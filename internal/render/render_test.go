package render

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dswisher/markban/internal/board"
)

func TestRenderHTML(t *testing.T) {
	b := &board.Board{
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

	f, err := os.CreateTemp("", "markban-test-*.html")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	err = renderHTML(b, f)
	require.NoError(t, err)

	// Re-read the file to inspect the output.
	data, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	html := string(data)

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
