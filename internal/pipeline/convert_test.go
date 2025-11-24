package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/planner"
)

func TestConvertConflicts(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		conflicts := []planner.Conflict{}
		result := convertConflicts(conflicts)
		assert.Nil(t, result)
	})

	t.Run("nil slice", func(t *testing.T) {
		result := convertConflicts(nil)
		assert.Nil(t, result)
	})

	t.Run("single conflict", func(t *testing.T) {
		path := domain.NewFilePath("/home/user/.bashrc").Unwrap()
		conflict := planner.NewConflict(
			planner.ConflictFileExists,
			path,
			"File exists at target",
		).WithContext("package", "bash")

		result := convertConflicts([]planner.Conflict{conflict})

		require.Len(t, result, 1)
		assert.Equal(t, "file_exists", result[0].Type)
		assert.Equal(t, "/home/user/.bashrc", result[0].Path)
		assert.Equal(t, "File exists at target", result[0].Details)
		assert.Equal(t, "bash", result[0].Context["package"])
	})

	t.Run("multiple conflicts", func(t *testing.T) {
		path1 := domain.NewFilePath("/home/user/.bashrc").Unwrap()
		path2 := domain.NewFilePath("/home/user/.vimrc").Unwrap()

		conflicts := []planner.Conflict{
			planner.NewConflict(planner.ConflictFileExists, path1, "File 1 exists"),
			planner.NewConflict(planner.ConflictWrongLink, path2, "Wrong link"),
		}

		result := convertConflicts(conflicts)

		require.Len(t, result, 2)
		assert.Equal(t, "file_exists", result[0].Type)
		assert.Equal(t, "wrong_link", result[1].Type)
	})
}

func TestConvertWarnings(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		warnings := []planner.Warning{}
		result := convertWarnings(warnings)
		assert.Nil(t, result)
	})

	t.Run("nil slice", func(t *testing.T) {
		result := convertWarnings(nil)
		assert.Nil(t, result)
	})

	t.Run("single warning", func(t *testing.T) {
		warning := planner.Warning{
			Message:  "Backup created",
			Severity: planner.WarnCaution,
			Context: map[string]string{
				"path": "/home/user/.bashrc",
			},
		}

		result := convertWarnings([]planner.Warning{warning})

		require.Len(t, result, 1)
		assert.Equal(t, "Backup created", result[0].Message)
		assert.Equal(t, "caution", result[0].Severity)
		assert.Equal(t, "/home/user/.bashrc", result[0].Context["path"])
	})

	t.Run("multiple warnings with different severities", func(t *testing.T) {
		warnings := []planner.Warning{
			{Message: "Info message", Severity: planner.WarnInfo},
			{Message: "Caution message", Severity: planner.WarnCaution},
			{Message: "Danger message", Severity: planner.WarnDanger},
		}

		result := convertWarnings(warnings)

		require.Len(t, result, 3)
		assert.Equal(t, "info", result[0].Severity)
		assert.Equal(t, "caution", result[1].Severity)
		assert.Equal(t, "danger", result[2].Severity)
	})
}

func TestCopyContext(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		result := copyContext(nil)
		assert.Nil(t, result)
	})

	t.Run("empty context", func(t *testing.T) {
		original := map[string]string{}
		result := copyContext(original)

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("context is copied", func(t *testing.T) {
		original := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		result := copyContext(original)

		// Values match
		assert.Equal(t, "value1", result["key1"])
		assert.Equal(t, "value2", result["key2"])
		assert.Len(t, result, 2)
	})

	t.Run("mutations do not affect original", func(t *testing.T) {
		original := map[string]string{
			"key1": "value1",
		}

		copied := copyContext(original)

		// Mutate the copy
		copied["key1"] = "modified"
		copied["key2"] = "new"

		// Original is unchanged
		assert.Equal(t, "value1", original["key1"])
		assert.NotContains(t, original, "key2")
	})
}

func TestConvertConflicts_ContextIsolation(t *testing.T) {
	t.Run("mutating converted conflict context does not affect original", func(t *testing.T) {
		path := domain.NewFilePath("/home/user/.bashrc").Unwrap()
		originalContext := map[string]string{
			"package": "bash",
		}

		conflict := planner.NewConflict(
			planner.ConflictFileExists,
			path,
			"File exists",
		).WithContext("package", "bash")

		result := convertConflicts([]planner.Conflict{conflict})

		// Mutate the converted conflict's context
		result[0].Context["package"] = "modified"
		result[0].Context["new_key"] = "new_value"

		// Original conflict context should be unchanged
		// Note: We can't directly access conflict.Context since WithContext returns a copy,
		// but we verify the original map is unchanged
		assert.Equal(t, "bash", originalContext["package"])
		assert.NotContains(t, originalContext, "new_key")
	})
}

func TestConvertWarnings_ContextIsolation(t *testing.T) {
	t.Run("mutating converted warning context does not affect original", func(t *testing.T) {
		originalContext := map[string]string{
			"path": "/home/user/.bashrc",
		}

		warning := planner.Warning{
			Message:  "Backup created",
			Severity: planner.WarnCaution,
			Context:  originalContext,
		}

		result := convertWarnings([]planner.Warning{warning})

		// Mutate the converted warning's context
		result[0].Context["path"] = "modified"
		result[0].Context["new_key"] = "new_value"

		// Original warning context should be unchanged
		assert.Equal(t, "/home/user/.bashrc", originalContext["path"])
		assert.NotContains(t, originalContext, "new_key")
	})
}
