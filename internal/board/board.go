package board

// Task represents a single Kanban card, parsed from a Markdown file.
type Task struct {
	Title    string   // from the first # heading
	Blurb    string   // the line immediately after the title (if any)
	Priority string   // from frontmatter (optional)
	Tags     []string // from frontmatter (optional)
	Slug     string   // filename without .md extension
}

// Column represents a single Kanban column, backed by a subdirectory.
type Column struct {
	Name  string // directory name with numeric prefix stripped
	Order int    // parsed numeric prefix, or inferred order
	Tasks []Task
}

// Board represents the full Kanban board.
type Board struct {
	Columns []Column
}
