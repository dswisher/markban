package board

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTask_WithFrontmatter(t *testing.T) {
	task, err := ParseTask("testdata/with-frontmatter.md")
	require.NoError(t, err)

	assert.Equal(t, "with-frontmatter", task.Slug)
	assert.Equal(t, "My Task", task.Title)
	assert.Equal(t, "A short blurb about the task.", task.Blurb)
	assert.Equal(t, "high", task.Priority)
	assert.Equal(t, []string{"go", "cli"}, task.Tags)
}

func TestParseTask_NoFrontmatter(t *testing.T) {
	task, err := ParseTask("testdata/no-frontmatter.md")
	require.NoError(t, err)

	assert.Equal(t, "no-frontmatter", task.Slug)
	assert.Equal(t, "Simple Task", task.Title)
	assert.Equal(t, "Just a blurb, no frontmatter.", task.Blurb)
	assert.Empty(t, task.Priority)
	assert.Empty(t, task.Tags)
}

func TestParseTask_TitleOnly(t *testing.T) {
	task, err := ParseTask("testdata/title-only.md")
	require.NoError(t, err)

	assert.Equal(t, "Title Only", task.Title)
	assert.Empty(t, task.Blurb)
}

func TestParseTask_WithColor(t *testing.T) {
	task, err := ParseTask("testdata/with-color.md")
	require.NoError(t, err)

	assert.Equal(t, "with-color", task.Slug)
	assert.Equal(t, "Task with Color", task.Title)
	assert.Equal(t, "This task has a yellow color in its frontmatter.", task.Blurb)
	assert.Equal(t, "high", task.Priority)
	assert.Equal(t, "yellow", task.Color)
	assert.Equal(t, []string{"test", "color"}, task.Tags)
}
