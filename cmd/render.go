package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dswisher/markban/internal/board"
	"github.com/dswisher/markban/internal/render"
)

var renderCmd = &cobra.Command{
	Use:   "render <board-dir>",
	Short: "Render a Kanban board and open it in the browser",
	Args:  cobra.ExactArgs(1),
	RunE:  runRender,
}

func runRender(cmd *cobra.Command, args []string) error {
	dir := args[0]

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("cannot access %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}

	b, err := board.LoadBoard(dir)
	if err != nil {
		return fmt.Errorf("failed to load board: %w", err)
	}

	if err := render.RenderAndOpen(b); err != nil {
		return fmt.Errorf("failed to render board: %w", err)
	}

	return nil
}
