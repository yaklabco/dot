package executor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

func TestExecutionResult_Success(t *testing.T) {
	t.Run("success with no failures", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []domain.OperationID{"op1", "op2"},
			Failed:     []domain.OperationID{},
			RolledBack: []domain.OperationID{},
			Errors:     []error{},
		}

		require.True(t, result.Success())
	})

	t.Run("failure with failed operations", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []domain.OperationID{"op1"},
			Failed:     []domain.OperationID{"op2"},
			RolledBack: []domain.OperationID{},
			Errors:     []error{errors.New("op2 failed")},
		}

		require.False(t, result.Success())
	})

	t.Run("failure with errors but no failed ops", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []domain.OperationID{"op1"},
			Failed:     []domain.OperationID{},
			RolledBack: []domain.OperationID{},
			Errors:     []error{errors.New("some error")},
		}

		require.False(t, result.Success())
	})
}

func TestExecutionResult_PartialFailure(t *testing.T) {
	t.Run("partial failure with some executed and some failed", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []domain.OperationID{"op1", "op2"},
			Failed:     []domain.OperationID{"op3"},
			RolledBack: []domain.OperationID{},
			Errors:     []error{errors.New("op3 failed")},
		}

		require.True(t, result.PartialFailure())
	})

	t.Run("not partial if all succeeded", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []domain.OperationID{"op1", "op2"},
			Failed:     []domain.OperationID{},
			RolledBack: []domain.OperationID{},
			Errors:     []error{},
		}

		require.False(t, result.PartialFailure())
	})

	t.Run("not partial if all failed", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []domain.OperationID{},
			Failed:     []domain.OperationID{"op1", "op2"},
			RolledBack: []domain.OperationID{},
			Errors:     []error{errors.New("all failed")},
		}

		require.False(t, result.PartialFailure())
	})

	t.Run("not partial if nothing executed", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []domain.OperationID{},
			Failed:     []domain.OperationID{},
			RolledBack: []domain.OperationID{},
			Errors:     []error{},
		}

		require.False(t, result.PartialFailure())
	})
}
