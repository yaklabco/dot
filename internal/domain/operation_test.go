package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

func TestLinkCreateOperation(t *testing.T) {
	source := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()

	op := domain.NewLinkCreate("link1", source, target)

	assert.Equal(t, domain.OpKindLinkCreate, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	// Validate checks operation structure, not filesystem state
	err := op.Validate()
	assert.NoError(t, err)

	// Dependencies should be empty for link creation
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestLinkDeleteOperation(t *testing.T) {
	target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()

	op := domain.NewLinkDelete("link1", target)

	assert.Equal(t, domain.OpKindLinkDelete, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestDirCreateOperation(t *testing.T) {
	path := domain.NewFilePath("/home/user/.vim").Unwrap()

	op := domain.NewDirCreate("dir1", path)

	assert.Equal(t, domain.OpKindDirCreate, op.Kind())
	assert.Contains(t, op.String(), ".vim")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestDirDeleteOperation(t *testing.T) {
	path := domain.NewFilePath("/home/user/.vim").Unwrap()

	op := domain.NewDirDelete("dir1", path)

	assert.Equal(t, domain.OpKindDirDelete, op.Kind())
	assert.Contains(t, op.String(), ".vim")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestFileMoveOperation(t *testing.T) {
	source := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()

	op := domain.NewFileMove("move1", source, dest)

	assert.Equal(t, domain.OpKindFileMove, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestFileBackupOperation(t *testing.T) {
	source := domain.NewFilePath("/home/user/.vimrc").Unwrap()
	backup := domain.NewFilePath("/home/user/.vimrc.backup").Unwrap()

	op := domain.NewFileBackup("backup1", source, backup)

	assert.Equal(t, domain.OpKindFileBackup, op.Kind())
	assert.Contains(t, op.String(), "vimrc")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestFileDeleteOperation(t *testing.T) {
	path := domain.NewFilePath("/home/user/.vimrc").Unwrap()

	op := domain.NewFileDelete("delete1", path)

	assert.Equal(t, domain.OpKindFileDelete, op.Kind())
	assert.Contains(t, op.String(), "delete")
	assert.Contains(t, op.String(), "vimrc")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestOperationEquality(t *testing.T) {
	source := domain.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()

	op1 := domain.NewLinkCreate("link1", source, target)
	op2 := domain.NewLinkCreate("link2", source, target)
	op3 := domain.NewLinkDelete("link3", target)

	assert.True(t, op1.Equals(op2))
	assert.False(t, op1.Equals(op3))
}

func TestLinkCreateEquals(t *testing.T) {
	source1 := domain.NewFilePath("/home/user/.dotfiles/vimrc").Unwrap()
	target1 := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	source2 := domain.NewFilePath("/home/user/.dotfiles/bashrc").Unwrap()

	op1 := domain.NewLinkCreate("link1", source1, target1)
	op2 := domain.NewLinkCreate("link2", source1, target1)
	op3 := domain.NewLinkCreate("link3", source2, target1)
	op4 := domain.NewLinkDelete("link4", target1)

	assert.True(t, op1.Equals(op2), "same source and target should be equal")
	assert.False(t, op1.Equals(op3), "different source should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestLinkDeleteEquals(t *testing.T) {
	target1 := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	target2 := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	dirPath := domain.NewFilePath("/home/user/.vim").Unwrap()

	op1 := domain.NewLinkDelete("link1", target1)
	op2 := domain.NewLinkDelete("link2", target1)
	op3 := domain.NewLinkDelete("link3", target2)
	op4 := domain.NewDirDelete("dir1", dirPath)

	assert.True(t, op1.Equals(op2), "same target should be equal")
	assert.False(t, op1.Equals(op3), "different target should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestDirCreateEquals(t *testing.T) {
	path1 := domain.NewFilePath("/home/user/.vim").Unwrap()
	path2 := domain.NewFilePath("/home/user/.config").Unwrap()

	op1 := domain.NewDirCreate("dir1", path1)
	op2 := domain.NewDirCreate("dir2", path1)
	op3 := domain.NewDirCreate("dir3", path2)
	op4 := domain.NewDirDelete("dir4", path1)

	assert.True(t, op1.Equals(op2), "same path should be equal")
	assert.False(t, op1.Equals(op3), "different path should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestDirDeleteEquals(t *testing.T) {
	path1 := domain.NewFilePath("/home/user/.vim").Unwrap()
	path2 := domain.NewFilePath("/home/user/.config").Unwrap()

	op1 := domain.NewDirDelete("dir1", path1)
	op2 := domain.NewDirDelete("dir2", path1)
	op3 := domain.NewDirDelete("dir3", path2)
	op4 := domain.NewDirCreate("dir4", path1)

	assert.True(t, op1.Equals(op2), "same path should be equal")
	assert.False(t, op1.Equals(op3), "different path should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestDirRemoveAllOperation(t *testing.T) {
	path := domain.NewFilePath("/home/user/.config").Unwrap()

	op := domain.NewDirRemoveAll("dir1", path)

	assert.Equal(t, domain.OpKindDirRemoveAll, op.Kind())
	assert.Contains(t, op.String(), ".config")
	assert.Contains(t, op.String(), "recursively")

	err := op.Validate()
	assert.NoError(t, err)

	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestDirRemoveAllEquals(t *testing.T) {
	path1 := domain.NewFilePath("/home/user/.vim").Unwrap()
	path2 := domain.NewFilePath("/home/user/.config").Unwrap()

	op1 := domain.NewDirRemoveAll("dir1", path1)
	op2 := domain.NewDirRemoveAll("dir2", path1)
	op3 := domain.NewDirRemoveAll("dir3", path2)
	op4 := domain.NewDirDelete("dir4", path1)

	assert.True(t, op1.Equals(op2), "same path should be equal")
	assert.False(t, op1.Equals(op3), "different path should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestDirRemoveAllOperationID(t *testing.T) {
	path := domain.NewFilePath("/home/user/.config").Unwrap()
	op := domain.NewDirRemoveAll("dir1", path)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestFileMoveEquals(t *testing.T) {
	source1 := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest1 := domain.NewFilePath("/home/user/.dotfiles/vimrc").Unwrap()
	source2 := domain.NewTargetPath("/home/user/.bashrc").Unwrap()
	backupSource := domain.NewFilePath("/home/user/.vimrc").Unwrap()

	op1 := domain.NewFileMove("move1", source1, dest1)
	op2 := domain.NewFileMove("move2", source1, dest1)
	op3 := domain.NewFileMove("move3", source2, dest1)
	op4 := domain.NewFileBackup("backup1", backupSource, dest1)

	assert.True(t, op1.Equals(op2), "same source and dest should be equal")
	assert.False(t, op1.Equals(op3), "different source should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestFileBackupEquals(t *testing.T) {
	source1 := domain.NewFilePath("/home/user/.vimrc").Unwrap()
	backup1 := domain.NewFilePath("/home/user/.vimrc.backup").Unwrap()
	source2 := domain.NewFilePath("/home/user/.bashrc").Unwrap()

	sourceTarget1 := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest1 := domain.NewFilePath("/home/user/.dotfiles/vimrc").Unwrap()
	op1 := domain.NewFileBackup("backup1", source1, backup1)
	op2 := domain.NewFileBackup("backup2", source1, backup1)
	op3 := domain.NewFileBackup("backup3", source2, backup1)
	op4 := domain.NewFileMove("move1", sourceTarget1, dest1)

	assert.True(t, op1.Equals(op2), "same source and backup should be equal")
	assert.False(t, op1.Equals(op3), "different source should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestFileDeleteEquals(t *testing.T) {
	path1 := domain.NewFilePath("/home/user/.vimrc").Unwrap()
	path2 := domain.NewFilePath("/home/user/.bashrc").Unwrap()

	op1 := domain.NewFileDelete("delete1", path1)
	op2 := domain.NewFileDelete("delete2", path1)
	op3 := domain.NewFileDelete("delete3", path2)
	op4 := domain.NewFileBackup("backup1", path1, path2)

	assert.True(t, op1.Equals(op2), "same path should be equal")
	assert.False(t, op1.Equals(op3), "different path should not be equal")
	assert.False(t, op1.Equals(op4), "different operation type should not be equal")
}

func TestOperationKindString(t *testing.T) {
	tests := []struct {
		kind domain.OperationKind
		want string
	}{
		{domain.OpKindLinkCreate, "LinkCreate"},
		{domain.OpKindLinkDelete, "LinkDelete"},
		{domain.OpKindDirCreate, "DirCreate"},
		{domain.OpKindDirDelete, "DirDelete"},
		{domain.OpKindDirRemoveAll, "DirRemoveAll"},
		{domain.OpKindFileMove, "FileMove"},
		{domain.OpKindFileBackup, "FileBackup"},
		{domain.OpKindFileDelete, "FileDelete"},
		{domain.OpKindDirCopy, "DirCopy"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kind.String())
		})
	}
}

func TestLinkCreateOperationID(t *testing.T) {
	source := domain.NewFilePath("/packages/vim/.vimrc").Unwrap()
	dest := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	op := domain.NewLinkCreate("link1", source, dest)

	id := op.ID()
	assert.NotEmpty(t, id)
	// ID should be deterministic
	assert.Equal(t, op.ID(), op.ID())
}

func TestLinkDeleteOperationID(t *testing.T) {
	link := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	op := domain.NewLinkDelete("link1", link)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestDirCreateOperationID(t *testing.T) {
	path := domain.NewFilePath("/home/user/.config").Unwrap()
	op := domain.NewDirCreate("dir1", path)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestDirDeleteOperationID(t *testing.T) {
	path := domain.NewFilePath("/home/user/.config").Unwrap()
	op := domain.NewDirDelete("dir1", path)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestFileMoveOperationID(t *testing.T) {
	source := domain.NewTargetPath("/home/user/.vimrc").Unwrap()
	dest := domain.NewFilePath("/packages/vim/.vimrc").Unwrap()
	op := domain.NewFileMove("move1", source, dest)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}

func TestFileBackupOperationID(t *testing.T) {
	source := domain.NewFilePath("/home/user/.vimrc").Unwrap()
	dest := domain.NewFilePath("/home/user/.vimrc.backup").Unwrap()
	op := domain.NewFileBackup("backup1", source, dest)

	id := op.ID()
	assert.NotEmpty(t, id)
	assert.Equal(t, op.ID(), op.ID())
}
