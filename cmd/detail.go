package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dswisher/markban/internal/board"
	"github.com/dswisher/markban/internal/config"
)

var detailCmd = &cobra.Command{
	Use:     "detail <card>",
	Aliases: []string{"details"},
	Short:   "Display the details of a card",
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

	var epicSummary string
	if result.Task.Type == "epic" {
		epicSummary, err = loadEpicSummary(dir, result.Task.Slug)
		if err != nil {
			return fmt.Errorf("loading epic subtasks: %w", err)
		}
	}

	// Display the card
	if userCfg.MarkdownViewer == "" {
		// No viewer configured, print raw contents
		content, err := os.ReadFile(result.Path)
		if err != nil {
			return fmt.Errorf("reading card file: %w", err)
		}
		fmt.Print(string(content))
		if epicSummary != "" {
			fmt.Print(epicSummary)
		}
	} else {
		// Use configured viewer
		viewerPath := result.Path
		if epicSummary != "" {
			content, err := os.ReadFile(result.Path)
			if err != nil {
				return fmt.Errorf("reading card file: %w", err)
			}
			file, err := os.CreateTemp(filepath.Dir(result.Path), ".markban-detail-*.md")
			if err != nil {
				return fmt.Errorf("creating temporary detail file: %w", err)
			}
			viewerPath = file.Name()
			defer os.Remove(viewerPath)
			if _, err := file.Write(append(content, []byte(epicSummary)...)); err != nil {
				file.Close()
				return fmt.Errorf("writing temporary detail file: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("closing temporary detail file: %w", err)
			}
		}
		if err := runViewer(userCfg.MarkdownViewer, viewerPath); err != nil {
			return fmt.Errorf("running viewer: %w", err)
		}
	}

	return nil
}

func loadEpicSummary(dir, epicSlug string) (string, error) {
	b, _, err := board.LoadBoard(dir)
	if err != nil {
		return "", err
	}
	archived, err := board.LoadArchive(dir)
	if err != nil {
		return "", err
	}

	var subtasks []board.Task
	for _, column := range b.Columns {
		for _, task := range column.Tasks {
			if strings.EqualFold(task.Epic, epicSlug) {
				subtasks = append(subtasks, task)
			}
		}
	}
	for _, task := range archived {
		if strings.EqualFold(task.Epic, epicSlug) {
			subtasks = append(subtasks, task)
		}
	}

	var summary strings.Builder
	summary.WriteString("\n\n## Subtasks\n\n")
	if len(subtasks) == 0 {
		summary.WriteString("No subtasks found.\n")
		return summary.String(), nil
	}
	summary.WriteString("| Subtask | Status | Slug |\n| --- | --- | --- |\n")
	for _, task := range subtasks {
		status := task.Column
		if task.Blocked {
			status += " (blocked)"
		}
		fmt.Fprintf(&summary, "| %s | %s | `%s` |\n", task.Title, status, task.Slug)
	}
	return summary.String(), nil
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
