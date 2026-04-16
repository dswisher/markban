package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "markban",
	Short: "A Markdown-based Kanban board viewer",
	Long: `Markban reads a directory-based Kanban board made up of Markdown task
files and renders it as a static HTML page opened in your default browser.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(browseCmd)
	rootCmd.AddCommand(viewCmd)

	// Make "view" the default command when no subcommand is specified
	// by copying view's RunE and Args settings to root
	rootCmd.RunE = viewCmd.RunE
	rootCmd.Args = viewCmd.Args
}
