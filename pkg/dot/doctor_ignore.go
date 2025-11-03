package dot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jamesainslie/dot/internal/manifest"
)

// IgnoreLink adds a symlink to the ignore list.
func (s *DoctorService) IgnoreLink(ctx context.Context, linkPath, reason string) error {
	targetPath, err := s.getTargetPath()
	if err != nil {
		return err
	}

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return manifestResult.UnwrapErr()
	}
	m := manifestResult.Unwrap()

	// Read link target
	fullPath := filepath.Join(s.targetDir, linkPath)
	target, err := s.fs.ReadLink(ctx, fullPath)
	if err != nil {
		return fmt.Errorf("cannot read link: %w", err)
	}

	m.AddIgnoredLink(linkPath, target, reason)

	s.logger.Info(ctx, "ignored_link", "path", linkPath, "target", target)

	return s.manifestSvc.Save(ctx, targetPath, m)
}

// IgnorePattern adds a glob pattern to ignore list.
func (s *DoctorService) IgnorePattern(ctx context.Context, pattern string) error {
	targetPath, err := s.getTargetPath()
	if err != nil {
		return err
	}

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return manifestResult.UnwrapErr()
	}
	m := manifestResult.Unwrap()

	m.AddIgnoredPattern(pattern)

	s.logger.Info(ctx, "ignored_pattern", "pattern", pattern)

	return s.manifestSvc.Save(ctx, targetPath, m)
}

// UnignoreLink removes a symlink from ignore list.
func (s *DoctorService) UnignoreLink(ctx context.Context, linkPath string) error {
	targetPath, err := s.getTargetPath()
	if err != nil {
		return err
	}

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return manifestResult.UnwrapErr()
	}
	m := manifestResult.Unwrap()

	if !m.RemoveIgnoredLink(linkPath) {
		return fmt.Errorf("link not in ignore list: %s", linkPath)
	}

	s.logger.Info(ctx, "unignored_link", "path", linkPath)

	return s.manifestSvc.Save(ctx, targetPath, m)
}

// ListIgnored returns all ignored links and patterns.
func (s *DoctorService) ListIgnored(ctx context.Context) (map[string]manifest.IgnoredLink, []string, error) {
	targetPath, err := s.getTargetPath()
	if err != nil {
		return nil, nil, err
	}

	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		return nil, nil, manifestResult.UnwrapErr()
	}
	m := manifestResult.Unwrap()

	if m.Doctor == nil {
		return make(map[string]manifest.IgnoredLink), []string{}, nil
	}

	return m.Doctor.IgnoredLinks, m.Doctor.IgnoredPatterns, nil
}
