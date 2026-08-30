package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dswisher/markban/internal/board"
)

func TestRenderCard_RelationshipIndicators(t *testing.T) {
	lines := renderCard(board.Task{
		Title:   "Child",
		Epic:    "main-epic",
		Blocked: true,
	}, 60, 5, true, false)

	assert.Contains(t, lines, "   [epic: main-epic, BLOCKED]")
}

func TestRenderCard_EpicIndicator(t *testing.T) {
	lines := renderCard(board.Task{Title: "Epic", Type: "epic"}, 40, 3, true, false)

	assert.Contains(t, lines, "   [EPIC]")
}
