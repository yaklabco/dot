package dot

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/yaklabco/dot/internal/doctor"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/manifest"
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
		adoptSvc:      nil,
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

// DiagnosticMode defines the depth of diagnostic checks to perform.
type DiagnosticMode string

const (
	// DiagnosticFast performs only essential checks (managed packages, manifest integrity).
	DiagnosticFast DiagnosticMode = "fast"
	// DiagnosticDeep performs comprehensive checks including orphan detection.
	DiagnosticDeep DiagnosticMode = "deep"
)

// Doctor performs health checks with default scan configuration.
func (s *DoctorService) Doctor(ctx context.Context) (DiagnosticReport, error) {
	return s.DoctorWithMode(ctx, DiagnosticDeep, DefaultScanConfig())
}

// DoctorWithScan performs health checks with explicit scan configuration.
func (s *DoctorService) DoctorWithScan(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error) {
	return s.DoctorWithMode(ctx, DiagnosticDeep, scanCfg)
}

// DoctorWithMode performs health checks with explicit mode and configuration.
func (s *DoctorService) DoctorWithMode(ctx context.Context, mode DiagnosticMode, scanCfg ScanConfig) (DiagnosticReport, error) {
	engine := doctor.NewDiagnosticEngine()

	// Helper adapters for check constructors
	newTargetPath := &doctorTargetPathCreatorAdapter{}

	// Adapter for ManifestLoader interface
	manifestLoader := &manifestLoaderAdapter{svc: s.manifestSvc}

	// Adapter for LinkHealthChecker interface
	healthChecker := &linkHealthCheckerAdapter{checker: s.healthChecker}

	// Adapter for FS interface
	fsAdapter := &doctorFSAdapter{fs: s.fs}

	// Convert scanCfg to doctor.ScanConfig
	doctorScanCfg := doctor.ScanConfig{
		Mode:         doctor.ScanMode(scanCfg.Mode),
		MaxWorkers:   scanCfg.MaxWorkers,
		MaxDepth:     scanCfg.MaxDepth,
		MaxIssues:    scanCfg.MaxIssues,
		ScopeToDirs:  scanCfg.ScopeToDirs,
		SkipPatterns: scanCfg.SkipPatterns,
	}

	// Fast mode: Essential checks only
	// 1. Manifest Integrity Check
	engine.RegisterCheck(doctor.NewManifestIntegrityCheck(fsAdapter, manifestLoader, s.targetDir, newTargetPath, IsManifestNotFoundError))

	// 2. Managed Packages Check
	engine.RegisterCheck(doctor.NewManagedPackageCheck(fsAdapter, manifestLoader, healthChecker, s.targetDir, newTargetPath, IsManifestNotFoundError))

	// Deep mode: Additional comprehensive checks
	if mode == DiagnosticDeep {
		// 3. Orphan Check (only if not disabled)
		if scanCfg.Mode != ScanOff {
			engine.RegisterCheck(doctor.NewOrphanCheck(
				doctor.WithFS(fsAdapter),
				doctor.WithManifestLoader(manifestLoader),
				doctor.WithTargetDir(s.targetDir),
				doctor.WithScanConfig(doctorScanCfg),
				doctor.WithTargetPathCreator(newTargetPath),
			))
		}

		// 4. Platform Compatibility Check
		engine.RegisterCheck(doctor.NewPlatformCheck(fsAdapter, manifestLoader, s.packageDir, s.targetDir, newTargetPath))
	}

	// Execute checks with parallel execution for performance
	report, err := engine.Run(ctx, doctor.RunOptions{
		Parallel: true,
	})
	if err != nil {
		return DiagnosticReport{}, err
	}

	// Check for system-level check execution errors and propagate them
	for _, result := range report.Results {
		for _, issue := range result.Issues {
			if issue.Code == "CHECK_EXECUTION_ERROR" {
				return DiagnosticReport{}, fmt.Errorf("%s: %s", result.CheckName, issue.Message)
			}
		}
	}

	// Transform report to legacy DiagnosticReport for CLI compatibility
	return s.transformReport(report), nil
}

// PreFlightCheck performs quick checks before an operation.
func (s *DoctorService) PreFlightCheck(ctx context.Context, packages []string) (DiagnosticReport, error) {
	engine := doctor.NewDiagnosticEngine()

	// Adapter for FS interface
	fsAdapter := &doctorFSAdapter{fs: s.fs}

	// Check permissions
	engine.RegisterCheck(doctor.NewPermissionCheck(fsAdapter, s.targetDir))

	// Check conflicts for specific packages if we knew their links
	// TODO: Implement ConflictCheck for the provided packages
	// For now, we ignore the 'packages' argument and just run the permission check
	// as an example of "PreFlight" capability.

	report, err := engine.Run(ctx, doctor.RunOptions{Parallel: true})
	if err != nil {
		return DiagnosticReport{}, err
	}
	return s.transformReport(report), nil
}

// aggregateStat adds an integer stat value to the total.
// It gracefully handles int, int64, and float64 types.
func aggregateStat(stats map[string]any, key string) int {
	if val, ok := stats[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// convertSeverity converts domain severity to public severity.
func convertSeverity(severity domain.IssueSeverity) IssueSeverity {
	switch severity {
	case domain.IssueSeverityError:
		return SeverityError
	case domain.IssueSeverityWarning:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// convertIssueType converts code string to IssueType.
func convertIssueType(code string) IssueType {
	// Normalize to lowercase for consistent mapping
	code = strings.ToLower(code)

	switch code {
	case "broken_link":
		return IssueBrokenLink
	case "orphaned_link":
		return IssueOrphanedLink
	case "wrong_target":
		return IssueWrongTarget
	case "permission", "permission_denied", "target_dir_not_writable", "target_dir_not_readable", "write_test_failed":
		return IssuePermission
	case "circular":
		return IssueCircular
	case "manifest_inconsistency", "no_manifest", "manifest_inconsistent", "check_execution_error":
		return IssueManifestInconsistency
	case "conflict_detected", "access_error":
		// Map conflict/access issues to a reasonable existing type
		return IssueManifestInconsistency
	case "metadata_read_error", "platform_incompatible", "test_fail", "test_issue", "target_dir_missing", "cleanup_failed":
		// Map additional error codes
		return IssueManifestInconsistency
	default:
		return IssueManifestInconsistency
	}
}

// extractSuggestion extracts suggestion from context.
func extractSuggestion(ctx map[string]any) string {
	if val, ok := ctx["suggestion"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// convertIssue converts domain issue to public issue.
func convertIssue(internalIssue domain.Issue) Issue {
	return Issue{
		Severity:   convertSeverity(internalIssue.Severity),
		Type:       convertIssueType(internalIssue.Code),
		Path:       internalIssue.Path,
		Message:    internalIssue.Message,
		Suggestion: extractSuggestion(internalIssue.Context),
	}
}

// determineOverallHealth determines overall health from status.
func determineOverallHealth(status domain.CheckStatus) HealthStatus {
	if status == domain.CheckStatusFail {
		return HealthErrors
	} else if status == domain.CheckStatusWarning {
		return HealthWarnings
	}
	return HealthOK
}

// transformReport converts internal engine report to public DiagnosticReport.
func (s *DoctorService) transformReport(internal doctor.DiagnosticReport) DiagnosticReport {
	// Count total issues for preallocation
	totalIssues := 0
	for _, res := range internal.Results {
		totalIssues += len(res.Issues)
	}

	issues := make([]Issue, 0, totalIssues)
	stats := DiagnosticStats{}

	for _, res := range internal.Results {
		stats.TotalLinks += aggregateStat(res.Stats, "total_links")
		stats.BrokenLinks += aggregateStat(res.Stats, "broken_links")
		stats.OrphanedLinks += aggregateStat(res.Stats, "orphaned_links")
		stats.ManagedLinks += aggregateStat(res.Stats, "managed_links")

		for _, internalIssue := range res.Issues {
			issues = append(issues, convertIssue(internalIssue))
		}
	}

	return DiagnosticReport{
		OverallHealth: determineOverallHealth(internal.OverallStatus),
		Issues:        issues,
		Statistics:    stats,
	}
}

// IsManifestNotFoundError checks if an error indicates a missing manifest.
func IsManifestNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, os.ErrNotExist)
}

// Note: Helper methods like getTargetPath, loadManifestOrCreateDefault are still useful if used by other methods
// but might be redundant if logic moved to checks.
// Keeping them if they are used by Triage or other methods not yet refactored.
// Triage still needs `loadManifestOrCreateDefault`.

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
		if IsManifestNotFoundError(err) {
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

// manifestLoaderAdapter adapts ManifestService to doctor.ManifestLoader interface.
type manifestLoaderAdapter struct {
	svc *ManifestService
}

func (a *manifestLoaderAdapter) Load(ctx context.Context, targetPath domain.TargetPath) domain.Result[manifest.Manifest] {
	return a.svc.Load(ctx, targetPath)
}

// linkHealthCheckerAdapter adapts HealthChecker to doctor.LinkHealthChecker interface.
type linkHealthCheckerAdapter struct {
	checker *HealthChecker
}

func (a *linkHealthCheckerAdapter) CheckLink(ctx context.Context, packageName, linkPath, packageDir string) doctor.LinkHealthResult {
	result := a.checker.CheckLink(ctx, packageName, linkPath, packageDir)

	var severity domain.IssueSeverity
	switch result.Severity {
	case SeverityError:
		severity = domain.IssueSeverityError
	case SeverityWarning:
		severity = domain.IssueSeverityWarning
	default:
		severity = domain.IssueSeverityInfo
	}

	return doctor.LinkHealthResult{
		IsHealthy:  result.IsHealthy,
		Severity:   severity,
		IssueType:  doctor.IssueType(result.IssueType.String()),
		Message:    result.Message,
		Suggestion: result.Suggestion,
	}
}

// doctorTargetPathCreatorAdapter implements doctor.TargetPathCreator.
type doctorTargetPathCreatorAdapter struct{}

func (a *doctorTargetPathCreatorAdapter) NewTargetPath(path string) domain.Result[domain.TargetPath] {
	r := NewTargetPath(path)
	if r.IsErr() {
		return domain.Err[domain.TargetPath](r.UnwrapErr())
	}
	return domain.Ok(r.Unwrap())
}

// doctorFSAdapter adapts pkg/dot/FS to internal/doctor/FS.
// Since domain.FileInfo and domain.DirEntry are now type aliases for fs.FileInfo
// and fs.DirEntry, this adapter can return values directly without wrapping.
type doctorFSAdapter struct {
	fs FS
}

func (a *doctorFSAdapter) Exists(ctx context.Context, path string) (bool, error) {
	return a.fs.Exists(ctx, path), nil
}

func (a *doctorFSAdapter) IsDir(ctx context.Context, path string) (bool, error) {
	return a.fs.IsDir(ctx, path)
}

func (a *doctorFSAdapter) Lstat(ctx context.Context, name string) (fs.FileInfo, error) {
	return a.fs.Lstat(ctx, name)
}

func (a *doctorFSAdapter) ReadDir(ctx context.Context, name string) ([]fs.DirEntry, error) {
	return a.fs.ReadDir(ctx, name)
}

func (a *doctorFSAdapter) ReadFile(ctx context.Context, name string) ([]byte, error) {
	return a.fs.ReadFile(ctx, name)
}

func (a *doctorFSAdapter) ReadLink(ctx context.Context, name string) (string, error) {
	return a.fs.ReadLink(ctx, name)
}

func (a *doctorFSAdapter) WriteFile(ctx context.Context, name string, data []byte, perm os.FileMode) error {
	return a.fs.WriteFile(ctx, name, data, perm)
}

func (a *doctorFSAdapter) Remove(ctx context.Context, name string) error {
	return a.fs.Remove(ctx, name)
}

func (a *doctorFSAdapter) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return a.fs.MkdirAll(ctx, path, perm)
}

func (a *doctorFSAdapter) Stat(ctx context.Context, name string) (fs.FileInfo, error) {
	return a.fs.Stat(ctx, name)
}
