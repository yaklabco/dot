// Package adopt provides interactive file adoption.
//
// This file contains Bubble Tea UI code which is excluded from coverage
// requirements as interactive terminal UI cannot be reliably unit tested.
package adopt

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaklabco/dot/internal/domain"
)

var debugLog *log.Logger

func init() {
	// Create debug log file
	f, err := os.OpenFile("/tmp/dot-selector-debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		debugLog = log.New(f, "", log.Ltime|log.Lmicroseconds)
	}
}

// ArrowSelector provides an arrow-key based multi-select interface.
type ArrowSelector struct {
	input     io.Reader
	output    io.Writer
	fs        domain.FS
	configDir string
}

// NewArrowSelector creates a new arrow-key selector.
func NewArrowSelector(input io.Reader, output io.Writer, fs domain.FS, configDir string) *ArrowSelector {
	return &ArrowSelector{
		input:     input,
		output:    output,
		fs:        fs,
		configDir: configDir,
	}
}

// bubbleModel represents the Bubble Tea model for the selector.
type bubbleModel struct {
	items       []string
	cursor      int
	selected    map[int]bool
	viewportTop int
	height      int
	width       int
	quitting    bool
	confirmed   bool
	ignoring    map[int]bool       // Items being ignored (for animation)
	ignoreTime  map[int]time.Time  // When ignore started
	viewModal   bool               // Whether view modal is open
	viewContent string             // Content to show in modal
	candidates  []DotfileCandidate // Original candidates
	fs          domain.FS          // Filesystem for operations
	configDir   string             // Config directory
}

// Message types for ignore animation and view modal
type ignoreStartMsg struct {
	itemIdx int
}

type ignoreCompleteMsg struct {
	itemIdx int
}

type ignoreTickMsg time.Time

type viewContentMsg struct {
	content string
}

// Init initializes the Bubble Tea model.
func (m bubbleModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return ignoreTickMsg(t)
	})
}

// Update handles messages and updates the model state.
func (m bubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.MouseMsg:
		return m.handleMouseMsg(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case ignoreStartMsg:
		return m.handleIgnoreStartMsg(msg)
	case ignoreTickMsg:
		return m.handleIgnoreTickMsg(msg)
	case ignoreCompleteMsg:
		return m.handleIgnoreCompleteMsg(msg)
	case viewContentMsg:
		return m.handleViewContentMsg(msg)
	}
	return m, nil
}

// handleKeyMsg processes keyboard input.
func (m bubbleModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle modal ESC first
	if m.viewModal && (msg.String() == "esc" || msg.String() == "q") {
		m.viewModal = false
		return m, nil
	}

	// Check for quit keys
	if cmd := m.handleQuitKeys(msg); cmd != nil {
		return m, cmd
	}

	// Handle navigation when not in modal
	if !m.viewModal {
		m.handleNavigationKeys(msg)
		return m, m.handleActionKeys(msg)
	}

	return m, nil
}

// handleQuitKeys processes quit and confirm keys.
func (m *bubbleModel) handleQuitKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return tea.Quit
	case "q", "esc":
		if !m.viewModal {
			m.quitting = true
			return tea.Quit
		}
	case "enter":
		if !m.viewModal {
			m.quitting = true
			m.confirmed = true
			return tea.Quit
		}
	}
	return nil
}

// handleNavigationKeys processes arrow keys for cursor movement.
func (m *bubbleModel) handleNavigationKeys(msg tea.KeyMsg) {
	switch msg.String() {
	case "up":
		m.moveToPreviousRow()
		m.updateViewport()
	case "down":
		m.moveToNextRow()
		m.updateViewport()
	case "left":
		m.moveToPreviousColumn()
		m.updateViewport()
	case "right":
		m.moveToNextColumn()
		m.updateViewport()
	}
}

// handleActionKeys processes action keys (select, ignore, view).
func (m *bubbleModel) handleActionKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case " ":
		if m.selected[m.cursor] {
			delete(m.selected, m.cursor)
		} else {
			m.selected[m.cursor] = true
		}
	case "a", "A":
		for i := range m.items {
			m.selected[i] = true
		}
	case "n", "N":
		m.selected = make(map[int]bool)
	case "i", "I":
		return m.ignoreItem(m.cursor)
	case "v", "V":
		return m.viewItem(m.cursor)
	}
	return nil
}

// handleMouseMsg processes mouse input.
func (m bubbleModel) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if debugLog != nil {
		debugLog.Printf("Mouse event: Type=%v, Action=%v, X=%d, Y=%d", msg.Type, msg.Action, msg.X, msg.Y)
	}

	// Ignore mouse events when in modal
	if m.viewModal {
		return m, nil
	}

	switch msg.Type {
	case tea.MouseLeft:
		return m.handleLeftClick(msg)
	case tea.MouseRight:
		return m.handleRightClick(msg)
	case tea.MouseWheelUp:
		return m.handleWheelUp()
	case tea.MouseWheelDown:
		return m.handleWheelDown()
	}

	return m, nil
}

// handleLeftClick processes left mouse button clicks.
func (m bubbleModel) handleLeftClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// ONLY handle the actual button press, not motion with button held
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}

	if debugLog != nil {
		debugLog.Printf("Left click detected at X=%d, Y=%d", msg.X, msg.Y)
	}

	// Left click: move cursor and toggle selection
	if idx := m.getItemIndexFromMouse(msg.X, msg.Y); idx >= 0 {
		m.cursor = idx
		m.updateViewport()
		// Toggle selection
		if m.selected[m.cursor] {
			delete(m.selected, m.cursor)
		} else {
			m.selected[m.cursor] = true
		}
	}

	return m, nil
}

// handleRightClick processes right mouse button clicks.
func (m bubbleModel) handleRightClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// ONLY handle the actual button press, not motion with button held
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}

	if debugLog != nil {
		debugLog.Printf("Right click detected at X=%d, Y=%d", msg.X, msg.Y)
	}

	// Right click: move cursor and open view modal
	if idx := m.getItemIndexFromMouse(msg.X, msg.Y); idx >= 0 {
		m.cursor = idx
		m.updateViewport()
		return m, m.viewItem(m.cursor)
	}

	return m, nil
}

// handleWheelUp processes mouse wheel up events.
func (m bubbleModel) handleWheelUp() (tea.Model, tea.Cmd) {
	if debugLog != nil {
		debugLog.Printf("Mouse wheel up event - scrolling viewport up 3 rows")
	}

	// Scroll viewport up by 3 rows (scroll the view, not just the cursor)
	numCols, _ := m.getGridLayout()
	if numCols == 0 {
		return m, nil
	}

	// Calculate how many items to scroll (3 rows * numCols)
	scrollAmount := 3 * numCols

	// Move viewport up
	newViewportTop := m.viewportTop - scrollAmount
	if newViewportTop < 0 {
		newViewportTop = 0
	}
	m.viewportTop = newViewportTop

	// Keep cursor in view - if it scrolled out, move it to the last visible row
	maxVisibleRows := m.height - 6
	if maxVisibleRows < 5 {
		maxVisibleRows = 5
	}
	viewportEnd := m.viewportTop + (maxVisibleRows * numCols)
	if viewportEnd > len(m.items) {
		viewportEnd = len(m.items)
	}

	if m.cursor >= viewportEnd {
		// Cursor is below visible area, move it to last visible item
		m.cursor = viewportEnd - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}

	if debugLog != nil {
		debugLog.Printf("Scrolled up: viewportTop=%d, cursor=%d", m.viewportTop, m.cursor)
	}

	return m, nil
}

// handleWheelDown processes mouse wheel down events.
func (m bubbleModel) handleWheelDown() (tea.Model, tea.Cmd) {
	if debugLog != nil {
		debugLog.Printf("Mouse wheel down event - scrolling viewport down 3 rows")
	}

	// Scroll viewport down by 3 rows (scroll the view, not just the cursor)
	numCols, totalRows := m.getGridLayout()
	if numCols == 0 {
		return m, nil
	}

	// Calculate how many items to scroll (3 rows * numCols)
	scrollAmount := 3 * numCols

	// Move viewport down
	maxViewportTop := len(m.items) - 1
	newViewportTop := m.viewportTop + scrollAmount
	if newViewportTop > maxViewportTop {
		// Align to row boundary
		lastRow := totalRows - 1
		newViewportTop = lastRow * numCols
		if newViewportTop < 0 {
			newViewportTop = 0
		}
	}
	m.viewportTop = newViewportTop

	// Keep cursor in view - if it scrolled out, move it to the first visible row
	if m.cursor < m.viewportTop {
		// Cursor is above visible area, move it to first visible item
		m.cursor = m.viewportTop
		if m.cursor >= len(m.items) {
			m.cursor = len(m.items) - 1
		}
	}

	if debugLog != nil {
		debugLog.Printf("Scrolled down: viewportTop=%d, cursor=%d", m.viewportTop, m.cursor)
	}

	return m, nil
}

// handleWindowSizeMsg processes terminal resize events.
func (m bubbleModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.height = msg.Height
	m.width = msg.Width
	m.updateViewport()
	return m, nil
}

// handleIgnoreStartMsg starts the ignore animation for an item.
func (m bubbleModel) handleIgnoreStartMsg(msg ignoreStartMsg) (tea.Model, tea.Cmd) {
	m.ignoring[msg.itemIdx] = true
	m.ignoreTime[msg.itemIdx] = time.Now()
	if len(m.ignoring) == 1 {
		return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return ignoreTickMsg(t)
		})
	}
	return m, nil
}

// handleIgnoreTickMsg updates the ignore animation state.
func (m bubbleModel) handleIgnoreTickMsg(msg ignoreTickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for idx, startTime := range m.ignoreTime {
		if time.Since(startTime) >= 500*time.Millisecond {
			cmds = append(cmds, func(i int) tea.Cmd {
				return func() tea.Msg {
					return ignoreCompleteMsg{itemIdx: i}
				}
			}(idx))
		}
	}
	if len(m.ignoring) > 0 {
		cmds = append(cmds, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return ignoreTickMsg(t)
		}))
	}
	return m, tea.Batch(cmds...)
}

// handleIgnoreCompleteMsg removes an ignored item from the list.
func (m bubbleModel) handleIgnoreCompleteMsg(msg ignoreCompleteMsg) (tea.Model, tea.Cmd) {
	idx := msg.itemIdx
	delete(m.ignoring, idx)
	delete(m.ignoreTime, idx)

	newItems := make([]string, 0, len(m.items)-1)
	newCandidates := make([]DotfileCandidate, 0, len(m.candidates)-1)
	for i := 0; i < len(m.items); i++ {
		if i != idx {
			newItems = append(newItems, m.items[i])
			newCandidates = append(newCandidates, m.candidates[i])
		}
	}
	m.items = newItems
	m.candidates = newCandidates

	newSelected := make(map[int]bool)
	for i, sel := range m.selected {
		if i < idx {
			newSelected[i] = sel
		} else if i > idx {
			newSelected[i-1] = sel
		}
	}
	m.selected = newSelected

	if m.cursor >= len(m.items) && len(m.items) > 0 {
		m.cursor = len(m.items) - 1
	}
	m.updateViewport()
	return m, nil
}

// handleViewContentMsg displays the view modal with content.
func (m bubbleModel) handleViewContentMsg(msg viewContentMsg) (tea.Model, tea.Cmd) {
	m.viewContent = msg.content
	m.viewModal = true
	return m, nil
}

// ignoreItem marks an item for ignoring and adds it to the .dotignore file.
func (m bubbleModel) ignoreItem(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.candidates) {
		return nil
	}

	candidate := m.candidates[idx]
	// Use basename as pattern (e.g., ".ollama")
	pattern := filepath.Base(candidate.Path)

	// Capture the index for the closure
	itemIdx := idx

	return func() tea.Msg {
		ctx := context.Background()
		if err := AppendToGlobalDotignore(ctx, m.fs, m.configDir, pattern); err != nil {
			// Silently fail for now - in a real implementation, we'd show an error
			return nil
		}

		// Return message to start the ignore animation
		return ignoreStartMsg{itemIdx: itemIdx}
	}
}

// viewItem loads and displays the item's details in a modal.
func (m bubbleModel) viewItem(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.candidates) {
		return nil
	}

	candidate := m.candidates[idx]

	return func() tea.Msg {
		ctx := context.Background()
		content := m.buildViewContent(ctx, candidate)
		return viewContentMsg{content: content}
	}
}

// buildViewContent creates the content string for the view modal.
func (m bubbleModel) buildViewContent(ctx context.Context, candidate DotfileCandidate) string {
	var b strings.Builder

	// Header with file info
	b.WriteString(fmt.Sprintf("Path: %s\n", candidate.Path))
	b.WriteString(fmt.Sprintf("Type: %s\n", map[bool]string{true: "Directory", false: "File"}[candidate.IsDir]))
	b.WriteString(fmt.Sprintf("Size: %s\n", formatSize(candidate.Size)))
	b.WriteString(fmt.Sprintf("Modified: %s\n", candidate.ModTime.Format("2006-01-02 15:04:05")))
	b.WriteString("\n")

	// Content
	if candidate.IsDir {
		// List directory contents
		entries, err := m.fs.ReadDir(ctx, candidate.Path)
		if err != nil {
			b.WriteString(fmt.Sprintf("Error reading directory: %v\n", err))
		} else {
			b.WriteString(fmt.Sprintf("Contents (%d items):\n", len(entries)))
			b.WriteString("\n")

			// Limit to first 50 entries
			maxEntries := len(entries)
			if maxEntries > 50 {
				maxEntries = 50
			}

			for i := 0; i < maxEntries; i++ {
				entry := entries[i]
				info, err := m.fs.Stat(ctx, filepath.Join(candidate.Path, entry.Name()))
				if err != nil {
					continue
				}

				typeIndicator := ""
				if entry.IsDir() {
					typeIndicator = "/"
				}

				b.WriteString(fmt.Sprintf("  %-40s %10s\n",
					entry.Name()+typeIndicator,
					formatSize(info.Size()),
				))
			}

			if len(entries) > 50 {
				b.WriteString(fmt.Sprintf("\n  ... and %d more items\n", len(entries)-50))
			}
		}
	} else {
		// Show file preview (first 50 lines)
		content, err := m.fs.ReadFile(ctx, candidate.Path)
		if err != nil {
			b.WriteString(fmt.Sprintf("Error reading file: %v\n", err))
		} else {
			// Check if file is binary
			if isBinaryContent(content) {
				b.WriteString("Binary file (cannot preview)\n")
				b.WriteString("\n")
				b.WriteString("File appears to be binary and cannot be displayed as text.\n")
				b.WriteString("Common binary file types include:\n")
				b.WriteString("  - Executables (.exe, .dll, .so)\n")
				b.WriteString("  - Images (.png, .jpg, .gif)\n")
				b.WriteString("  - Archives (.zip, .tar, .gz)\n")
				b.WriteString("  - Compiled code (.pyc, .o, .a)\n")
			} else {
				// Apply syntax highlighting based on file extension
				highlighted := m.highlightContent(candidate.Path, content)
				lines := strings.Split(highlighted, "\n")

				// Smart preview message
				totalLines := len(lines)
				maxLines := totalLines
				if maxLines > 50 {
					maxLines = 50
				}

				if totalLines <= 50 {
					b.WriteString(fmt.Sprintf("Preview (%d lines):\n", totalLines))
				} else {
					b.WriteString(fmt.Sprintf("Preview (first %d of %d lines):\n", maxLines, totalLines))
				}
				b.WriteString("\n")

				for i := 0; i < maxLines; i++ {
					// Truncate long lines (accounting for ANSI codes)
					line := lines[i]
					visualLen := len(stripANSI(line))
					if visualLen > 80 {
						// Find position to truncate (need to handle ANSI codes)
						line = truncateWithANSI(line, 77) + "..."
					}
					b.WriteString(fmt.Sprintf("%4d | %s\n", i+1, line))
				}

				if totalLines > 50 {
					b.WriteString(fmt.Sprintf("\n... and %d more lines\n", totalLines-50))
				}
			}
		}
	}

	return b.String()
}

// isBinaryContent checks if the content appears to be binary.
// Returns true if the content contains null bytes or has high ratio of non-printable characters.
func isBinaryContent(content []byte) bool {
	// Check first 8KB (or entire file if smaller)
	sampleSize := 8192
	if len(content) < sampleSize {
		sampleSize = len(content)
	}

	if sampleSize == 0 {
		return false
	}

	sample := content[:sampleSize]

	// Check for null bytes (very strong indicator of binary content)
	for _, b := range sample {
		if b == 0 {
			return true
		}
	}

	// Count non-printable characters
	nonPrintable := 0
	for _, b := range sample {
		// Allow common text characters: printable ASCII, tabs, newlines, carriage returns
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			nonPrintable++
		} else if b > 126 && b < 128 {
			// DEL and other control characters
			nonPrintable++
		}
	}

	// If more than 30% non-printable, consider it binary
	ratio := float64(nonPrintable) / float64(sampleSize)
	return ratio > 0.30
}

// highlightContent applies syntax highlighting to file content based on file extension.
func (m bubbleModel) highlightContent(filePath string, content []byte) string {
	var highlighted strings.Builder

	// Use chroma to highlight with terminal256 formatter
	err := quick.Highlight(&highlighted, string(content), filepath.Ext(filePath), "terminal256", "monokai")
	if err != nil {
		// If highlighting fails, return plain content
		return string(content)
	}

	return highlighted.String()
}

// stripANSI removes ANSI escape codes from a string to get visual length.
func stripANSI(s string) string {
	// Simple ANSI stripper - matches ESC [ ... m
	inEscape := false
	var result strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // Skip the '['
			continue
		}

		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}

		result.WriteByte(s[i])
	}

	return result.String()
}

// truncateWithANSI truncates a string containing ANSI codes to a visual width.
func truncateWithANSI(s string, maxVisualLen int) string {
	visualLen := 0
	var result strings.Builder
	inEscape := false

	for i := 0; i < len(s); i++ {
		// Detect start of ANSI sequence
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			result.WriteByte(s[i])
			continue
		}

		// Copy ANSI codes without counting them
		if inEscape {
			result.WriteByte(s[i])
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}

		// Count and copy visible characters
		if visualLen >= maxVisualLen {
			break
		}

		result.WriteByte(s[i])
		visualLen++
	}

	return result.String()
}

// getGridLayout calculates the grid layout parameters.
// Returns (numCols, totalRows) for row-major layout.
func (m *bubbleModel) getGridLayout() (numCols, totalRows int) {
	if len(m.items) == 0 {
		return 1, 0
	}

	maxItemLen := m.getMaxItemLength()
	colWidth := 2 + 4 + 1 + maxItemLen + 2
	numCols = m.width / colWidth
	if numCols < 1 {
		numCols = 1
	}
	if numCols > 4 {
		numCols = 4
	}

	totalItems := len(m.items)
	totalRows = (totalItems + numCols - 1) / numCols
	return numCols, totalRows
}

// getItemIndexFromMouse converts mouse coordinates to item index.
// Returns -1 if the mouse is not over a valid item.
func (m *bubbleModel) getItemIndexFromMouse(mouseX, mouseY int) int {
	// Header takes 3 lines (title + separator + blank)
	headerLines := 3

	// Check if click is in the content area (after header, before footer)
	if mouseY < headerLines {
		return -1 // Clicked in header
	}

	// Calculate which visual row was clicked (0-based, relative to viewport)
	visualRow := mouseY - headerLines

	// Calculate viewport row (actual row in the full grid)
	numCols, _ := m.getGridLayout()
	if numCols == 0 {
		return -1
	}

	viewportTopRow := m.viewportTop / numCols
	actualRow := viewportTopRow + visualRow

	// Calculate which column was clicked
	maxItemLen := m.getMaxItemLength()
	colWidth := 2 + 4 + 1 + maxItemLen + 2

	col := mouseX / colWidth
	if col >= numCols {
		col = numCols - 1
	}
	if col < 0 {
		col = 0
	}

	// Calculate item index using row-major layout
	idx := (actualRow * numCols) + col

	if debugLog != nil {
		debugLog.Printf("getItemIndexFromMouse: mouseX=%d, mouseY=%d, visualRow=%d, actualRow=%d, col=%d, idx=%d, valid=%v",
			mouseX, mouseY, visualRow, actualRow, col, idx, idx >= 0 && idx < len(m.items))
	}

	// Validate index
	if idx < 0 || idx >= len(m.items) {
		return -1
	}

	return idx
}

// moveToPreviousColumn moves the cursor to the previous column (staying on same visual row).
// Algorithm: Row-major layout: idx = (row * numCols) + col
// We reverse this to find current position, then move to previous column.
func (m *bubbleModel) moveToPreviousColumn() {
	if len(m.items) == 0 {
		return
	}

	numCols, _ := m.getGridLayout()
	if numCols == 0 {
		return
	}

	// Current position in the grid (row-major)
	// idx = (row * numCols) + col
	// So: row = idx / numCols, col = idx % numCols
	currentRow := m.cursor / numCols
	currentCol := m.cursor % numCols

	if debugLog != nil {
		debugLog.Printf("LEFT: cursor=%d, totalItems=%d, numCols=%d",
			m.cursor, len(m.items), numCols)
		debugLog.Printf("LEFT: currentRow=%d, currentCol=%d", currentRow, currentCol)
	}

	// Try to move to previous column (same row)
	if currentCol > 0 {
		targetIdx := m.cursor - 1

		if debugLog != nil {
			debugLog.Printf("LEFT: targetIdx=%d (exists=%v)",
				targetIdx, targetIdx >= 0 && targetIdx < len(m.items))
		}

		// Ensure target exists
		if targetIdx >= 0 && targetIdx < len(m.items) {
			m.cursor = targetIdx
			if debugLog != nil {
				debugLog.Printf("LEFT: moved to cursor=%d", m.cursor)
			}
		}
	} else if debugLog != nil {
		debugLog.Printf("LEFT: already in first column, cannot move left")
	}
}

// moveToNextColumn moves the cursor to the next column (staying on same visual row).
// Algorithm: Row-major layout: idx = (row * numCols) + col
// We reverse this to find current position, then move to next column.
func (m *bubbleModel) moveToNextColumn() {
	if len(m.items) == 0 {
		return
	}

	numCols, _ := m.getGridLayout()
	if numCols == 0 {
		return
	}

	// Current position in the grid (row-major)
	// idx = (row * numCols) + col
	// So: row = idx / numCols, col = idx % numCols
	currentRow := m.cursor / numCols
	currentCol := m.cursor % numCols

	if debugLog != nil {
		debugLog.Printf("RIGHT: cursor=%d, totalItems=%d, numCols=%d",
			m.cursor, len(m.items), numCols)
		debugLog.Printf("RIGHT: currentRow=%d, currentCol=%d", currentRow, currentCol)
	}

	// Try to move to next column (same row)
	if currentCol < numCols-1 {
		targetIdx := m.cursor + 1

		if debugLog != nil {
			debugLog.Printf("RIGHT: targetIdx=%d (exists=%v)",
				targetIdx, targetIdx >= 0 && targetIdx < len(m.items))
		}

		// Ensure target exists (important for last row which may be incomplete)
		if targetIdx >= 0 && targetIdx < len(m.items) {
			m.cursor = targetIdx
			if debugLog != nil {
				debugLog.Printf("RIGHT: moved to cursor=%d", m.cursor)
			}
		} else if debugLog != nil {
			debugLog.Printf("RIGHT: target doesn't exist (incomplete last row)")
		}
	} else if debugLog != nil {
		debugLog.Printf("RIGHT: already in last column, cannot move right")
	}
}

// moveToPreviousRow moves the cursor up one row (staying in same column).
// Algorithm: Row-major layout: idx = (row * numCols) + col
// To move up: subtract numCols from current index.
func (m *bubbleModel) moveToPreviousRow() {
	if len(m.items) == 0 {
		return
	}

	numCols, _ := m.getGridLayout()
	if numCols == 0 {
		return
	}

	// Current position in the grid (row-major)
	currentRow := m.cursor / numCols
	currentCol := m.cursor % numCols

	if debugLog != nil {
		debugLog.Printf("UP: cursor=%d, totalItems=%d, numCols=%d",
			m.cursor, len(m.items), numCols)
		debugLog.Printf("UP: currentRow=%d, currentCol=%d", currentRow, currentCol)
	}

	// Try to move up one row
	if currentRow > 0 {
		targetIdx := m.cursor - numCols

		if debugLog != nil {
			debugLog.Printf("UP: targetIdx=%d (exists=%v)",
				targetIdx, targetIdx >= 0 && targetIdx < len(m.items))
		}

		// Move up
		if targetIdx >= 0 {
			m.cursor = targetIdx
			if debugLog != nil {
				debugLog.Printf("UP: moved to cursor=%d", m.cursor)
			}
		}
	} else if debugLog != nil {
		debugLog.Printf("UP: already in first row, cannot move up")
	}
}

// moveToNextRow moves the cursor down one row (staying in same column).
// Algorithm: Row-major layout: idx = (row * numCols) + col
// To move down: add numCols to current index.
func (m *bubbleModel) moveToNextRow() {
	if len(m.items) == 0 {
		return
	}

	numCols, totalRows := m.getGridLayout()
	if numCols == 0 {
		return
	}

	// Current position in the grid (row-major)
	currentRow := m.cursor / numCols
	currentCol := m.cursor % numCols

	if debugLog != nil {
		debugLog.Printf("DOWN: cursor=%d, totalItems=%d, numCols=%d, totalRows=%d",
			m.cursor, len(m.items), numCols, totalRows)
		debugLog.Printf("DOWN: currentRow=%d, currentCol=%d", currentRow, currentCol)
	}

	// Try to move down one row
	if currentRow < totalRows-1 {
		targetIdx := m.cursor + numCols

		if debugLog != nil {
			debugLog.Printf("DOWN: targetIdx=%d (exists=%v)",
				targetIdx, targetIdx >= 0 && targetIdx < len(m.items))
		}

		// Ensure target exists (last row might be incomplete)
		if targetIdx < len(m.items) {
			m.cursor = targetIdx
			if debugLog != nil {
				debugLog.Printf("DOWN: moved to cursor=%d", m.cursor)
			}
		} else if debugLog != nil {
			debugLog.Printf("DOWN: target doesn't exist (incomplete last row)")
		}
	} else if debugLog != nil {
		debugLog.Printf("DOWN: already in last row, cannot move down")
	}
}

// updateViewport adjusts the viewport to keep the cursor visible.
func (m *bubbleModel) updateViewport() {
	if len(m.items) == 0 {
		return
	}

	// Reserve space for header (3 lines) and footer (3 lines)
	maxVisibleRows := m.height - 6
	if maxVisibleRows < 5 {
		maxVisibleRows = 5
	}

	numCols, _ := m.getGridLayout()
	if numCols == 0 {
		return
	}

	// Calculate which row the cursor is on
	cursorRow := m.cursor / numCols

	// Calculate which row the viewportTop is on
	viewportTopRow := m.viewportTop / numCols

	// Calculate viewport bounds in terms of rows
	viewportBottomRow := viewportTopRow + maxVisibleRows - 1

	if debugLog != nil {
		debugLog.Printf("updateViewport: cursor=%d, cursorRow=%d, viewportTopRow=%d, viewportBottomRow=%d, maxVisibleRows=%d",
			m.cursor, cursorRow, viewportTopRow, viewportBottomRow, maxVisibleRows)
	}

	// Scroll up if cursor is above viewport
	if cursorRow < viewportTopRow {
		m.viewportTop = cursorRow * numCols
		if debugLog != nil {
			debugLog.Printf("updateViewport: scrolled up, viewportTop=%d", m.viewportTop)
		}
	} else if cursorRow > viewportBottomRow {
		// Scroll down if cursor is below viewport
		newTopRow := cursorRow - maxVisibleRows + 1
		if newTopRow < 0 {
			newTopRow = 0
		}
		m.viewportTop = newTopRow * numCols
		if debugLog != nil {
			debugLog.Printf("updateViewport: scrolled down, viewportTop=%d", m.viewportTop)
		}
	}
}

// View renders the UI.
func (m bubbleModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	styles := m.getStyles()
	separatorWidth := m.getSeparatorWidth()

	// Header
	m.renderHeader(&b, styles, separatorWidth)

	// Items
	m.renderItems(&b, styles)

	// Footer
	m.renderFooter(&b, styles, separatorWidth)

	// Overlay view modal if active
	if m.viewModal {
		return m.renderViewModal(b.String(), styles)
	}

	return b.String()
}

// viewStyles holds the lipgloss styles for rendering.
type viewStyles struct {
	header      lipgloss.Style
	cursor      lipgloss.Style
	selected    lipgloss.Style
	dim         lipgloss.Style
	instruction lipgloss.Style
	ignoring    lipgloss.Style
	modal       lipgloss.Style
	modalBorder lipgloss.Style
	highlight   lipgloss.Style
}

// getStyles returns the lipgloss styles for rendering.
func (m bubbleModel) getStyles() viewStyles {
	return viewStyles{
		header:      lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Bold(true),
		cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Bold(true),
		selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		dim:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		instruction: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		ignoring:    lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("109")).
			Padding(1, 2).
			Width(80),
		modalBorder: lipgloss.NewStyle().Foreground(lipgloss.Color("109")),
		highlight:   lipgloss.NewStyle().Background(lipgloss.Color("235")),
	}
}

// getSeparatorWidth returns the width for separators.
func (m bubbleModel) getSeparatorWidth() int {
	width := m.width
	if width < 40 {
		width = 40
	}
	return width
}

// renderHeader renders the header section.
func (m bubbleModel) renderHeader(b *strings.Builder, styles viewStyles, separatorWidth int) {
	title := fmt.Sprintf("Select Dotfiles (%d/%d selected)", len(m.selected), len(m.items))
	b.WriteString(styles.header.Render(title))
	b.WriteString("\n")
	b.WriteString(styles.dim.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n\n")
}

// renderFooter renders the footer section.
func (m bubbleModel) renderFooter(b *strings.Builder, styles viewStyles, separatorWidth int) {
	b.WriteString("\n")
	b.WriteString(styles.dim.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n")
	b.WriteString(styles.instruction.Render("↑↓←→/mouse: navigate | Click/space: toggle | Right-click/v: view | i: ignore | a: all | n: none | Enter: confirm | q: cancel"))
}

// renderItems renders the items in columns.
func (m bubbleModel) renderItems(b *strings.Builder, styles viewStyles) {
	maxVisibleRows := m.height - 6
	if maxVisibleRows < 5 {
		maxVisibleRows = 5
	}

	// Calculate column layout
	maxItemLen := m.getMaxItemLength()
	colWidth := 2 + 4 + 1 + maxItemLen + 2
	numCols := m.width / colWidth
	if numCols < 1 {
		numCols = 1
	}
	if numCols > 4 {
		numCols = 4
	}

	// Calculate viewport end based on rows and columns
	// viewportTop is the first item index to show
	// Show maxVisibleRows rows, each with numCols items
	viewportEnd := m.viewportTop + (maxVisibleRows * numCols)
	if viewportEnd > len(m.items) {
		viewportEnd = len(m.items)
	}

	// Calculate which row viewportTop is on
	viewportTopRow := m.viewportTop / numCols

	// Calculate how many rows to actually render
	itemsInView := viewportEnd - m.viewportTop
	rowsNeeded := (itemsInView + numCols - 1) / numCols

	if debugLog != nil {
		debugLog.Printf("renderItems: viewportTop=%d, viewportEnd=%d, viewportTopRow=%d, rowsNeeded=%d, numCols=%d",
			m.viewportTop, viewportEnd, viewportTopRow, rowsNeeded, numCols)
	}

	for row := 0; row < rowsNeeded; row++ {
		m.renderRow(b, styles, row, numCols, rowsNeeded, viewportEnd, colWidth)
		b.WriteString("\n")
	}
}

// getMaxItemLength returns the visual length of the longest item (rune count).
func (m bubbleModel) getMaxItemLength() int {
	maxLen := 0
	for _, item := range m.items {
		runeLen := len([]rune(item))
		if runeLen > maxLen {
			maxLen = runeLen
		}
	}
	return maxLen
}

// renderViewModal renders a modal overlay with content.
func (m bubbleModel) renderViewModal(baseView string, styles viewStyles) string {
	// Limit content height to fit on screen
	maxHeight := m.height - 8 // Reserve space for borders and instructions
	if maxHeight < 10 {
		maxHeight = 10
	}

	// Split content into lines and truncate if needed
	contentLines := strings.Split(m.viewContent, "\n")
	if len(contentLines) > maxHeight {
		contentLines = contentLines[:maxHeight]
		contentLines = append(contentLines, "... (content truncated)")
	}

	// Rebuild content
	truncatedContent := strings.Join(contentLines, "\n")

	// Create modal with fixed width
	modalWidth := m.width - 10
	if modalWidth > 100 {
		modalWidth = 100
	}
	if modalWidth < 60 {
		modalWidth = 60
	}

	// Create the modal style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("109")).
		Padding(1, 2).
		Width(modalWidth).
		MaxHeight(maxHeight + 2) // Account for padding

	// Render the modal content
	modal := modalStyle.Render(truncatedContent)

	// Use lipgloss Place to center the modal
	centered := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)

	// Add instruction at bottom
	instruction := styles.instruction.Render("Press ESC to close")
	instructionCentered := lipgloss.Place(
		m.width,
		1,
		lipgloss.Center,
		lipgloss.Top,
		instruction,
	)

	// Combine centered modal with instruction at bottom
	lines := strings.Split(centered, "\n")
	if len(lines) > 0 {
		// Replace last line with instruction
		lines[len(lines)-1] = instructionCentered
	}

	return strings.Join(lines, "\n")
}

// renderRow renders a single row of items.
func (m bubbleModel) renderRow(b *strings.Builder, styles viewStyles, row, numCols, rowsNeeded, viewportEnd, colWidth int) {
	for col := 0; col < numCols; col++ {
		// Use row-major layout: items go left-to-right, then down
		idx := m.viewportTop + (row * numCols) + col
		if idx >= viewportEnd {
			continue
		}

		isCursor := idx == m.cursor

		// Get components - if this is the cursor row, apply highlight background to styles
		var prefix, checkbox, itemText string
		var prefixPlain, checkboxPlain string

		if isCursor {
			// Apply highlight background to all components
			cursorStyle := styles.cursor.Copy().Background(lipgloss.Color("235"))
			selectedStyle := styles.selected.Copy().Background(lipgloss.Color("235"))
			ignoringStyle := styles.ignoring.Copy().Background(lipgloss.Color("235"))
			normalStyle := lipgloss.NewStyle().Background(lipgloss.Color("235"))

			// Prefix with highlight (always cursor for highlighted row)
			prefix = cursorStyle.Render("❯ ")
			prefixPlain = "❯ "

			// Checkbox with highlight
			if m.selected[idx] {
				checkbox = selectedStyle.Render("[✓]")
				checkboxPlain = "[✓]"
			} else {
				checkbox = normalStyle.Render("[ ]")
				checkboxPlain = "[ ]"
			}

			// Item text with highlight
			if m.ignoring[idx] {
				itemText = ignoringStyle.Render(m.items[idx])
			} else {
				itemText = normalStyle.Render(m.items[idx])
			}
		} else {
			// No highlight
			prefix, prefixPlain = m.getPrefix(idx, styles)
			checkbox, checkboxPlain = m.getCheckbox(idx, styles)

			// Apply grey style to item text if ignoring
			itemText = m.items[idx]
			if m.ignoring[idx] {
				itemText = styles.ignoring.Render(itemText)
			}
		}

		// Build the full item text
		fullItemText := fmt.Sprintf("%s %s %s", prefix, checkbox, itemText)

		// Calculate visual width using rune count (handles Unicode properly)
		visualWidth := len([]rune(prefixPlain)) + 1 + len([]rune(checkboxPlain)) + 1 + len([]rune(m.items[idx]))

		// Add padding
		if col < numCols-1 {
			padding := colWidth - visualWidth
			if padding > 0 {
				if isCursor {
					// Add highlighted padding
					fullItemText += styles.highlight.Render(strings.Repeat(" ", padding))
				} else {
					fullItemText += strings.Repeat(" ", padding)
				}
			}
		}

		b.WriteString(fullItemText)
	}
}

// getPrefix returns the styled and plain prefix for an item.
func (m bubbleModel) getPrefix(idx int, styles viewStyles) (string, string) {
	if idx == m.cursor {
		return styles.cursor.Render("❯ "), "❯ "
	}
	return "  ", "  "
}

// getCheckbox returns the styled and plain checkbox for an item.
func (m bubbleModel) getCheckbox(idx int, styles viewStyles) (string, string) {
	if m.selected[idx] {
		return styles.selected.Render("[✓]"), "[✓]"
	}
	return "[ ]", "[ ]"
}

// SelectMultiple displays items and allows arrow key navigation with spacebar to toggle selection.
// Returns indices of selected items.
func (s *ArrowSelector) SelectMultiple(items []string, candidates []DotfileCandidate) ([]int, error) {
	if len(items) == 0 {
		return []int{}, nil
	}

	m := bubbleModel{
		items:      items,
		candidates: candidates,
		selected:   make(map[int]bool),
		ignoring:   make(map[int]bool),
		ignoreTime: make(map[int]time.Time),
		height:     24, // Default, will be updated by WindowSizeMsg
		width:      80, // Default, will be updated by WindowSizeMsg
		fs:         s.fs,
		configDir:  s.configDir,
	}

	// Use tea.WithAltScreen() for proper alternate screen buffer handling
	// Use tea.WithInput() to use custom input reader
	// Use tea.WithMouseCellMotion() for mouse support
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}

	// Only set custom input if it's not stdin (for testing)
	if s.input != nil {
		opts = append(opts, tea.WithInput(s.input))
	}

	p := tea.NewProgram(m, opts...)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running selector: %w", err)
	}

	// Extract selected indices
	final := finalModel.(bubbleModel)

	// If user quit without confirming, return empty
	if !final.confirmed {
		return []int{}, nil
	}

	indices := make([]int, 0, len(final.selected))
	for idx := range final.selected {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	return indices, nil
}
