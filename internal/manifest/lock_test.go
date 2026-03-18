package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileLock_LockUnlock(t *testing.T) {
	tmpDir := t.TempDir()

	lock := NewFileLock(tmpDir)
	err := lock.Lock(1 * time.Second)
	require.NoError(t, err)

	// Lock file should exist
	lockPath := filepath.Join(tmpDir, lockFileName)
	_, err = os.Stat(lockPath)
	assert.NoError(t, err, "lock file should exist while locked")

	err = lock.Unlock()
	require.NoError(t, err)
}

func TestFileLock_SecondLockTimesOut(t *testing.T) {
	tmpDir := t.TempDir()

	lock1 := NewFileLock(tmpDir)
	err := lock1.Lock(1 * time.Second)
	require.NoError(t, err)
	defer lock1.Unlock()

	// Second lock should time out
	lock2 := NewFileLock(tmpDir)
	err = lock2.Lock(200 * time.Millisecond)
	assert.Error(t, err, "second lock should fail with timeout")
	assert.Contains(t, err.Error(), "timeout")
}

func TestFileLock_UnlockAllowsReacquire(t *testing.T) {
	tmpDir := t.TempDir()

	lock1 := NewFileLock(tmpDir)
	err := lock1.Lock(1 * time.Second)
	require.NoError(t, err)

	err = lock1.Unlock()
	require.NoError(t, err)

	// Should be able to acquire again
	lock2 := NewFileLock(tmpDir)
	err = lock2.Lock(1 * time.Second)
	require.NoError(t, err)

	err = lock2.Unlock()
	require.NoError(t, err)
}

func TestFileLock_UnlockIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	lock := NewFileLock(tmpDir)
	err := lock.Lock(1 * time.Second)
	require.NoError(t, err)

	err = lock.Unlock()
	require.NoError(t, err)

	// Second unlock should be no-op
	err = lock.Unlock()
	assert.NoError(t, err)
}
