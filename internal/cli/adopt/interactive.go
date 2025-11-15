// Package adopt provides interactive file adoption.
//
// This file contains interactive workflow logic that is tightly coupled to
// Bubble Tea UI components and cannot be reliably unit tested. It is excluded
// from coverage requirements.
package adopt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jamesainslie/dot/internal/domain"
)

// InteractiveAdopter manages the interactive adoption workflow.
type InteractiveAdopter struct {
	input      io.Reader
	output     io.Writer
	candidates []DotfileCandidate
	colorize   bool
	fs         domain.FS
	configDir  string
}

// NewInteractiveAdopter creates a new interactive adopter.
func NewInteractiveAdopter(input io.Reader, output io.Writer, colorize bool, fs domain.FS, configDir string) *InteractiveAdopter {
	return &InteractiveAdopter{
		input:     input,
		output:    output,
		colorize:  colorize,
		fs:        fs,
		configDir: configDir,
	}
}

// Run executes the interactive adoption workflow.
// Returns selected groups ready for adoption.
func (ia *InteractiveAdopter) Run(ctx context.Context, candidates []DotfileCandidate) ([]AdoptGroup, error) {
	ia.candidates = candidates

	if len(candidates) == 0 {
		fmt.Fprintln(ia.output, "No adoptable dotfiles found.")
		return nil, nil
	}

	// Step 1: Display and select files
	selectedIndices, err := ia.selectFiles(ctx)
	if err != nil {
		return nil, err
	}

	if len(selectedIndices) == 0 {
		fmt.Fprintln(ia.output, "No files selected.")
		return nil, nil
	}

	// Step 2: Group and review package names
	groups, err := ia.organizeIntoPackages(selectedIndices)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		fmt.Fprintln(ia.output, "No packages created.")
		return nil, nil
	}

	// Step 3: Preview changes (dry-run style)
	if !ia.confirmAdoption(groups) {
		fmt.Fprintln(ia.output, "Adoption cancelled.")
		return nil, nil
	}

	return groups, nil
}

// selectFiles displays candidates and prompts for selection using arrow keys.
func (ia *InteractiveAdopter) selectFiles(ctx context.Context) ([]int, error) {
	// Use arrow-key selector
	sel := NewArrowSelector(ia.input, ia.output, ia.fs, ia.configDir)

	// Format candidates as display strings
	displayItems := make([]string, len(ia.candidates))
	for i, c := range ia.candidates {
		sizeStr := formatSize(c.Size)
		typeStr := ""
		if c.IsDir {
			typeStr = " (dir)"
		}

		// Format: "~/.bashrc                 [shell]  2.3KB"
		displayItems[i] = fmt.Sprintf("%-35s [%-6s] %s%s",
			c.RelPath,
			c.Category,
			sizeStr,
			typeStr,
		)
	}

	// Get selection using arrow-key interface
	indices, err := sel.SelectMultiple(displayItems, ia.candidates)
	if err != nil {
		return nil, err
	}

	return indices, nil
}

// organizeIntoPackages groups selections and prompts for package names.
func (ia *InteractiveAdopter) organizeIntoPackages(selections []int) ([]AdoptGroup, error) {
	// Auto-group by suggested package name
	groups := GroupByCategory(ia.candidates, selections)

	// Display groups and allow editing
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Bold(true)
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("109"))

	fmt.Fprintln(ia.output, "")
	fmt.Fprintln(ia.output, headerStyle.Render("Package Organization"))
	fmt.Fprintln(ia.output, strings.Repeat("─", 60))

	finalGroups := make([]AdoptGroup, 0, len(groups))
	scanner := bufio.NewScanner(ia.input)

	// Sort group names for consistent display
	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, pkgName := range groupNames {
		candidates := groups[pkgName]

		fmt.Fprintf(ia.output, "\nPackage: %s (%d files)\n",
			headerStyle.Render(pkgName), len(candidates))

		for _, c := range candidates {
			fmt.Fprintf(ia.output, "  • %s\n", c.RelPath)
		}

		fmt.Fprint(ia.output, promptStyle.Render("❯")+" Accept package name? [Y/n/edit]: ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("read input: %w", err)
			}
			return nil, fmt.Errorf("unexpected end of input")
		}

		response := strings.TrimSpace(strings.ToLower(scanner.Text()))

		finalPkgName := pkgName
		if response == "edit" || response == "e" {
			fmt.Fprint(ia.output, promptStyle.Render("❯")+" Enter package name: ")
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return nil, fmt.Errorf("read input: %w", err)
				}
				return nil, fmt.Errorf("unexpected end of input")
			}
			finalPkgName = strings.TrimSpace(scanner.Text())
			if finalPkgName == "" {
				finalPkgName = pkgName // Keep default if empty
			}
		} else if response == "n" || response == "no" {
			continue // Skip this group
		}

		// Build group
		files := make([]string, len(candidates))
		for i, c := range candidates {
			files[i] = c.Path
		}

		finalGroups = append(finalGroups, AdoptGroup{
			PackageName: finalPkgName,
			Files:       files,
			Category:    candidates[0].Category,
		})
	}

	return finalGroups, nil
}

// confirmAdoption displays preview and confirms.
func (ia *InteractiveAdopter) confirmAdoption(groups []AdoptGroup) bool {
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Bold(true)
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("109"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("179"))

	fmt.Fprintln(ia.output, "")
	fmt.Fprintln(ia.output, headerStyle.Render("Adoption Preview"))
	fmt.Fprintln(ia.output, strings.Repeat("─", 60))

	totalFiles := 0
	for _, group := range groups {
		totalFiles += len(group.Files)
		fmt.Fprintf(ia.output, "\nPackage: %s\n", accentStyle.Render(group.PackageName))
		for _, file := range group.Files {
			baseName := filepath.Base(file)
			fmt.Fprintf(ia.output, "  %s → %s\n",
				baseName,
				accentStyle.Render(filepath.Join(group.PackageName, baseName)),
			)
		}
	}

	fmt.Fprintln(ia.output, "")
	fmt.Fprintf(ia.output, "Total: %s files into %s packages\n",
		accentStyle.Render(strconv.Itoa(totalFiles)),
		accentStyle.Render(strconv.Itoa(len(groups))),
	)
	fmt.Fprintln(ia.output, "")
	fmt.Fprintln(ia.output, warningStyle.Render("⚠")+" Files will be moved and symlinks created.")
	fmt.Fprintln(ia.output, "")

	fmt.Fprint(ia.output, "Proceed with adoption? [y/N]: ")

	scanner := bufio.NewScanner(ia.input)
	if !scanner.Scan() {
		return false
	}

	response := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return response == "y" || response == "yes"
}

// formatSize formats bytes into human-readable size.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
