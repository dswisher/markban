package render

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/dswisher/markban/internal/board"
)

// RenderAndOpen executes the board template into a temporary HTML file and
// opens it in the system default browser.
func RenderAndOpen(b *board.Board) error {
	f, err := os.CreateTemp("", "markban-*.html")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := renderHTML(b, f); err != nil {
		return err
	}

	return OpenBrowser(f.Name())
}

// RenderToDir renders the board as HTML and writes it to <buildDir>/index.html,
// creating the directory if it does not exist.
func RenderToDir(b *board.Board, buildDir string) error {
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return fmt.Errorf("creating build dir: %w", err)
	}

	indexPath := filepath.Join(buildDir, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("creating index.html: %w", err)
	}
	defer f.Close()

	return renderHTML(b, f)
}

// renderHTML executes the board template, writing the result to w.
func renderHTML(b *board.Board, w io.Writer) error {
	tmpl, err := template.New("board").Parse(boardTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, b)
}

// OpenBrowser opens the given URL or file path in the system default browser.
func OpenBrowser(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}
