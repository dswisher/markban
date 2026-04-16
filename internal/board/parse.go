package board

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// frontmatter holds the optional YAML fields at the top of a task file.
type frontmatter struct {
	Priority string   `yaml:"priority"`
	Tags     []string `yaml:"tags"`
	Color    string   `yaml:"color"`
}

// ParseTask reads a single Markdown task file and returns a Task.
func ParseTask(path string) (Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Task{}, err
	}

	slug := strings.TrimSuffix(filepath.Base(path), ".md")

	content := string(data)
	var fm frontmatter

	// Split off YAML frontmatter if the file starts with "---".
	if strings.HasPrefix(strings.TrimLeft(content, "\r\n"), "---") {
		// Find the start of the opening "---".
		start := strings.Index(content, "---")
		// Find the closing "---" after the opening marker.
		rest := content[start+3:]
		if yamlBlock, after, found := strings.Cut(rest, "---"); found {
			_ = yaml.Unmarshal([]byte(yamlBlock), &fm)
			content = after
		}
	}

	title, blurb := extractTitleAndBlurb(content)

	return Task{
		Title:    title,
		Blurb:    blurb,
		Priority: fm.Priority,
		Tags:     fm.Tags,
		Color:    fm.Color,
		Slug:     slug,
	}, nil
}

// extractTitleAndBlurb scans markdown text for the first "# " heading and the
// first non-empty line that follows it.
func extractTitleAndBlurb(content string) (title, blurb string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	foundTitle := false

	for scanner.Scan() {
		line := scanner.Text()

		if !foundTitle {
			if rest, ok := strings.CutPrefix(line, "# "); ok {
				title = strings.TrimSpace(rest)
				foundTitle = true
			}
			continue
		}

		// We have the title; look for the next non-empty line as the blurb.
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Skip lines that start a new heading — those are sections, not blurbs.
			if !strings.HasPrefix(trimmed, "#") {
				blurb = trimmed
			}
			break
		}
	}

	return title, blurb
}
