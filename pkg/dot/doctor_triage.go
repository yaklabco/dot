package dot

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesainslie/dot/internal/doctor"
	"github.com/jamesainslie/dot/internal/manifest"
)

// TriageOptions configures triage behavior.
type TriageOptions struct {
	AutoIgnoreHighConfidence bool // Automatically ignore high confidence categories
}

// TriageResult contains the results of a triage operation.
type TriageResult struct {
	Ignored  []string          // Links ignored
	Patterns []string          // Patterns added
	Adopted  map[string]string // Link -> package name
	Skipped  []string          // Links skipped
	Errors   map[string]error  // Link -> error
}

// OrphanGroup groups orphaned symlinks by category.
type OrphanGroup struct {
	Category        *doctor.PatternCategory
	Links           []Issue
	Confidence      string
	Pattern         string // Suggested ignore pattern
	IsUncategorized bool
}

// Triage performs interactive triage of orphaned symlinks.
func (s *DoctorService) Triage(ctx context.Context, scanCfg ScanConfig, opts TriageOptions) (TriageResult, error) {
	result := TriageResult{
		Adopted: make(map[string]string),
		Errors:  make(map[string]error),
	}

	// Run doctor to get issues
	report, err := s.DoctorWithScan(ctx, scanCfg)
	if err != nil {
		return result, err
	}

	// Load manifest
	targetPath, err := s.getTargetPath()
	if err != nil {
		return result, err
	}

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return result, manifestResult.UnwrapErr()
	}
	m := manifestResult.Unwrap()

	// Filter for orphaned links only
	orphanedIssues := filterIssuesByType(report.Issues, IssueOrphanedLink)
	if len(orphanedIssues) == 0 {
		return result, nil
	}

	// Group by category
	groups := s.groupOrphansByCategory(ctx, orphanedIssues)

	// Present overview and get processing choice
	choice := s.promptTriageOverview(orphanedIssues, groups)

	switch choice {
	case "c": // Process by category
		s.processTriageByCategory(ctx, &m, groups, opts, &result)
	case "l": // Process linearly
		s.processTriageLinearly(ctx, &m, orphanedIssues, groups, &result)
	case "a": // Auto-ignore high confidence
		s.autoIgnoreHighConfidence(ctx, &m, groups, &result)
	case "q": // Quit
		return result, nil
	}

	// Save manifest if changes made
	if len(result.Ignored) > 0 || len(result.Patterns) > 0 || len(result.Adopted) > 0 {
		if err := s.manifestSvc.Save(ctx, targetPath, m); err != nil {
			return result, fmt.Errorf("failed to save manifest: %w", err)
		}
	}

	return result, nil
}

// groupOrphansByCategory groups orphaned links by their category.
func (s *DoctorService) groupOrphansByCategory(ctx context.Context, issues []Issue) []OrphanGroup {
	categories := doctor.DefaultPatternCategories()
	categoryMap := make(map[string]*OrphanGroup)
	var uncategorized []Issue

	for _, issue := range issues {
		// Read the link target
		fullPath := filepath.Join(s.targetDir, issue.Path)
		target, err := s.fs.ReadLink(ctx, fullPath)
		if err != nil {
			uncategorized = append(uncategorized, issue)
			continue
		}

		// Try to categorize
		cat := doctor.CategorizeSymlink(target, categories)
		if cat == nil {
			uncategorized = append(uncategorized, issue)
			continue
		}

		// Add to category group
		key := cat.Name
		if _, exists := categoryMap[key]; !exists {
			pattern := s.generateIgnorePattern(cat, issue.Path)
			categoryMap[key] = &OrphanGroup{
				Category:   cat,
				Confidence: cat.Confidence,
				Pattern:    pattern,
			}
		}
		categoryMap[key].Links = append(categoryMap[key].Links, issue)
	}

	// Convert map to slice
	groups := make([]OrphanGroup, 0, len(categoryMap)+1)
	for _, group := range categoryMap {
		groups = append(groups, *group)
	}

	// Add uncategorized if any
	if len(uncategorized) > 0 {
		groups = append(groups, OrphanGroup{
			Links:           uncategorized,
			Confidence:      "unknown",
			IsUncategorized: true,
		})
	}

	return groups
}

// generateIgnorePattern creates a suggested ignore pattern for a category.
func (s *DoctorService) generateIgnorePattern(cat *doctor.PatternCategory, examplePath string) string {
	// Use the first pattern from the category
	if len(cat.Patterns) > 0 {
		return cat.Patterns[0]
	}

	// Fallback: generate from path
	parts := strings.Split(examplePath, "/")
	if len(parts) >= 2 {
		return fmt.Sprintf("*/%s/*", parts[0])
	}
	return examplePath
}

// promptTriageOverview shows overview and prompts for processing mode.
func (s *DoctorService) promptTriageOverview(allIssues []Issue, groups []OrphanGroup) string {
	fmt.Printf("\nFound %d orphaned links", len(allIssues))

	if len(groups) > 1 || (len(groups) == 1 && !groups[0].IsUncategorized) {
		fmt.Printf(" in %d categories:\n", len(groups))
		for i, group := range groups {
			if group.IsUncategorized {
				fmt.Printf("  [%d] Other (%d links)\n", i+1, len(group.Links))
			} else {
				fmt.Printf("  [%d] %s (%d links) - %s confidence\n",
					i+1, group.Category.Description, len(group.Links), group.Confidence)
			}
		}
	} else {
		fmt.Printf("\n")
	}

	fmt.Printf("\nProcess:\n")
	fmt.Printf("  c - Process by category\n")
	fmt.Printf("  l - Process linearly (one by one)\n")
	fmt.Printf("  a - Auto-ignore high confidence categories\n")
	fmt.Printf("  q - Quit\n")
	fmt.Printf("\nChoice [c]: ")

	var choice string
	fmt.Scanln(&choice)
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "" {
		choice = "c"
	}

	return choice
}

// processTriageByCategory processes orphans grouped by category.
func (s *DoctorService) processTriageByCategory(ctx context.Context, m *manifest.Manifest, groups []OrphanGroup, opts TriageOptions, result *TriageResult) {
	for _, group := range groups {
		if group.IsUncategorized {
			fmt.Printf("\n=== Category: Other (%d links) ===\n", len(group.Links))
		} else {
			fmt.Printf("\n=== Category: %s (%d links) ===\n", group.Category.Description, len(group.Links))
		}

		// Show sample links (up to 5)
		sampleCount := len(group.Links)
		if sampleCount > 5 {
			sampleCount = 5
		}
		for i := 0; i < sampleCount; i++ {
			fmt.Printf("  • %s\n", group.Links[i].Path)
		}
		if len(group.Links) > 5 {
			fmt.Printf("  ... and %d more\n", len(group.Links)-5)
		}

		// Prompt for category action
		action := s.promptCategoryAction(group)

		switch action {
		case "i": // Ignore this category
			pattern := group.Pattern
			if pattern == "" {
				// Prompt for pattern
				fmt.Printf("Enter ignore pattern: ")
				fmt.Scanln(&pattern)
				pattern = strings.TrimSpace(pattern)
			}

			if pattern != "" {
				m.AddIgnoredPattern(pattern)
				result.Patterns = append(result.Patterns, pattern)
				fmt.Printf("Added ignore pattern: %s\n", pattern)
			}

		case "r": // Review individually
			s.processLinksIndividually(ctx, m, group.Links, result)

		case "s": // Skip
			for _, link := range group.Links {
				result.Skipped = append(result.Skipped, link.Path)
			}

		case "q": // Quit
			return
		}
	}
}

// promptCategoryAction prompts for action on a category.
func (s *DoctorService) promptCategoryAction(group OrphanGroup) string {
	fmt.Printf("\nActions:\n")

	if !group.IsUncategorized && group.Pattern != "" {
		fmt.Printf("  i - Ignore this category (pattern: %s)\n", group.Pattern)
	} else {
		fmt.Printf("  i - Ignore with custom pattern\n")
	}

	fmt.Printf("  r - Review each link individually\n")
	fmt.Printf("  s - Skip this category\n")
	fmt.Printf("  q - Quit\n")
	fmt.Printf("\nChoice [i]: ")

	var choice string
	fmt.Scanln(&choice)
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "" {
		choice = "i"
	}

	return choice
}

// processTriageLinearly processes orphans one by one.
func (s *DoctorService) processTriageLinearly(ctx context.Context, m *manifest.Manifest, issues []Issue, groups []OrphanGroup, result *TriageResult) {
	// Create a map of issue path to category for quick lookup
	categoryMap := make(map[string]*doctor.PatternCategory)
	for _, group := range groups {
		if !group.IsUncategorized {
			for _, link := range group.Links {
				categoryMap[link.Path] = group.Category
			}
		}
	}

	s.processLinksIndividually(ctx, m, issues, result)
}

// processLinksIndividually processes each link with individual prompts.
func (s *DoctorService) processLinksIndividually(ctx context.Context, m *manifest.Manifest, issues []Issue, result *TriageResult) {
	applyToAll := false
	applyToAllAction := ""

	for i, issue := range issues {
		if applyToAll {
			s.applyTriageAction(ctx, m, issue, applyToAllAction, result)
			continue
		}

		action, all := s.promptLinkAction(ctx, issue, i+1, len(issues))
		if all {
			applyToAll = true
			applyToAllAction = action
		}

		s.applyTriageAction(ctx, m, issue, action, result)

		if action == "q" {
			break
		}
	}
}

// promptLinkAction prompts for action on an individual link.
// Returns (action, applyToAll).
func (s *DoctorService) promptLinkAction(ctx context.Context, issue Issue, current, total int) (string, bool) {
	fullPath := filepath.Join(s.targetDir, issue.Path)
	target, err := s.fs.ReadLink(ctx, fullPath)
	if err != nil {
		target = "(unable to read target)"
	}

	// Try to categorize
	categories := doctor.DefaultPatternCategories()
	cat := doctor.CategorizeSymlink(target, categories)

	fmt.Printf("\nOrphaned symlink [%d/%d]: %s\n", current, total, issue.Path)
	fmt.Printf("  Target: %s\n", target)

	if cat != nil {
		fmt.Printf("  Category: %s\n", cat.Description)
	} else {
		fmt.Printf("  Category: Unknown\n")
	}

	fmt.Printf("\nActions:\n")
	fmt.Printf("  i - Ignore this link\n")
	fmt.Printf("  p - Ignore with pattern\n")

	if cat != nil {
		suggestedPattern := s.generateIgnorePattern(cat, issue.Path)
		fmt.Printf("  P - Auto-ignore pattern (%s)\n", suggestedPattern)
		fmt.Printf("  c - Ignore all in category \"%s\"\n", cat.Description)
	}

	fmt.Printf("  a - Adopt into dot\n")
	fmt.Printf("  s - Skip\n")
	fmt.Printf("  A - Apply to all remaining\n")
	fmt.Printf("  q - Quit\n")
	fmt.Printf("\nChoice [s]: ")

	var choice string
	fmt.Scanln(&choice)
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "" {
		choice = "s"
	}

	// Check for "apply to all"
	if choice == "a" && strings.Contains(choice, "A") {
		return choice, true
	}

	return choice, false
}

// applyTriageAction applies the chosen action to a link.
func (s *DoctorService) applyTriageAction(ctx context.Context, m *manifest.Manifest, issue Issue, action string, result *TriageResult) {
	fullPath := filepath.Join(s.targetDir, issue.Path)
	target, _ := s.fs.ReadLink(ctx, fullPath)

	switch action {
	case "i": // Ignore this link
		m.AddIgnoredLink(issue.Path, target, "user triage")
		result.Ignored = append(result.Ignored, issue.Path)

	case "p": // Ignore with custom pattern
		fmt.Printf("Enter ignore pattern: ")
		var pattern string
		fmt.Scanln(&pattern)
		pattern = strings.TrimSpace(pattern)

		if pattern != "" {
			m.AddIgnoredPattern(pattern)
			result.Patterns = append(result.Patterns, pattern)
			fmt.Printf("Added ignore pattern: %s\n", pattern)
		}

	case "P": // Auto-ignore pattern
		categories := doctor.DefaultPatternCategories()
		cat := doctor.CategorizeSymlink(target, categories)
		if cat != nil {
			pattern := s.generateIgnorePattern(cat, issue.Path)
			m.AddIgnoredPattern(pattern)
			result.Patterns = append(result.Patterns, pattern)
			fmt.Printf("Added ignore pattern: %s\n", pattern)
		}

	case "c": // Ignore all in category
		categories := doctor.DefaultPatternCategories()
		cat := doctor.CategorizeSymlink(target, categories)
		if cat != nil {
			// Add all patterns from this category
			for _, pattern := range cat.Patterns {
				m.AddIgnoredPattern(pattern)
				result.Patterns = append(result.Patterns, pattern)
			}
			fmt.Printf("Added %d patterns for category %s\n", len(cat.Patterns), cat.Description)
		}

	case "a": // Adopt
		pkgName := s.promptPackageName()
		if pkgName != "" {
			result.Adopted[issue.Path] = pkgName
			fmt.Printf("Marked for adoption into package: %s\n", pkgName)
			// Note: Actual adoption would need to be done in a separate phase
			// For now, just track it
		}

	case "s": // Skip
		result.Skipped = append(result.Skipped, issue.Path)

	case "q": // Quit
		// Just return, caller will handle
		return
	}
}

// promptPackageName prompts user for package name for adoption.
func (s *DoctorService) promptPackageName() string {
	fmt.Printf("Enter package name (or press Enter to cancel): ")
	var pkgName string
	fmt.Scanln(&pkgName)
	return strings.TrimSpace(pkgName)
}

// autoIgnoreHighConfidence automatically ignores high confidence categories.
func (s *DoctorService) autoIgnoreHighConfidence(ctx context.Context, m *manifest.Manifest, groups []OrphanGroup, result *TriageResult) {
	fmt.Printf("\nAuto-ignoring high confidence categories...\n")

	for _, group := range groups {
		if group.Confidence == "high" && !group.IsUncategorized {
			// Add all patterns for this category
			if group.Category != nil {
				for _, pattern := range group.Category.Patterns {
					m.AddIgnoredPattern(pattern)
					result.Patterns = append(result.Patterns, pattern)
				}
				fmt.Printf("  • Ignored %s (%d links, %d patterns)\n",
					group.Category.Description, len(group.Links), len(group.Category.Patterns))
			}
		}
	}

	fmt.Printf("\nAuto-ignore complete. Other links not affected.\n")
}

// filterIssuesByType returns issues matching the given type.
func filterIssuesByType(issues []Issue, issueType IssueType) []Issue {
	var filtered []Issue
	for _, issue := range issues {
		if issue.Type == issueType {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}
