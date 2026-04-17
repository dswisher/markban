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
	err := renderHTML(b, &buf, true, false)
	require.NoError(t, err)

	assertBoardHTML(t, buf.String())
}

func TestRenderHTML_ContainsSSEScript(t *testing.T) {
	b := makeTestBoard()

	var buf strings.Builder
	err := renderHTML(b, &buf, true, false)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `new EventSource("/events")`)
}

func TestRenderHTML_HyphenToSpace(t *testing.T) {
	b := &board.Board{
		Columns: []board.Column{
			{
				Name:  "in progress",
				Order: 1,
				Tasks: []board.Task{{Title: "Task"}},
			},
			{
				Name:  "to do",
				Order: 2,
				Tasks: []board.Task{{Title: "Task 2"}},
			},
		},
	}

	var buf strings.Builder
	err := renderHTML(b, &buf, true, false)
	require.NoError(t, err)

	html := buf.String()
	// Should contain the spaced version, not hyphenated
	assert.Contains(t, html, "in progress")
	assert.Contains(t, html, "to do")
	// Should not contain the hyphenated versions
	assert.NotContains(t, html, "in-progress")
	assert.NotContains(t, html, "to-do")
}

func TestRenderToDir(t *testing.T) {
	b := makeTestBoard()

	dir := t.TempDir()
	buildDir := filepath.Join(dir, ".build")

	err := RenderToDir(b, buildDir, true, false)
	require.NoError(t, err)

	indexPath := filepath.Join(buildDir, "index.html")
	data, err := os.ReadFile(indexPath)
	require.NoError(t, err)

	assertBoardHTML(t, string(data))
}

func TestRenderHTML_HasArchiveLink(t *testing.T) {
	b := makeTestBoard()

	var buf strings.Builder
	err := renderHTML(b, &buf, true, true)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `href="/archive"`)
	assert.Contains(t, buf.String(), "Archive")
}

func TestRenderHTML_NoArchiveLink(t *testing.T) {
	b := makeTestBoard()

	var buf strings.Builder
	err := renderHTML(b, &buf, true, false)
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), `href="/archive"`)
}

func TestRenderArchive(t *testing.T) {
	tasks := []board.Task{
		{Title: "Old Task", Blurb: "This was archived."},
		{Title: "Another Old Task"},
	}

	dir := t.TempDir()
	buildDir := filepath.Join(dir, ".build")

	err := RenderArchive(tasks, buildDir)
	require.NoError(t, err)

	archivePath := filepath.Join(buildDir, "archive.html")
	data, err := os.ReadFile(archivePath)
	require.NoError(t, err)

	html := string(data)
	assert.Contains(t, html, "Old Task")
	assert.Contains(t, html, "This was archived.")
	assert.Contains(t, html, "Another Old Task")
	assert.Contains(t, html, `href="/"`)
	assert.Contains(t, html, "Archive")
}

func TestRenderArchive_Empty(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, ".build")

	err := RenderArchive(nil, buildDir)
	require.NoError(t, err)

	archivePath := filepath.Join(buildDir, "archive.html")
	data, err := os.ReadFile(archivePath)
	require.NoError(t, err)

	assert.Contains(t, string(data), "No archived tasks")
}
