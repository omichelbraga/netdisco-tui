package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResizableTable manages column widths and mouse-based resizing
type ResizableTable struct {
	ColumnWidths []int
	MinWidths    []int
	resizing     bool
	resizeCol    int
	resizeStartX int
	resizeStartW int
	hoverCol     int
	hovering     bool
}

// NewResizableTable creates a new resizable table with default column widths
func NewResizableTable(defaultWidths []int) ResizableTable {
	minWidths := make([]int, len(defaultWidths))
	for i, w := range defaultWidths {
		minWidths[i] = max(8, w/2) // Min width is half of default, at least 8
	}
	return ResizableTable{
		ColumnWidths: defaultWidths,
		MinWidths:    minWidths,
	}
}

// HandleMouse processes mouse events for column resizing
// xOffset: horizontal position where the table starts (usually 0, but may have left padding)
func (t *ResizableTable) HandleMouse(msg tea.MouseMsg, xOffset int) {
	// If we're currently resizing, update the column width for any mouse movement
	if t.resizing {
		if msg.Type == tea.MouseRelease {
			t.resizing = false
			return
		}
		// Update column width based on current mouse position
		adjustedX := msg.X - xOffset
		delta := adjustedX - t.resizeStartX
		newWidth := t.resizeStartW + delta

		// Enforce minimum width
		if newWidth < t.MinWidths[t.resizeCol] {
			newWidth = t.MinWidths[t.resizeCol]
		}

		t.ColumnWidths[t.resizeCol] = newWidth
		return
	}

	// Not currently resizing - check if starting a resize
	switch msg.Type {
	case tea.MouseLeft:
		// Check if clicking near a column border (the │ separators)
		adjustedX := msg.X - xOffset
		x := 0
		for i := 0; i < len(t.ColumnWidths)-1; i++ {
			x += t.ColumnWidths[i]
			// Check if clicking on the separator or within ±2 chars of it
			// The separator itself is at position x, and is 1 char wide
			if adjustedX >= x-2 && adjustedX <= x+3 {
				t.resizing = true
				t.resizeCol = i
				t.resizeStartX = adjustedX
				t.resizeStartW = t.ColumnWidths[i]
				return
			}
			// Move past the separator for the next column
			x += 1
		}
	}
}

// RenderHeader renders a table header with the current column widths
func (t *ResizableTable) RenderHeader(headers []string) string {
	var parts []string
	for i, header := range headers {
		width := t.ColumnWidths[i]
		text := padToWidth(header, width)
		col := lipgloss.NewStyle().
			Bold(true).
			Foreground(CurrentTheme.Secondary).
			Render(text)
		parts = append(parts, col)

		// Add separator between columns (except after last)
		if i < len(headers)-1 {
			sep := lipgloss.NewStyle().
				Foreground(CurrentTheme.BorderMuted).
				Render("│")
			parts = append(parts, sep)
		}
	}
	return strings.Join(parts, "")
}

// RenderSeparator renders the separator line between header and data
// Shows visual handles (╪) at borders to indicate they're draggable
func (t *ResizableTable) RenderSeparator() string {
	var parts []string
	for i, width := range t.ColumnWidths {
		parts = append(parts, strings.Repeat("━", width))
		if i < len(t.ColumnWidths)-1 {
			// Use ╪ to show draggable border
			var handle string
			if t.resizing && t.resizeCol == i {
				handle = lipgloss.NewStyle().Foreground(CurrentTheme.Primary).Bold(true).Render("╋")
			} else {
				handle = lipgloss.NewStyle().Foreground(CurrentTheme.Secondary).Render("╪")
			}
			parts = append(parts, handle)
		}
	}
	return lipgloss.NewStyle().Foreground(CurrentTheme.Border).Render(strings.Join(parts, ""))
}

// RenderRow renders a data row with the current column widths
func (t *ResizableTable) RenderRow(values []string, colors []lipgloss.Color) string {
	var parts []string
	for i, value := range values {
		width := t.ColumnWidths[i]
		color := CurrentTheme.Text
		if i < len(colors) {
			color = colors[i]
		}
		text := padToWidth(value, width)
		col := lipgloss.NewStyle().
			Foreground(color).
			Render(text)
		parts = append(parts, col)

		// Add separator between columns (except after last)
		if i < len(values)-1 {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(CurrentTheme.BorderMuted).
				Render("│"))
		}
	}
	return strings.Join(parts, "")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// padToWidth ensures a string is exactly the specified width
// Truncates if too long (with ...), pads with spaces if too short
func padToWidth(s string, width int) string {
	if len(s) > width {
		if width <= 3 {
			return strings.Repeat(".", width)
		}
		return s[:width-3] + "..."
	}
	if len(s) < width {
		return s + strings.Repeat(" ", width-len(s))
	}
	return s
}
