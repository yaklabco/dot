package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestLinkCreateOperation(t *testing.T) {
	source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := dot.NewTargetPath("/home/user/.vimrc").Unwrap()

	op := dot.NewLinkCreate("link1", source, target)

	assert.Equal(t, dot.OpKindLinkCreate, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	// Validate checks operation structure, not filesystem state
	err := op.Validate()
	assert.NoError(t, err)

	// Dependencies should be empty for link creation
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestLinkDeleteOperation(t *testing.T) {
	target := dot.NewTargetPath("/home/user/.vimrc").Unwrap()

	op := dot.NewLinkDelete("link1", target)

	assert.Equal(t, dot.OpKindLinkDelete, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestDirCreateOperation(t *testing.T) {
	path := dot.NewFilePath("/home/user/.vim").Unwrap()

	op := dot.NewDirCreate("dir1", path)

	assert.Equal(t, dot.OpKindDirCreate, op.Kind())
	assert.Contains(t, op.String(), ".vim")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestDirDeleteOperation(t *testing.T) {
	path := dot.NewFilePath("/home/user/.vim").Unwrap()

	op := dot.NewDirDelete("dir1", path)

	assert.Equal(t, dot.OpKindDirDelete, op.Kind())
	assert.Contains(t, op.String(), ".vim")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestFileMoveOperation(t *testing.T) {
	source := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()

	op := dot.NewFileMove("move1", source, dest)

	assert.Equal(t, dot.OpKindFileMove, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestFileBackupOperation(t *testing.T) {
	source := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	backup := dot.NewFilePath("/home/user/.vimrc.backup").Unwrap()

	op := dot.NewFileBackup("backup1", source, backup)

	assert.Equal(t, dot.OpKindFileBackup, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestOperationEquality(t *testing.T) {
	source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := dot.NewTargetPath("/home/user/.vimrc").Unwrap()

	op1 := dot.NewLinkCreate("link1", source, target)
	op2 := dot.NewLinkCreate("link2", source, target)
	op3 := dot.NewLinkDelete("link3", target)

	assert.True(t, op1.Equals(op2))
	assert.False(t, op1.Equals(op3))
}

func TestLinkCreateEquals(t *testing.T) {
	source1 := dot.NewFilePath("/home/user/.dotfiles/vimrc").Unwrap()
	target1 := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	source2 := dot.NewFilePath("/home/user/.dotfiles/bashrc").Unwrap()

	op1 := dot.NewLinkCreate("link1", source1, target1)
	op2 := dot.NewLinkCreate("link2", source1, target1)
	op3 := dot.NewLinkCreate("link3", source2, target1)
	op4 := dot.NewLinkDelete("link4", target1)

	assert.True(t, op1.Equals(op2), "same source and target should be equal")
	assert.False(t, op1.Equals(op3), "different source should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestLinkDeleteEquals(t *testing.T) {
	target1 := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	target2 := dot.NewTargetPath("/home/user/.bashrc").Unwrap()
	dirPath := dot.NewFilePath("/home/user/.vim").Unwrap()

	op1 := dot.NewLinkDelete("link1", target1)
	op2 := dot.NewLinkDelete("link2", target1)
	op3 := dot.NewLinkDelete("link3", target2)
	op4 := dot.NewDirDelete("dir1", dirPath)

	assert.True(t, op1.Equals(op2), "same target should be equal")
	assert.False(t, op1.Equals(op3), "different target should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestDirCreateEquals(t *testing.T) {
	path1 := dot.NewFilePath("/home/user/.vim").Unwrap()
	path2 := dot.NewFilePath("/home/user/.config").Unwrap()

	op1 := dot.NewDirCreate("dir1", path1)
	op2 := dot.NewDirCreate("dir2", path1)
	op3 := dot.NewDirCreate("dir3", path2)
	op4 := dot.NewDirDelete("dir4", path1)

	assert.True(t, op1.Equals(op2), "same path should be equal")
	assert.False(t, op1.Equals(op3), "different path should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestDirDeleteEquals(t *testing.T) {
	path1 := dot.NewFilePath("/home/user/.vim").Unwrap()
	path2 := dot.NewFilePath("/home/user/.config").Unwrap()

	op1 := dot.NewDirDelete("dir1", path1)
	op2 := dot.NewDirDelete("dir2", path1)
	op3 := dot.NewDirDelete("dir3", path2)
	op4 := dot.NewDirCreate("dir4", path1)

	assert.True(t, op1.Equals(op2), "same path should be equal")
	assert.False(t, op1.Equals(op3), "different path should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestFileMoveEquals(t *testing.T) {
	source1 := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest1 := dot.NewFilePath("/home/user/.dotfiles/vimrc").Unwrap()
	source2 := dot.NewTargetPath("/home/user/.bashrc").Unwrap()
	backupSource := dot.NewFilePath("/home/user/.vimrc").Unwrap()

	op1 := dot.NewFileMove("move1", source1, dest1)
	op2 := dot.NewFileMove("move2", source1, dest1)
	op3 := dot.NewFileMove("move3", source2, dest1)
	op4 := dot.NewFileBackup("backup1", backupSource, dest1)

	assert.True(t, op1.Equals(op2), "same source and dest should be equal")
	assert.False(t, op1.Equals(op3), "different source should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestFileBackupEquals(t *testing.T) {
	source1 := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	backup1 := dot.NewFilePath("/home/user/.vimrc.backup").Unwrap()
	source2 := dot.NewFilePath("/home/user/.bashrc").Unwrap()

	sourceTarget1 := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest1 := dot.NewFilePath("/home/user/.dotfiles/vimrc").Unwrap()
	op1 := dot.NewFileBackup("backup1", source1, backup1)
	op2 := dot.NewFileBackup("backup2", source1, backup1)
	op3 := dot.NewFileBackup("backup3", source2, backup1)
	op4 := dot.NewFileMove("move1", sourceTarget1, dest1)

	assert.True(t, op1.Equals(op2), "same source and backup should be equal")
	assert.False(t, op1.Equals(op3), "different source should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestOperationKindString(t *testing.T) {
	tests := []struct {
		kind dot.OperationKind
		want string
	}{
		{dot.OpKindLinkCreate, "LinkCreate"},
		{dot.OpKindLinkDelete, "LinkDelete"},
		{dot.OpKindDirCreate, "DirCreate"},
		{dot.OpKindDirDelete, "DirDelete"},
		{dot.OpKindFileMove, "FileMove"},
		{dot.OpKindFileBackup, "FileBackup"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kind.String())
		})
	}
}

func TestLinkCreateOperationID(t *testing.T) {
	source := dot.NewFilePath("/packages/vim/.vimrc").Unwrap()
	dest := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	op := dot.NewLinkCreate("link1", source, dest)

	id := op.ID()
	assert.NotEmpty(t, id)
	// ID should be deterministic
	assert.Equal(t, op.ID(), op.ID())
}

func TestLinkDeleteOperationID(t *testing.T) {
	link := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	op := dot.NewLinkDelete("link1", link)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestDirCreateOperationID(t *testing.T) {
	path := dot.NewFilePath("/home/user/.config").Unwrap()
	op := dot.NewDirCreate("dir1", path)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestDirDeleteOperationID(t *testing.T) {
	path := dot.NewFilePath("/home/user/.config").Unwrap()
	op := dot.NewDirDelete("dir1", path)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestFileMoveOperationID(t *testing.T) {
	source := dot.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest := dot.NewFilePath("/packages/vim/.vimrc").Unwrap()
	op := dot.NewFileMove("move1", source, dest)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestFileBackupOperationID(t *testing.T) {
	source := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	dest := dot.NewFilePath("/home/user/.vimrc.backup").Unwrap()
	op := dot.NewFileBackup("backup1", source, dest)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}
