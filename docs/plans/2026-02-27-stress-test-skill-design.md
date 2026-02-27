# Design: Full App Stress Test Skill

## Purpose

A reusable project skill (`/stress-test`) that orchestrates 7 parallel subagents to comprehensively test every feature of the `dot` CLI. Each agent operates in an isolated sandbox, finds bugs, UX issues, and feature gaps, and files them as beads issues.

## Invocation

```
/stress-test
```

Presents an interactive scope selection prompt before launching. Options:

- **All agents** (full run, ~7 parallel agents)
- **Pick agents** (multi-select from the 7 agent types)

Flags embedded in the skill prompt, not CLI args.

## Safety Model

Every agent creates and operates exclusively inside its own sandbox:

```
/tmp/dot-test-<agent-name>-<timestamp>/
  home/           # $HOME for this agent
  packages/       # --dir (package source directory)
```

Environment overrides per agent:

- `HOME=/tmp/dot-test-<name>/home`
- `XDG_CONFIG_HOME=/tmp/dot-test-<name>/home/.config`
- `DOT_CONFIG=` (unset, use defaults)
- `NO_COLOR=` (unset unless testing color)

Agents **never** reference the real `$HOME`, `~/.config`, or any path outside their sandbox. The skill MUST include this constraint in every agent prompt.

## Build Step

Every agent starts by building and installing `dot` from source:

```bash
cd /Volumes/Development/dot && go build -o /tmp/dot-test-<name>/bin/dot ./cmd/dot/
```

Then uses `/tmp/dot-test-<name>/bin/dot` for all operations, with `PATH` prepended.

## Agent Specifications

### Agent 1: Core Lifecycle

Domain: manage, unmanage, remanage fundamental operations.

Tests:
- Manage single package, verify symlinks created
- Manage multiple packages at once
- Unmanage single package, verify symlinks removed + empty dirs cleaned
- Unmanage `--all` with `--yes`, verify all removed
- Unmanage with `--purge` vs default behavior
- Remanage with no changes (expect "No changes detected" info message)
- Remanage after modifying package contents (expect success message)
- Manage reserved names (`dot`, `.dot`, `dot-config`) - expect clear errors
- Manage nonexistent package - expect clear error
- Unmanage nonexistent package - expect ErrPackageNotFound
- `--dry-run` on all operations - verify no filesystem changes occur
- Edge cases: empty packages, packages with only ignored files, deeply nested dirs

### Agent 2: Adopt + Clone Workflow

Domain: file adoption and repository cloning.

Tests:
- Create real files in sandbox HOME, adopt them into a package
- Verify files moved to package dir and symlinks created back
- Unmanage adopted package - verify files restored to original location
- Unmanage adopted with `--no-restore` - verify files stay in package dir
- Unmanage adopted with `--purge` - verify package directory deleted
- Adopt with explicit package name vs auto-derived name
- Adopt with relative, absolute, and `~/` paths
- Clone a local git repo (create one in sandbox) with `.dotbootstrap.yaml`
- Clone with `--profile` selection
- Clone with `--force` to overwrite existing
- Bootstrap generation: `clone bootstrap`, validate YAML output
- Edge cases: adopt a file that already exists in package, adopt a symlink

### Agent 3: Config + Doctor + Status

Domain: inspection and configuration commands.

Tests:
- `config init` creates valid YAML, `config init --force` overwrites
- `config get/set` for representative keys from each section
- `config list` shows all sections
- `config path` shows path with existence status
- `config upgrade` with old-style config containing `ignore.overrides`
- `config upgrade --yes` and `-y` flags skip prompt
- `status` with no packages (expect "No packages installed")
- `status <installed-pkg>` shows details
- `status <nonexistent>` shows "not installed" message
- `status --format json/yaml/table/text` all produce valid output
- `list` with `--sort name/links/date`, `--format` variations, `--show-target`
- `doctor` on healthy installation (exit code 0)
- `doctor` with deliberately broken symlinks (exit code 2)
- `doctor` with orphaned symlinks (exit code 1)
- `doctor --scan-mode off/scoped/deep`
- `doctor --format json/yaml/table/text`
- `doctor --verbose`
- `--no-color` and `NO_COLOR=1` produce uncolored output

### Agent 4: Ignore + Translation + Edge Cases

Domain: dotfile translation rules and ignore patterns.

Tests:
- `dot-vimrc` translates to `.vimrc` in target
- `dot-ssh` creates `.ssh/` directory mapping
- Plain `vim` package (no dot- prefix) links without translation
- Package with `.dotignore` - ignored files not linked
- `.dotignore` with negation patterns (`!keep-this`)
- Default ignore patterns (`.git`, `.DS_Store`, `.dotignore`, `.dotbootstrap.yaml`)
- `--no-defaults` disables default patterns
- `--no-dotignore` disables per-package ignore
- `--ignore "*.bak"` adds custom patterns from CLI
- `--max-file-size` blocks files exceeding limit
- Deeply nested package structures (3-4 levels)
- Package with special characters in filenames (spaces, unicode)
- Package containing symlinks (symlink within package dir)
- Empty directories in packages
- Package name mapping on vs off (config toggle)

### Agent 5: Real Shell Configs

Domain: actual shell config files verified by running the shells.

Tests:
- `dot-bashrc` with aliases/exports, manage, run `bash --rcfile` and verify
- `dot-zshrc` with real config, manage, run `zsh -c 'source ...; echo $VAR'`
- `dot-profile` with env vars, manage, run `sh -c '. ...; echo $VAR'`
- `dot-fishconfig` (fish `config.fish`), manage, verify with `fish -c`
- Remanage after modifying shell config, verify changes picked up by shell
- Full cycle: create config -> manage -> verify shell loads it -> modify -> remanage -> verify again

### Agent 6: Real Editor + Git + Tool Configs

Domain: configs for git, vim, tmux, ssh, gpg verified by running those programs.

Tests:
- `dot-gitconfig` with `[user]` section, manage, run `git -c include.path=... config user.name`
- `dot-vimrc` with settings, manage, run `vim` in ex mode to verify settings loaded
- `dot-tmux` with tmux.conf, manage, verify `tmux -f <path> start-server` parses
- `dot-ssh` with config file, manage, verify `ssh -F <path> -G hostname` parses it
- `dot-gnupg` with gpg.conf, manage, verify gpg reads config from managed homedir
- Full adopt workflow: create gitconfig in HOME -> adopt -> verify git still reads it -> unmanage -> verify restored

### Agent 7: Stress + Error Recovery

Domain: robustness, error handling, and adversarial conditions.

Tests:
- Manage 20+ packages simultaneously
- Rapidly manage/unmanage same package in sequence
- Manage package, manually delete symlink, then doctor + remanage to recover
- Manage package, manually replace symlink with real file, attempt remanage
- Two packages creating conflicting symlinks
- Manage with read-only target directory (expect clear error)
- Manage with read-only package directory (expect clear error)
- Corrupt manifest JSON, then run doctor/status/manage (graceful recovery)
- Very long package name (200+ chars)
- Package with 100+ files
- `--batch` mode across all commands
- Exit codes: 0 for success, non-zero for errors
- Doctor exit codes: 0 (healthy), 1 (warnings), 2 (errors)

## Issue Filing

Each agent files beads issues using `bd create`:

| Category | Type | Priority |
|----------|------|----------|
| Data loss / corruption risk | bug | P1 |
| Broken feature / wrong output | bug | P2 |
| Confusing UX / misleading message | task | P3 |
| Missing feature a user would expect | feature | P3 |
| Nice-to-have polish | task | P4 |

Title format: `[agent-N:name] description`

Each issue includes in notes:
- Reproduction steps (exact commands)
- Expected behavior
- Actual behavior (with output)
- Sandbox path for debugging

Agents should NOT file issues for:
- Known limitations documented in help text
- Behaviors that match `--help` descriptions exactly
- Pre-existing issues already in beads

## Skill Structure

```
.claude/skills/stress-test/
  SKILL.md          # Skill definition with prompts and agent specs
```

The skill:
1. Presents scope selection (all or pick)
2. Builds `dot` binary once
3. Launches selected agents in parallel via Task tool
4. Each agent runs in background
5. Collects results and summarizes findings
6. Reports total issues filed by category
