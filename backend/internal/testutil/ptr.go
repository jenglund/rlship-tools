package testutil

import "time"

// IntPtr returns a pointer to the given int value
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to the given float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}

// TimePtr returns a pointer to the given time.Time value
func TimePtr(t time.Time) *time.Time {
	return &t
}
