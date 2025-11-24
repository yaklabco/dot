package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestMustParsePath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected panic for invalid path
			assert.NotNil(t, r)
		}
	}()

	// This should panic with relative path
	dot.MustParsePath("relative/path")

	// If we get here without panic, test should fail
	t.Fatal("Expected panic for relative path")
}

func TestMustParsePath_ValidPath(t *testing.T) {
	path := dot.MustParsePath("/absolute/path")
	assert.Equal(t, "/absolute/path", path.String())
}
