package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestDirCopy_Operation(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Create source directory with files
	require.NoError(t, fs.MkdirAll(ctx, "/source/dir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/dir/file1.txt", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/source/dir/file2.txt", []byte("content2"), 0644))
	require.NoError(t, fs.MkdirAll(ctx, "/source/dir/subdir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/dir/subdir/file3.txt", []byte("content3"), 0644))

	// Create destination parent
	require.NoError(t, fs.MkdirAll(ctx, "/dest", 0755))

	source := dot.MustParsePath("/source/dir")
	dest := dot.MustParsePath("/dest/dir")

	op := dot.NewDirCopy("copy1", source, dest)

	// Execute copy
	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify destination has all files
	assert.True(t, fs.Exists(ctx, "/dest/dir/file1.txt"))
	assert.True(t, fs.Exists(ctx, "/dest/dir/file2.txt"))
	assert.True(t, fs.Exists(ctx, "/dest/dir/subdir/file3.txt"))

	// Verify SOURCE still exists (copied, not moved!)
	assert.True(t, fs.Exists(ctx, "/source/dir/file1.txt"), "Source should still exist after copy")
	assert.True(t, fs.Exists(ctx, "/source/dir/file2.txt"), "Source should still exist after copy")
	assert.True(t, fs.Exists(ctx, "/source/dir/subdir/file3.txt"), "Source should still exist after copy")

	// Verify contents
	data, _ := fs.ReadFile(ctx, "/dest/dir/file1.txt")
	assert.Equal(t, []byte("content1"), data)

	data, _ = fs.ReadFile(ctx, "/source/dir/file1.txt")
	assert.Equal(t, []byte("content1"), data, "Source should have same content")
}
