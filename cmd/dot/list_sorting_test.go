package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestSortPackages_ByName(t *testing.T) {
	packages := []dot.PackageInfo{
		{Name: "zsh"},
		{Name: "bash"},
		{Name: "vim"},
	}

	sortPackages(packages, "name")

	assert.Equal(t, "bash", packages[0].Name)
	assert.Equal(t, "vim", packages[1].Name)
	assert.Equal(t, "zsh", packages[2].Name)
}

func TestSortPackages_ByLinks(t *testing.T) {
	packages := []dot.PackageInfo{
		{Name: "a", LinkCount: 5},
		{Name: "b", LinkCount: 10},
		{Name: "c", LinkCount: 3},
	}

	sortPackages(packages, "links")

	assert.Equal(t, 10, packages[0].LinkCount)
	assert.Equal(t, 5, packages[1].LinkCount)
	assert.Equal(t, 3, packages[2].LinkCount)
}

func TestSortPackages_ByDate(t *testing.T) {
	now := time.Now()
	packages := []dot.PackageInfo{
		{Name: "a", InstalledAt: now.Add(-2 * time.Hour)},
		{Name: "b", InstalledAt: now},
		{Name: "c", InstalledAt: now.Add(-1 * time.Hour)},
	}

	sortPackages(packages, "date")

	assert.Equal(t, "b", packages[0].Name) // Most recent first
	assert.Equal(t, "c", packages[1].Name)
	assert.Equal(t, "a", packages[2].Name)
}

func TestSortPackages_InvalidSortBy(t *testing.T) {
	packages := []dot.PackageInfo{
		{Name: "zsh"},
		{Name: "bash"},
	}

	// Should default to name sorting
	sortPackages(packages, "invalid")

	assert.Equal(t, "bash", packages[0].Name)
	assert.Equal(t, "zsh", packages[1].Name)
}
