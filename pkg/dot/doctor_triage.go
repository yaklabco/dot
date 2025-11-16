package dot

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesainslie/dot/internal/doctor"
	"github.com/jamesainslie/dot/internal/manifest"
)

// TriageOptions configures triage behavior.
type TriageOptions struct {
	AutoIgnoreHighConfidence bool // Automatically ignore high confidence categories
	DryRun                   bool // Show what would change without modifying
	AutoConfirm              bool // Skip confirmation prompts (--yes flag)
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

	// Filter out already-ignored items
	orphanedIssues = s.filterAlreadyIgnored(orphanedIssues, &m)

	if len(orphanedIssues) == 0 {
		return result, nil
	}

	// Group by category
	groups := s.groupOrphansByCategory(ctx, orphanedIssues)

	// If auto-ignore flag is set, automatically ignore high confidence categories
	if opts.AutoIgnoreHighConfidence {
		s.autoIgnoreHighConfidence(ctx, &m, groups, &result)
	} else {
		// Present overview and get processing choice
		choice := s.promptTriageOverview(orphanedIssues, groups)

		switch choice {
		case "c": // Process by category
			s.processTriageByCategory(ctx, &m, groups, opts, &result)
		case "l": // Process linearly
			s.processTriageLinearly(ctx, &m, orphanedIssues, groups, opts, &result)
		case "a": // Auto-ignore high confidence
			s.autoIgnoreHighConfidence(ctx, &m, groups, &result)
		case "q": // Quit
			return result, nil
		}
	}

	// Save changes
	if err := s.saveTriageResults(ctx, targetPath, m, opts, result); err != nil {
		return result, err
	}

	return result, nil
}

// saveTriageResults saves the manifest if changes were made.
func (s *DoctorService) saveTriageResults(ctx context.Context, targetPath TargetPath, m manifest.Manifest, opts TriageOptions, result TriageResult) error {
	hasChanges := len(result.Ignored) > 0 || len(result.Patterns) > 0 || len(result.Adopted) > 0
	if !hasChanges {
		return nil
	}

	if opts.DryRun {
		fmt.Println("\n[DRY RUN] No changes were made")
		return nil
	}

	if !opts.AutoConfirm && !s.confirmTriageChanges(result) {
		fmt.Println("\nChanges cancelled")
		return nil
	}

	if err := s.manifestSvc.Save(ctx, targetPath, m); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	fmt.Println("\nChanges saved successfully")
	return nil
}

// addIgnorePatternIfNew adds a pattern to the manifest if it doesn't already exist.
// Returns true if the pattern was added, false if it already existed.
func (s *DoctorService) addIgnorePatternIfNew(m *manifest.Manifest, pattern string, result *TriageResult) bool {
	if m.Doctor == nil {
		m.Doctor = &manifest.DoctorState{
			IgnoredLinks:    make(map[string]manifest.IgnoredLink),
			IgnoredPatterns: []string{},
		}
	}

	// Check if pattern already exists
	for _, existing := range m.Doctor.IgnoredPatterns {
		if existing == pattern {
			return false
		}
	}

	m.AddIgnoredPattern(pattern)
	result.Patterns = append(result.Patterns, pattern)
	return true
}

// confirmTriageChanges shows a summary and asks for confirmation before saving.
func (s *DoctorService) confirmTriageChanges(result TriageResult) bool {
	fmt.Printf("\nSummary of changes:\n")
	if len(result.Ignored) > 0 {
		fmt.Printf("  • %d links to ignore\n", len(result.Ignored))
	}
	if len(result.Patterns) > 0 {
		fmt.Printf("  • %d patterns to add\n", len(result.Patterns))
	}
	if len(result.Adopted) > 0 {
		fmt.Printf("  • %d links to adopt\n", len(result.Adopted))
	}
	if len(result.Errors) > 0 {
		fmt.Printf("  • %d errors occurred\n", len(result.Errors))
	}

	fmt.Printf("\nSave these changes? [Y/n]: ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Handle EOF or input error - default to yes
		return true
	}
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "" || response == "y" || response == "yes"
}

// filterAlreadyIgnored filters out issues that are already ignored.
func (s *DoctorService) filterAlreadyIgnored(issues []Issue, m *manifest.Manifest) []Issue {
	if m.Doctor == nil {
		return issues
	}

	ignoreSet := s.buildIgnoreSet(m)
	filtered := []Issue{}

	for _, issue := range issues {
		// Check explicit ignores
		if _, ignored := m.Doctor.IgnoredLinks[issue.Path]; ignored {
			continue
		}

		// Check patterns
		if ignoreSet.ShouldIgnore(issue.Path) {
			continue
		}

		filtered = append(filtered, issue)
	}

	return filtered
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

	// Sort groups by category name for deterministic output
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Category == nil {
			return false
		}
		if groups[j].Category == nil {
			return true
		}
		return groups[i].Category.Name < groups[j].Category.Name
	})

	// Add uncategorized if any (always last)
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
	if _, err := fmt.Scanln(&choice); err != nil {
		// Handle EOF or input error
		choice = "q"
	}
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "" {
		choice = "c"
	}

	return choice
}

// processTriageByCategory processes orphans grouped by category.
func (s *DoctorService) processTriageByCategory(ctx context.Context, m *manifest.Manifest, groups []OrphanGroup, opts TriageOptions, result *TriageResult) {
	for _, group := range groups {
		s.displayCategoryInfo(group)
		action := s.getCategoryAction(group, opts)
		if !s.handleCategoryAction(ctx, m, group, action, opts, result) {
			return
		}
	}
}

// displayCategoryInfo displays category header and sample links.
func (s *DoctorService) displayCategoryInfo(group OrphanGroup) {
	if group.IsUncategorized {
		fmt.Printf("\n=== Category: Other (%d links) ===\n", len(group.Links))
	} else {
		fmt.Printf("\n=== Category: %s (%d links) ===\n", group.Category.Description, len(group.Links))
	}

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
}

// getCategoryAction gets the action for a category (auto or prompted).
func (s *DoctorService) getCategoryAction(group OrphanGroup, opts TriageOptions) string {
	if opts.AutoConfirm {
		fmt.Printf("\n[AUTO] Action: ignore this category\n")
		return "i"
	}
	return s.promptCategoryAction(group)
}

// handleCategoryAction handles the action for a category. Returns false if user quit.
func (s *DoctorService) handleCategoryAction(ctx context.Context, m *manifest.Manifest, group OrphanGroup, action string, opts TriageOptions, result *TriageResult) bool {
	switch action {
	case "i":
		s.handleIgnoreCategory(m, group.Pattern, opts.DryRun, result)
	case "r":
		s.processLinksIndividually(ctx, m, group.Links, result, opts.DryRun)
	case "s":
		s.handleSkipCategory(group, result)
	case "q":
		return false
	}
	return true
}

// handleSkipCategory handles skipping a category.
func (s *DoctorService) handleSkipCategory(group OrphanGroup, result *TriageResult) {
	for _, link := range group.Links {
		result.Skipped = append(result.Skipped, link.Path)
	}
}

// handleIgnoreCategory handles the ignore action for a category with dry-run support.
func (s *DoctorService) handleIgnoreCategory(m *manifest.Manifest, pattern string, dryRun bool, result *TriageResult) {
	if pattern == "" {
		// Prompt for pattern
		fmt.Printf("Enter ignore pattern: ")
		if _, err := fmt.Scanln(&pattern); err != nil {
			fmt.Printf("\nInput cancelled\n")
			return
		}
		pattern = strings.TrimSpace(pattern)
	}

	if pattern == "" {
		return
	}

	// Validate pattern is a valid glob
	if _, err := filepath.Match(pattern, "test"); err != nil {
		fmt.Printf("Invalid pattern: %s\n", err)
		return
	}

	if dryRun {
		fmt.Printf("[DRY RUN] Would add ignore pattern: %s\n", pattern)
		return
	}

	if s.addIgnorePatternIfNew(m, pattern, result) {
		fmt.Printf("Added ignore pattern: %s\n", pattern)
	} else {
		fmt.Printf("Pattern already exists: %s\n", pattern)
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
	if _, err := fmt.Scanln(&choice); err != nil {
		// Handle EOF or input error
		choice = "q"
	}
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "" {
		choice = "i"
	}

	return choice
}

// processTriageLinearly processes orphans one by one.
func (s *DoctorService) processTriageLinearly(ctx context.Context, m *manifest.Manifest, issues []Issue, groups []OrphanGroup, opts TriageOptions, result *TriageResult) {
	s.processLinksIndividually(ctx, m, issues, result, opts.DryRun)
}

// processLinksIndividually processes each link with individual prompts.
func (s *DoctorService) processLinksIndividually(ctx context.Context, m *manifest.Manifest, issues []Issue, result *TriageResult, dryRun bool) {
	applyToAll := false
	applyToAllAction := ""

	for i, issue := range issues {
		if applyToAll {
			s.applyTriageAction(ctx, m, issue, applyToAllAction, result, dryRun)
			continue
		}

		action, all := s.promptLinkAction(ctx, issue, i+1, len(issues))
		if all {
			applyToAll = true
			applyToAllAction = action
		}

		s.applyTriageAction(ctx, m, issue, action, result, dryRun)

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
	if _, err := fmt.Scanln(&choice); err != nil {
		// Handle EOF or input error
		choice = "q"
	}
	choice = strings.TrimSpace(choice)

	// Check for "apply to all" BEFORE lowercasing
	if choice == "A" {
		return "a", true
	}

	choice = strings.ToLower(choice)

	if choice == "" {
		choice = "s"
	}

	return choice, false
}

// actionDescription returns a human-readable description of an action.
func actionDescription(action string) string {
	switch action {
	case "i":
		return "ignore link"
	case "p", "P":
		return "add ignore pattern"
	case "c":
		return "ignore category"
	case "a":
		return "adopt link"
	case "s":
		return "skip"
	default:
		return "unknown action"
	}
}

// applyTriageAction applies the chosen action to a link.
func (s *DoctorService) applyTriageAction(ctx context.Context, m *manifest.Manifest, issue Issue, action string, result *TriageResult, dryRun bool) {
	if dryRun {
		fmt.Printf("[DRY RUN] Would %s: %s\n", actionDescription(action), issue.Path)
		return
	}

	fullPath := filepath.Join(s.targetDir, issue.Path)
	target, _ := s.fs.ReadLink(ctx, fullPath)

	switch action {
	case "i": // Ignore this link
		s.applyIgnoreLink(m, issue, target, result)
	case "p": // Ignore with custom pattern
		s.applyIgnoreCustomPattern(m, result)
	case "P": // Auto-ignore pattern
		s.applyAutoIgnorePattern(m, issue, target, result)
	case "c": // Ignore all in category
		s.applyIgnoreCategory(m, target, result)
	case "a": // Adopt
		s.applyAdoptLink(ctx, m, issue, result)
	case "s": // Skip
		result.Skipped = append(result.Skipped, issue.Path)
	case "q": // Quit
		// Just return, caller will handle
		return
	}
}

func (s *DoctorService) applyIgnoreLink(m *manifest.Manifest, issue Issue, target string, result *TriageResult) {
	m.AddIgnoredLink(issue.Path, target, "user triage")
	result.Ignored = append(result.Ignored, issue.Path)
}

func (s *DoctorService) applyIgnoreCustomPattern(m *manifest.Manifest, result *TriageResult) {
	fmt.Printf("Enter ignore pattern: ")
	var pattern string
	if _, err := fmt.Scanln(&pattern); err != nil {
		fmt.Printf("\nInput cancelled\n")
		return
	}
	pattern = strings.TrimSpace(pattern)

	if pattern == "" {
		return
	}

	// Validate pattern is a valid glob
	if _, err := filepath.Match(pattern, "test"); err != nil {
		fmt.Printf("Invalid pattern: %s\n", err)
		return
	}

	if s.addIgnorePatternIfNew(m, pattern, result) {
		fmt.Printf("Added ignore pattern: %s\n", pattern)
	} else {
		fmt.Printf("Pattern already exists: %s\n", pattern)
	}
}

func (s *DoctorService) applyAutoIgnorePattern(m *manifest.Manifest, issue Issue, target string, result *TriageResult) {
	categories := doctor.DefaultPatternCategories()
	cat := doctor.CategorizeSymlink(target, categories)
	if cat != nil {
		pattern := s.generateIgnorePattern(cat, issue.Path)
		if s.addIgnorePatternIfNew(m, pattern, result) {
			fmt.Printf("Added ignore pattern: %s\n", pattern)
		} else {
			fmt.Printf("Pattern already exists: %s\n", pattern)
		}
	}
}

func (s *DoctorService) applyIgnoreCategory(m *manifest.Manifest, target string, result *TriageResult) {
	categories := doctor.DefaultPatternCategories()
	cat := doctor.CategorizeSymlink(target, categories)
	if cat != nil {
		addedCount := 0
		for _, pattern := range cat.Patterns {
			if s.addIgnorePatternIfNew(m, pattern, result) {
				addedCount++
			}
		}
		if addedCount > 0 {
			fmt.Printf("Added %d patterns for category %s\n", addedCount, cat.Description)
		} else {
			fmt.Printf("All patterns for category %s already exist\n", cat.Description)
		}
	}
}

func (s *DoctorService) applyAdoptLink(ctx context.Context, m *manifest.Manifest, issue Issue, result *TriageResult) {
	pkgName := s.promptPackageName()
	if pkgName != "" {
		if err := s.executeAdoption(ctx, issue.Path, pkgName); err != nil {
			result.Errors[issue.Path] = err
			fmt.Printf("Failed to adopt: %s\n", err)
		} else {
			result.Adopted[issue.Path] = pkgName
			fmt.Printf("Successfully adopted into package: %s\n", pkgName)
		}
	}
}

// executeAdoption actually adopts a symlink into a package.
func (s *DoctorService) executeAdoption(ctx context.Context, linkPath, pkgName string) error {
	if s.adoptSvc == nil {
		return fmt.Errorf("adoption not supported (adoptSvc not initialized)")
	}

	// Call the AdoptService to handle the adoption
	// The AdoptService will:
	// 1. Move the file/dir to the package directory
	// 2. Create a symlink back to the original location
	// 3. Update the manifest
	err := s.adoptSvc.Adopt(ctx, []string{linkPath}, pkgName)
	if err != nil {
		return fmt.Errorf("adoption failed: %w", err)
	}

	s.logger.Info(ctx, "adopted_via_triage", "link", linkPath, "package", pkgName)
	return nil
}

// promptPackageName prompts user for package name for adoption.
func (s *DoctorService) promptPackageName() string {
	fmt.Printf("Enter package name (or press Enter to cancel): ")
	var pkgName string
	if _, err := fmt.Scanln(&pkgName); err != nil {
		// Handle EOF or input error
		return ""
	}
	return strings.TrimSpace(pkgName)
}

// autoIgnoreHighConfidence automatically ignores high confidence categories.
func (s *DoctorService) autoIgnoreHighConfidence(ctx context.Context, m *manifest.Manifest, groups []OrphanGroup, result *TriageResult) {
	fmt.Printf("\nAuto-ignoring high confidence categories...\n")

	for _, group := range groups {
		if group.Confidence == "high" && !group.IsUncategorized {
			// Add all patterns for this category
			if group.Category != nil {
				addedCount := 0
				for _, pattern := range group.Category.Patterns {
					if s.addIgnorePatternIfNew(m, pattern, result) {
						addedCount++
					}
				}
				if addedCount > 0 {
					fmt.Printf("  • Ignored %s (%d links, %d new patterns)\n",
						group.Category.Description, len(group.Links), addedCount)
				}
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
