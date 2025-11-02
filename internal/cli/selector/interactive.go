// Package selector provides package selection interfaces and implementations.
package selector

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// PackageSelector defines the interface for selecting packages.
type PackageSelector interface {
	// Select prompts the user to select packages from the provided list.
	// Returns the selected package names.
	Select(ctx context.Context, packages []string) ([]string, error)
}

// InteractiveSelector implements PackageSelector with interactive prompts.
type InteractiveSelector struct {
	input  io.Reader
	output io.Writer
}

// NewInteractiveSelector creates a new interactive selector.
func NewInteractiveSelector(input io.Reader, output io.Writer) *InteractiveSelector {
	return &InteractiveSelector{
		input:  input,
		output: output,
	}
}

// Select prompts the user to select packages interactively.
func (s *InteractiveSelector) Select(ctx context.Context, packages []string) ([]string, error) {
	// Handle empty package list
	if len(packages) == 0 {
		return []string{}, nil
	}

	// Get terminal width for layout
	termWidth := getTerminalWidth()

	// Calculate content width to make separators match column width
	contentWidth := calculateContentWidth(packages, termWidth)

	// Define color styles
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Bold(true) // Muted blue, bold
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))         // Dark gray
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))             // Gray
	instructionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))       // Gray
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Bold(true) // Muted cyan, bold

	// Display header
	fmt.Fprintln(s.output, headerStyle.Render("Package Selection"))
	fmt.Fprintln(s.output, separatorStyle.Render(strings.Repeat("─", contentWidth)))
	fmt.Fprintf(s.output, "%s\n\n", countStyle.Render(fmt.Sprintf("%d packages available", len(packages))))

	// Display packages in columns
	formatted := formatPackagesMultiColumn(packages, termWidth)
	fmt.Fprint(s.output, formatted)

	// Display footer
	fmt.Fprintln(s.output, "")
	fmt.Fprintln(s.output, separatorStyle.Render(strings.Repeat("─", contentWidth)))
	fmt.Fprintln(s.output, instructionStyle.Render("Select: numbers (1,2,3), ranges (1-5), all, none"))
	fmt.Fprintln(s.output, "")
	fmt.Fprint(s.output, promptStyle.Render("❯")+" ")

	// Read user input
	scanner := bufio.NewScanner(s.input)
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Read input
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("read input: %w", err)
			}
			return nil, fmt.Errorf("unexpected end of input")
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Bold(true)
			fmt.Fprint(s.output, promptStyle.Render("❯")+" ")
			continue
		}

		// Parse selection
		indices, err := parseSelection(input, len(packages))
		if err != nil {
			warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("179")) // Muted gold
			promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Bold(true)
			fmt.Fprintf(s.output, "\n%s Invalid selection: %v\n", warningStyle.Render("⚠"), err)
			fmt.Fprint(s.output, promptStyle.Render("❯")+" ")
			continue
		}

		// Build selected package list
		selected := make([]string, 0, len(indices))
		for _, idx := range indices {
			selected = append(selected, packages[idx])
		}

		return selected, nil
	}
}

// parseSelection parses user input into package indices.
//
// Supported formats:
//   - "1" - single number
//   - "1,3,5" - comma-separated numbers
//   - "1-3" - range
//   - "1, 3-5, 7" - mixed
//   - "all" - all packages
//   - "none" - no packages
//
// Returns zero-based indices.
func parseSelection(input string, maxIndex int) ([]int, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	// Handle special keywords
	if input == "all" {
		indices := make([]int, maxIndex)
		for i := range indices {
			indices[i] = i
		}
		return indices, nil
	}

	if input == "none" {
		return []int{}, nil
	}

	// Parse comma-separated parts
	parts := strings.Split(input, ",")
	indices := make(map[int]bool) // Use map to deduplicate

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check if it's a range
		if strings.Contains(part, "-") {
			rangeIndices, err := parseRange(part, maxIndex)
			if err != nil {
				return nil, err
			}
			for _, idx := range rangeIndices {
				indices[idx] = true
			}
			continue
		}

		// Parse single number
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", part)
		}

		if num < 1 || num > maxIndex {
			return nil, fmt.Errorf("number %d out of range (1-%d)", num, maxIndex)
		}

		indices[num-1] = true // Convert to zero-based index
	}

	// Convert map to sorted slice
	result := make([]int, 0, len(indices))
	for idx := range indices {
		result = append(result, idx)
	}
	sort.Ints(result)

	return result, nil
}

// parseRange parses a range like "1-3" into indices.
func parseRange(rangeStr string, maxIndex int) ([]int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid range start: %s", parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid range end: %s", parts[1])
	}

	if start < 1 || start > maxIndex {
		return nil, fmt.Errorf("range start %d out of range (1-%d)", start, maxIndex)
	}

	if end < 1 || end > maxIndex {
		return nil, fmt.Errorf("range end %d out of range (1-%d)", end, maxIndex)
	}

	if start > end {
		return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
	}

	indices := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		indices = append(indices, i-1) // Convert to zero-based index
	}

	return indices, nil
}

// getTerminalWidth returns the width of the terminal.
// Returns 80 as a default fallback if detection fails.
func getTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	width, _, err := term.GetSize(fd)
	if err != nil || width == 0 {
		return 80 // Default fallback
	}
	return width
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatPackagesMultiColumn formats packages in a multi-column grid layout.
// Returns a formatted string with packages arranged in columns for better readability.
func formatPackagesMultiColumn(packages []string, termWidth int) string {
	if len(packages) == 0 {
		return ""
	}

	// Calculate number width (e.g., "  1", " 23", "123")
	numWidth := len(fmt.Sprintf("%d", len(packages)))
	if numWidth < 2 {
		numWidth = 2 // Minimum 2 for alignment
	}

	// Find maximum package name length
	maxNameLen := 0
	for _, pkg := range packages {
		if len(pkg) > maxNameLen {
			maxNameLen = len(pkg)
		}
	}

	// Calculate entry width: number + 2 spaces + package name + 2 spaces padding
	entryWidth := numWidth + 2 + maxNameLen + 2

	// Calculate number of columns (minimum 1, maximum 4)
	numCols := termWidth / entryWidth
	if numCols < 1 {
		numCols = 1
	}
	if numCols > 4 {
		numCols = 4
	}

	// For very narrow terminals or long package names, use single column
	if entryWidth > termWidth-10 {
		numCols = 1
	}

	// Calculate number of rows needed
	numRows := (len(packages) + numCols - 1) / numCols

	// Define color styles
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("109"))  // Muted cyan
	packageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")) // Light gray/white

	var result strings.Builder
	for row := 0; row < numRows; row++ {
		for col := 0; col < numCols; col++ {
			idx := row*numCols + col
			if idx >= len(packages) {
				break
			}

			// Format: colored right-aligned number + spaces + colored package name
			number := idx + 1
			numberStr := fmt.Sprintf("%*d", numWidth, number)
			packageStr := fmt.Sprintf("%-*s", maxNameLen, packages[idx])

			result.WriteString(numberStyle.Render(numberStr))
			result.WriteString("  ")
			result.WriteString(packageStyle.Render(packageStr))

			// Add spacing between columns (but not after last column)
			if col < numCols-1 && idx < len(packages)-1 {
				result.WriteString("  ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// calculateContentWidth calculates the actual width of the multi-column content.
func calculateContentWidth(packages []string, termWidth int) int {
	if len(packages) == 0 {
		return 0
	}

	numWidth := len(fmt.Sprintf("%d", len(packages)))
	if numWidth < 2 {
		numWidth = 2
	}

	maxNameLen := 0
	for _, pkg := range packages {
		if len(pkg) > maxNameLen {
			maxNameLen = len(pkg)
		}
	}

	entryWidth := numWidth + 2 + maxNameLen + 2

	numCols := termWidth / entryWidth
	if numCols < 1 {
		numCols = 1
	}
	if numCols > 4 {
		numCols = 4
	}

	if entryWidth > termWidth-10 {
		numCols = 1
	}

	// Calculate actual content width: (entry width * cols) + (spacing between cols)
	contentWidth := (numWidth+2+maxNameLen)*numCols + (2 * (numCols - 1))

	// Don't exceed terminal width
	if contentWidth > termWidth {
		contentWidth = termWidth
	}

	return contentWidth
}
