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
// Card colors are enabled by default.
func RenderAndOpen(b *board.Board) error {
	f, err := os.CreateTemp("", "markban-*.html")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := renderHTML(b, f, true, false); err != nil {
		return err
	}

	return OpenBrowser(f.Name())
}

// RenderToDir renders the board as HTML and writes it to <buildDir>/index.html,
// creating the directory if it does not exist.
// useColor determines whether to render card background colors.
func RenderToDir(b *board.Board, buildDir string, useColor bool, hasArchive bool) error {
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return fmt.Errorf("creating build dir: %w", err)
	}

	indexPath := filepath.Join(buildDir, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("creating index.html: %w", err)
	}
	defer f.Close()

	return renderHTML(b, f, useColor, hasArchive)
}

// RenderArchive renders the archive page as HTML and writes it to <buildDir>/archive.html.
func RenderArchive(tasks []board.Task, buildDir string) error {
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return fmt.Errorf("creating build dir: %w", err)
	}

	archivePath := filepath.Join(buildDir, "archive.html")
	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("creating archive.html: %w", err)
	}
	defer f.Close()

	return renderArchiveHTML(tasks, f)
}

// templateData holds the data passed to the HTML template.
type templateData struct {
	*board.Board
	UseColor   bool
	HasArchive bool
}

// archiveData holds the data passed to the archive HTML template.
type archiveData struct {
	Tasks []board.Task
}

// renderHTML executes the board template, writing the result to w.
// useColor determines whether to render card background colors.
func renderHTML(b *board.Board, w io.Writer, useColor bool, hasArchive bool) error {
	tmpl, err := template.New("board").Funcs(template.FuncMap{
		"cardStyle": cardStyleFunc(useColor),
	}).Parse(boardTemplate)
	if err != nil {
		return err
	}

	data := templateData{
		Board:      b,
		UseColor:   useColor,
		HasArchive: hasArchive,
	}

	return tmpl.Execute(w, data)
}

// renderArchiveHTML executes the archive template, writing the result to w.
func renderArchiveHTML(tasks []board.Task, w io.Writer) error {
	tmpl, err := template.New("archive").Parse(archiveTemplate)
	if err != nil {
		return err
	}

	data := archiveData{
		Tasks: tasks,
	}

	return tmpl.Execute(w, data)
}

// cardStyleFunc returns a function that generates CSS styles for cards.
// If useColor is false, it returns an empty string.
func cardStyleFunc(useColor bool) func(string) template.CSS {
	return func(color string) template.CSS {
		if !useColor || color == "" {
			return ""
		}
		bg := colorToCSS(color)
		if bg == "" {
			return ""
		}
		return template.CSS(fmt.Sprintf("background-color: %s;", bg))
	}
}

// colorToCSS maps color names to CSS color values.
func colorToCSS(color string) string {
	switch color {
	case "yellow":
		return "#fff9c4"
	case "green":
		return "#c8e6c9"
	case "blue":
		return "#bbdefb"
	case "red":
		return "#ffcdd2"
	case "orange":
		return "#ffe0b2"
	case "purple":
		return "#e1bee7"
	case "magenta":
		return "#f8bbd0"
	case "cyan":
		return "#b2ebf2"
	default:
		return ""
	}
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
