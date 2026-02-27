package scanner

import (
	"path/filepath"
	"strings"
)

// TranslateDotfile converts "dot-filename" to ".filename".
// Files with "dot-" prefix become dotfiles in the target directory.
//
// Examples:
//   - "dot-vimrc" -> ".vimrc"
//   - "dot-bashrc" -> ".bashrc"
//   - "README.md" -> "README.md" (no change)
func TranslateDotfile(name string) string {
	if strings.HasPrefix(name, "dot-") {
		return "." + name[4:] // Replace "dot-" with "."
	}
	return name
}

// UntranslateDotfile converts ".filename" to "dot-filename".
// This is the reverse operation of TranslateDotfile.
//
// Examples:
//   - ".vimrc" -> "dot-vimrc"
//   - ".bashrc" -> "dot-bashrc"
//   - "README.md" -> "README.md" (no change)
func UntranslateDotfile(name string) string {
	if strings.HasPrefix(name, ".") && len(name) > 1 {
		return "dot-" + name[1:] // Replace "." with "dot-"
	}
	return name
}

// TranslatePath translates the last component of a path if it has dot- prefix.
// This handles paths like "vim/dot-vimrc" -> "vim/.vimrc".
//
// The function only translates the final component (base name), leaving
// directory components unchanged.
func TranslatePath(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	translated := TranslateDotfile(base)

	if dir == "." {
		return translated
	}

	return filepath.Join(dir, translated)
}

// TranslatePathAll translates ALL components of a path that have the dot- prefix.
// Unlike TranslatePath which only translates the leaf (base name), this function
// translates every path component, including intermediate directories.
//
// Examples:
//   - "dot-config/a/b/c/file.txt" -> ".config/a/b/c/file.txt"
//   - "deep/dot-config/nested/dot-file" -> "deep/.config/nested/.file"
//   - "dot-vimrc" -> ".vimrc"
func TranslatePathAll(path string) string {
	components := splitPathComponents(path)
	for i, comp := range components {
		components[i] = TranslateDotfile(comp)
	}
	return filepath.Join(components...)
}

// splitPathComponents splits a file path into its individual components.
func splitPathComponents(path string) []string {
	var components []string
	for {
		dir, file := filepath.Split(path)
		if file != "" {
			components = append([]string{file}, components...)
		}
		if dir == "" || dir == "/" {
			break
		}
		path = filepath.Clean(dir)
	}
	return components
}

// UntranslatePath translates the last component of a path if it starts with dot.
// This is the reverse of TranslatePath.
func UntranslatePath(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	untranslated := UntranslateDotfile(base)

	if dir == "." {
		return untranslated
	}

	return filepath.Join(dir, untranslated)
}

// TranslatePackageName converts package names to target directory names.
// Package names with "dot-" prefix become dotfiles in the target directory.
//
// This enables intuitive package naming where "dot-gnupg" targets ~/.gnupg/
// instead of requiring redundant nesting like dot-gnupg/dot-gnupg/.
//
// Examples:
//   - "dot-gnupg" -> ".gnupg"
//   - "dot-config" -> ".config"
//   - "vim" -> "vim"
//   - "" -> ""
func TranslatePackageName(name string) string {
	if strings.HasPrefix(name, "dot-") {
		return "." + name[4:] // Replace "dot-" with "."
	}
	return name
}
