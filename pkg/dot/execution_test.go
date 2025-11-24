package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestExecutionResult_Success(t *testing.T) {
	result := dot.ExecutionResult{
		Executed: []dot.OperationID{"op1", "op2"},
		Failed:   []dot.OperationID{},
	}

	assert.True(t, result.Success())
	assert.False(t, result.PartialFailure())
}

func TestExecutionResult_PartialFailure(t *testing.T) {
	result := dot.ExecutionResult{
		Executed: []dot.OperationID{"op1"},
		Failed:   []dot.OperationID{"op2"},
	}

	assert.False(t, result.Success())
	assert.True(t, result.PartialFailure())
}
