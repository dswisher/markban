//go:build windows

package terminal

import (
	"golang.org/x/sys/windows"
)

// EnableVirtualTerminalProcessing enables ANSI escape sequence support on Windows.
// This is required for modern terminals like WezTerm, Windows Terminal, etc.
// Returns an error if the console mode cannot be set.
func EnableVirtualTerminalProcessing() error {
	handle := windows.Handle(windows.Stdout)

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}

	// Enable virtual terminal processing for ANSI escape sequences
	// ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING

	if err := windows.SetConsoleMode(handle, mode); err != nil {
		return err
	}

	return nil
}
