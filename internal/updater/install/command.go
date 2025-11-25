package install

import (
	"fmt"
	"regexp"
	"strings"
)

// Command represents a validated, executable command for package managers.
// Fields are unexported to prevent modification after construction.
type Command struct {
	name   string   // e.g., "brew", "sudo"
	args   []string // e.g., "upgrade", "yaklabco/dot/dot"
	source Source   // The installation source this command is for
}

// commandSpec defines the allowed structure for a command.
type commandSpec struct {
	// executable is the command name (e.g., "brew", "sudo").
	executable string
	// staticArgs are fixed arguments that don't change.
	staticArgs []string
	// dynamicArgPattern validates the package name or other dynamic arg.
	dynamicArgPattern *regexp.Regexp
}

// commandSpecs maps sources to their allowed command structures.
// The dynamic argument (package name) is always appended at the end.
var commandSpecs = map[Source]commandSpec{
	SourceHomebrew: {
		executable:        "brew",
		staticArgs:        []string{"upgrade"},
		dynamicArgPattern: regexp.MustCompile(`^[a-z][a-z0-9-]*/[a-z][a-z0-9-]*/[a-z][a-z0-9-]*$|^[a-z][a-z0-9-]*$`),
	},
	SourceApt: {
		executable:        "sudo",
		staticArgs:        []string{"apt-get", "install", "--only-upgrade", "-y"},
		dynamicArgPattern: regexp.MustCompile(`^[a-z][a-z0-9+.-]*$`),
	},
	SourcePacman: {
		executable:        "sudo",
		staticArgs:        []string{"pacman", "-S", "--noconfirm"},
		dynamicArgPattern: regexp.MustCompile(`^[a-z][a-z0-9@._+-]*$`),
	},
	SourceChocolatey: {
		executable:        "choco",
		staticArgs:        []string{"upgrade", "-y"},
		dynamicArgPattern: regexp.MustCompile(`^[A-Za-z][A-Za-z0-9.-]*$`),
	},
	SourceGoInstall: {
		executable:        "go",
		staticArgs:        []string{"install"},
		dynamicArgPattern: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9./-]*@[a-zA-Z0-9./-]+$`),
	},
}

// shellMetachars contains characters that could enable shell injection.
const shellMetachars = "|;&$`\"'\\<>(){}[]!*?~#"

// NewCommand creates a validated Command for the given source and package name.
func NewCommand(source Source, pkgName string) (*Command, error) {
	spec, ok := commandSpecs[source]
	if !ok {
		return nil, fmt.Errorf("no command spec for source: %s", source)
	}

	// Validate package name
	if err := validateArgument(pkgName, spec.dynamicArgPattern); err != nil {
		return nil, fmt.Errorf("invalid package name %q: %w", pkgName, err)
	}

	// Build args: static args + package name at end
	args := make([]string, 0, len(spec.staticArgs)+1)
	args = append(args, spec.staticArgs...)
	args = append(args, pkgName)

	return &Command{
		name:   spec.executable,
		args:   args,
		source: source,
	}, nil
}

// validateArgument checks that an argument is safe and matches the expected pattern.
func validateArgument(arg string, pattern *regexp.Regexp) error {
	if arg == "" {
		return fmt.Errorf("empty argument")
	}

	// Check for shell metacharacters
	if containsShellMetachars(arg) {
		return fmt.Errorf("contains shell metacharacters")
	}

	// Check for null bytes
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("contains null byte")
	}

	// Check pattern
	if pattern != nil && !pattern.MatchString(arg) {
		return fmt.Errorf("does not match expected pattern")
	}

	return nil
}

// containsShellMetachars checks if a string contains shell metacharacters.
func containsShellMetachars(s string) bool {
	for _, c := range s {
		if strings.ContainsRune(shellMetachars, c) {
			return true
		}
	}
	return false
}

// Name returns the command executable name.
func (c *Command) Name() string {
	return c.name
}

// Args returns a copy of the command arguments.
func (c *Command) Args() []string {
	args := make([]string, len(c.args))
	copy(args, c.args)
	return args
}

// Source returns the installation source this command is for.
func (c *Command) Source() Source {
	return c.source
}

// String returns a human-readable representation of the command.
func (c *Command) String() string {
	parts := make([]string, 0, 1+len(c.args))
	parts = append(parts, c.name)
	parts = append(parts, c.args...)
	return strings.Join(parts, " ")
}
