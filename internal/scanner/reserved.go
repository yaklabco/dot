package scanner

import "strings"

// IsReservedPackageName checks if the given package name is reserved for dot's internal use.
// Reserved names cannot be managed as packages.
func IsReservedPackageName(name string) bool {
	reserved := []string{
		"dot",
		".dot",
		"dot-config",
	}

	nameLower := strings.ToLower(name)
	for _, r := range reserved {
		if nameLower == r {
			return true
		}
	}

	return false
}

// GetReservedPackageReason returns a human-readable reason why a package name is reserved.
func GetReservedPackageReason(name string) string {
	return "Package name is reserved for dot's internal use"
}
