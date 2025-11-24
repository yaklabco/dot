package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

func TestDirCopy_Execute(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Create source directory with files
	require.NoError(t, fs.MkdirAll(ctx, "/source/mydir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/mydir/file1.txt", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/source/mydir/file2.txt", []byte("content2"), 0644))

	// Create destination parent
	require.NoError(t, fs.MkdirAll(ctx, "/dest", 0755))

	source := domain.MustParsePath("/source/mydir")
	dest := domain.MustParsePath("/dest/mydir")

	op := domain.NewDirCopy("copy1", source, dest)

	// Execute
	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify destination exists with files
	assert.True(t, fs.Exists(ctx, "/dest/mydir/file1.txt"))
	assert.True(t, fs.Exists(ctx, "/dest/mydir/file2.txt"))

	// Verify source still exists (copy, not move!)
	assert.True(t, fs.Exists(ctx, "/source/mydir/file1.txt"))
	assert.True(t, fs.Exists(ctx, "/source/mydir/file2.txt"))
}

func TestDirCopy_Rollback(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Create destination directory
	require.NoError(t, fs.MkdirAll(ctx, "/dest/copied", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/dest/copied/file.txt", []byte("data"), 0644))

	source := domain.MustParsePath("/source/dir")
	dest := domain.MustParsePath("/dest/copied")

	op := domain.NewDirCopy("copy1", source, dest)

	// Rollback should remove destination
	err := op.Rollback(ctx, fs)
	require.NoError(t, err)

	// Verify destination was removed
	assert.False(t, fs.Exists(ctx, "/dest/copied"))
}

func TestDirCopy_Validate(t *testing.T) {
	source := domain.MustParsePath("/source")
	dest := domain.MustParsePath("/dest")

	op := domain.NewDirCopy("copy1", source, dest)
	err := op.Validate()
	assert.NoError(t, err)

	// Empty ID should fail
	op2 := domain.NewDirCopy("", source, dest)
	err = op2.Validate()
	assert.Error(t, err)
}

func TestDirCopy_String(t *testing.T) {
	source := domain.MustParsePath("/source/dir")
	dest := domain.MustParsePath("/dest/dir")

	op := domain.NewDirCopy("copy1", source, dest)

	str := op.String()
	assert.Contains(t, str, "copy")
	assert.Contains(t, str, "/source/dir")
	assert.Contains(t, str, "/dest/dir")
}

func TestDirCopy_Equals(t *testing.T) {
	source1 := domain.MustParsePath("/source")
	dest1 := domain.MustParsePath("/dest")
	source2 := domain.MustParsePath("/other")

	op1 := domain.NewDirCopy("copy1", source1, dest1)
	op2 := domain.NewDirCopy("copy2", source1, dest1)
	op3 := domain.NewDirCopy("copy3", source2, dest1)

	assert.True(t, op1.Equals(op2), "Same source and dest should be equal")
	assert.False(t, op1.Equals(op3), "Different source should not be equal")
}

func TestDirCopy_NestedDirectories(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Create nested structure
	require.NoError(t, fs.MkdirAll(ctx, "/source/dir/sub1/sub2", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/dir/file.txt", []byte("root"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/source/dir/sub1/file.txt", []byte("sub1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/source/dir/sub1/sub2/file.txt", []byte("sub2"), 0644))

	require.NoError(t, fs.MkdirAll(ctx, "/dest", 0755))

	source := domain.MustParsePath("/source/dir")
	dest := domain.MustParsePath("/dest/dir")

	op := domain.NewDirCopy("copy1", source, dest)

	err := op.Execute(ctx, fs)
	require.NoError(t, err)

	// Verify all levels copied
	assert.True(t, fs.Exists(ctx, "/dest/dir/file.txt"))
	assert.True(t, fs.Exists(ctx, "/dest/dir/sub1/file.txt"))
	assert.True(t, fs.Exists(ctx, "/dest/dir/sub1/sub2/file.txt"))

	// Verify source still exists
	assert.True(t, fs.Exists(ctx, "/source/dir/file.txt"))
	assert.True(t, fs.Exists(ctx, "/source/dir/sub1/sub2/file.txt"))
}

func TestDirCopy_ID(t *testing.T) {
	source := domain.MustParsePath("/src")
	dest := domain.MustParsePath("/dst")

	op := domain.NewDirCopy("my-copy-id", source, dest)

	assert.Equal(t, domain.OperationID("my-copy-id"), op.ID())
}

func TestDirCopy_Kind(t *testing.T) {
	source := domain.MustParsePath("/src")
	dest := domain.MustParsePath("/dst")

	op := domain.NewDirCopy("copy1", source, dest)

	assert.Equal(t, domain.OpKindDirCopy, op.Kind())
}

func TestDirCopy_Dependencies(t *testing.T) {
	source := domain.MustParsePath("/src")
	dest := domain.MustParsePath("/dst")

	op := domain.NewDirCopy("copy1", source, dest)

	deps := op.Dependencies()
	assert.Empty(t, deps, "DirCopy should have no dependencies")
}
