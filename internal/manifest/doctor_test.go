package manifest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManifest_DoctorState(t *testing.T) {
	m := New()

	// Test EnsureDoctorState
	assert.Nil(t, m.Doctor)
	m.EnsureDoctorState()
	assert.NotNil(t, m.Doctor)
	assert.NotNil(t, m.Doctor.IgnoredLinks)
	assert.NotNil(t, m.Doctor.IgnoredPatterns)

	// Test AddIgnoredLink
	m.AddIgnoredLink(".cargo/bin/rustup", "/usr/local/cargo/bin/rustup", "System managed")
	assert.Len(t, m.Doctor.IgnoredLinks, 1)

	link, exists := m.Doctor.IgnoredLinks[".cargo/bin/rustup"]
	assert.True(t, exists)
	assert.Equal(t, "/usr/local/cargo/bin/rustup", link.Target)
	assert.Equal(t, "System managed", link.Reason)
	assert.NotEmpty(t, link.TargetHash)
	assert.False(t, link.AcknowledgedAt.IsZero())

	// Test AddIgnoredPattern
	m.AddIgnoredPattern("*/node_modules/*")
	assert.Len(t, m.Doctor.IgnoredPatterns, 1)
	assert.Equal(t, "*/node_modules/*", m.Doctor.IgnoredPatterns[0])

	// Test RemoveIgnoredLink
	removed := m.RemoveIgnoredLink(".cargo/bin/rustup")
	assert.True(t, removed)
	assert.Len(t, m.Doctor.IgnoredLinks, 0)

	// Test RemoveIgnoredLink on non-existent link
	removed = m.RemoveIgnoredLink("nonexistent")
	assert.False(t, removed)

	// Test RemoveIgnoredLink on nil doctor state
	m.Doctor = nil
	removed = m.RemoveIgnoredLink("anything")
	assert.False(t, removed)
}

func TestIgnoredLink_TargetHash(t *testing.T) {
	m := New()
	m.AddIgnoredLink("test", "/target/path", "")

	link1 := m.Doctor.IgnoredLinks["test"]
	assert.NotEmpty(t, link1.TargetHash)

	// Add same path again with different target - should have different hash
	m.AddIgnoredLink("test", "/different/target/path", "")
	link2 := m.Doctor.IgnoredLinks["test"]
	assert.NotEqual(t, link1.TargetHash, link2.TargetHash)
}

func TestManifest_DoctorState_MultipleOperations(t *testing.T) {
	m := New()

	// Add multiple ignored links
	m.AddIgnoredLink("link1", "/target1", "reason1")
	m.AddIgnoredLink("link2", "/target2", "reason2")
	m.AddIgnoredLink("link3", "/target3", "")

	assert.Len(t, m.Doctor.IgnoredLinks, 3)

	// Add multiple patterns
	m.AddIgnoredPattern("pattern1")
	m.AddIgnoredPattern("pattern2")

	assert.Len(t, m.Doctor.IgnoredPatterns, 2)

	// Verify UpdatedAt is updated
	oldTime := m.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure time difference
	m.AddIgnoredLink("link4", "/target4", "")
	assert.True(t, m.UpdatedAt.After(oldTime))

	// Remove a link
	oldTime = m.UpdatedAt
	time.Sleep(time.Millisecond)
	removed := m.RemoveIgnoredLink("link2")
	assert.True(t, removed)
	assert.Len(t, m.Doctor.IgnoredLinks, 3)
	assert.True(t, m.UpdatedAt.After(oldTime))
}

func TestManifest_AddIgnoredPattern_Idempotent(t *testing.T) {
	m := New()

	m.AddIgnoredPattern("Code/*")
	m.AddIgnoredPattern("Code/*")

	assert.Equal(t, []string{"Code/*"}, m.Doctor.IgnoredPatterns)
}

func TestManifest_RemoveIgnoredPattern(t *testing.T) {
	m := New()
	m.AddIgnoredPattern("Code/*")
	m.AddIgnoredPattern("Bootstrap/*")

	assert.True(t, m.RemoveIgnoredPattern("Code/*"))
	assert.Equal(t, []string{"Bootstrap/*"}, m.Doctor.IgnoredPatterns)

	assert.False(t, m.RemoveIgnoredPattern("Code/*"))
	assert.False(t, m.RemoveIgnoredPattern("never-added/*"))
}

func TestManifest_AddIgnoredLink_AfterPartialDoctorState(t *testing.T) {
	// Simulates a manifest loaded from JSON where doctor state exists with
	// patterns but ignored_links was absent (nil map after unmarshal).
	m := New()
	m.Doctor = &DoctorState{IgnoredPatterns: []string{"Code/*"}}

	m.AddIgnoredLink(".nix-profile", "/nix/store/abc", "nix managed")

	assert.Len(t, m.Doctor.IgnoredLinks, 1)
}
