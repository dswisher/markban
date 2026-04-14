package render

import (
	"html/template"
	"os"
	"os/exec"
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

	return openBrowser(f.Name())
}

// renderHTML executes the board template, writing the result to w.
func renderHTML(b *board.Board, w *os.File) error {
	tmpl, err := template.New("board").Parse(boardTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, b)
}

// openBrowser opens the given file path in the system default browser.
func openBrowser(path string) error {
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
