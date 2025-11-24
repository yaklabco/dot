package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

func TestOperationKind_String(t *testing.T) {
	tests := []struct {
		kind     domain.OperationKind
		expected string
	}{
		{domain.OpKindLinkCreate, "LinkCreate"},
		{domain.OpKindLinkDelete, "LinkDelete"},
		{domain.OpKindDirCreate, "DirCreate"},
		{domain.OpKindDirDelete, "DirDelete"},
		{domain.OpKindFileMove, "FileMove"},
		{domain.OpKindFileBackup, "FileBackup"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.kind.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLinkCreate_Validate_EmptyID(t *testing.T) {
	source := domain.MustParsePath("/source")
	targetResult := domain.NewTargetPath("/target")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkCreate("", source, target)

	err := op.Validate()
	assert.Error(t, err)
}

func TestLinkDelete_Validate_EmptyID(t *testing.T) {
	targetResult := domain.NewTargetPath("/target")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := domain.NewLinkDelete("", target)

	err := op.Validate()
	assert.Error(t, err)
}

func TestDirCreate_Validate_EmptyID(t *testing.T) {
	path := domain.MustParsePath("/dir")
	op := domain.NewDirCreate("", path)

	err := op.Validate()
	assert.Error(t, err)
}

func TestDirDelete_Validate_EmptyID(t *testing.T) {
	path := domain.MustParsePath("/dir")
	op := domain.NewDirDelete("", path)

	err := op.Validate()
	assert.Error(t, err)
}

func TestFileMove_Validate_EmptyID(t *testing.T) {
	sourceResult := domain.NewTargetPath("/source")
	require.True(t, sourceResult.IsOk())
	source := sourceResult.Unwrap()

	dest := domain.MustParsePath("/dest")

	op := domain.NewFileMove("", source, dest)

	err := op.Validate()
	assert.Error(t, err)
}

func TestFileBackup_Validate_EmptyID(t *testing.T) {
	source := domain.MustParsePath("/source")
	backup := domain.MustParsePath("/backup")

	op := domain.NewFileBackup("", source, backup)

	err := op.Validate()
	assert.Error(t, err)
}

func TestLinkCreate_String(t *testing.T) {
	source := domain.MustParsePath("/pkg/file")
	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())

	op := domain.NewLinkCreate("link1", source, targetResult.Unwrap())

	str := op.String()
	assert.Contains(t, str, "link")
	assert.Contains(t, str, "/pkg/file")
	assert.Contains(t, str, "/target/link")
}

func TestLinkDelete_String(t *testing.T) {
	targetResult := domain.NewTargetPath("/target/link")
	require.True(t, targetResult.IsOk())

	op := domain.NewLinkDelete("del1", targetResult.Unwrap())

	str := op.String()
	assert.Contains(t, str, "delete")
	assert.Contains(t, str, "/target/link")
}

func TestDirCreate_String(t *testing.T) {
	path := domain.MustParsePath("/dir")
	op := domain.NewDirCreate("dir1", path)

	str := op.String()
	assert.Contains(t, str, "create")
	assert.Contains(t, str, "/dir")
}

func TestDirDelete_String(t *testing.T) {
	path := domain.MustParsePath("/dir")
	op := domain.NewDirDelete("del1", path)

	str := op.String()
	assert.Contains(t, str, "delete")
	assert.Contains(t, str, "/dir")
}

func TestFileMove_String(t *testing.T) {
	sourceResult := domain.NewTargetPath("/source/file")
	require.True(t, sourceResult.IsOk())

	dest := domain.MustParsePath("/dest/file")

	op := domain.NewFileMove("move1", sourceResult.Unwrap(), dest)

	str := op.String()
	assert.Contains(t, str, "move")
	assert.Contains(t, str, "/source/file")
	assert.Contains(t, str, "/dest/file")
}

func TestFileBackup_String(t *testing.T) {
	source := domain.MustParsePath("/file")
	backup := domain.MustParsePath("/file.bak")

	op := domain.NewFileBackup("bak1", source, backup)

	str := op.String()
	assert.Contains(t, str, "backup")
	assert.Contains(t, str, "/file")
	assert.Contains(t, str, "/file.bak")
}
