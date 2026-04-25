package cmd

import (
	"fmt"
	"os"

	"github.com/dswisher/markban/internal/board"
)

// resolveBoardDir determines the board directory to use.
// If args contains a directory, it is used directly.
// Otherwise, auto-discovery is attempted by finding the git root
// and looking for a board subdirectory.
func resolveBoardDir(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	// Auto-discover the board directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current directory: %w", err)
	}

	boardDir, err := board.FindProjectBoard(currentDir)
	if err != nil {
		return "", fmt.Errorf("cannot auto-discover board directory: %w", err)
	}

	return boardDir, nil
}
