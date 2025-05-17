package models

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSyncError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "sync config error",
			err:      ErrInvalidSyncConfig,
			expected: true,
		},
		{
			name:     "sync source error",
			err:      ErrInvalidSyncSource,
			expected: true,
		},
		{
			name:     "missing sync ID error",
			err:      ErrMissingSyncID,
			expected: true,
		},
		{
			name:     "invalid sync status error",
			err:      ErrInvalidSyncStatus,
			expected: true,
		},
		{
			name:     "invalid sync transition error",
			err:      ErrInvalidSyncTransition,
			expected: true,
		},
		{
			name:     "sync disabled error",
			err:      ErrSyncDisabled,
			expected: true,
		},
		{
			name:     "sync already enabled error",
			err:      ErrSyncAlreadyEnabled,
			expected: true,
		},
		{
			name:     "conflict not found error",
			err:      ErrConflictNotFound,
			expected: true,
		},
		{
			name:     "conflict already resolved error",
			err:      ErrConflictAlreadyResolved,
			expected: true,
		},
		{
			name:     "invalid resolution error",
			err:      ErrInvalidResolution,
			expected: true,
		},
		{
			name:     "external source unavailable error",
			err:      ErrExternalSourceUnavailable,
			expected: true,
		},
		{
			name:     "external source error",
			err:      ErrExternalSourceError,
			expected: true,
		},
		{
			name:     "external source timeout error",
			err:      ErrExternalSourceTimeout,
			expected: true,
		},
		{
			name:     "wrapped sync error",
			err:      fmt.Errorf("something went wrong: %w", ErrInvalidSyncConfig),
			expected: true,
		},
		{
			name:     "non-sync error",
			err:      ErrNotFound,
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSyncError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
