package dot

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/bootstrap"
	"github.com/yaklabco/dot/internal/cli/selector"
)

func TestNewCloneService(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	manageSvc := &ManageService{}
	cloner := adapters.NewGoGitCloner()
	sel := selector.NewInteractiveSelector(os.Stdin, os.Stdout)

	svc := newCloneService(fs, logger, manageSvc, cloner, sel, "/packages", "/home", false)

	assert.NotNil(t, svc)
	assert.Equal(t, "/packages", svc.packageDir)
	assert.Equal(t, "/home", svc.targetDir)
}

func TestCloneService_ValidatePackageDir(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	t.Run("empty directory is valid", func(t *testing.T) {
		err := fs.MkdirAll(ctx, "/packages", 0755)
		require.NoError(t, err)

		err = validatePackageDir(ctx, fs, "/packages", false)
		assert.NoError(t, err)
	})

	t.Run("non-existent directory is valid", func(t *testing.T) {
		err := validatePackageDir(ctx, fs, "/nonexistent", false)
		assert.NoError(t, err)
	})

	t.Run("directory with files fails", func(t *testing.T) {
		err := fs.MkdirAll(ctx, "/packages2", 0755)
		require.NoError(t, err)
		err = fs.WriteFile(ctx, "/packages2/file.txt", []byte("test"), 0644)
		require.NoError(t, err)

		err = validatePackageDir(ctx, fs, "/packages2", false)
		assert.Error(t, err)
		assert.IsType(t, ErrPackageDirNotEmpty{}, err)
	})

	t.Run("force flag allows non-empty directory", func(t *testing.T) {
		err := fs.MkdirAll(ctx, "/packages3", 0755)
		require.NoError(t, err)
		err = fs.WriteFile(ctx, "/packages3/file.txt", []byte("test"), 0644)
		require.NoError(t, err)

		err = validatePackageDir(ctx, fs, "/packages3", true)
		assert.NoError(t, err)
	})
}

func TestCloneService_SelectPackages_WithProfile(t *testing.T) {
	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim", "dot-zsh"},
			},
		},
	}

	packages, err := selectPackagesFromProfile(config, "minimal")
	require.NoError(t, err)
	assert.Equal(t, []string{"dot-vim", "dot-zsh"}, packages)
}

func TestCloneService_SelectPackages_ProfileNotFound(t *testing.T) {
	config := bootstrap.Config{
		Version:  "1.0",
		Packages: []bootstrap.PackageSpec{{Name: "dot-vim"}},
	}

	_, err := selectPackagesFromProfile(config, "nonexistent")
	assert.Error(t, err)
	assert.IsType(t, ErrProfileNotFound{}, err)
}

func TestCloneService_SelectPackages_Interactive(t *testing.T) {
	input := strings.NewReader("1,2\n")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	packages := []string{"dot-vim", "dot-zsh", "dot-tmux"}
	selected, err := sel.Select(context.Background(), packages)
	require.NoError(t, err)

	assert.Equal(t, []string{"dot-vim", "dot-zsh"}, selected)
}

func TestCloneService_FilterByPlatform(t *testing.T) {
	packages := []bootstrap.PackageSpec{
		{Name: "all-platforms"},
		{Name: "linux-only", Platform: []string{"linux"}},
		{Name: "darwin-only", Platform: []string{"darwin"}},
	}

	currentPlatform := runtime.GOOS

	filtered := bootstrap.FilterPackagesByPlatform(packages, currentPlatform)

	// Verify "all-platforms" is always included
	names := make([]string, 0, len(filtered))
	for _, p := range filtered {
		names = append(names, p.Name)
	}
	assert.Contains(t, names, "all-platforms")

	// Verify platform-specific packages are filtered correctly
	if currentPlatform == "linux" {
		assert.Contains(t, names, "linux-only")
		assert.NotContains(t, names, "darwin-only")
	} else if currentPlatform == "darwin" {
		assert.Contains(t, names, "darwin-only")
		assert.NotContains(t, names, "linux-only")
	}
}

func TestCloneService_BuildRepositoryInfo(t *testing.T) {
	url := "https://github.com/user/dotfiles"
	branch := "main"
	beforeClone := time.Now()

	info := buildRepositoryInfo(url, branch, "abc123def456")

	assert.Equal(t, url, info.URL)
	assert.Equal(t, branch, info.Branch)
	assert.Equal(t, "abc123def456", info.CommitSHA)
	assert.True(t, info.ClonedAt.After(beforeClone.Add(-time.Second)))
	assert.True(t, info.ClonedAt.Before(time.Now().Add(time.Second)))
}

func TestCloneService_ExtractPackageNames(t *testing.T) {
	packages := []bootstrap.PackageSpec{
		{Name: "dot-vim"},
		{Name: "dot-zsh"},
		{Name: "dot-tmux"},
	}

	names := extractPackageNames(packages)
	assert.Equal(t, []string{"dot-vim", "dot-zsh", "dot-tmux"}, names)
}

func TestIntersectPackages(t *testing.T) {
	t.Run("filters to only allowed packages", func(t *testing.T) {
		packages := []string{"vim", "zsh", "tmux", "emacs"}
		allowed := []string{"vim", "tmux"}

		result := intersectPackages(packages, allowed)
		assert.Equal(t, []string{"vim", "tmux"}, result)
	})

	t.Run("preserves order from first list", func(t *testing.T) {
		packages := []string{"c", "a", "b"}
		allowed := []string{"a", "b", "c"}

		result := intersectPackages(packages, allowed)
		assert.Equal(t, []string{"c", "a", "b"}, result)
	})

	t.Run("returns empty when no matches", func(t *testing.T) {
		packages := []string{"vim", "zsh"}
		allowed := []string{"tmux", "emacs"}

		result := intersectPackages(packages, allowed)
		assert.Empty(t, result)
	})

	t.Run("returns all when all match", func(t *testing.T) {
		packages := []string{"vim", "zsh", "tmux"}
		allowed := []string{"vim", "zsh", "tmux", "emacs"}

		result := intersectPackages(packages, allowed)
		assert.Equal(t, []string{"vim", "zsh", "tmux"}, result)
	})

	t.Run("handles empty packages list", func(t *testing.T) {
		packages := []string{}
		allowed := []string{"vim", "zsh"}

		result := intersectPackages(packages, allowed)
		assert.Empty(t, result)
	})

	t.Run("handles empty allowed list", func(t *testing.T) {
		packages := []string{"vim", "zsh"}
		allowed := []string{}

		result := intersectPackages(packages, allowed)
		assert.Empty(t, result)
	})
}

func TestCloneService_LoadBootstrapConfig_Found(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory first
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Create bootstrap config
	configContent := `version: "1.0"
packages:
  - name: dot-vim
    required: true
`
	err = fs.WriteFile(ctx, "/packages/.dotbootstrap.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	config, found, err := loadBootstrapConfig(ctx, fs, "/packages")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "1.0", config.Version)
	assert.Len(t, config.Packages, 1)
}

func TestCloneService_LoadBootstrapConfig_NotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	config, found, err := loadBootstrapConfig(ctx, fs, "/packages")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, bootstrap.Config{}, config)
}

func TestCloneService_LoadBootstrapConfig_Invalid(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directory first
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Create invalid config
	invalidConfig := `this is not valid yaml: [unclosed`
	err = fs.WriteFile(ctx, "/packages/.dotbootstrap.yaml", []byte(invalidConfig), 0644)
	require.NoError(t, err)

	_, _, err = loadBootstrapConfig(ctx, fs, "/packages")
	assert.Error(t, err)
	assert.IsType(t, ErrInvalidBootstrap{}, err)
}

func TestCloneService_DiscoverPackages(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create package directories
	err := fs.MkdirAll(ctx, "/packages/dot-vim", 0755)
	require.NoError(t, err)
	err = fs.MkdirAll(ctx, "/packages/dot-zsh", 0755)
	require.NoError(t, err)
	err = fs.WriteFile(ctx, "/packages/README.md", []byte("test"), 0644)
	require.NoError(t, err)

	packages, err := discoverPackages(ctx, fs, "/packages")
	require.NoError(t, err)

	// Should only find directories, not files
	assert.Contains(t, packages, "dot-vim")
	assert.Contains(t, packages, "dot-zsh")
	assert.NotContains(t, packages, "README.md")
}

func TestCloneService_SelectPackagesWithBootstrap_DefaultProfile(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Defaults: bootstrap.Defaults{
			Profile: "minimal",
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim", "dot-zsh"},
			},
		},
	}

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{})
	require.NoError(t, err)
	assert.Equal(t, []string{"dot-vim", "dot-zsh"}, packages)
}

func TestCloneService_SelectPackagesWithBootstrap_ExplicitProfile(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim"},
			},
		},
	}

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{Profile: "minimal"})
	require.NoError(t, err)
	assert.Equal(t, []string{"dot-vim"}, packages)
}

func TestCloneService_SelectPackagesWithBootstrap_ProfileWithPlatformFilter(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	currentOS := runtime.GOOS
	otherOS := "windows"
	if currentOS == "windows" {
		otherOS = "linux"
	}

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "cross-platform"},
			{Name: "current-os-only", Platform: []string{currentOS}},
			{Name: "other-os-only", Platform: []string{otherOS}},
		},
		Profiles: map[string]bootstrap.Profile{
			"all": {
				Description: "All packages including platform-specific",
				Packages:    []string{"cross-platform", "current-os-only", "other-os-only"},
			},
		},
	}

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{Profile: "all"})
	require.NoError(t, err)

	// Should only include cross-platform and current-os-only, not other-os-only
	assert.ElementsMatch(t, []string{"cross-platform", "current-os-only"}, packages)
	assert.NotContains(t, packages, "other-os-only")
}

func TestCloneService_SelectPackagesWithBootstrap_DefaultProfileWithPlatformFilter(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	currentOS := runtime.GOOS

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "common"},
			{Name: "linux-specific", Platform: []string{"linux"}},
			{Name: "darwin-specific", Platform: []string{"darwin"}},
			{Name: "windows-specific", Platform: []string{"windows"}},
		},
		Defaults: bootstrap.Defaults{
			Profile: "complete",
		},
		Profiles: map[string]bootstrap.Profile{
			"complete": {
				Description: "All packages",
				Packages:    []string{"common", "linux-specific", "darwin-specific", "windows-specific"},
			},
		},
	}

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{})
	require.NoError(t, err)

	// Should always include common
	assert.Contains(t, packages, "common")

	// Should only include platform-specific package for current OS
	switch currentOS {
	case "linux":
		assert.Contains(t, packages, "linux-specific")
		assert.NotContains(t, packages, "darwin-specific")
		assert.NotContains(t, packages, "windows-specific")
	case "darwin":
		assert.Contains(t, packages, "darwin-specific")
		assert.NotContains(t, packages, "linux-specific")
		assert.NotContains(t, packages, "windows-specific")
	case "windows":
		assert.Contains(t, packages, "windows-specific")
		assert.NotContains(t, packages, "linux-specific")
		assert.NotContains(t, packages, "darwin-specific")
	}
}

func TestCloneService_SelectPackagesWithBootstrap_ProfileNotFoundError(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
		},
		Profiles: map[string]bootstrap.Profile{
			"existing": {
				Description: "Existing profile",
				Packages:    []string{"dot-vim"},
			},
		},
	}

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	// Test with explicit non-existent profile
	_, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{Profile: "nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile not found")

	// Test with default non-existent profile
	config.Defaults.Profile = "nonexistent-default"
	_, err = svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile not found")
}

func TestCloneService_SelectPackagesWithBootstrap_DefaultProfilePriority(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
			{Name: "dot-git"},
		},
		Defaults: bootstrap.Defaults{
			Profile: "minimal",
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim", "dot-zsh"},
			},
		},
	}

	// Use empty input to simulate what would happen if selector was called
	// (it would fail or wait for input)
	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	// With Interactive=false and a default profile configured,
	// the default profile should be used even if terminal is interactive.
	// The selector.Select should NOT be called.
	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{Interactive: false})
	require.NoError(t, err)
	// Should use default profile packages, not all packages
	assert.Equal(t, []string{"dot-vim", "dot-zsh"}, packages)
	// Verify selector was not called by checking output is empty
	assert.Empty(t, output.String())
}

func TestCloneService_SelectPackagesWithBootstrap_ExplicitInteractiveOverridesDefault(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	config := bootstrap.Config{
		Version: "1.0",
		Packages: []bootstrap.PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Defaults: bootstrap.Defaults{
			Profile: "minimal",
		},
		Profiles: map[string]bootstrap.Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim"},
			},
		},
	}

	// Provide input to simulate user selecting packages
	input := strings.NewReader("1,2\n") // Select vim and zsh
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	// With Interactive=true, should prompt even if default profile exists
	packages, err := svc.selectPackagesWithBootstrap(ctx, config, CloneOptions{Interactive: true})
	require.NoError(t, err)
	// Should use user selection, not default profile
	assert.ElementsMatch(t, []string{"dot-vim", "dot-zsh"}, packages)
	// Verify selector was called by checking output contains prompt
	assert.Contains(t, output.String(), "Package Selection")
}

func TestCloneService_SelectPackagesWithoutBootstrap_AllPackages(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create package directories
	err := fs.MkdirAll(ctx, "/packages/dot-vim", 0755)
	require.NoError(t, err)
	err = fs.MkdirAll(ctx, "/packages/dot-zsh", 0755)
	require.NoError(t, err)

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	// Non-interactive should install all
	packages, err := svc.selectPackagesWithoutBootstrap(ctx, CloneOptions{})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"dot-vim", "dot-zsh"}, packages)
}

func TestCloneService_SelectPackagesWithoutBootstrap_NoPackages(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create empty directory
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	input := strings.NewReader("")
	output := &strings.Builder{}
	sel := selector.NewInteractiveSelector(input, output)

	svc := newCloneService(fs, logger, nil, nil, sel, "/packages", "/home", false)

	packages, err := svc.selectPackagesWithoutBootstrap(ctx, CloneOptions{})
	require.NoError(t, err)
	assert.Empty(t, packages)
}

func TestCloneOptions_Defaults(t *testing.T) {
	opts := CloneOptions{}

	assert.Empty(t, opts.Profile)
	assert.False(t, opts.Interactive)
	assert.False(t, opts.Force)
	assert.Empty(t, opts.Branch)
}

func TestCloneOptions_WithValues(t *testing.T) {
	opts := CloneOptions{
		Profile:     "minimal",
		Interactive: true,
		Force:       true,
		Branch:      "develop",
	}

	assert.Equal(t, "minimal", opts.Profile)
	assert.True(t, opts.Interactive)
	assert.True(t, opts.Force)
	assert.Equal(t, "develop", opts.Branch)
}

func TestCloneService_Clone_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create managed package directories to simulate a successful manage operation
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Mock git cloner
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			// Simulate cloning by creating a package directory
			return fs.MkdirAll(ctx, dest+"/dot-vim", 0755)
		},
	}

	// Mock selector
	selector := &mockPackageSelector{
		selectFn: func(ctx context.Context, packages []string) ([]string, error) {
			return []string{"dot-vim"}, nil
		},
	}

	// Create a simple ManageService that doesn't do anything
	manageSvc := &ManageService{
		fs:         fs,
		logger:     logger,
		packageDir: "/packages",
		targetDir:  "/home",
		dryRun:     true, // Dry run to avoid actual file operations
	}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", true)

	err = svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{
		Branch: "main",
	})

	require.NoError(t, err)
}

func TestCloneService_Clone_PackageDirNotEmpty(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create non-empty package directory
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)
	err = fs.WriteFile(ctx, "/packages/existing-file.txt", []byte("test"), 0644)
	require.NoError(t, err)

	cloner := &mockGitCloner{}
	selector := &mockPackageSelector{}
	manageSvc := &ManageService{}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", false)

	err = svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{})

	assert.Error(t, err)
	assert.IsType(t, ErrPackageDirNotEmpty{}, err)
}

func TestCloneService_Clone_CloneFails(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Mock git cloner that fails
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			return assert.AnError
		},
	}

	selector := &mockPackageSelector{}
	manageSvc := &ManageService{}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", false)

	err := svc.Clone(ctx, "https://github.com/user/invalid", CloneOptions{})

	assert.Error(t, err)
	assert.IsType(t, ErrCloneFailed{}, err)
}

func TestCloneService_Clone_WithBootstrap(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Create package directory
	err := fs.MkdirAll(ctx, "/packages", 0755)
	require.NoError(t, err)

	// Mock git cloner that creates bootstrap config
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			// Create package directories
			_ = fs.MkdirAll(ctx, dest+"/dot-vim", 0755)
			_ = fs.MkdirAll(ctx, dest+"/dot-zsh", 0755)

			// Create bootstrap config
			bootstrapContent := `version: "1.0"
packages:
  - name: dot-vim
    required: true
  - name: dot-zsh
    required: false
profiles:
  minimal:
    description: "Minimal setup"
    packages:
      - dot-vim
`
			return fs.WriteFile(ctx, dest+"/.dotbootstrap.yaml", []byte(bootstrapContent), 0644)
		},
	}

	// Mock selector (shouldn't be called because profile is specified)
	selector := &mockPackageSelector{}

	manageSvc := &ManageService{
		fs:         fs,
		logger:     logger,
		packageDir: "/packages",
		targetDir:  "/home",
		dryRun:     true,
	}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", true)

	err = svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{
		Profile: "minimal",
		Branch:  "main",
	})

	require.NoError(t, err)
}

func TestCloneService_Clone_NoPackagesSelected(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Mock git cloner
	cloner := &mockGitCloner{
		cloneFn: func(ctx context.Context, url string, dest string, opts adapters.CloneOptions) error {
			// Create empty clone
			return fs.MkdirAll(ctx, dest, 0755)
		},
	}

	// Mock selector that returns no packages
	selector := &mockPackageSelector{
		selectFn: func(ctx context.Context, packages []string) ([]string, error) {
			return []string{}, nil
		},
	}

	manageSvc := &ManageService{}

	svc := newCloneService(fs, logger, manageSvc, cloner, selector, "/packages", "/home", false)

	err := svc.Clone(ctx, "https://github.com/user/dotfiles", CloneOptions{
		Interactive: true,
	})

	require.NoError(t, err) // Should succeed even with no packages
}

func TestCloneService_GetCommitSHA(t *testing.T) {
	t.Skip("getCommitSHA requires git repository - tested in integration tests")
}

func TestGetCurrentBranch(t *testing.T) {
	t.Run("reads main branch from HEAD", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		headContent := "ref: refs/heads/main\n"
		headPath := gitDir + "/HEAD"
		require.NoError(t, os.WriteFile(headPath, []byte(headContent), 0644))

		branch, err := getCurrentBranch(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("reads master branch from HEAD", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		headContent := "ref: refs/heads/master\n"
		headPath := gitDir + "/HEAD"
		require.NoError(t, os.WriteFile(headPath, []byte(headContent), 0644))

		branch, err := getCurrentBranch(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "master", branch)
	})

	t.Run("reads feature branch from HEAD", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		headContent := "ref: refs/heads/feature-branch-name\n"
		headPath := gitDir + "/HEAD"
		require.NoError(t, os.WriteFile(headPath, []byte(headContent), 0644))

		branch, err := getCurrentBranch(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "feature-branch-name", branch)
	})

	t.Run("handles HEAD without trailing newline", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		headContent := "ref: refs/heads/develop"
		headPath := gitDir + "/HEAD"
		require.NoError(t, os.WriteFile(headPath, []byte(headContent), 0644))

		branch, err := getCurrentBranch(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "develop", branch)
	})

	t.Run("returns error for detached HEAD with SHA", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		headContent := "a1b2c3d4e5f6789012345678901234567890abcd\n"
		headPath := gitDir + "/HEAD"
		require.NoError(t, os.WriteFile(headPath, []byte(headContent), 0644))

		branch, err := getCurrentBranch(tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "detached HEAD")
		assert.Empty(t, branch)
	})

	t.Run("returns error when HEAD file missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		branch, err := getCurrentBranch(tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read HEAD file")
		assert.Empty(t, branch)
	})

	t.Run("returns error when .git directory missing", func(t *testing.T) {
		tmpDir := t.TempDir()

		branch, err := getCurrentBranch(tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read HEAD file")
		assert.Empty(t, branch)
	})

	t.Run("returns error for empty branch name", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		headContent := "ref: refs/heads/\n"
		headPath := gitDir + "/HEAD"
		require.NoError(t, os.WriteFile(headPath, []byte(headContent), 0644))

		branch, err := getCurrentBranch(tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty branch name")
		assert.Empty(t, branch)
	})

	t.Run("returns error for malformed HEAD content", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := tmpDir + "/.git"
		require.NoError(t, os.Mkdir(gitDir, 0755))

		headContent := "invalid content\n"
		headPath := gitDir + "/HEAD"
		require.NoError(t, os.WriteFile(headPath, []byte(headContent), 0644))

		branch, err := getCurrentBranch(tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "detached HEAD or unexpected format")
		assert.Empty(t, branch)
	})
}

func TestGetAuthMethodName(t *testing.T) {
	tests := []struct {
		name     string
		auth     adapters.AuthMethod
		expected string
	}{
		{
			name:     "nil auth returns none",
			auth:     nil,
			expected: "none",
		},
		{
			name:     "NoAuth returns none",
			auth:     adapters.NoAuth{},
			expected: "none",
		},
		{
			name:     "TokenAuth returns token",
			auth:     adapters.TokenAuth{Token: "ghp_test123"},
			expected: "token",
		},
		{
			name:     "SSHAuth returns ssh",
			auth:     adapters.SSHAuth{},
			expected: "ssh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAuthMethodName(tt.auth)
			assert.Equal(t, tt.expected, result)
		})
	}
}
