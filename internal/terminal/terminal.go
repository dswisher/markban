// Package terminal provides terminal-related utilities.
package terminal

import (
	"os"

	"golang.org/x/term"
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
