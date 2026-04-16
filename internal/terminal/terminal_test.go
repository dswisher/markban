package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCardColor_DarkMode_ValidColors(t *testing.T) {
	tests := []struct {
		colorName string
		wantEmpty bool
	}{
		{"yellow", false},
		{"green", false},
		{"blue", false},
		{"red", false},
		{"orange", false},
		{"purple", false},
		{"magenta", false},
		{"cyan", false},
		{"YELLOW", false}, // case insensitive
		{"Green", false},  // case insensitive
		{"", true},        // empty returns empty
		{"invalid", true}, // invalid returns empty
	}

	for _, tt := range tests {
		t.Run(tt.colorName, func(t *testing.T) {
			got := CardColor(tt.colorName, true) // dark mode
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.NotEmpty(t, got)
				assert.Contains(t, got, "\x1b[") // Should be an ANSI escape sequence
			}
		})
	}
}

func TestCardColor_LightMode_ValidColors(t *testing.T) {
	tests := []struct {
		colorName string
		wantEmpty bool
	}{
		{"yellow", false},
		{"green", false},
		{"blue", false},
		{"red", false},
		{"orange", false},
		{"purple", false},
		{"magenta", false},
		{"cyan", false},
		{"", true},        // empty returns empty
		{"invalid", true}, // invalid returns empty
	}

	for _, tt := range tests {
		t.Run(tt.colorName, func(t *testing.T) {
			got := CardColor(tt.colorName, false) // light mode
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.NotEmpty(t, got)
				assert.Contains(t, got, "\x1b[") // Should be an ANSI escape sequence
			}
		})
	}
}

func TestCardColor_DifferentModes(t *testing.T) {
	// Test that dark mode and light mode return different colors
	darkYellow := CardColor("yellow", true)
	lightYellow := CardColor("yellow", false)
	assert.NotEqual(t, darkYellow, lightYellow, "dark mode and light mode should return different colors")
}

func TestCardForeground_WithColor(t *testing.T) {
	result := CardForeground("test", "yellow")
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "\x1b[")
}

func TestCardForeground_NoColor(t *testing.T) {
	result := CardForeground("test", "")
	assert.Equal(t, "test", result)
}

func TestCardForegroundWithMode_WithColor(t *testing.T) {
	result := CardForegroundWithMode("test", "green", true)
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "\x1b[")
}

func TestCardForegroundWithMode_NoColor(t *testing.T) {
	result := CardForegroundWithMode("test", "", true)
	assert.Equal(t, "test", result)
}

func TestReset(t *testing.T) {
	assert.Equal(t, "\x1b[0m", Reset())
}
