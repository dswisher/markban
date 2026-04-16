// Package terminal provides terminal-related utilities.
package terminal

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// ANSI escape sequences
const (
	escape          = "\x1b"
	boldStart       = escape + "[1m"
	boldEnd         = escape + "[0m"
	reset           = escape + "[0m"
	fgBlack         = escape + "[30m"
	fgRed           = escape + "[31m"
	fgGreen         = escape + "[32m"
	fgYellow        = escape + "[33m"
	fgBlue          = escape + "[34m"
	fgMagenta       = escape + "[35m"
	fgCyan          = escape + "[36m"
	fgWhite         = escape + "[37m"
	fgBrightBlack   = escape + "[90m"
	fgBrightRed     = escape + "[91m"
	fgBrightGreen   = escape + "[92m"
	fgBrightYellow  = escape + "[93m"
	fgBrightBlue    = escape + "[94m"
	fgBrightMagenta = escape + "[95m"
	fgBrightCyan    = escape + "[96m"
	fgBrightWhite   = escape + "[97m"
)

// ValidColors are the acceptable color names for task cards.
var ValidColors = []string{"yellow", "green", "blue", "red", "orange", "purple", "magenta", "cyan"}

// IsDarkMode attempts to detect if the terminal is using a dark background.
// It checks the COLORFGBG environment variable which some terminals set.
// Returns true if dark mode is detected or if we can't determine the mode.
func IsDarkMode() bool {
	// COLORFGBG is set by some terminals in the format "fg;bg" or "fg;bg;attr"
	// where values are 0-15 (0-7 are standard colors, 8-15 are bright)
	// 0=black, 7=white, 8=bright black, 15=bright white
	colorfgbg := os.Getenv("COLORFGBG")
	if colorfgbg != "" {
		parts := strings.Split(colorfgbg, ";")
		if len(parts) >= 2 {
			bg, err := strconv.Atoi(parts[1])
			if err == nil {
				// Background colors 0-7: 0=black, 1=red, 2=green, 3=yellow, 4=blue, 5=magenta, 6=cyan, 7=white
				// If background is black (0) or dark colors, it's dark mode
				// If background is white (7 or 15) or light colors, it's light mode
				return bg <= 6 || bg == 8 // black, dark colors, or bright black (gray)
			}
		}
	}

	// Default to dark mode if we can't detect
	return true
}

// CardColor returns a card foreground color for the terminal based on the color name.
// It returns appropriate colors for dark or light mode terminals.
// Returns empty string if the color is invalid or empty.
func CardColor(colorName string, darkMode bool) string {
	if darkMode {
		// For dark backgrounds, use bright colors
		switch strings.ToLower(colorName) {
		case "yellow":
			return fgBrightYellow
		case "green":
			return fgBrightGreen
		case "blue":
			return fgBrightBlue
		case "red":
			return fgBrightRed
		case "orange":
			return escape + "[38;5;214m" // ANSI 256-color for orange
		case "purple":
			return escape + "[38;5;141m" // ANSI 256-color for light purple
		case "magenta":
			return fgBrightMagenta
		case "cyan":
			return fgBrightCyan
		default:
			return ""
		}
	} else {
		// For light backgrounds, use dark colors
		switch strings.ToLower(colorName) {
		case "yellow":
			return escape + "[38;5;130m" // Dark yellow/olive
		case "green":
			return fgGreen
		case "blue":
			return fgBlue
		case "red":
			return fgRed
		case "orange":
			return escape + "[38;5;166m" // Dark orange
		case "purple":
			return escape + "[38;5;93m" // Dark purple
		case "magenta":
			return fgMagenta
		case "cyan":
			return fgCyan
		default:
			return ""
		}
	}
}

// CardForeground applies a foreground color to text for terminal display.
// Automatically detects dark/light mode and chooses appropriate color intensity.
// If colorName is empty or invalid, returns the text unchanged.
func CardForeground(text, colorName string) string {
	return CardForegroundWithMode(text, colorName, IsDarkMode())
}

// CardForegroundWithMode applies a foreground color to text for terminal display.
// If colorName is empty or invalid, returns the text unchanged.
func CardForegroundWithMode(text, colorName string, darkMode bool) string {
	colorCode := CardColor(colorName, darkMode)
	if colorCode == "" {
		return text
	}
	return colorCode + text + reset
}

// Reset returns the ANSI reset code.
func Reset() string {
	return reset
}

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
