package testutil

import (
	"time"
)

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to the given float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// TimePtr returns a pointer to the given time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}
