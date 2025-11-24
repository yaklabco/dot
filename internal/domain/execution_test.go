package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/internal/domain"
)

func TestExecutionResult_Success(t *testing.T) {
	result := domain.ExecutionResult{
		Executed: []domain.OperationID{"op1", "op2"},
		Failed:   []domain.OperationID{},
	}

	assert.True(t, result.Success())
	assert.False(t, result.PartialFailure())
}

func TestExecutionResult_PartialFailure(t *testing.T) {
	result := domain.ExecutionResult{
		Executed: []domain.OperationID{"op1"},
		Failed:   []domain.OperationID{"op2"},
	}

	assert.False(t, result.Success())
	assert.True(t, result.PartialFailure())
}
