package board

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Common errors returned by the finder.
var (
	ErrNoMatch         = errors.New("no matching card found")
	ErrMultipleMatches = errors.New("multiple cards match the query")
)

// multipleMatchesError is a custom error type for multiple card matches.
type multipleMatchesError struct {
	msg   string
	query string
}

func (e *multipleMatchesError) Error() string {
	return e.msg
}

func (e *multipleMatchesError) Is(target error) bool {
	return target == ErrMultipleMatches
}

// MatchResult represents a found card with its metadata.
type MatchResult struct {
	Task   Task
	Path   string
	Column string
}

// CardFinder provides methods to locate cards within a board.
type CardFinder struct {
	rootDir string
}

// NewCardFinder creates a new card finder for the given board directory.
func NewCardFinder(rootDir string) *CardFinder {
	return &CardFinder{rootDir: rootDir}
}

// FindCard searches for a card using fuzzy matching:
//  1. Exact slug match (case-insensitive)
//  2. Title substring match (case-insensitive, only if unique)
//
// The archive directory is always excluded from searches.
func (f *CardFinder) FindCard(query string) (*MatchResult, error) {
	// First, try exact slug match
	result, err := f.FindBySlug(query)
	if err == nil {
		return result, nil
	}
	if !errors.Is(err, ErrNoMatch) {
		return nil, err
	}

	// No slug match, try title substring match
	matches, err := f.FindByTitleSubstring(query)
	if err != nil {
		return nil, err
	}

	switch len(matches) {
	case 0:
		return nil, ErrNoMatch
	case 1:
		return &matches[0], nil
	default:
		// Build error message with all matches
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("multiple cards match '%s':\n", query))
		for _, m := range matches {
			sb.WriteString(fmt.Sprintf("  - %s (in %s)\n", m.Task.Title, m.Column))
		}
		return nil, &multipleMatchesError{
			msg:   sb.String(),
			query: query,
		}
	}
}

// FindBySlug searches for a card by its slug (filename without .md extension).
// The search is case-insensitive.
func (f *CardFinder) FindBySlug(slug string) (*MatchResult, error) {
	queryLower := strings.ToLower(slug)

	entries, err := os.ReadDir(f.rootDir)
	if err != nil {
		return nil, fmt.Errorf("reading board directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		// Skip archive directory
		if isArchiveDir(entry.Name()) {
			continue
		}

		columnPath := filepath.Join(f.rootDir, entry.Name())
		result, err := f.findBySlugInColumn(columnPath, entry.Name(), queryLower)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}

	return nil, ErrNoMatch
}

// findBySlugInColumn searches for a card by slug within a specific column directory.
func (f *CardFinder) findBySlugInColumn(columnPath, columnDirName, queryLower string) (*MatchResult, error) {
	entries, err := os.ReadDir(columnPath)
	if err != nil {
		return nil, fmt.Errorf("reading column directory %q: %w", columnPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		slug := strings.TrimSuffix(entry.Name(), ".md")
		if strings.ToLower(slug) == queryLower {
			// Found match
			taskPath := filepath.Join(columnPath, entry.Name())
			task, err := ParseTask(taskPath)
			if err != nil {
				return nil, fmt.Errorf("parsing task %q: %w", taskPath, err)
			}

			return &MatchResult{
				Task:   task,
				Path:   taskPath,
				Column: columnDirName,
			}, nil
		}
	}

	return nil, nil
}

// FindByTitleSubstring searches for cards whose titles contain the given substring.
// The search is case-insensitive.
func (f *CardFinder) FindByTitleSubstring(substring string) ([]MatchResult, error) {
	queryLower := strings.ToLower(substring)
	var matches []MatchResult

	entries, err := os.ReadDir(f.rootDir)
	if err != nil {
		return nil, fmt.Errorf("reading board directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		// Skip archive directory
		if isArchiveDir(entry.Name()) {
			continue
		}

		columnPath := filepath.Join(f.rootDir, entry.Name())
		columnMatches, err := f.findByTitleInColumn(columnPath, entry.Name(), queryLower)
		if err != nil {
			return nil, err
		}
		matches = append(matches, columnMatches...)
	}

	return matches, nil
}

// findByTitleInColumn searches for cards by title substring within a specific column directory.
func (f *CardFinder) findByTitleInColumn(columnPath, columnDirName, queryLower string) ([]MatchResult, error) {
	var matches []MatchResult

	entries, err := os.ReadDir(columnPath)
	if err != nil {
		return nil, fmt.Errorf("reading column directory %q: %w", columnPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		taskPath := filepath.Join(columnPath, entry.Name())
		task, err := ParseTask(taskPath)
		if err != nil {
			return nil, fmt.Errorf("parsing task %q: %w", taskPath, err)
		}

		if strings.Contains(strings.ToLower(task.Title), queryLower) {
			matches = append(matches, MatchResult{
				Task:   task,
				Path:   taskPath,
				Column: columnDirName,
			})
		}
	}

	return matches, nil
}
