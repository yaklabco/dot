package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

func TestConflictInfo(t *testing.T) {
	t.Run("basic construction", func(t *testing.T) {
		info := domain.ConflictInfo{
			Type:    "file_exists",
			Path:    "/home/user/.bashrc",
			Details: "File exists at target location",
			Context: map[string]string{
				"package": "bash",
			},
		}

		assert.Equal(t, "file_exists", info.Type)
		assert.Equal(t, "/home/user/.bashrc", info.Path)
		assert.Equal(t, "File exists at target location", info.Details)
		assert.Equal(t, "bash", info.Context["package"])
	})

	t.Run("without context", func(t *testing.T) {
		info := domain.ConflictInfo{
			Type:    "wrong_link",
			Path:    "/home/user/.vimrc",
			Details: "Link points to wrong location",
		}

		assert.Equal(t, "wrong_link", info.Type)
		assert.Nil(t, info.Context)
	})
}

func TestWarningInfo(t *testing.T) {
	t.Run("basic construction", func(t *testing.T) {
		info := domain.WarningInfo{
			Message:  "Overwriting existing file",
			Severity: "danger",
			Context: map[string]string{
				"path": "/home/user/.bashrc",
			},
		}

		assert.Equal(t, "Overwriting existing file", info.Message)
		assert.Equal(t, "danger", info.Severity)
		assert.Equal(t, "/home/user/.bashrc", info.Context["path"])
	})

	t.Run("without context", func(t *testing.T) {
		info := domain.WarningInfo{
			Message:  "Informational message",
			Severity: "info",
		}

		assert.Equal(t, "Informational message", info.Message)
		assert.Equal(t, "info", info.Severity)
		assert.Nil(t, info.Context)
	})
}
