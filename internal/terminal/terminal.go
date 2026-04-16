// Package terminal provides terminal-related utilities.
package terminal

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI escape sequences
const (
	escape    = "\x1b"
	boldStart = escape + "[1m"
	boldEnd   = escape + "[0m"
)

// Size returns the width and height of the terminal.
// If the terminal size cannot be determined, it returns default values.
func Size() (width, height int) {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return 80, 24 // Default terminal size
	}
	w, h, err := term.GetSize(fd)
	if err != nil {
		return 80, 24 // Default terminal size
	}
	return w, h
}

// IsTerminal returns true if stdout is a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Bold returns the string with ANSI bold formatting.
func Bold(s string) string {
	return boldStart + s + boldEnd
}

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, ch := range s {
		if ch == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if ch == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(ch)
	}
	return result.String()
}

// VisibleLength returns the visible length of a string (excluding ANSI codes).
func VisibleLength(s string) int {
	return len(StripANSI(s))
}
