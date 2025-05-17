package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "zero value",
			input:    0,
			expected: 0,
		},
		{
			name:     "positive value",
			input:    42,
			expected: 42,
		},
		{
			name:     "negative value",
			input:    -42,
			expected: -42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntPtr(tt.input)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestFloat64Ptr(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "zero value",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "positive value",
			input:    42.5,
			expected: 42.5,
		},
		{
			name:     "negative value",
			input:    -42.5,
			expected: -42.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Float64Ptr(tt.input)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestTimePtr(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "zero value",
			input:    time.Time{},
			expected: time.Time{},
		},
		{
			name:     "current time",
			input:    time.Now(),
			expected: time.Now(),
		},
		{
			name:     "specific time",
			input:    time.Date(2024, 3, 14, 15, 9, 26, 0, time.UTC),
			expected: time.Date(2024, 3, 14, 15, 9, 26, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimePtr(tt.input)
			assert.NotNil(t, result)
			if tt.name != "current time" {
				assert.Equal(t, tt.expected, *result)
			} else {
				assert.WithinDuration(t, tt.expected, *result, time.Second)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "non-empty string",
			input:    "test",
			expected: "test",
		},
		{
			name:     "string with spaces",
			input:    "test with spaces",
			expected: "test with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringPtr(tt.input)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}
