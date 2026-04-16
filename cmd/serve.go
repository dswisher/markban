package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/dswisher/markban/internal/board"
	"github.com/dswisher/markban/internal/render"
	"github.com/dswisher/markban/internal/server"
)

var serveNoColor bool

var serveCmd = &cobra.Command{
	Use:   "serve [board-dir]",
	Short: "Serve a Kanban board via HTTP and open it in your web browser",
	Long: `Serve a Kanban board via HTTP and open it in your web browser.

If board-dir is not specified, the command will attempt to auto-discover
the board by finding the git root and looking for a subdirectory containing
board.toml or with "board" in its name.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runServe,
}

func init() {
	serveCmd.Flags().BoolVar(&serveNoColor, "no-color", false, "Disable colored card backgrounds")
}

func runServe(cmd *cobra.Command, args []string) error {
	dir, err := resolveBoardDir(args)
	if err != nil {
		return err
	}

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("cannot access %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}

	fmt.Fprintf(os.Stderr, "Loading board from: %s\n", dir)

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.New(dir, !serveNoColor)

	// Start the server in a goroutine so we can open the browser once it's up.
	runErr := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		close(started)
		runErr <- srv.Run(ctx)
	}()

	// Wait for the server goroutine to have been scheduled before opening.
	<-started

	// Give the server a moment to bind its port.
	// We poll until the port is set (Run sets it just before Serve).
	url := waitForPort(srv)

	if err := render.OpenBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not open browser: %v\n", err)
	}

	// Block until Ctrl-C or server error.
	if err := <-runErr; err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Shutting down.")
	return nil
}

// waitForPort polls srv.Port() until it is non-zero and returns the URL.
func waitForPort(srv *server.Server) string {
	for {
		if p := srv.Port(); p != 0 {
			return fmt.Sprintf("http://localhost:%d", p)
		}
	}
}

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
