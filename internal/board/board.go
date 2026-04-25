package board

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Task represents a single Kanban card, parsed from a Markdown file.
type Task struct {
	Title    string    // from the first # heading
	Blurb    string    // the line immediately after the title (if any)
	Priority string    // from frontmatter (optional)
	Tags     []string  // from frontmatter (optional)
	Color    string    // from frontmatter (optional): yellow, green, blue, red, orange, purple, magenta, cyan
	Done     time.Time // from frontmatter (optional): completion date
	Slug     string    // filename without .md extension
}

// Column represents a single Kanban column, backed by a subdirectory.
type Column struct {
	Name  string // directory name with numeric prefix stripped
	Order int    // parsed numeric prefix, or inferred order
	Tasks []Task
}

// Board represents the full Kanban board.
type Board struct {
	Name    string
	Columns []Column
}

// Config represents the board.toml configuration file.
type Config struct {
	Name string `toml:"name"`
}

// conventionalOrder assigns a sort order to columns whose names lack a numeric
// prefix, based on common Kanban conventions.
var conventionalOrder = map[string]int{
	"backlog":     0,
	"todo":        1,
	"to-do":       1,
	"in-progress": 2,
	"doing":       2,
	"review":      3,
	"done":        100,
	"archive":     101,
}

// numericPrefix matches a leading sequence of digits followed by a hyphen,
// e.g. "1-", "42-".
var numericPrefix = regexp.MustCompile(`^(\d+)-`)

// columnName strips a leading numeric prefix and hyphen from a directory name,
// returning the display name (with hyphens replaced by spaces) and the parsed
// order (or -1 if no prefix).
func columnName(dirName string) (name string, order int) {
	if m := numericPrefix.FindStringSubmatch(dirName); m != nil {
		n, _ := strconv.Atoi(m[1])
		name = dirName[len(m[0]):]
		return strings.ReplaceAll(name, "-", " "), n
	}
	return strings.ReplaceAll(dirName, "-", " "), -1
}

// inferOrder returns a sort key for a column that had no numeric prefix.
// It consults conventionalOrder first, then falls back to alphabetic via the
// caller sorting on Name.
func inferOrder(name string) int {
	// Normalize: lowercase and replace spaces with hyphens for lookup
	normalized := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	if o, ok := conventionalOrder[normalized]; ok {
		return o
	}
	// 50 sits between "in-progress" (2) and "done" (100) as a neutral fallback;
	// ties are broken alphabetically by the caller.
	return 50
}

// LoadBoard reads rootDir, discovers column subdirectories and their task
// files, and returns a fully populated Board sorted by column order.
// If a board.toml file exists in rootDir, it is loaded for configuration.
// The archive directory (named "archive") is excluded from columns.
func LoadBoard(rootDir string) (*Board, error) {
	config := loadConfig(rootDir)

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	var columns []Column

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories.
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		// Skip archive directory.
		if isArchiveDir(entry.Name()) {
			continue
		}

		name, order := columnName(entry.Name())

		tasks, err := loadTasks(filepath.Join(rootDir, entry.Name()))
		if err != nil {
			return nil, err
		}

		// Apply special sorting for "done" columns
		if isDoneColumn(name) {
			sortDoneTasks(tasks, rootDir, entry.Name())
		}

		if order == -1 {
			order = inferOrder(name)
		}

		columns = append(columns, Column{
			Name:  name,
			Order: order,
			Tasks: tasks,
		})
	}

	sort.Slice(columns, func(i, j int) bool {
		if columns[i].Order != columns[j].Order {
			return columns[i].Order < columns[j].Order
		}
		return columns[i].Name < columns[j].Name
	})

	return &Board{Name: config.Name, Columns: columns}, nil
}

// isArchiveDir returns true if the directory name (after stripping numeric
// prefix) is "archive".
func isArchiveDir(dirName string) bool {
	name, _ := columnName(dirName)
	return strings.EqualFold(name, "archive")
}

// LoadArchive reads the archive directory (named "archive") under rootDir
// and returns all tasks found within. Returns nil if no archive directory exists.
func LoadArchive(rootDir string) ([]Task, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if isArchiveDir(entry.Name()) {
			return loadTasks(filepath.Join(rootDir, entry.Name()))
		}
	}

	return nil, nil
}

// loadConfig reads board.toml from rootDir if it exists.
// Returns an empty Config if the file does not exist.
func loadConfig(rootDir string) Config {
	var config Config
	configPath := filepath.Join(rootDir, "board.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	_ = toml.Unmarshal(data, &config)
	return config
}

// loadTasks reads all .md files from a directory and parses each into a Task.
// Tasks are sorted alphabetically by slug (filename) for consistent display.
func loadTasks(dir string) ([]Task, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tasks []Task

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		task, err := ParseTask(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	// Sort tasks by priority (high -> medium -> low -> other), then title, then slug
	sortTasksByPriority(tasks)

	return tasks, nil
}

// sortTasksByPriority sorts tasks by priority (high -> medium -> low -> other), then title, then slug
func sortTasksByPriority(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		priorityOrder := map[string]int{
			"high":   0,
			"medium": 1,
			"low":    2,
		}
		pi, oki := priorityOrder[strings.ToLower(tasks[i].Priority)]
		if !oki {
			pi = 3 // anything else comes last
		}
		pj, okj := priorityOrder[strings.ToLower(tasks[j].Priority)]
		if !okj {
			pj = 3 // anything else comes last
		}
		if pi != pj {
			return pi < pj
		}
		if tasks[i].Title != tasks[j].Title {
			return tasks[i].Title < tasks[j].Title
		}
		return tasks[i].Slug < tasks[j].Slug
	})
}

// isDoneColumn returns true if the column name (normalized) is "done"
func isDoneColumn(name string) bool {
	normalized := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	return normalized == "done"
}

// sortDoneTasks sorts tasks in a done column by completion date (most recent first).
// Tasks without a done date are logged as warnings and sorted to the end.
// Secondary sort is by priority, then slug.
func sortDoneTasks(tasks []Task, rootDir, columnName string) {
	// Check for tasks without done dates and log warnings
	for _, task := range tasks {
		if task.Done.IsZero() {
			log.Printf("Warning: Task %q in column %q/%s has no done date", task.Slug, rootDir, columnName)
		}
	}

	// Very old date to use for sorting tasks without done dates to the end
	veryOldDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

	sort.Slice(tasks, func(i, j int) bool {
		// Get effective done dates (use very old date if zero)
		di := tasks[i].Done
		if di.IsZero() {
			di = veryOldDate
		}
		dj := tasks[j].Done
		if dj.IsZero() {
			dj = veryOldDate
		}

		// Sort by done date descending (most recent first)
		if !di.Equal(dj) {
			return di.After(dj)
		}

		// If dates are equal, fall back to priority sort
		priorityOrder := map[string]int{
			"high":   0,
			"medium": 1,
			"low":    2,
		}
		pi, oki := priorityOrder[strings.ToLower(tasks[i].Priority)]
		if !oki {
			pi = 3
		}
		pj, okj := priorityOrder[strings.ToLower(tasks[j].Priority)]
		if !okj {
			pj = 3
		}
		if pi != pj {
			return pi < pj
		}

		// Finally sort by slug
		return tasks[i].Slug < tasks[j].Slug
	})
}
