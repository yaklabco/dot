package dot

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/manifest"
)

// DoctorService handles health check and diagnostic operations.
type DoctorService struct {
	fs            FS
	logger        Logger
	manifestSvc   *ManifestService
	packageDir    string
	targetDir     string
	healthChecker *HealthChecker
	adoptSvc      *AdoptService
}

// scanResult holds the results from scanning a single directory.
type scanResult struct {
	issues []Issue
	stats  DiagnosticStats
}

// newDoctorService creates a new doctor service (for tests).
func newDoctorService(
	fs FS,
	logger Logger,
	manifestSvc *ManifestService,
	packageDir string,
	targetDir string,
) *DoctorService {
	return &DoctorService{
		fs:            fs,
		logger:        logger,
		manifestSvc:   manifestSvc,
		packageDir:    packageDir,
		targetDir:     targetDir,
		healthChecker: newHealthChecker(fs, targetDir),
		adoptSvc:      nil, // Not needed for basic tests
	}
}

// newDoctorServiceWithAdopt creates a new doctor service with adoption support.
func newDoctorServiceWithAdopt(
	fs FS,
	logger Logger,
	manifestSvc *ManifestService,
	adoptSvc *AdoptService,
	packageDir string,
	targetDir string,
) *DoctorService {
	return &DoctorService{
		fs:            fs,
		logger:        logger,
		manifestSvc:   manifestSvc,
		packageDir:    packageDir,
		targetDir:     targetDir,
		healthChecker: newHealthChecker(fs, targetDir),
		adoptSvc:      adoptSvc,
	}
}

// Doctor performs health checks with default scan configuration.
func (s *DoctorService) Doctor(ctx context.Context) (DiagnosticReport, error) {
	return s.DoctorWithScan(ctx, DefaultScanConfig())
}

// DoctorWithScan performs health checks with explicit scan configuration.
func (s *DoctorService) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
	targetPath, err := s.getTargetPath()
	if err != nil {
		return DiagnosticReport{}, err
	}

	m, issues, stats, err := s.loadManifestOrCreateDefault(ctx, targetPath)
	if err != nil {
		return DiagnosticReport{}, err
	}
	// If manifest doesn't exist, return early with info issue
	if m == nil {
		return DiagnosticReport{
			OverallHealth: HealthOK,
			Issues:        issues,
			Statistics:    stats,
		}, nil
	}

	s.checkManagedPackages(ctx, m, &issues, &stats)

	if scanCfg.Mode != ScanOff {
		s.performOrphanScan(ctx, m, scanCfg, &issues, &stats)
	}

	health := s.determineOverallHealth(issues)

	return DiagnosticReport{
		OverallHealth: health,
		Issues:        issues,
		Statistics:    stats,
	}, nil
}

// getTargetPath constructs and validates target path.
func (s *DoctorService) getTargetPath() (TargetPath, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return TargetPath{}, targetPathResult.UnwrapErr()
	}
	return targetPathResult.Unwrap(), nil
}

// loadManifestOrCreateDefault loads manifest or returns default state if not found.
func (s *DoctorService) loadManifestOrCreateDefault(ctx context.Context, targetPath TargetPath) (*manifest.Manifest, []Issue, DiagnosticStats, error) {
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	issues := make([]Issue, 0)
	stats := DiagnosticStats{}

	if !manifestResult.IsOk() {
		err := manifestResult.UnwrapErr()
		if isManifestNotFoundError(err) {
			issues = append(issues, Issue{
				Severity:   SeverityInfo,
				Type:       IssueManifestInconsistency,
				Message:    "No manifest found - no packages are currently managed",
				Suggestion: "Run 'dot manage' to install packages",
			})
			return nil, issues, stats, nil
		}
		return nil, nil, stats, err
	}

	m := manifestResult.Unwrap()
	return &m, issues, stats, nil
}

// checkManagedPackages validates all packages in the manifest.
func (s *DoctorService) checkManagedPackages(ctx context.Context, m *manifest.Manifest, issues *[]Issue, stats *DiagnosticStats) {
	for pkgName, pkgInfo := range m.Packages {
		stats.ManagedLinks += pkgInfo.LinkCount
		for _, linkPath := range pkgInfo.Links {
			stats.TotalLinks++
			s.checkLink(ctx, pkgName, linkPath, pkgInfo, issues, stats)
		}
	}
}

// determineOverallHealth computes health status from issues.
func (s *DoctorService) determineOverallHealth(issues []Issue) HealthStatus {
	health := HealthOK
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			return HealthErrors
		}
		if issue.Severity == SeverityWarning && health == HealthOK {
			health = HealthWarnings
		}
	}
	return health
}

// checkLink validates a single link from the manifest.
func (s *DoctorService) checkLink(ctx context.Context, pkgName string, linkPath string, pkgInfo manifest.PackageInfo, issues *[]Issue, stats *DiagnosticStats) {
	// Use unified health checker
	result := s.healthChecker.CheckLink(ctx, pkgName, linkPath, pkgInfo.PackageDir)

	if !result.IsHealthy {
		// Count broken links for statistics
		if result.IssueType == IssueBrokenLink {
			stats.BrokenLinks++
		}

		// Add issue to list
		*issues = append(*issues, Issue{
			Severity:   result.Severity,
			Type:       result.IssueType,
			Path:       linkPath,
			Message:    result.Message,
			Suggestion: result.Suggestion,
		})
	}
}

// performOrphanScan executes orphaned link scanning based on configuration.
// Scans directories in parallel using worker pool for improved performance.
func (s *DoctorService) performOrphanScan(
	ctx context.Context,
	m *manifest.Manifest,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	scanDirs := s.determineScanDirectories(m, scanCfg)
	rootDirs := s.normalizeAndDeduplicateDirs(scanDirs, scanCfg.Mode)
	linkSet := buildManagedLinkSet(m)
	ignoreSet := s.buildIgnoreSet(m)

	// Determine worker count
	workers := scanCfg.MaxWorkers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	// If only 1 worker or 1 directory, use sequential scan (no overhead)
	if workers == 1 || len(rootDirs) == 1 {
		for _, dir := range rootDirs {
			if s.shouldStopScan(scanCfg, issues) {
				break
			}
			s.scanDirectory(ctx, dir, m, linkSet, ignoreSet, scanCfg, issues, stats)
		}
		return
	}

	// Parallel scan with worker pool
	resultChan := make(chan scanResult, len(rootDirs))
	dirChan := make(chan string, len(rootDirs))
	var wg sync.WaitGroup

	// Create cancellable context for early termination
	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go s.scanWorker(workerCtx, &wg, dirChan, resultChan, m, linkSet, ignoreSet, scanCfg)
	}

	// Feed directories to workers with cancellation support
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

	// Wait for workers and close results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, respecting MaxIssues budget
	for result := range resultChan {
		if s.collectScanResult(result, scanCfg, issues, stats, cancelWorkers, resultChan) {
			break
		}
	}
}

// scanWorker processes directories from dirChan and sends results to resultChan.
func (s *DoctorService) scanWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	dirChan chan string,
	resultChan chan scanResult,
	m *manifest.Manifest,
	linkSet map[string]bool,
	ignoreSet *ignore.IgnoreSet,
	scanCfg ScanConfig,
) {
	defer wg.Done()
	for dir := range dirChan {
		if ctx.Err() != nil {
			return
		}

		localIssues := []Issue{}
		localStats := DiagnosticStats{}
		s.scanDirectory(ctx, dir, m, linkSet, ignoreSet, scanCfg, &localIssues, &localStats)

		// Only send result if context not cancelled
		select {
		case resultChan <- scanResult{
			issues: localIssues,
			stats:  localStats,
		}:
		case <-ctx.Done():
			return
		}
	}
}

// collectScanResult processes a single scan result and returns true if collection should stop.
func (s *DoctorService) collectScanResult(
	result scanResult,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
	cancelWorkers context.CancelFunc,
	resultChan chan scanResult,
) bool {
	// Respect remaining budget before appending
	if scanCfg.MaxIssues > 0 {
		remaining := scanCfg.MaxIssues - len(*issues)
		if remaining <= 0 {
			// Budget exhausted, cancel workers and drain results
			cancelWorkers()
			for range resultChan {
			}
			return true
		}
		// Truncate result to remaining budget
		if len(result.issues) > remaining {
			*issues = append(*issues, result.issues[:remaining]...)
		} else {
			*issues = append(*issues, result.issues...)
		}
	} else {
		// No limit, append all
		*issues = append(*issues, result.issues...)
	}

	stats.TotalLinks += result.stats.TotalLinks
	stats.BrokenLinks += result.stats.BrokenLinks
	stats.OrphanedLinks += result.stats.OrphanedLinks
	stats.ManagedLinks += result.stats.ManagedLinks
	return false
}

// shouldStopScan checks if scanning should stop early based on MaxIssues limit.
func (s *DoctorService) shouldStopScan(scanCfg ScanConfig, issues *[]Issue) bool {
	return scanCfg.MaxIssues > 0 && len(*issues) >= scanCfg.MaxIssues
}

// determineScanDirectories determines which directories to scan based on configuration.
func (s *DoctorService) determineScanDirectories(m *manifest.Manifest, scanCfg ScanConfig) []string {
	if len(scanCfg.ScopeToDirs) > 0 {
		return scanCfg.ScopeToDirs
	}
	if scanCfg.Mode == ScanScoped {
		return extractManagedDirectories(m)
	}
	return []string{s.targetDir}
}

// normalizeAndDeduplicateDirs converts scan directories to absolute paths and removes descendants.
func (s *DoctorService) normalizeAndDeduplicateDirs(dirs []string, mode ScanMode) []string {
	absDirs := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		fullPath := dir
		if mode == ScanScoped {
			fullPath = filepath.Join(s.targetDir, dir)
		}
		absDirs = append(absDirs, fullPath)
	}
	return filterDescendants(absDirs)
}

// scanDirectory scans a single directory for orphaned links with limit checks.
func (s *DoctorService) scanDirectory(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	ignoreSet *ignore.IgnoreSet,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	err := s.scanForOrphanedLinksWithLimits(ctx, dir, m, linkSet, ignoreSet, scanCfg, issues, stats)
	if err != nil {
		// Log but continue - orphan detection is best-effort
		s.logger.Warn(ctx, "scan_directory_failed", "dir", dir, "error", err)
	}
}

// scanForOrphanedLinksWithLimits wraps scanForOrphanedLinks with depth and skip checks.
func (s *DoctorService) scanForOrphanedLinksWithLimits(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	ignoreSet *ignore.IgnoreSet,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	depth := calculateDepth(dir, s.targetDir)
	if scanCfg.MaxDepth > 0 && depth > scanCfg.MaxDepth {
		return nil
	}

	if shouldSkipDirectory(dir, scanCfg.SkipPatterns) {
		return nil
	}

	return s.scanForOrphanedLinks(ctx, dir, m, linkSet, ignoreSet, scanCfg, issues, stats)
}

// scanForOrphanedLinks recursively scans for symlinks not in the manifest.
// Optimized to check symlink type from DirEntry without extra syscalls.
func (s *DoctorService) scanForOrphanedLinks(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	ignoreSet *ignore.IgnoreSet,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) error {
	entries, err := s.fs.ReadDir(ctx, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if s.shouldSkipEntry(entry) {
			continue
		}

		// Check MaxIssues limit
		if s.shouldStopScan(scanCfg, issues) {
			return nil
		}

		fullPath := filepath.Join(dir, entry.Name())

		// Performance optimization: check type from DirEntry (no Lstat syscall)
		entryType := entry.Type()

		if entryType&os.ModeSymlink != 0 {
			// It's a symlink - check if orphaned
			s.checkForOrphanedLink(ctx, fullPath, m, linkSet, ignoreSet, issues, stats)
		} else if entry.IsDir() {
			// It's a directory - recurse
			s.scanDirectoryRecursive(ctx, fullPath, m, linkSet, ignoreSet, scanCfg, issues, stats)
		}
		// Regular files are ignored (no need to check)
	}
	return nil
}

// shouldSkipEntry checks if directory entry should be skipped.
func (s *DoctorService) shouldSkipEntry(entry DirEntry) bool {
	return entry.Name() == ".dot-manifest.json"
}

// scanDirectoryRecursive recursively scans subdirectory.
func (s *DoctorService) scanDirectoryRecursive(
	ctx context.Context,
	dir string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	ignoreSet *ignore.IgnoreSet,
	scanCfg ScanConfig,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	err := s.scanForOrphanedLinksWithLimits(ctx, dir, m, linkSet, ignoreSet, scanCfg, issues, stats)
	if err != nil {
		// Continue on error - best effort scanning
		s.logger.Warn(ctx, "recursive_scan_failed", "dir", dir, "error", err)
	}
}

// checkForOrphanedLink checks if symlink is orphaned (not in manifest) and validates target.
// Note: This function assumes fullPath is already confirmed to be a symlink by the caller.
func (s *DoctorService) checkForOrphanedLink(
	ctx context.Context,
	fullPath string,
	m *manifest.Manifest,
	linkSet map[string]bool,
	ignoreSet *ignore.IgnoreSet,
	issues *[]Issue,
	stats *DiagnosticStats,
) {
	relPath, err := filepath.Rel(s.targetDir, fullPath)
	if err != nil {
		relPath = fullPath
	}

	normalizedRel := filepath.ToSlash(relPath)
	normalizedFull := filepath.ToSlash(fullPath)
	managed := linkSet[normalizedRel] || linkSet[normalizedFull]

	if !managed {
		// Check if this link is explicitly ignored
		if m.Doctor != nil {
			if _, ignored := m.Doctor.IgnoredLinks[relPath]; ignored {
				return // Skip ignored link
			}
		}

		// Check if this link matches an ignore pattern
		if ignoreSet != nil && ignoreSet.ShouldIgnore(relPath) {
			return // Skip link matching ignore pattern
		}

		stats.TotalLinks++
		stats.OrphanedLinks++

		// Check if the orphaned symlink's target exists
		target, err := s.fs.ReadLink(ctx, fullPath)
		if err == nil {
			// Resolve target to absolute path
			var absTarget string
			if filepath.IsAbs(target) {
				absTarget = target
			} else {
				absTarget = filepath.Join(filepath.Dir(fullPath), target)
			}

			// Check if target exists
			_, err = s.fs.Stat(ctx, absTarget)
			if err != nil {
				if os.IsNotExist(err) {
					// Orphaned and broken - still classify as orphaned since it's unmanaged
					stats.BrokenLinks++
					*issues = append(*issues, Issue{
						Severity:   SeverityError,
						Type:       IssueOrphanedLink,
						Path:       relPath,
						Message:    "Unmanaged symlink with broken target: " + target,
						Suggestion: "Remove manually or fix target, then use 'dot adopt' to manage",
					})
					return
				}
			}
		}

		// Orphaned but target exists (or couldn't check)
		*issues = append(*issues, Issue{
			Severity:   SeverityWarning,
			Type:       IssueOrphanedLink,
			Path:       relPath,
			Message:    "Symlink not managed by dot",
			Suggestion: "Remove manually or use 'dot adopt' to bring under management",
		})
	}
}

// extractManagedDirectories returns unique directories containing managed links.
func extractManagedDirectories(m *manifest.Manifest) []string {
	dirSet := make(map[string]bool)
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			dir := filepath.Dir(link)
			for dir != "." && dir != "/" && dir != "" {
				dirSet[dir] = true
				dir = filepath.Dir(dir)
			}
			dirSet["."] = true
		}
	}
	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	return dirs
}

// filterDescendants removes directories that are descendants of other directories.
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
			if err == nil && rel != "." && !filepath.IsAbs(rel) && rel[0] != '.' {
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

// buildManagedLinkSet creates a set for O(1) link lookup.
func buildManagedLinkSet(m *manifest.Manifest) map[string]bool {
	linkSet := make(map[string]bool)
	for _, pkgInfo := range m.Packages {
		for _, link := range pkgInfo.Links {
			normalized := filepath.ToSlash(link)
			linkSet[normalized] = true
		}
	}
	return linkSet
}

// buildIgnoreSet creates an ignore set from manifest doctor state.
func (s *DoctorService) buildIgnoreSet(m *manifest.Manifest) *ignore.IgnoreSet {
	ignoreSet := ignore.NewIgnoreSet()

	if m.Doctor == nil {
		return ignoreSet
	}

	// Add patterns from manifest
	for _, pattern := range m.Doctor.IgnoredPatterns {
		// Ignore errors - best effort matching
		_ = ignoreSet.Add(pattern)
	}

	return ignoreSet
}

// calculateDepth returns the directory depth relative to target directory.
func calculateDepth(path, targetDir string) int {
	path = filepath.Clean(path)
	targetDir = filepath.Clean(targetDir)
	if path == targetDir {
		return 0
	}
	rel, err := filepath.Rel(targetDir, path)
	if err != nil || rel == "." {
		return 0
	}
	depth := 0
	for _, c := range rel {
		if c == filepath.Separator {
			depth++
		}
	}
	if rel != "" && rel != "." {
		depth++
	}
	return depth
}

// shouldSkipDirectory checks if a directory should be skipped based on patterns.
func shouldSkipDirectory(path string, skipPatterns []string) bool {
	base := filepath.Base(path)
	for _, pattern := range skipPatterns {
		if base == pattern {
			return true
		}
		if filepath.Base(filepath.Dir(path)) == pattern {
			return true
		}
	}
	return false
}
