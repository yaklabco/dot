package doctor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/manifest"
)

// OrphanCheck scans for symlinks not managed by dot.
type OrphanCheck struct {
	fs            FS
	targetDir     string
	manifestSvc   ManifestLoader
	config        ScanConfig
	newTargetPath TargetPathCreator
}

func NewOrphanCheck(
	fs FS,
	manifestSvc ManifestLoader,
	targetDir string,
	config ScanConfig,
	newTargetPath TargetPathCreator,
) *OrphanCheck {
	return &OrphanCheck{
		fs:            fs,
		targetDir:     targetDir,
		manifestSvc:   manifestSvc,
		config:        config,
		newTargetPath: newTargetPath,
	}
}

func (c *OrphanCheck) Name() string {
	return "orphaned_links"
}

func (c *OrphanCheck) Description() string {
	return "Scans for symlinks that are not managed by dot"
}

// loadManifestOrEmpty loads manifest or returns empty if missing.
func (c *OrphanCheck) loadManifestOrEmpty(ctx context.Context) (manifest.Manifest, error) {
	targetPathResult := c.newTargetPath.NewTargetPath(c.targetDir)
	if !targetPathResult.IsOk() {
		return manifest.Manifest{}, targetPathResult.UnwrapErr()
	}

	manifestResult := c.manifestSvc.Load(ctx, targetPathResult.Unwrap())
	if manifestResult.IsOk() {
		return manifestResult.Unwrap(), nil
	}
	return manifest.Manifest{}, nil
}

// performScan executes sequential or parallel directory scan.
func (c *OrphanCheck) performScan(ctx context.Context, rootDirs []string, m *manifest.Manifest, linkSet map[string]struct{}, ignoreSet *ignore.IgnoreSet) ([]Issue, DiagnosticStats) {
	workers := c.config.MaxWorkers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	stats := DiagnosticStats{}
	localIssues := make([]Issue, 0)

	if workers == 1 || len(rootDirs) == 1 {
		c.scanSequential(ctx, rootDirs, m, linkSet, ignoreSet, &localIssues, &stats)
	} else {
		c.scanParallel(ctx, rootDirs, workers, m, linkSet, ignoreSet, &localIssues, &stats)
	}

	return localIssues, stats
}

// scanSequential performs sequential directory scanning.
func (c *OrphanCheck) scanSequential(ctx context.Context, rootDirs []string, m *manifest.Manifest, linkSet map[string]struct{}, ignoreSet *ignore.IgnoreSet, localIssues *[]Issue, stats *DiagnosticStats) {
	for _, dir := range rootDirs {
		c.scanDirectory(ctx, dir, m, linkSet, ignoreSet, localIssues, stats)
	}
}

// scanParallel performs parallel directory scanning.
func (c *OrphanCheck) scanParallel(ctx context.Context, rootDirs []string, workers int, m *manifest.Manifest, linkSet map[string]struct{}, ignoreSet *ignore.IgnoreSet, localIssues *[]Issue, stats *DiagnosticStats) {
	resultChan := make(chan scanResult, len(rootDirs))
	dirChan := make(chan string, len(rootDirs))
	var wg sync.WaitGroup

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go c.scanWorker(workerCtx, &wg, dirChan, resultChan, m, linkSet, ignoreSet)
	}

	go func() {
		for _, dir := range rootDirs {
			select {
			case dirChan <- dir:
			case <-workerCtx.Done():
				close(dirChan)
				return
			}
		}
		close(dirChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		// Check if adding these issues would exceed MaxIssues
		if c.config.MaxIssues > 0 && len(*localIssues) >= c.config.MaxIssues {
			// Stop collecting, but continue draining channel
			cancelWorkers()
			continue
		}

		// Add issues, but cap to MaxIssues if specified
		if c.config.MaxIssues > 0 {
			remaining := c.config.MaxIssues - len(*localIssues)
			if remaining > 0 {
				if len(res.issues) <= remaining {
					*localIssues = append(*localIssues, res.issues...)
				} else {
					*localIssues = append(*localIssues, res.issues[:remaining]...)
					cancelWorkers() // Signal workers to stop
				}
			}
		} else {
			*localIssues = append(*localIssues, res.issues...)
		}

		stats.TotalLinks += res.stats.TotalLinks
		stats.BrokenLinks += res.stats.BrokenLinks
		stats.OrphanedLinks += res.stats.OrphanedLinks
	}
}

// convertIssuesToDomain converts local issues to domain issues.
func convertIssuesToDomain(localIssues []Issue) []domain.Issue {
	domainIssues := make([]domain.Issue, 0, len(localIssues))
	for _, issue := range localIssues {
		severity := domain.IssueSeverityWarning
		if issue.Severity == domain.IssueSeverityError {
			severity = domain.IssueSeverityError
		}

		domainIssues = append(domainIssues, domain.Issue{
			Code:     string(issue.Type),
			Message:  issue.Message,
			Severity: severity,
			Path:     issue.Path,
			Context: map[string]any{
				"suggestion": issue.Suggestion,
			},
		})
	}
	return domainIssues
}

func (c *OrphanCheck) Run(ctx context.Context) (domain.CheckResult, error) {
	result := domain.CheckResult{
		CheckName: c.Name(),
		Status:    domain.CheckStatusPass,
		Issues:    make([]domain.Issue, 0),
		Stats:     make(map[string]any),
	}

	if c.config.Mode == ScanOff {
		result.Status = domain.CheckStatusSkipped
		return result, nil
	}

	m, err := c.loadManifestOrEmpty(ctx)
	if err != nil {
		return result, err
	}

	scanDirs := c.determineScanDirectories(&m)
	rootDirs := c.normalizeAndDeduplicateDirs(scanDirs)
	linkSet := c.buildManagedLinkSet(&m)
	ignoreSet := c.buildIgnoreSet(&m)

	localIssues, stats := c.performScan(ctx, rootDirs, &m, linkSet, ignoreSet)

	result.Issues = convertIssuesToDomain(localIssues)
	result.Stats["orphaned_links"] = stats.OrphanedLinks
	result.Stats["total_links"] = stats.TotalLinks
	result.Stats["broken_links"] = stats.BrokenLinks

	// Set status based on issue severity
	if len(result.Issues) > 0 {
		hasError := false
		for _, issue := range result.Issues {
			if issue.Severity == domain.IssueSeverityError {
				hasError = true
				break
			}
		}
		if hasError {
			result.Status = domain.CheckStatusFail
		} else {
			result.Status = domain.CheckStatusWarning
		}
	}

	return result, nil
}

// Helper methods ported from DoctorService

// Issue represents a diagnostic issue found during scanning.
type Issue struct {
	Severity   domain.IssueSeverity
	Type       IssueType
	Path       string
	Message    string
	Suggestion string
}

type scanResult struct {
	issues []Issue
	stats  DiagnosticStats
}

func (c *OrphanCheck) scanWorker(ctx context.Context, wg *sync.WaitGroup, dirChan chan string, resultChan chan scanResult, m *manifest.Manifest, linkSet map[string]struct{}, ignoreSet *ignore.IgnoreSet) {
	defer wg.Done()
	for dir := range dirChan {
		if ctx.Err() != nil {
			return
		}
		localIssues := []Issue{}
		localStats := DiagnosticStats{}
		c.scanDirectory(ctx, dir, m, linkSet, ignoreSet, &localIssues, &localStats)

		select {
		case resultChan <- scanResult{issues: localIssues, stats: localStats}:
		case <-ctx.Done():
			return
		}
	}
}

func (c *OrphanCheck) scanDirectory(ctx domain.Context, dir string, m *manifest.Manifest, linkSet map[string]struct{}, ignoreSet *ignore.IgnoreSet, issues *[]Issue, stats *DiagnosticStats) {
	// Recursive scan logic
	entries, err := c.fs.ReadDir(ctx, dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.Name() == ".dot-manifest.json" {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())

		// Check skip patterns
		if c.shouldSkipDirectory(fullPath) {
			continue
		}

		if entry.Type()&os.ModeSymlink != 0 {
			c.checkForOrphanedLink(ctx, fullPath, m, linkSet, ignoreSet, issues, stats)
		} else if entry.IsDir() {
			// Check max depth
			depth := c.calculateDepth(fullPath)
			if c.config.MaxDepth > 0 && depth > c.config.MaxDepth {
				continue
			}
			c.scanDirectory(ctx, fullPath, m, linkSet, ignoreSet, issues, stats)
		}
	}
}

func (c *OrphanCheck) checkForOrphanedLink(ctx context.Context, fullPath string, m *manifest.Manifest, linkSet map[string]struct{}, ignoreSet *ignore.IgnoreSet, issues *[]Issue, stats *DiagnosticStats) {
	// Check if we've hit the MaxIssues limit
	if c.config.MaxIssues > 0 && len(*issues) >= c.config.MaxIssues {
		return
	}

	relPath, err := filepath.Rel(c.targetDir, fullPath)
	if err != nil {
		relPath = fullPath
	}

	normalizedRel := filepath.ToSlash(relPath)
	normalizedFull := filepath.ToSlash(fullPath)
	_, managedRel := linkSet[normalizedRel]
	_, managedFull := linkSet[normalizedFull]
	managed := managedRel || managedFull

	if !managed {
		if m.Doctor != nil {
			if _, ignored := m.Doctor.IgnoredLinks[relPath]; ignored {
				return
			}
		}
		if ignoreSet != nil && ignoreSet.ShouldIgnore(relPath) {
			return
		}

		stats.TotalLinks++
		stats.OrphanedLinks++

		target, err := c.fs.ReadLink(ctx, fullPath)
		if err == nil {
			// Resolve target
			var absTarget string
			if filepath.IsAbs(target) {
				absTarget = target
			} else {
				absTarget = filepath.Join(filepath.Dir(fullPath), target)
			}

			// fmt.Printf("DEBUG: Checking broken link for %s -> %s (abs: %s)\n", fullPath, target, absTarget)
			_, err = c.fs.Stat(ctx, absTarget)
			// if err != nil {
			// 	fmt.Printf("DEBUG: Stat error for %s: %v (IsNotExist: %v)\n", absTarget, err, os.IsNotExist(err))
			// }
			if err != nil && os.IsNotExist(err) {
				stats.BrokenLinks++
				*issues = append(*issues, Issue{
					Severity:   domain.IssueSeverityError,
					Type:       IssueOrphanedLink,
					Path:       relPath,
					Message:    "Unmanaged symlink with broken target: " + target,
					Suggestion: "Remove manually or fix target",
				})
				return
			}
		}

		*issues = append(*issues, Issue{
			Severity:   domain.IssueSeverityWarning,
			Type:       IssueOrphanedLink,
			Path:       relPath,
			Message:    "Symlink not managed by dot",
			Suggestion: "Use 'dot adopt' to bring under management",
		})
	}
}

func (c *OrphanCheck) determineScanDirectories(m *manifest.Manifest) []string {
	if len(c.config.ScopeToDirs) > 0 {
		return c.config.ScopeToDirs
	}
	if c.config.Mode == ScanScoped {
		return extractManagedDirectories(m)
	}
	return []string{c.targetDir}
}

func (c *OrphanCheck) normalizeAndDeduplicateDirs(dirs []string) []string {
	absDirs := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		fullPath := dir
		if c.config.Mode == ScanScoped {
			fullPath = filepath.Join(c.targetDir, dir)
		}
		absDirs = append(absDirs, fullPath)
	}
	return filterDescendants(absDirs)
}

func (c *OrphanCheck) buildManagedLinkSet(m *manifest.Manifest) map[string]struct{} {
	linkSet := make(map[string]struct{})
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			normalized := filepath.ToSlash(link)
			linkSet[normalized] = struct{}{}
		}
	}
	return linkSet
}

func (c *OrphanCheck) buildIgnoreSet(m *manifest.Manifest) *ignore.IgnoreSet {
	ignoreSet := ignore.NewIgnoreSet()
	if m.Doctor != nil {
		for _, pattern := range m.Doctor.IgnoredPatterns {
			_ = ignoreSet.Add(pattern)
		}
	}
	return ignoreSet
}

func (c *OrphanCheck) calculateDepth(path string) int {
	path = filepath.Clean(path)
	targetDir := filepath.Clean(c.targetDir)
	if path == targetDir {
		return 0
	}
	rel, err := filepath.Rel(targetDir, path)
	if err != nil || rel == "." {
		return 0
	}
	// Depth is number of separators + 1 (e.g. "a/b" has 1 separator, depth 2)
	return strings.Count(rel, string(filepath.Separator)) + 1
}

func (c *OrphanCheck) shouldSkipDirectory(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range c.config.SkipPatterns {
		if base == pattern {
			return true
		}
	}
	return false
}

// Helpers duplicated from doctor_service.go (since they weren't exported)
func extractManagedDirectories(m *manifest.Manifest) []string {
	dirSet := make(map[string]struct{})
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			dir := filepath.Dir(link)
			for dir != "." && dir != "/" && dir != "" {
				dirSet[dir] = struct{}{}
				dir = filepath.Dir(dir)
			}
			dirSet["."] = struct{}{}
		}
	}
	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	return dirs
}

func filterDescendants(dirs []string) []string {
	if len(dirs) <= 1 {
		return dirs
	}
	cleaned := make([]string, len(dirs))
	for i, dir := range dirs {
		cleaned[i] = filepath.Clean(dir)
	}
	roots := make([]string, 0, len(cleaned))
	for _, dir := range cleaned {
		isDescendant := false
		for _, other := range cleaned {
			if dir == other {
				continue
			}
			rel, err := filepath.Rel(other, dir)
			if err == nil && rel != "." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..") {
				isDescendant = true
				break
			}
		}
		if !isDescendant {
			roots = append(roots, dir)
		}
	}
	return roots
}
