package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestDoctorExitCode(t *testing.T) {
	tests := []struct {
		name     string
		status   dot.HealthStatus
		expected int
	}{
		{
			name:     "healthy returns 0",
			status:   dot.HealthOK,
			expected: 0,
		},
		{
			name:     "warnings returns 1",
			status:   dot.HealthWarnings,
			expected: 1,
		},
		{
			name:     "errors returns 2",
			status:   dot.HealthErrors,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DoctorExitCode(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDoctorResultHolder(t *testing.T) {
	t.Run("not executed returns false", func(t *testing.T) {
		holder := &DoctorResultHolder{}
		assert.False(t, holder.Executed, "should be false when not executed")
	})

	t.Run("executed has correct status", func(t *testing.T) {
		holder := &DoctorResultHolder{
			Executed: true,
			Status:   dot.HealthWarnings,
		}
		assert.True(t, holder.Executed, "should be true when executed")
		assert.Equal(t, dot.HealthWarnings, holder.Status)
	})

	t.Run("context storage and retrieval", func(t *testing.T) {
		holder := &DoctorResultHolder{}
		ctx := WithDoctorResultHolder(context.Background(), holder)

		retrieved := DoctorResultHolderFromContext(ctx)
		assert.NotNil(t, retrieved, "should retrieve holder from context")
		assert.Same(t, holder, retrieved, "should be the same instance")

		// Modify through retrieved pointer
		retrieved.Executed = true
		retrieved.Status = dot.HealthErrors

		// Original should be modified too
		assert.True(t, holder.Executed)
		assert.Equal(t, dot.HealthErrors, holder.Status)
	})

	t.Run("nil context returns nil", func(t *testing.T) {
		retrieved := DoctorResultHolderFromContext(nil)
		assert.Nil(t, retrieved)
	})

	t.Run("context without holder returns nil", func(t *testing.T) {
		ctx := context.Background()
		retrieved := DoctorResultHolderFromContext(ctx)
		assert.Nil(t, retrieved)
	})
}
