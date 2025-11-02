<a name="unreleased"></a>
## [Unreleased]


<a name="v0.4.4"></a>
## [v0.4.4] - 2025-11-02
### Chore
- fix the bootstrap config
- fix the bootstrap config
- update IDE exclusions in .gitignore
- **build:** add fuzz and bench targets to Makefile

### Ci
- add vulnerability checking with govulncheck

### Docs
- add security policy with vulnerability disclosure
- remove stale and broken links from documentation
- **changelog:** update for v0.4.4 release
- **cli:** improve clone help text scannability
- **readme:** update adopt command documentation

### Feat
- **adapters:** integrate GitHub CLI authentication
- **api:** integrate backup system with manage service
- **cli:** add batch mode and async version check
- **cli:** add color helper for dynamic color detection
- **cli:** add profiling and diagnostics support
- **cli:** add command aliases for config
- **cli:** add pluralization helpers for output formatting
- **cli:** improve package selection UI with colors and layout
- **cli:** add interactive prompt utilities
- **clone:** derive directory from repository name like git clone
- **config:** add network configuration support
- **config:** add backup and overwrite configuration options
- **config:** add validation for network timeout fields
- **domain:** add FileDelete operation and fix FileBackup permissions
- **manifest:** add backup tracking to package metadata
- **pipeline:** add current state scanner for conflict detection
- **planner:** implement backup and overwrite conflict policies
- **retry:** add exponential backoff retry utility
- **updater:** add security validation for package managers
- **updater:** improve version checking with better error handling

### Fix
- **cli:** improve async version check and test file cleanup
- **cli:** resolve gosec G602 array bounds warnings in pager
- **test:** update tests for new conflict detection behavior
- **vuln:** exclude GO-2024-3295 from vulnerability checks

### Refactor
- **cli:** make cleanup grace period a constant
- **cli:** remove redundant goroutine wrapper
- **cli:** enhance unmanage output and add -y flag
- **cli:** improve success messages with proper pluralization

### Test
- add fuzz tests for config, domain, and ignore packages
- **backup:** add comprehensive integration tests for backup workflow
- **cli:** fix verification test expectations
- **cli:** improve signal handling test isolation
- **cli:** add golden tests for adopt and manage commands
- **cli:** add golden file testing framework
- **cli:** add comprehensive main package test coverage
- **cli:** add signal handling integration tests
- **clone:** add coverage for auth method name formatting
- **pipeline:** add coverage for operation mapping and state scanner
- **scanner:** add benchmarks for package scanning performance

### Pull Requests
- Merge pull request [#36](https://github.com/jamesainslie/dot/issues/36) from jamesainslie/feature-improve-testing
- Merge pull request [#35](https://github.com/jamesainslie/dot/issues/35) from jamesainslie/feature-improve-ux
- Merge pull request [#34](https://github.com/jamesainslie/dot/issues/34) from jamesainslie/feature-use-go-gh-sdk-for-auth

### BREAKING CHANGE

```
Remove automatic glob mode detection to eliminate ambiguous behavior. Users must now provide explicit package names when adopting multiple files.

Changes:
- Remove fileExists(), deriveCommonPackageName(), and commonPrefix()
- Simplify logic: single file = auto-naming, multiple = explicit
- Update help text with section headers for clarity
- Remove tests for deleted helper functions
- Update success message format

Before:
  dot adopt .git*  # Auto-detected package name from files

After:
  dot adopt git .git*  # Explicit package name required

This ensures predictable behavior and clearer package organization.
```



<a name="v0.4.3"></a>
## [v0.4.3] - 2025-10-13
### Docs
- **changelog:** update for v0.4.3 release
- **user:** add comprehensive upgrade and version management documentation

### Feat
- **cli:** integrate startup version check into root command
- **cli:** implement dot upgrade command with comprehensive testing
- **config:** add update configuration for package manager and version checking
- **updater:** add colorized output to update notification
- **updater:** implement startup version checking system
- **updater:** add version checker and package manager services

### Fix
- **updater:** address critical bugs and security issues
- **updater:** correct notification box alignment and version truncation

### Refactor
- **updater:** optimize color detection and ANSI stripping

### Test
- **quality:** increase coverage to 80.2% and fix quality issues

### Pull Requests
- Merge pull request [#33](https://github.com/jamesainslie/dot/issues/33) from jamesainslie/feature-upgrade-command


<a name="v0.4.2"></a>
## [v0.4.2] - 2025-10-12
### Docs
- **changelog:** update for v0.4.2 release
- **readme:** add clone and bootstrap commands to documentation

### Feat
- **cli:** refactor table rendering with go-pretty for professional UX
- **cli:** add progress tracking using go-pretty
- **cli:** add list and text formatting utilities
- **cli:** add table infrastructure using go-pretty
- **cli:** add muted colorization to doctor and unmanage commands
- **config:** wire up table_style configuration to all commands
- **doctor:** enable scoped orphan scanning by default and detect broken unmanaged links
- **pager:** add keyboard controls for interactive pagination
- **ui:** add automatic pagination to dot doctor output
- **ui:** add configuration toggle between modern and legacy table styles
- **unmanage:** add --all flag to unmanage all packages at once

### Fix
- **cli:** replace time.Sleep race with sync.WaitGroup in ProgressTracker
- **clone:** resolve silent errors and add comprehensive logging
- **pager:** remove blank lines left by status indicator after paging
- **ui:** add newline after table output for better terminal spacing
- **unmanage:** use filepath.Join for cross-platform path handling

### Perf
- **doctor:** optimize scan performance with parallel execution and smart filtering

### Refactor
- **cli:** improve config loading, progress tracker, and command UX
- **cli:** migrate from go-pretty to lipgloss v1.1.0
- **config:** simplify repository configuration loading
- **doctor:** enhance orphan scan with worker context and result collection

### Test
- **terminal:** add comprehensive tests for terminal detection

### Yak
- **ci:** have a simple - pass fail on test coverage

### Pull Requests
- Merge pull request [#32](https://github.com/jamesainslie/dot/issues/32) from jamesainslie/fix-some-missed-bugs
- Merge pull request [#31](https://github.com/jamesainslie/dot/issues/31) from jamesainslie/fix-silent-errors-when-cloning


<a name="v0.4.1"></a>
## [v0.4.1] - 2025-10-10
### Docs
- **changelog:** update for v0.4.1 release
- **clone:** add bootstrap subcommand documentation

### Feat
- **bootstrap:** implement bootstrap config generator
- **clone:** implement clone bootstrap subcommand
- **dot:** add bootstrap generation to Client facade

### Fix
- **bootstrap:** use installed parameter in YAML comments
- **bootstrap:** implement manifest parsing for from-manifest flag
- **bootstrap:** correct documentation URL in generated config header

### Refactor
- **bootstrap:** remove unused makeSet helper function

### Test
- **bootstrap:** add manifest filtering verification assertions
- **dot:** add tests for BootstrapService

### Pull Requests
- Merge pull request [#30](https://github.com/jamesainslie/dot/issues/30) from jamesainslie/feature-bootstrap-generation


<a name="v0.4.0"></a>
## [v0.4.0] - 2025-10-10
### Build
- **deps:** add go-git and golang.org/x/term dependencies

### Change
- **docs:** logo > transparent background

### Docs
- **changelog:** update for v0.4.0 release
- **changelog:** wrap BREAKING CHANGE content in code blocks
- **changelog:** regenerate from cleaned commit history
- **readme:** add clone command to Quick Start
- **user:** add clone command and bootstrap config documentation

### Feat
- **bootstrap:** add configuration schema and loader
- **cli:** implement dot clone command
- **cli:** add interactive package selector and terminal detection
- **client:** integrate CloneService into Client facade
- **clone:** improve branch detection and profile filtering
- **clone:** implement CloneService orchestrator
- **errors:** add clone-specific error types
- **git:** add git cloning with authentication support
- **manifest:** add repository tracking support

### Fix
- **adapters:** use correct HTTP Basic Auth format for tokens
- **clone:** improve string handling safety in git SHA parsing
- **clone:** use errors.As for wrapped error handling

### Test
- **adapters:** replace network-dependent tests with hermetic fixtures
- **integration:** add clone feature tests and fixtures

### Pull Requests
- Merge pull request [#29](https://github.com/jamesainslie/dot/issues/29) from jamesainslie/feature-clone-command


<a name="v0.3.1"></a>
## [v0.3.1] - 2025-10-09
### Build
- **makefile:** add uninstall target to remove installed binary

### Docs
- **changelog:** update for v0.3.1 release
- **readme:** update current version to v0.3.1


<a name="v0.3.0"></a>
## [v0.3.0] - 2025-10-09
### Build
- **makefile:** add coverage threshold validation to check target

### Docs
- update documentation for v0.3 flat structure
- **adopt:** add glob expansion examples to documentation
- **changelog:** update for v0.3.0 release
- **changelog:** fix BREAKING CHANGE formatting for v0.2.0
- **pkg:** update implementation status and test comments

### Feat
- **adopt:** add auto-naming and glob expansion modes
- **cli:** add comprehensive tab-completion and unmanage restoration

### Fix
- **ci:** remove path to golangci-lint
- **ci:** remove path to golangci-lint
- **planner:** resolve directory creation dependency ordering
- **tests:** skip permission test in CI environments
- **tests:** improve permission conflict test robustness

### Refactor
- **adopt:** implement flat package structure with consistent dot-prefix
- **adopt:** preserve leading dots in package naming
- **dotprefix:** work in progress on dot prefix refactoring
- **pkg:** replace MustParsePath with error handling in production code

### Pull Requests
- Merge pull request [#28](https://github.com/jamesainslie/dot/issues/28) from jamesainslie/refactor-dotprefix
- Merge pull request [#27](https://github.com/jamesainslie/dot/issues/27) from jamesainslie/fix-changelog-v0.2.0-formatting
- Merge pull request [#26](https://github.com/jamesainslie/dot/issues/26) from jamesainslie/pre-release-niggles

### BREAKING CHANGE

```
Adopt now creates flat package structure

Change adopt behavior to store directory contents at package root with
consistent 'dot-' prefix application.

Before:
~/dotfiles/ssh/dot-ssh/config → ~/.ssh

After:
~/dotfiles/dot-ssh/config → ~/.ssh

Changes:
- Package names preserve leading dots: .ssh → dot-ssh
- Directory contents stored at package root (flat structure)
- Apply dotfile translation to each file/directory
- Symlinks point to package root (not nested subdirectory)

Implementation:
- Add createDirectoryAdoptOperations for directory handling
- Add collectDirectoryFiles for recursive collection
- Add translatePathComponents for per-component translation
- Update unmanage restoration to handle both structures

Testing:
- All existing tests updated for new structure
- New tests for flat structure and nested dotfiles
- Regression tests preserve backward compatibility checks
- 80%+ coverage maintained

Refs: docs/planning/dot-prefix-refactoring-plan.md
```



<a name="v0.2.0"></a>
## [v0.2.0] - 2025-10-08
### Docs
- **changelog:** update for v0.2.0 release
- **developer:** add a mascot..because gopher
- **packages:** update user documentation for package name mapping

### Feat
- **cli:** display usage help on invalid flags and arguments
- **packages:** enable package name to target directory mapping

### Test
- **cli:** complete runtime error test assertions

### BREAKING CHANGE

```
Package structure requirements changed.

Before (v0.1.x):
dot-gnupg/
├── common.conf → ~/common.conf
└── public-keys.d/ → ~/public-keys.d/

After (v0.2.0):
dot-gnupg/
├── common.conf → ~/.gnupg/common.conf
└── public-keys.d/ → ~/.gnupg/public-keys.d/

Migration: Restructure packages to remove redundant nesting, or
opt-out by setting dotfile.package_name_mapping: false in config.

Rationale: Project is pre-1.0 (v0.1.1), establishing intuitive
design before API stability commitment in 1.0.0 release.
```



<a name="v0.1.1"></a>
## [v0.1.1] - 2025-10-08
### Chore
- remove documentation from source control
- remove reference docs from source control
- remove migration docs from source control
- add planning docs to gitignore
- remove planning and archive docs from source control
- **ci:** expunge emojis
- **ci:** expunge emojis
- **ci:** ignore some planning docs
- **ci:** expunge emojis
- **docs:** remove unwanted docs
- **hooks:** add pre-commit hook for test coverage enforcement

### Docs
- document refactoring
- add executive summary of completion
- update progress doc to reflect completion
- document Core completion
- add complete summary
- **architecture:** add comprehensive architecture documentation
- **architecture:** update for service-based architecture
- **changelog:** update README.md
- **changelog:** update for v0.1.1 release
- **changelog:** update for v0.1.1 release
- **changelog:** update for v0.1.1 release
- **developer:** add mermaid diagrams and comprehensive testing documentation
- **index:** update documentation index to reflect current structure
- **navigation:** add root README links to all child documentation
- **planning:** add progress checkpoint
- **planning:** add code quality improvements plan
- **readme:** fix broken documentation links
- **test:** update benchmark template to use testing.TB pattern

### Feat
- **config:** implement TOML marshal strategy
- **config:** implement JSON marshal strategy
- **config:** implement YAML marshal strategy
- **domain:** add chainable Result methods
- **pkg:** extract DoctorService from Client
- **pkg:** extract AdoptService from Client
- **pkg:** extract StatusService from Client
- **pkg:** extract UnmanageService from Client
- **pkg:** extract ManageService from Client
- **pkg:** extract ManifestService from Client
- **release:** automate release workflow with integrated changelog generation

### Fix
- **client:** properly propagate manifest errors in Doctor
- **client:** populate packages from plan when empty in updateManifest
- **config:** add missing KeyDoctorCheckPermissions constant
- **config:** honor MarshalOptions.Indent in TOML strategy
- **doctor:** detect and report permission errors on link targets
- **domain:** make path validator tests OS-aware for Windows
- **hooks:** check overall project coverage to match CI
- **hooks:** show linting output in pre-commit hook
- **hooks:** show test output in pre-commit hook
- **manage:** implement proper unmanage in planFullRemanage
- **manifest:** propagate non-not-found errors in Update
- **path:** add method forwarding to Path wrapper type
- **release:** move tag to amended commit in release workflow
- **status:** propagate non-not-found manifest errors
- **test:** strengthen PackageOperations assertion in exhaustive test
- **test:** rename ExecutionFailure test to match actual behavior
- **test:** correct comment in PlanOperationsEmpty test
- **test:** add Windows build constraints to Unix-specific tests
- **test:** add proper error handling to CLI integration tests
- **test:** add Windows compatibility to testutil symlink tests
- **test:** skip file mode test on Windows
- **test:** correct mock variadic parameter handling in ports_test

### Refactor
- **api:** replace Client interface with concrete struct
- **config:** migrate writer to use strategy pattern
- **config:** use permission constants
- **domain:** clean up temporary migration scripts
- **domain:** complete internal package migration and simplify pkg/dot
- **domain:** create internal/domain package structure
- **domain:** move Result monad to internal/domain
- **domain:** move Path and errors types to internal/domain
- **domain:** use validators in path constructors
- **domain:** move MustParsePath to testing.go
- **domain:** improve TraversalFreeValidator implementation
- **domain:** move all domain types to internal/domain
- **domain:** update all internal package imports to use internal/domain
- **domain:** use TargetPath for operation targets
- **domain:** format code and fix linter issues
- **hooks:** eliminate duplicate test run in pre-commit
- **path:** remove Path generic wrapper to eliminate code quality issue
- **pkg:** simplify scanForOrphanedLinks method
- **pkg:** convert Client to facade pattern
- **pkg:** extract helper methods in DoctorService
- **pkg:** simplify DoctorWithScan method
- **test:** improve benchmark tests with proper error handling

### Style
- fix goimports formatting

### Test
- **client:** add exhaustive tests for increased coverage margin
- **client:** add edge case tests for coverage buffer
- **client:** add comprehensive tests for pkg/dot Client struct
- **config:** add marshal strategy interface tests
- **config:** add default value constant tests
- **config:** add configuration key constant tests
- **domain:** add path validator tests
- **domain:** add Result unwrap helper tests
- **domain:** add error helper tests
- **domain:** add permission constant tests
- **integration:** implement integration test categories
- **integration:** implement comprehensive integration test infrastructure

### Pull Requests
- Merge pull request [#25](https://github.com/jamesainslie/dot/issues/25) from jamesainslie/feature-documentation-shiznickle
- Merge pull request [#24](https://github.com/jamesainslie/dot/issues/24) from jamesainslie/docs-update-document-index
- Merge pull request [#23](https://github.com/jamesainslie/dot/issues/23) from jamesainslie/docs-add-root-links
- Merge pull request [#22](https://github.com/jamesainslie/dot/issues/22) from jamesainslie/integration-testing
- Merge pull request [#21](https://github.com/jamesainslie/dot/issues/21) from jamesainslie/feature-tech-debt
- Merge pull request [#20](https://github.com/jamesainslie/dot/issues/20) from jamesainslie/feature-implement-git-changelog
- Merge pull request [#19](https://github.com/jamesainslie/dot/issues/19) from jamesainslie/feature-domain-refactor

### BREAKING CHANGE

```
(internal only): internal/api package removed.
This only affects code that directly imported internal/api, which
should not exist since it was an internal package.
```



<a name="v0.1.0"></a>
## v0.1.0 - 2025-10-07
### Build
- **make:** add buildvcs flag for reproducible builds
- **makefile:** add build infrastructure with semantic versioning
- **release:** add Homebrew tap integration

### Chore
- update README and clean up whitespace in planner files
- **ci:** ignore reviews directory
- **ci:** ignore reviews directory
- **ci:** remove args from golangci
- **ci:** move initial design docs to docs folder, and keep ignoring them
- **ci:** ignore control files
- **deps:** update go.mod dependency classification
- **docs:** add planning docs
- **init:** initialize Go module and project structure

### Ci
- **github:** add GitHub Actions workflows and goreleaser configuration
- **lint:** replace golangci-lint-action with direct installation for v2.x compatibility
- **lint:** update golangci-lint version to v2.5.0
- **release:** install golangci-lint before running linters
- **release:** use GORELEASER_TOKEN for tap updates

### Docs
- mark core implementation complete
- replace tabs with spaces in Makefile code block
- complete with ignore patterns and scanner
- complete planner foundation with 100% coverage
- implement comprehensive documentation suite
- add and completion documents
- add completion summary and update changelog
- mark complete
- add completion documentation
- add completion summary
- add implementation plan for API enhancements
- add completion summary and update changelog
- add completion summary and update changelog
- add completion summary and update changelog
- add final implementation summary
- document code review improvements
- **adr:** add ADR-003 and ADR-004 for future enhancements
- **api:** remove GNU symlink reference from package documentation
- **api:** document Doctor breaking change and migration path
- **changelog:** update changelog for completion
- **changelog:** update with features and fixes
- **cli:** remove GNU symlink references from user-facing text
- **config:** add configuration guide and update README
- **config:** add configuration management design
- **dot:** enhance Result monad documentation with usage guidance
- **dot:** clarify ScanConfig field behavior in comments
- **executor:** update completion document with improved coverage
- **executor:** add completion document
- **install:** add Homebrew installation guide and release process
- **plan:** update plan with new CLI verb terminology
- **planner:** add implementation plan
- **planner:** document completion
- **plans:** add language hints to code blocks and fix formatting
- **readme:** update documentation for completion
- **review:** add code review remediation progress tracking
- **review:** add final coverage status and analysis
- **review:** add final remediation summary
- **review:** add language identifier to commit list code block
- **terminology:** adopt manage/unmanage/remanage command naming

### Feat
- **adapters:** implement slog logger and no-op adapters
- **adapters:** implement OS filesystem adapter
- **api:** implement Unmanage, Remanage, and Adopt operations
- **api:** add foundational types for Client API
- **api:** define Client interface for public API
- **api:** implement Client with Manage operation
- **api:** add comprehensive tests and documentation
- **api:** implement directory extraction and link set optimization
- **api:** update Doctor API to accept ScanConfig parameter
- **api:** update Doctor API to accept ScanConfig parameter
- **api:** implement incremental remanage with hash-based change detection
- **api:** add depth calculation and directory skip logic
- **api:** implement link count extraction from plan
- **api:** add DoctorWithScan for explicit scan configuration
- **api:** wire up orphaned link detection with safety limits
- **cli:** implement list command for package inventory
- **cli:** add scan control flags to doctor command
- **cli:** implement help system with examples and completion
- **cli:** implement progress indicators for operation feedback
- **cli:** implement terminal styling and layout system
- **cli:** implement output renderer infrastructure
- **cli:** add minimal CLI entry point for build validation
- **cli:** add config command for XDG configuration management
- **cli:** implement error formatting foundation for
- **cli:** implement status command for installation state inspection
- **cli:** implement doctor command for health checks
- **cli:** implement CLI infrastructure with core commands
- **cli:** implement UX polish with output formatting
- **cli:** implement command handlers for manage, unmanage, remanage, adopt
- **cli:** show complete operation breakdown in table summary
- **config:** wire backup directory through system
- **config:** implement extended configuration infrastructure
- **config:** implement configuration management with Viper and XDG compliance
- **config:** add Config struct with validation
- **domain:** implement operation type hierarchy
- **domain:** implement error taxonomy with user-facing messages
- **domain:** implement Result monad for functional error handling
- **domain:** implement phantom-typed paths for compile-time safety
- **domain:** implement domain value objects
- **domain:** add package-operation mapping to Plan
- **dot:** add ScanConfig types for orphaned link detection
- **executor:** add metrics instrumentation wrapper
- **executor:** implement parallel batch execution
- **executor:** implement executor with two-phase commit
- **ignore:** implement pattern matching engine and ignore sets
- **manifest:** implement FSManifestStore persistence
- **manifest:** add core manifest domain types
- **manifest:** implement content hashing for packages
- **manifest:** define ManifestStore interface
- **manifest:** implement manifest validation
- **operation:** add Execute and Rollback methods to operations
- **pipeline:** track package ownership in operation plans
- **pipeline:** surface conflicts and warnings in plan metadata
- **pipeline:** enhance context cancellation handling in pipeline stages
- **pipeline:** implement symlink pipeline with scanning, planning, resolution, and sorting stages
- **planner:** implement suggestion generation and conflict enrichment
- **planner:** implement conflict detection for links and directories
- **planner:** define conflict type enumeration
- **planner:** implement real desired state computation
- **planner:** implement desired state computation foundation
- **planner:** define resolution status types
- **planner:** implement resolve result type
- **planner:** implement conflict value object
- **planner:** implement resolution policy types and basic policies
- **planner:** implement main resolver function and policy dispatcher
- **planner:** integrate resolver with planning pipeline
- **planner:** implement dependency graph construction
- **planner:** implement parallelization analysis
- **planner:** implement topological sort with cycle detection
- **ports:** define infrastructure port interfaces
- **scanner:** implement tree scanning with recursive traversal
- **scanner:** implement dotfile translation logic
- **scanner:** implement package scanner with ignore support
- **types:** add Status and PackageInfo types

### Fix
- **api:** address CodeRabbit feedback on
- **api:** use configured skip patterns in recursive orphan scanning
- **api:** improve error handling and test robustness
- **api:** use package-operation mapping for accurate manifest tracking
- **api:** normalize paths for cross-platform link lookup
- **api:** enforce depth and context limits in recursive orphan scanning
- **cli:** resolve critical bugs in progress, config, and rendering
- **cli:** handle both pointer and value operation types in renderers
- **cli:** improve config format detection and help text indentation
- **cli:** correct scan flag variable scope in NewDoctorCommand
- **cli:** add error templates for checkpoint and not implemented errors
- **cli:** respect NO_COLOR environment variable in shouldColorize
- **cli:** improve JSON/YAML output and doctor performance
- **cli:** render execution plan in dry-run mode
- **cli:** improve TTY detection portability and path truncation
- **config:** enable CodeRabbit auto-review for all pull requests
- **executor:** make Checkpoint operations map thread-safe
- **executor:** address code review feedback for concurrent safety and error handling
- **manifest:** add security guards and prevent hash collisions
- **pipeline:** prevent shared mutation of context maps in metadata conversion
- **release:** separate archive configs for Homebrew compatibility
- **scanner:** implement real package tree scanning with ignore filtering
- **test:** improve test isolation and cross-platform compatibility
- **test:** make Adopt execution error test deterministic

### Refactor
- **adopt:** update Adopt and PlanAdopt methods to use files-first signature
- **api:** reduce cyclomatic complexity in PlanRemanage
- **api:** extract orphan scan logic to reduce complexity
- **cli:** address code review nitpicks for improved code quality
- **cli:** reduce cyclomatic complexity in table renderer
- **cli:** add default case and eliminate type assertion duplication
- **pipeline:** use safe unwrap pattern in path construction tests
- **pipeline:** improve test quality and organization
- **quality:** improve error handling documentation and panic messages
- **terminology:** update suggestion text from unmanage to unmanage
- **terminology:** replace symlink with package directory terminology
- **terminology:** complete symlink removal from test fixtures
- **terminology:** rename symlink-prefixed variables to package/manage

### Style
- **all:** apply goimports formatting
- **all:** apply goimports formatting
- **domain:** fix linting issues and apply formatting
- **manifest:** apply goimports formatting
- **planner:** fix linting issues in implementation
- **scanner:** apply goimports formatting
- **test:** format test files with goimports

### Test
- **adapters:** add comprehensive MemFS tests to achieve 80%+ coverage
- **api:** add manifest helper tests and document remediation
- **api:** add comprehensive test coverage for all API methods
- **cli:** fix help text assertion after symlink removal
- **cli:** increase cmd/dot test coverage to 88.6%
- **cli:** add comprehensive tests to restore coverage above 80%
- **cmd:** add basic command constructor tests
- **config:** add comprehensive loader and precedence tests
- **config:** add aggressive coverage boost tests
- **config:** add validation edge case tests
- **config:** improve test coverage to 83%
- **coverage:** increase test coverage from 73.8% to 83.7%
- **dot:** add comprehensive error and operation tests
- **executor:** add comprehensive tests to exceed 80% coverage threshold
- **planner:** add coverage tests to exceed 80 percent threshold

### Pull Requests
- Merge pull request [#18](https://github.com/jamesainslie/dot/issues/18) from jamesainslie/feature-homebrew-tap
- Merge pull request [#17](https://github.com/jamesainslie/dot/issues/17) from jamesainslie/fix-dry-run-output
- Merge pull request [#16](https://github.com/jamesainslie/dot/issues/16) from jamesainslie/feature-implement-stubs
- Merge pull request [#15](https://github.com/jamesainslie/dot/issues/15) from jamesainslie/feature-remove-symlink-terminology
- Merge pull request [#14](https://github.com/jamesainslie/dot/issues/14) from jamesainslie/feature-remove-symlink-references
- Merge pull request [#13](https://github.com/jamesainslie/dot/issues/13) from jamesainslie/feature-api-enhancements
- Merge pull request [#12](https://github.com/jamesainslie/dot/issues/12) from jamesainslie/feature-code-review-remediation
- Merge pull request [#11](https://github.com/jamesainslie/dot/issues/11) from jamesainslie/feature-error-handling-ux
- Merge pull request [#10](https://github.com/jamesainslie/dot/issues/10) from jamesainslie/feature-implement-cli-query
- Merge pull request [#9](https://github.com/jamesainslie/dot/issues/9) from jamesainslie/feature-implement-cli
- Merge pull request [#7](https://github.com/jamesainslie/dot/issues/7) from jamesainslie/feature-implement-api
- Merge pull request [#6](https://github.com/jamesainslie/dot/issues/6) from jamesainslie/feature-manifests-state-management
- Merge pull request [#5](https://github.com/jamesainslie/dot/issues/5) from jamesainslie/feature-executor
- Merge pull request [#4](https://github.com/jamesainslie/dot/issues/4) from jamesainslie/jamesainslie-implement-pipeline-orchestration
- Merge pull request [#3](https://github.com/jamesainslie/dot/issues/3) from jamesainslie/jamesainslie-implement-topological-sorter
- Merge pull request [#2](https://github.com/jamesainslie/dot/issues/2) from jamesainslie/jamesainslie-implement-resolver
- Merge pull request [#1](https://github.com/jamesainslie/dot/issues/1) from jamesainslie/jamesainslie-implement-func-scanner


[Unreleased]: https://github.com/jamesainslie/dot/compare/v0.4.4...HEAD
[v0.4.4]: https://github.com/jamesainslie/dot/compare/v0.4.3...v0.4.4
[v0.4.3]: https://github.com/jamesainslie/dot/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/jamesainslie/dot/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/jamesainslie/dot/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/jamesainslie/dot/compare/v0.3.1...v0.4.0
[v0.3.1]: https://github.com/jamesainslie/dot/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/jamesainslie/dot/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/jamesainslie/dot/compare/v0.1.1...v0.2.0
[v0.1.1]: https://github.com/jamesainslie/dot/compare/v0.1.0...v0.1.1
