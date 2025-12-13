package main

import (
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

func TestDoctorResultState(t *testing.T) {
	// Reset state before test
	ResetDoctorResult()

	t.Run("not executed returns false", func(t *testing.T) {
		ResetDoctorResult()
		_, ok := GetDoctorResult()
		assert.False(t, ok, "should return false when doctor was not executed")
	})

	t.Run("executed returns true with status", func(t *testing.T) {
		ResetDoctorResult()
		setDoctorResult(dot.HealthWarnings)

		status, ok := GetDoctorResult()
		assert.True(t, ok, "should return true when doctor was executed")
		assert.Equal(t, dot.HealthWarnings, status)
	})

	t.Run("reset clears state", func(t *testing.T) {
		setDoctorResult(dot.HealthErrors)
		ResetDoctorResult()

		_, ok := GetDoctorResult()
		assert.False(t, ok, "should return false after reset")
	})
}
