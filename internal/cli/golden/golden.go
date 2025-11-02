// Package golden provides utilities for golden file testing.
package golden

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// Golden provides utilities for comparing test output with golden files.
type Golden struct {
	t       *testing.T
	dir     string
	fixture string
}

// New creates a new Golden instance for the given test and fixture name.
// The fixture name is used to organize golden files into subdirectories.
func New(t *testing.T, fixture string) *Golden {
	t.Helper()
	return &Golden{
		t:       t,
		dir:     filepath.Join("testdata", "golden"),
		fixture: fixture,
	}
}

// Assert compares the given output with the golden file.
// If the --update flag is set, it updates the golden file instead.
func (g *Golden) Assert(name string, got []byte) {
	g.t.Helper()

	goldenPath := filepath.Join(g.dir, g.fixture, name+".golden")

	if *update {
		err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
		require.NoError(g.t, err, "failed to create golden directory")

		err = os.WriteFile(goldenPath, got, 0600)
		require.NoError(g.t, err, "failed to write golden file")
		g.t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(g.t, err, "failed to read golden file: %s (run with -update to create)", goldenPath)

	if !assert.Equal(g.t, string(want), string(got), "output mismatch for %s", name) {
		g.t.Logf("To update golden files: go test -update")
		g.t.Logf("Golden file: %s", goldenPath)
	}
}

// AssertString is a convenience wrapper around Assert for string output.
func (g *Golden) AssertString(name string, got string) {
	g.t.Helper()
	g.Assert(name, []byte(got))
}

// Path returns the path to a golden file for the given name.
func (g *Golden) Path(name string) string {
	return filepath.Join(g.dir, g.fixture, name+".golden")
}

// Exists checks if a golden file exists for the given name.
func (g *Golden) Exists(name string) bool {
	_, err := os.Stat(g.Path(name))
	return err == nil
}
