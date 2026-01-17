package scanner

import "strings"

var reservedNames = map[string]struct{}{
	"dot":        {},
	".dot":       {},
	"dot-config": {},
}

// IsReservedPackageName checks if a package name is reserved for dot's use.
func IsReservedPackageName(name string) bool {
	_, exists := reservedNames[strings.ToLower(name)]
	return exists
}

// GetReservedPackageReason returns a human-readable reason why a package name is reserved.
func GetReservedPackageReason(name string) string {
	return "Package name is reserved for dot's internal use"
}
