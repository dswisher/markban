package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dswisher/markban/internal/board"
	"github.com/dswisher/markban/internal/config"
)

var detailCmd = &cobra.Command{
	Use:   "detail <card>",
	Short: "Display the details of a card",
	Long: `Display the full contents of a card.

The card can be specified by:
  - Slug (filename without .md): "markban detail live-reload"
  - Substring of title (if unique): "markban detail dark"

If no matching card is found, an error is displayed.

The output can be piped through a markdown viewer configured in:
  ~/.config/markban/markban.toml

Example config:
  markdown_viewer = "bat"
  # or
  markdown_viewer = "bat -l md --paging=always"
  # or leave empty/unset to print raw markdown`,
	Args: cobra.ExactArgs(1),
	RunE: runDetail,
}

func runDetail(cmd *cobra.Command, args []string) error {
	// Load user configuration
	userCfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("loading user config: %w", err)
	}

	// Resolve board directory
	dir, err := resolveBoardDir([]string{})
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

	// Create card finder and search for the card
	finder := board.NewCardFinder(dir)
	result, err := finder.FindCard(args[0])
	if err != nil {
		if errors.Is(err, board.ErrNoMatch) {
			fmt.Fprintln(os.Stderr, "no matching card found")
			return nil
		}
		if errors.Is(err, board.ErrMultipleMatches) {
			fmt.Fprintln(os.Stderr, err.Error())
			return nil
		}
		return err
	}

	// Display the card
	if userCfg.MarkdownViewer == "" {
		// No viewer configured, print raw contents
		content, err := os.ReadFile(result.Path)
		if err != nil {
			return fmt.Errorf("reading card file: %w", err)
		}
		fmt.Print(string(content))
	} else {
		// Use configured viewer
		if err := runViewer(userCfg.MarkdownViewer, result.Path); err != nil {
			return fmt.Errorf("running viewer: %w", err)
		}
	}

	return nil
}

// runViewer executes the markdown viewer with the given file.
// The viewer command can include arguments (e.g., "bat -l md").
func runViewer(viewerCmd, filePath string) error {
	parts := strings.Fields(viewerCmd)
	if len(parts) == 0 {
		return errors.New("empty viewer command")
	}

	cmd := exec.Command(parts[0], append(parts[1:], filePath)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
