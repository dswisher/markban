//go:build !windows

package terminal

// EnableVirtualTerminalProcessing is a no-op on non-Windows systems.
// Unix-like systems support ANSI escape sequences by default.
func EnableVirtualTerminalProcessing() error {
	return nil
}
