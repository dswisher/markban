package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dswisher/markban/internal/board"
	"github.com/dswisher/markban/internal/terminal"
)

var viewWidth int
var viewNoColor bool

var viewCmd = &cobra.Command{
	Use:   "view [board-dir]",
	Short: "View a Kanban board in the terminal",
	Long: `View a Kanban board in the terminal.

If board-dir is not specified, the command will attempt to auto-discover
the board by finding the git root and looking for a subdirectory containing
board.toml or with "board" in its name.

The display is automatically adjusted to fit within the terminal screen,
truncating cards and columns as needed.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runView,
}

func init() {
	viewCmd.Flags().IntVarP(&viewWidth, "width", "w", 0, "Override the terminal width (0 = auto-detect)")
	viewCmd.Flags().BoolVar(&viewNoColor, "no-color", false, "Disable colored output")
}

// Constants for layout
const (
	minColumnWidth = 20 // Minimum width for a column
	columnPadding  = 3  // Padding between columns
	headerHeight   = 2  // Column header lines (name + separator)
	blurbIndent    = 3  // Number of spaces to indent blurb
)

func runView(cmd *cobra.Command, args []string) error {
	dir, err := resolveBoardDir(args)
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

	b, err := board.LoadBoard(dir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	screenWidth, screenHeight := terminal.Size()
	if viewWidth > 0 {
		screenWidth = viewWidth
	}

	useColor := !viewNoColor
	renderBoard(b, screenWidth, screenHeight, useColor)

	return nil
}

// renderBoard displays the board in the terminal, fitting it to the screen.
func renderBoard(b *board.Board, screenWidth, screenHeight int, useColor bool) {
	if len(b.Columns) == 0 {
		fmt.Println("No columns found in board.")
		return
	}

	// Calculate column width based on screen width
	numColumns := len(b.Columns)
	availableWidth := screenWidth - (numColumns-1)*columnPadding
	columnWidth := availableWidth / numColumns

	// Ensure minimum column width
	if columnWidth < minColumnWidth {
		columnWidth = minColumnWidth
	}

	// Calculate available height for cards (excluding header)
	availableHeight := screenHeight - headerHeight
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Render the board
	renderColumns(b.Columns, columnWidth, availableHeight, screenWidth, useColor)
}

// renderColumns renders all columns side by side.
func renderColumns(columns []board.Column, columnWidth, availableHeight, screenWidth int, useColor bool) {
	// Calculate how many columns fit on screen
	maxColumns := (screenWidth + columnPadding) / (columnWidth + columnPadding)
	if maxColumns < 1 {
		maxColumns = 1
	}
	if maxColumns > len(columns) {
		maxColumns = len(columns)
	}

	visibleColumns := columns[:maxColumns]
	truncatedColumns := len(columns) - maxColumns

	// Print column headers
	for i, col := range visibleColumns {
		if i > 0 {
			fmt.Print(strings.Repeat(" ", columnPadding))
		}
		header := truncateVisible(col.Name, columnWidth)
		fmt.Print(padRight(header, columnWidth))
	}
	if truncatedColumns > 0 {
		fmt.Print(strings.Repeat(" ", columnPadding))
		msg := fmt.Sprintf("(%d more)", truncatedColumns)
		fmt.Print(truncateVisible(msg, columnWidth))
	}
	fmt.Println()

	// Print separator line
	for i := 0; i < maxColumns; i++ {
		if i > 0 {
			fmt.Print(strings.Repeat(" ", columnPadding))
		}
		fmt.Print(strings.Repeat("-", columnWidth))
	}
	if truncatedColumns > 0 {
		fmt.Print(strings.Repeat(" ", columnPadding))
		fmt.Print(strings.Repeat("-", columnWidth))
	}
	fmt.Println()

	// Build card buffers for each column
	cardBuffers := make([][][]string, len(visibleColumns))
	maxCards := 0
	for i, col := range visibleColumns {
		cards := buildCards(col.Tasks, columnWidth, availableHeight, useColor)
		cardBuffers[i] = cards
		if len(cards) > maxCards {
			maxCards = len(cards)
		}
	}

	// Render cards line by line
	currentCardIdx := make([]int, len(visibleColumns))
	currentLineIdx := make([]int, len(visibleColumns))
	linesRendered := 0

	for linesRendered < availableHeight {
		allColumnsDone := true
		var line strings.Builder

		for colIdx, cards := range cardBuffers {
			if colIdx > 0 {
				line.WriteString(strings.Repeat(" ", columnPadding))
			}

			cardIdx := currentCardIdx[colIdx]
			lineIdx := currentLineIdx[colIdx]

			if cardIdx < len(cards) {
				allColumnsDone = false
				card := cards[cardIdx]

				if lineIdx < len(card) {
					// Output this line of the card
					line.WriteString(card[lineIdx])
					// Pad to column width
					visibleLen := terminal.VisibleLength(card[lineIdx])
					if visibleLen < columnWidth {
						line.WriteString(strings.Repeat(" ", columnWidth-visibleLen))
					}
					currentLineIdx[colIdx] = lineIdx + 1
				} else {
					// Card is done, move to next card
					currentCardIdx[colIdx] = cardIdx + 1
					currentLineIdx[colIdx] = 0

					// Check if there's a next card
					if cardIdx+1 < len(cards) {
						// Output blank line separator
						line.WriteString(strings.Repeat(" ", columnWidth))
					} else {
						// No more cards in this column
						line.WriteString(strings.Repeat(" ", columnWidth))
					}
				}
			} else {
				// Column has no more cards
				line.WriteString(strings.Repeat(" ", columnWidth))
			}
		}

		if allColumnsDone {
			break
		}

		fmt.Println(line.String())
		linesRendered++
	}

	// Check for overflow in any column
	overflowMsgs := make([]string, len(visibleColumns))
	hasOverflow := false
	for i, col := range visibleColumns {
		if currentCardIdx[i] < len(col.Tasks) {
			hasOverflow = true
			remaining := len(col.Tasks) - currentCardIdx[i]
			overflowMsgs[i] = fmt.Sprintf("(%d more)", remaining)
		}
	}

	if hasOverflow {
		var line strings.Builder
		for i, msg := range overflowMsgs {
			if i > 0 {
				line.WriteString(strings.Repeat(" ", columnPadding))
			}
			if msg != "" {
				line.WriteString(center(msg, columnWidth))
			} else {
				line.WriteString(strings.Repeat(" ", columnWidth))
			}
		}
		fmt.Println(line.String())
	}
}

// buildCards builds the card content for each task that fits.
// Returns a slice of cards, where each card is a slice of lines.
func buildCards(tasks []board.Task, columnWidth, availableHeight int, useColor bool) [][]string {
	var cards [][]string
	linesUsed := 0

	for i, task := range tasks {
		// Add blank line separator before cards (except first)
		if i > 0 {
			linesUsed++
		}

		// Check if we have room for at least the title
		if linesUsed >= availableHeight {
			break
		}

		isLastCard := true // Assume last until proven otherwise
		remainingHeight := availableHeight - linesUsed

		// Render this card
		card := renderCard(task, columnWidth, remainingHeight, isLastCard, useColor)
		cardHeight := len(card)

		// Check if card fits
		if linesUsed+cardHeight > availableHeight {
			// Doesn't fit, stop here
			break
		}

		cards = append(cards, card)
		linesUsed += cardHeight
	}

	return cards
}

// renderCard renders a single card and returns its lines.
func renderCard(task board.Task, columnWidth, remainingHeight int, isLastCard bool, useColor bool) []string {
	var lines []string

	// Determine if we should use color and detect terminal mode
	useCardColor := useColor && task.Color != ""

	// Title (bold, with foreground color if specified)
	if remainingHeight <= 0 {
		return lines
	}
	titleText := truncate(task.Title, columnWidth)
	if useCardColor {
		titleLine := terminal.CardForeground(terminal.Bold(titleText), task.Color)
		lines = append(lines, titleLine)
	} else {
		titleLine := terminal.Bold(titleText)
		lines = append(lines, titleLine)
	}
	remainingHeight--

	// Blurb (indented, wrapped or truncated)
	if task.Blurb != "" && remainingHeight > 0 {
		blurbWidth := columnWidth - blurbIndent
		if blurbWidth > 0 {
			var blurbLines []string
			if isLastCard {
				// Last card: truncate if needed
				blurbLines = renderBlurbTruncated(task.Blurb, blurbWidth, remainingHeight)
			} else {
				// Not last card: wrap fully
				blurbLines = renderBlurbWrapped(task.Blurb, blurbWidth)
			}
			for _, blurbLine := range blurbLines {
				if remainingHeight <= 0 {
					break
				}
				lineText := strings.Repeat(" ", blurbIndent) + blurbLine
				if useCardColor {
					// Apply foreground color to the blurb text as well
					lineText = terminal.CardForeground(lineText, task.Color)
				}
				lines = append(lines, lineText)
				remainingHeight--
			}
		}
	}

	// Slug (right-aligned, on a line by itself, wrapped in square brackets)
	if remainingHeight > 0 && task.Slug != "" {
		// Account for the brackets in width calculation
		maxSlugWidth := columnWidth - 2
		if maxSlugWidth < 0 {
			maxSlugWidth = 0
		}
		slugWithBrackets := "[" + truncate(task.Slug, maxSlugWidth) + "]"
		// Right-align the slug
		visibleLen := len(slugWithBrackets)
		if visibleLen < columnWidth {
			slugWithBrackets = strings.Repeat(" ", columnWidth-visibleLen) + slugWithBrackets
		}
		if useCardColor {
			slugWithBrackets = terminal.CardForeground(slugWithBrackets, task.Color)
		}
		lines = append(lines, slugWithBrackets)
	}

	return lines
}

// renderBlurbWrapped wraps the blurb text to fit within maxWidth.
// Returns as many lines as needed to display the full blurb.
func renderBlurbWrapped(blurb string, maxWidth int) []string {
	if maxWidth <= 0 {
		return nil
	}

	var lines []string
	words := strings.Fields(blurb)
	if len(words) == 0 {
		return lines
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return lines
}

// renderBlurbTruncated renders the blurb, truncating with "..." if it doesn't fit.
// Fits at most maxLines lines.
func renderBlurbTruncated(blurb string, maxWidth, maxLines int) []string {
	if maxWidth <= 0 || maxLines <= 0 {
		return nil
	}

	lines := renderBlurbWrapped(blurb, maxWidth)
	if len(lines) <= maxLines {
		return lines
	}

	// Need to truncate
	result := lines[:maxLines]
	lastLine := result[maxLines-1]
	if len(lastLine) > maxWidth-3 {
		lastLine = lastLine[:maxWidth-3]
	}
	result[maxLines-1] = lastLine + "..."
	return result
}

// truncate truncates a string to fit within maxWidth.
func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}

// truncateVisible truncates a string to fit within maxWidth (for display columns).
func truncateVisible(s string, maxWidth int) string {
	return truncate(s, maxWidth)
}

// padRight pads a string with spaces on the right to reach the desired width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// center centers a string within the given width.
func center(s string, width int) string {
	visibleLen := terminal.VisibleLength(s)
	if visibleLen >= width {
		return truncateVisible(s, width)
	}
	padding := (width - visibleLen) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", width-padding-visibleLen)
}
