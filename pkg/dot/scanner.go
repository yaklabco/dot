package dot

import "github.com/yaklabco/dot/internal/scanner"

// UntranslateDotfile converts a dotfile name to its untranslated package form.
// Example: ".ssh" -> "dot-ssh"
func UntranslateDotfile(name string) string {
	return scanner.UntranslateDotfile(name)
}
