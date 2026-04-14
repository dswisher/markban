package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/dswisher/markban/internal/render"
	"github.com/dswisher/markban/internal/server"
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

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.New(dir)

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
