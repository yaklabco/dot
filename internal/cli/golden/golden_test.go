package golden

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoldenAssert(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Override golden dir for testing
	g := &Golden{
		t:       t,
		dir:     tmpDir,
		fixture: "test",
	}

	// Create golden file
	goldenPath := filepath.Join(tmpDir, "test", "example.golden")
	err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
	require.NoError(t, err)

	expected := []byte("test output\n")
	err = os.WriteFile(goldenPath, expected, 0644)
	require.NoError(t, err)

	// Test matching output
	g.Assert("example", expected)
}

func TestGoldenAssertString(t *testing.T) {
	tmpDir := t.TempDir()

	g := &Golden{
		t:       t,
		dir:     tmpDir,
		fixture: "test",
	}

	goldenPath := filepath.Join(tmpDir, "test", "string.golden")
	err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
	require.NoError(t, err)

	expected := "string test output\n"
	err = os.WriteFile(goldenPath, []byte(expected), 0644)
	require.NoError(t, err)

	g.AssertString("string", expected)
}

func TestGoldenPath(t *testing.T) {
	g := New(t, "fixture")

	path := g.Path("test")
	assert.Contains(t, path, "testdata")
	assert.Contains(t, path, "golden")
	assert.Contains(t, path, "fixture")
	assert.Contains(t, path, "test.golden")
}

func TestGoldenExists(t *testing.T) {
	tmpDir := t.TempDir()

	g := &Golden{
		t:       t,
		dir:     tmpDir,
		fixture: "test",
	}

	// Should not exist initially
	assert.False(t, g.Exists("nonexistent"))

	// Create file
	goldenPath := filepath.Join(tmpDir, "test", "exists.golden")
	err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(goldenPath, []byte("test"), 0644)
	require.NoError(t, err)

	// Should exist now
	assert.True(t, g.Exists("exists"))
}
