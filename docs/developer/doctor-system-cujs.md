# Doctor System Customer Usage Journeys (CUJs)

## Executive Summary

This document analyzes potential failure modes and diagnostic needs for the dot dotfile manager from first principles. It establishes Customer Usage Journeys that drive the design of a robust, safe, and user-friendly health checking and repair system.

The current doctor implementation has critical gaps in UX, safety, and assumptions that make it brittle in real-world scenarios. This analysis rebuilds from the ground up based on how users actually interact with their dotfiles lifecycle.

## Core Principles

1. **Defensive Verification**: Never assume the filesystem matches expectations
2. **Non-Destructive by Default**: Repairs should require explicit confirmation
3. **Clear Communication**: Users must understand what is wrong and why
4. **Transparent Actions**: Users must understand what actions will be taken
5. **Graceful Degradation**: Partial success is better than complete failure
6. **Idempotency**: Running diagnostics multiple times should be safe
7. **Audit Trail**: All changes should be logged and reversible

## User Personas

### Persona 1: The New User (Alex)
- Just installed dot
- Migrating from manual symlinks or another tool (stow, homesick)
- Limited understanding of symlink mechanics
- Wants confidence that operations are safe

### Persona 2: The Daily Driver (Blake)
- Uses dot for 6+ months
- Multiple machines with dotfiles
- Occasional issues after system updates or file moves
- Wants fast diagnostics and repair

### Persona 3: The Power User (Casey)
- Complex dotfile setup with 20+ packages
- Uses advanced features (hooks, ignore patterns, folding)
- Experiments frequently
- Wants detailed control and visibility

### Persona 4: The Team Lead (Drew)
- Maintains dotfiles repo for entire team
- Uses bootstrap configuration for onboarding
- Needs to ensure dotfiles work across heterogeneous environments
- Wants to catch issues before distribution

## CUJ Categories

### Category A: Initial Setup and Migration
### Category B: Routine Operations
### Category C: System Changes and Drift
### Category D: Error Recovery
### Category E: Multi-Machine Synchronization
### Category F: Advanced Scenarios

---

## Category A: Initial Setup and Migration

### CUJ-A1: First Time Installation

**Actor**: Alex (New User)

**Scenario**:
Alex has just installed dot and wants to manage their first dotfiles package.

**What Can Go Wrong**:
1. **Pre-existing files**: Target locations already contain files
2. **Pre-existing symlinks**: Symlinks from previous tools point elsewhere
3. **Permission issues**: Target directory not writable
4. **Filesystem limitations**: Target filesystem doesn't support symlinks
5. **Package directory structure**: User expectations don't match dot conventions

**Current Doctor Gaps**:
- Doesn't distinguish between user files and symlinks from other tools
- Doesn't detect filesystem symlink capability
- Doesn't validate package directory structure before operations
- Doesn't check write permissions proactively

**Diagnostic Needs**:
- Pre-flight check: Can we create symlinks in target directory?
- Conflict detection: What exists at target locations?
- Conflict classification: User file vs foreign symlink vs managed symlink
- Safety assessment: What is safe to overwrite vs dangerous?

**Repair Needs**:
- Backup strategy for user files
- Ignore strategy for foreign symlinks
- Clear conflict resolution workflow

**Success Criteria**:
- User understands exactly what exists and what will happen
- User can choose resolution strategy per-file or per-pattern
- No data loss occurs
- System provides undo capability

### CUJ-A2: Migration from GNU Stow

**Actor**: Alex (New User)

**Scenario**:
Alex has 15 packages managed by GNU Stow and wants to migrate to dot.

**What Can Go Wrong**:
1. **Stow symlinks detected as orphans**: Doctor flags all stow links as unmanaged
2. **Double management**: Both tools try to manage same files
3. **Stow-specific patterns**: Stow uses different directory flattening
4. **Mixed state**: Some packages migrated, some still in Stow

**Current Doctor Gaps**:
- Cannot detect stow-managed symlinks specifically
- No migration workflow for bulk symlink adoption
- Triage mode suggests ignore, not adopt
- No detection of double-management conflicts

**Diagnostic Needs**:
- Detect stow directory structure pattern
- Identify stow-managed symlinks by target pattern
- Assess migration safety per package
- Detect if stow is still installed and active

**Repair Needs**:
- Bulk adoption workflow for stow symlinks
- Automatic stow symlink detection and flagging
- Migration verification (did stow links get replaced correctly?)
- Rollback capability if migration fails

**Success Criteria**:
- Can migrate incrementally (package by package)
- No orphaned links after migration
- Can verify migration completeness
- Can coexist temporarily during transition

### CUJ-A3: Existing Dotfile Repository Adoption

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake has managed dotfiles manually with symlinks for years. Repository structure doesn't follow dot conventions exactly.

**What Can Go Wrong**:
1. **Unconventional structure**: Packages don't use dot- prefix
2. **Nested packages**: Some logical packages are nested
3. **Mixed content**: Some files should be symlinked, others shouldn't
4. **External dependencies**: Some dotfiles reference other dotfiles

**Current Doctor Gaps**:
- Assumes dot- prefix convention strictly
- Cannot handle non-standard package structures
- No guidance on restructuring repository
- Cannot detect interdependencies

**Diagnostic Needs**:
- Repository structure analysis
- Convention violation detection
- Suggestion for restructuring
- Dependency graph analysis

**Repair Needs**:
- Guided restructuring workflow
- Batch rename with prefix addition
- Selective file management (include/exclude patterns per package)
- Validation that restructuring succeeded

**Success Criteria**:
- Can adopt existing repos with minimal changes
- Clear guidance on required restructuring
- Validation that structure is manageable
- No broken references after restructuring

---

## Category B: Routine Operations

### CUJ-B1: Daily Health Check

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake runs `dot doctor` as part of morning routine to ensure everything is healthy.

**What Can Go Wrong**:
1. **False positives**: Flags issues that aren't real problems
2. **Noise**: Too many warnings obscure real issues
3. **Slow execution**: Takes too long for routine check
4. **Transient issues**: Flags temporary filesystem states

**Current Doctor Gaps**:
- Orphan scanning is slow and noisy
- No quick health check mode (skip expensive scans)
- Cannot suppress known non-issues
- No trend analysis (new issues vs known issues)

**Diagnostic Needs**:
- Fast mode: Only check managed links
- Comparison: What changed since last check?
- Severity filtering: Errors only, hide warnings
- Known issue suppression: Don't re-report acknowledged issues

**Repair Needs**:
- Not applicable (informational use)

**Success Criteria**:
- Completes in under 2 seconds for typical setup
- Only shows new or critical issues
- Clear indication of system health status
- Exit code suitable for scripting

### CUJ-B2: Post-Update Verification

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake just pulled latest changes from dotfiles repo and ran `dot remanage vim`. Wants to verify it worked.

**What Can Go Wrong**:
1. **Remanage failed silently**: Some links weren't updated
2. **New files not linked**: Added files in repo not detected
3. **Removed files still linked**: Deleted repo files leave dangling links
4. **Ignored files incorrectly linked**: Ignore patterns not respected

**Current Doctor Gaps**:
- Cannot compare current state to previous state
- Cannot detect partial remanage failures
- Cannot verify ignore patterns are working
- Cannot detect missing new files

**Diagnostic Needs**:
- Diff mode: What changed in this package?
- Completeness check: Are all package files linked?
- Consistency check: Do links match package content?
- Ignore verification: Are ignore patterns working?

**Repair Needs**:
- Complete partial remanage
- Add missing links
- Remove extra links
- Fix ignore pattern issues

**Success Criteria**:
- Can verify remanage succeeded completely
- Can detect and fix partial failures
- Clear report of what changed
- Confidence that system is in good state

### CUJ-B3: Package Addition

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake creates a new package for neovim config and runs `dot manage neovim`.

**What Can Go Wrong**:
1. **Conflicts with existing files**: Neovim config already exists
2. **Partial success**: Some files conflicted, others linked
3. **Wrong conflict strategy**: Used wrong --on-conflict flag
4. **Unexpected file overwrites**: Lost existing configuration

**Current Doctor Gaps**:
- Cannot analyze conflicts before manage operation
- Cannot verify manage operation succeeded completely
- Cannot detect if conflict strategy was appropriate
- Cannot recommend undo/rollback

**Diagnostic Needs**:
- Pre-manage conflict analysis
- Post-manage verification
- Conflict strategy recommendation
- Rollback feasibility assessment

**Repair Needs**:
- Undo recent manage operation
- Restore backed-up files
- Complete partial manage
- Switch conflict strategy and re-manage

**Success Criteria**:
- Can preview conflicts before manage
- Can verify manage succeeded
- Can rollback if needed
- No surprise data loss

---

## Category C: System Changes and Drift

### CUJ-C1: Operating System Upgrade

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake upgrades from Ubuntu 22.04 to 24.04. System restarts and various paths change.

**What Can Go Wrong**:
1. **Absolute links broken**: Package directory path changed
2. **System tools relocated**: Links to system binaries broken
3. **Permission changes**: Home directory permissions changed
4. **XDG paths changed**: Config directory structure evolved

**Current Doctor Gaps**:
- Cannot detect systematic path changes
- Cannot distinguish upgrade-related issues from random breakage
- Cannot suggest bulk fixes for path changes
- Cannot verify links still point to correct logical targets

**Diagnostic Needs**:
- Detect systematic link breakage patterns
- Identify path migration issues
- Assess if links point to semantically correct targets
- Recommend bulk path updates

**Repair Needs**:
- Bulk link target updates
- Remanage all packages with new paths
- Verify semantic correctness after repair
- Update manifest with new paths

**Success Criteria**:
- Can detect and fix systematic path changes
- Minimal manual intervention required
- Verification that fixes are correct
- Audit trail of changes made

### CUJ-C2: Dotfiles Directory Moved

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake moves dotfiles from `~/dotfiles` to `~/.local/share/dotfiles` for XDG compliance.

**What Can Go Wrong**:
1. **All relative links broken**: Every managed link now points nowhere
2. **Manifest still references old paths**: Package_dir field incorrect
3. **Absolute links partially work**: Mixed relative/absolute setup breaks differently
4. **Cannot run dot commands**: Package directory not found

**Current Doctor Gaps**:
- Cannot detect repository relocation
- Cannot automatically update configuration
- Cannot distinguish relocation from deletion
- Cannot suggest recovery strategy

**Diagnostic Needs**:
- Detect package directory relocation
- Assess link type implications (relative vs absolute)
- Verify new location is valid
- Check if old location still accessible

**Repair Needs**:
- Update configuration with new package directory
- Remanage all packages from new location
- Verify all links recreated correctly
- Clean up old location if desired

**Success Criteria**:
- Can recover from directory moves
- Clear guidance on required steps
- Automated recovery if possible
- Verification that recovery succeeded

### CUJ-C3: External Tool Interference

**Actor**: Casey (Power User)

**Scenario**:
Casey installed a system package that created symlinks in home directory, conflicting with dot-managed locations.

**What Can Go Wrong**:
1. **Managed link overwritten**: System package replaced dot symlink
2. **Competing management**: Both systems try to manage same file
3. **Link target hijacked**: Link still exists but points elsewhere
4. **Manifest out of sync**: Manifest says link exists, but it's wrong

**Current Doctor Gaps**:
- Cannot detect link target hijacking
- Cannot identify external tool ownership
- Cannot recommend resolution strategy
- Cannot prevent future conflicts

**Diagnostic Needs**:
- Detect links that point to wrong targets
- Identify target pattern (is it from known tool?)
- Assess safety of restoring dot management
- Recommend ignore or restore strategy

**Repair Needs**:
- Restore correct link target
- Add to ignore list if external management preferred
- Remove from manifest if no longer managed
- Verify restoration succeeded

**Success Criteria**:
- Can detect external interference
- Clear recommendation for resolution
- Can prevent future conflicts via ignore
- Audit trail of interference

### CUJ-C4: Filesystem Corruption

**Actor**: Blake (Daily Driver)

**Scenario**:
System crashed during write operation. Some symlinks partially created or corrupted.

**What Can Go Wrong**:
1. **Circular symlinks**: Corruption created circular reference
2. **Symlinks to invalid targets**: Target path corrupted
3. **Manifest corrupted**: JSON file truncated or malformed
4. **Mixed states**: Some packages corrupted, others fine

**Current Doctor Gaps**:
- Circular symlink detection exists but not highlighted
- Cannot detect manifest corruption
- Cannot recommend manifest rebuild strategy
- Cannot assess which packages are affected

**Diagnostic Needs**:
- Circular reference detection
- Manifest integrity check
- Per-package health assessment
- Corruption pattern analysis

**Repair Needs**:
- Break circular references
- Rebuild manifest from filesystem
- Recreate corrupted links
- Verify repair completeness

**Success Criteria**:
- Can detect corruption
- Can rebuild manifest if corrupted
- Can fix circular references safely
- No data loss from repair

---

## Category D: Error Recovery

### CUJ-D1: Accidental Unmanage

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake accidentally ran `dot unmanage vim` and lost all vim symlinks. Wants to restore.

**What Can Go Wrong**:
1. **Links removed**: All symlinks deleted
2. **Manifest updated**: Package removed from manifest
3. **No backup**: Original files still in package, but links gone
4. **User edits lost**: If user edited symlink targets, those edits lost

**Current Doctor Gaps**:
- Cannot detect recent unmanage operations
- Cannot suggest undo
- Cannot verify if manage will restore correct state
- No operation history

**Diagnostic Needs**:
- Detect missing packages (in repo but not manifest)
- Suggest re-manage for orphaned packages
- Verify package directory still exists
- Check if target locations now occupied

**Repair Needs**:
- Re-manage package
- Verify links restored correctly
- Check for conflicts before restoration
- Audit trail of restoration

**Success Criteria**:
- Can detect and suggest recovery
- Re-manage restores state
- Conflicts handled appropriately
- User informed of recovery success

### CUJ-D2: Broken Package After Edit

**Actor**: Casey (Power User)

**Scenario**:
Casey edited files in package directory directly, including renaming some files. Links now broken.

**What Can Go Wrong**:
1. **Links point to old names**: Renamed files break links
2. **Manifest incorrect**: Manifest still references old names
3. **Mixed working/broken**: Some links work, others don't
4. **Cannot determine intent**: Did user want to stop managing those files?

**Current Doctor Gaps**:
- Cannot detect renames (sees broken + new orphan)
- Cannot suggest remanage as solution
- No guidance on fixing mixed states
- Cannot detect package structure changes

**Diagnostic Needs**:
- Detect package structure changes
- Suggest remanage to sync manifest
- Identify likely renames (similar names)
- Assess if remanage will fix issue

**Repair Needs**:
- Remanage to resync with package content
- Verify manifest now correct
- Confirm all links now working
- Report what changed

**Success Criteria**:
- Can detect package changes
- Remanage suggestion is clear
- Remanage fixes issue
- User understands what happened

### CUJ-D3: Manifest Manually Edited

**Actor**: Casey (Power User)

**Scenario**:
Casey manually edited `.dot-manifest.json` to fix an issue. Manifest may now be inconsistent with filesystem.

**What Can Go Wrong**:
1. **Manifest references non-existent links**: Links in manifest don't exist on filesystem
2. **Filesystem has untracked links**: Links exist but not in manifest
3. **Count mismatches**: LinkCount field incorrect
4. **Invalid JSON**: Syntax errors in manifest

**Current Doctor Gaps**:
- Cannot detect manifest/filesystem inconsistency
- Cannot rebuild manifest from filesystem
- No validation of manifest integrity
- Cannot recommend fix strategy

**Diagnostic Needs**:
- Manifest integrity check (JSON syntax)
- Manifest consistency check (counts, references)
- Filesystem vs manifest comparison
- Recommend rebuild vs targeted fixes

**Repair Needs**:
- Fix JSON syntax errors
- Update counts to match reality
- Add missing links to manifest
- Remove phantom links from manifest
- Rebuild manifest from scratch if needed

**Success Criteria**:
- Can validate manifest integrity
- Can detect inconsistencies
- Can rebuild if necessary
- Minimal disruption to user

### CUJ-D4: Orphaned Links After Package Removal

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake deleted a package directory without unmanaging it first. Links now broken and manifest references non-existent package.

**What Can Go Wrong**:
1. **Broken links everywhere**: All package links broken
2. **Manifest references missing package**: Cannot remanage or unmanage
3. **Cannot easily clean up**: No package directory to reference
4. **User doesn't remember what was managed**: Uncertain what to clean

**Current Doctor Gaps**:
- Cannot detect package directory deletion
- Cannot suggest clean removal of phantom packages
- Cannot identify all affected links easily
- No bulk cleanup for broken package

**Diagnostic Needs**:
- Detect packages with missing directories
- Identify all links for phantom package
- Assess if package directory just moved or truly deleted
- Recommend cleanup strategy

**Repair Needs**:
- Remove phantom package from manifest
- Remove/clean up broken links
- Verify no links left behind
- Audit trail of cleanup

**Success Criteria**:
- Can detect phantom packages
- Can clean up completely
- User informed of cleanup actions
- No broken links left behind

---

## Category E: Multi-Machine Synchronization

### CUJ-E1: Dotfiles Across Different Machines

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake maintains dotfiles across work laptop (Linux), personal laptop (macOS), and server (Linux). Different paths and tools available.

**What Can Go Wrong**:
1. **Platform-specific paths**: Links reference platform-specific locations
2. **Missing tools**: Some dotfiles reference tools not installed everywhere
3. **Different home structures**: XDG vs traditional layout differs by machine
4. **Absolute links break**: Absolute links to user-specific paths differ

**Current Doctor Gaps**:
- Cannot assess platform compatibility
- Cannot detect platform-specific issues
- Cannot recommend conditional management
- No cross-machine health comparison

**Diagnostic Needs**:
- Platform compatibility analysis
- Detect platform-specific assumptions
- Identify machine-specific issues
- Compare health across machines

**Repair Needs**:
- Not directly doctor's job, but can inform user

**Success Criteria**:
- Can identify platform issues
- Clear reporting of machine-specific problems
- Guidance on platform-conditional setup
- Can verify health per platform

### CUJ-E2: Fresh Clone on New Machine

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake clones dotfiles repo to new machine and runs bootstrap.

**What Can Go Wrong**:
1. **Bootstrap assumptions wrong**: Assumes paths that don't exist
2. **Dependency ordering**: Some packages depend on others being installed first
3. **Pre-existing config**: New machine has conflicting default config
4. **Permission differences**: New machine has different home directory permissions

**Current Doctor Gaps**:
- Cannot validate bootstrap configuration before use
- Cannot detect dependency ordering issues
- No dry-run for bootstrap
- Cannot assess environment compatibility

**Diagnostic Needs**:
- Bootstrap configuration validation
- Environment compatibility check
- Dependency order verification
- Pre-flight conflict detection

**Repair Needs**:
- Fix bootstrap configuration issues
- Resolve conflicts found during bootstrap
- Verify bootstrap succeeded completely

**Success Criteria**:
- Can validate bootstrap before running
- Can detect bootstrap failures
- Can fix incomplete bootstrap
- Reliable first-time setup

### CUJ-E3: Selective Package Sync

**Actor**: Blake (Daily Driver)

**Scenario**:
Blake manages some packages on all machines (vim, git) but others only on specific machines (work tools on work laptop).

**What Can Go Wrong**:
1. **Manifest drift**: Manifests differ between machines
2. **Orphan detection confusion**: Doctor on machine A flags packages from machine B
3. **Accidental cross-contamination**: Pull from wrong machine overwrites manifest
4. **Cannot track what should be where**: No clear record of per-machine intent

**Current Doctor Gaps**:
- Cannot understand per-machine package expectations
- No concept of machine profiles
- Cannot suppress warnings for intentionally missing packages
- Cannot validate manifest is appropriate for machine

**Diagnostic Needs**:
- Machine profile awareness
- Expected vs actual package comparison
- Manifest validation for this machine
- Cross-machine consistency check

**Repair Needs**:
- Align manifest with machine profile
- Remove packages not intended for this machine
- Add packages missing for this machine
- Verify alignment succeeded

**Success Criteria**:
- Can define machine profiles
- Doctor understands per-machine expectations
- No false positives from missing packages
- Clear alignment verification

---

## Category F: Advanced Scenarios

### CUJ-F1: Security Audit

**Actor**: Drew (Team Lead)

**Scenario**:
Drew needs to audit team dotfiles for security issues before distributing to team.

**What Can Go Wrong**:
1. **Secrets in dotfiles**: API keys, passwords, tokens committed
2. **Dangerous symlinks**: Links to system files that could cause issues
3. **Insecure permissions**: Package files world-readable
4. **Supply chain issues**: Dotfiles reference external, untrusted sources

**Current Doctor Gaps**:
- No secret scanning
- No permission auditing
- No security-focused checks
- Cannot assess risk of symlinks

**Diagnostic Needs**:
- Secret pattern detection
- Permission vulnerability scan
- Dangerous symlink detection (links to system files)
- External reference analysis

**Repair Needs**:
- Flag issues (not auto-fix for security)
- Recommend remediation
- Block dangerous operations

**Success Criteria**:
- Can detect potential secrets
- Can identify security risks
- Clear risk reporting
- Actionable remediation guidance

### CUJ-F2: Large-Scale Deployment

**Actor**: Drew (Team Lead)

**Scenario**:
Drew deploys dotfiles to 50 developer machines via configuration management.

**What Can Go Wrong**:
1. **Partial failures**: Some machines fail, others succeed
2. **Environment variations**: Different environments need different handling
3. **No visibility**: Cannot assess deployment health across fleet
4. **Rollback complexity**: Cannot easily rollback failed deployments

**Current Doctor Gaps**:
- No batch health reporting
- No JSON/structured output for parsing
- No deployment verification mode
- Cannot assess fleet health

**Diagnostic Needs**:
- Structured output for automation
- Deployment verification mode
- Aggregate health reporting
- Failure pattern analysis

**Repair Needs**:
- Batch repair operations
- Idempotent repair scripts
- Audit trail for fleet changes

**Success Criteria**:
- Machine-readable output
- Can verify deployment succeeded
- Can aggregate health across fleet
- Can identify common failure patterns

### CUJ-F3: Complex Package Interdependencies

**Actor**: Casey (Power User)

**Scenario**:
Casey's shell config sources other managed dotfiles. Vim config references git config. Complex web of dependencies.

**What Can Go Wrong**:
1. **Dependency order matters**: Must manage in specific order
2. **Broken references**: One package references another's files
3. **Circular dependencies**: Packages reference each other
4. **Partial management**: Some dependencies not managed yet

**Current Doctor Gaps**:
- No dependency graph analysis
- Cannot detect cross-package references
- Cannot suggest management order
- Cannot verify dependency integrity

**Diagnostic Needs**:
- Dependency graph construction
- Reference validation across packages
- Circular dependency detection
- Missing dependency detection

**Repair Needs**:
- Recommend management order
- Flag circular dependencies
- Suggest managing missing dependencies

**Success Criteria**:
- Can analyze dependencies
- Clear visualization of dependency graph
- Warnings for dependency issues
- Recommended fix order

### CUJ-F4: Performance Issues

**Actor**: Casey (Power User)

**Scenario**:
Casey has 30 packages with 500+ total managed files. Doctor takes minutes to run.

**What Can Go Wrong**:
1. **Slow orphan scanning**: Deep directory scans take forever
2. **Repeated checks**: Running doctor multiple times very slow
3. **Unnecessary work**: Scans directories that can't have orphans
4. **No progress indication**: Appears hung for minutes

**Current Doctor Gaps**:
- Orphan scanning not optimized
- No caching of scan results
- No progress indication
- No selective checking

**Diagnostic Needs**:
- Fast mode: managed links only
- Selective checking: specific packages only
- Cached results: incremental checking
- Progress indication: show what's happening

**Repair Needs**:
- Not applicable

**Success Criteria**:
- Can complete health check in under 5 seconds
- Progress clearly indicated
- Can selectively check subsets
- Cached results for repeated runs

---

## Analysis: Current System vs Needed Capabilities

### Critical Gaps

1. **No Pre-Flight Checks**: Cannot assess safety before operations
2. **Limited Conflict Understanding**: Cannot distinguish conflict types
3. **No Undo Capability**: Cannot rollback operations
4. **Poor State Comparison**: Cannot detect what changed
5. **Weak Manifest Validation**: Cannot detect inconsistencies
6. **No Dependency Analysis**: Cannot understand package relationships
7. **Limited Security Scanning**: No secret or permission checks
8. **Poor Performance**: Slow for routine checks
9. **Confusing UX**: Unclear what doctor will do
10. **Dangerous Repairs**: Auto-fix could cause data loss

### Unsafe Assumptions

1. **Assumption**: All orphaned links should be ignored or adopted
   - **Reality**: May be managed by other tools intentionally
   - **Risk**: Breaking external tool management

2. **Assumption**: Broken links can be safely removed
   - **Reality**: Target may return (network mount, external drive)
   - **Risk**: Removing links user wants to keep

3. **Assumption**: Package directory structure is always correct
   - **Reality**: Users edit packages, files move
   - **Risk**: Manifest and reality diverge

4. **Assumption**: Remanage fixes all issues
   - **Reality**: May not fix external interference or path changes
   - **Risk**: False confidence

5. **Assumption**: Manifest is always correct
   - **Reality**: Can be manually edited or corrupted
   - **Risk**: Operating on false information

6. **Assumption**: User understands symlink mechanics
   - **Reality**: Many users don't fully understand
   - **Risk**: Confusion about repairs and side effects

7. **Assumption**: Filesystem is stable
   - **Reality**: Can change during operations
   - **Risk**: Race conditions, inconsistent state

### Brittle Areas

1. **Triage Mode**: Complex interactive flow with bugs
2. **Fix Mode**: Can destroy data without clear warnings
3. **Orphan Scanning**: Slow and generates noise
4. **Pattern Matching**: Doesn't handle all symlink target patterns
5. **Error Messages**: Not actionable
6. **Exit Codes**: Not suitable for scripting

---

## Required Capabilities Summary

### Diagnostic Capabilities (Read-Only)

1. **Health Checking**
   - Fast managed link validation
   - Comprehensive deep scan mode
   - Selective package checking
   - Comparison with previous state

2. **State Analysis**
   - Manifest integrity validation
   - Manifest vs filesystem consistency
   - Package structure validation
   - Dependency graph analysis

3. **Conflict Detection**
   - Pre-operation conflict scanning
   - Conflict type classification
   - Conflict ownership detection
   - Resolution strategy recommendation

4. **Problem Classification**
   - Severity assessment
   - Impact analysis
   - Root cause identification
   - Fix complexity estimation

5. **Security Scanning**
   - Secret pattern detection
   - Permission vulnerability scan
   - Dangerous symlink detection
   - External reference analysis

6. **Platform Analysis**
   - Platform compatibility checking
   - Path portability analysis
   - Tool dependency validation
   - Environment requirement checking

### Repair Capabilities (Write Operations)

1. **Link Management**
   - Recreate missing links
   - Fix broken links
   - Remove orphaned links
   - Update link targets

2. **Manifest Operations**
   - Sync manifest with filesystem
   - Rebuild manifest from scratch
   - Fix count mismatches
   - Remove phantom entries

3. **Bulk Operations**
   - Remanage all packages
   - Clean up broken package
   - Migrate from other tools
   - Update after path changes

4. **Interactive Resolution**
   - Guided conflict resolution
   - Pattern-based bulk decisions
   - Adoption workflow
   - Ignore management

5. **Rollback and Undo**
   - Operation history
   - Undo recent changes
   - Restore from backup
   - Verify rollback success

### UX Capabilities

1. **Reporting**
   - Clear problem descriptions
   - Actionable suggestions
   - Severity indication
   - Impact assessment

2. **Modes**
   - Quick health check
   - Deep comprehensive scan
   - Focused package check
   - Security audit mode

3. **Outputs**
   - Human-readable text
   - Machine-readable JSON
   - Structured table format
   - Comparison diff format

4. **Interactivity**
   - Dry-run preview
   - Interactive confirmation
   - Batch/auto-confirm mode
   - Step-by-step guidance

---

## Recommendations for Redesign

### Immediate Priorities

1. **Safety First**: All repair operations must be opt-in with clear warnings
2. **Clear Communication**: Every issue must explain what, why, impact, and how to fix
3. **Manifest Validation**: Always validate manifest before trusting it
4. **Performance**: Fast default mode suitable for routine use
5. **Exit Codes**: Reliable exit codes for CI/CD integration

### Architectural Changes

1. **Separate Concerns**: Diagnostic, analysis, repair, and reporting as distinct phases
2. **Pluggable Checks**: Modular check system for extensibility
3. **State Tracking**: Remember previous check results for comparison
4. **Operation History**: Audit trail of all changes made
5. **Dry-Run Always**: All repairs support dry-run mode

### User Experience Improvements

1. **Progressive Disclosure**: Show summary first, details on demand
2. **Guided Workflows**: Step-by-step resolution for complex issues
3. **Smart Defaults**: Sensible defaults for common scenarios
4. **Escape Hatches**: Advanced users can bypass guidance
5. **Learning Mode**: Explain concepts and mechanics to users

---

## Success Metrics

A successful doctor system redesign will achieve:

1. **Safety**: Zero data loss from doctor operations in 99.9% of cases
2. **Clarity**: 90%+ of users understand issue descriptions without documentation
3. **Performance**: Routine health check completes in under 2 seconds
4. **Reliability**: Exit codes accurate 100% of time for scripting
5. **Completeness**: Detects 95%+ of real-world dotfile issues
6. **Confidence**: Users trust doctor recommendations 90%+ of time

---

## Next Steps

1. Review and validate these CUJs with actual users
2. Prioritize CUJs by frequency and impact
3. Design check architecture based on diagnostic needs
4. Design repair workflows based on repair needs
5. Prototype new UX for highest-priority CUJs
6. Implement with comprehensive test coverage
7. Iterate based on user feedback

---

## Document Metadata

- **Version**: 1.0
- **Date**: 2025-11-16
- **Author**: System Analysis
- **Status**: Draft for Review
- **Next Review**: After user validation

