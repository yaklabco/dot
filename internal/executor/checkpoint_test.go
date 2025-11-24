package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

func TestCheckpoint_Create(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCheckpointStore()

	checkpoint := store.Create(ctx)

	require.NotEmpty(t, checkpoint.ID)
	require.NotZero(t, checkpoint.CreatedAt)
	require.Equal(t, 0, checkpoint.Len())
}

func TestCheckpoint_Record(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCheckpointStore()
	checkpoint := store.Create(ctx)

	source := domain.MustParsePath("/source")
	target := domain.MustParseTargetPath("/target")
	op := domain.NewLinkCreate("link1", source, target)

	checkpoint.Record("link1", op)

	retrieved := checkpoint.Lookup("link1")
	require.NotNil(t, retrieved)
	require.Equal(t, op.ID(), retrieved.ID())
}

func TestCheckpoint_Lookup_NotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCheckpointStore()
	checkpoint := store.Create(ctx)

	retrieved := checkpoint.Lookup("nonexistent")
	require.Nil(t, retrieved)
}

func TestMemoryCheckpointStore_Restore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCheckpointStore()

	// Create checkpoint
	checkpoint := store.Create(ctx)
	id := checkpoint.ID

	// Add operation
	source := domain.MustParsePath("/source")
	target := domain.MustParseTargetPath("/target")
	op := domain.NewLinkCreate("link1", source, target)
	checkpoint.Record("link1", op)

	// Restore checkpoint
	restored, err := store.Restore(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, restored)
	require.Equal(t, id, restored.ID)
	require.NotNil(t, restored.Lookup("link1"))
}

func TestMemoryCheckpointStore_Restore_NotFound(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCheckpointStore()

	// Try to restore non-existent checkpoint
	_, err := store.Restore(ctx, "nonexistent")
	require.Error(t, err)
	require.IsType(t, domain.ErrCheckpointNotFound{}, err)
}

func TestMemoryCheckpointStore_Delete(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCheckpointStore()

	// Create and delete checkpoint
	checkpoint := store.Create(ctx)
	id := checkpoint.ID

	err := store.Delete(ctx, id)
	require.NoError(t, err)

	// Verify it's gone
	_, err = store.Restore(ctx, id)
	require.Error(t, err)
}
