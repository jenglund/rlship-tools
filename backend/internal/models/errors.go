package models

import "errors"

// Common application-level errors
var (
	// ErrNotFound is returned when a resource could not be found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when a user is not authenticated
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when a user does not have permission to access a resource
	ErrForbidden = errors.New("forbidden")

	// ErrConflict is returned when there is a conflict with the current state of a resource
	ErrConflict = errors.New("conflict with current state")

	// ErrInternal is returned when an unexpected internal error occurs
	ErrInternal = errors.New("internal server error")

	// ErrDuplicate is returned when a resource with the same identifier already exists
	ErrDuplicate = errors.New("resource already exists")
)

// Sync-specific errors
var (
	// Configuration errors
	ErrInvalidSyncConfig = errors.New("invalid sync configuration")
	ErrInvalidSyncSource = errors.New("invalid sync source")
	ErrMissingSyncID     = errors.New("missing sync ID")
	ErrInvalidSyncStatus = errors.New("invalid sync status")

	// State transition errors
	ErrInvalidSyncTransition = errors.New("invalid sync state transition")
	ErrSyncDisabled          = errors.New("sync is disabled")
	ErrSyncAlreadyEnabled    = errors.New("sync is already enabled")

	// Conflict errors
	ErrConflictNotFound        = errors.New("sync conflict not found")
	ErrConflictAlreadyResolved = errors.New("sync conflict already resolved")
	ErrInvalidResolution       = errors.New("invalid conflict resolution")

	// External source errors
	ErrExternalSourceUnavailable = errors.New("external sync source unavailable")
	ErrExternalSourceError       = errors.New("external sync source error")
	ErrExternalSourceTimeout     = errors.New("external sync source timeout")

	// ErrConcurrentModification is returned when a concurrent modification is detected
	ErrConcurrentModification = errors.New("concurrent modification detected")
)

// Activity Photo errors
var (
	ErrInvalidID         = errors.New("invalid ID")
	ErrInvalidActivityID = errors.New("invalid activity ID")
	ErrInvalidURL        = errors.New("invalid URL")
	ErrInvalidCreatedAt  = errors.New("invalid created at time")
	ErrInvalidUpdatedAt  = errors.New("invalid updated at time")
	ErrInvalidDeletedAt  = errors.New("deleted at time cannot be before created at time")
)

// IsSyncError checks if an error is a sync-related error
func IsSyncError(err error) bool {
	return errors.Is(err, ErrInvalidSyncConfig) ||
		errors.Is(err, ErrInvalidSyncSource) ||
		errors.Is(err, ErrMissingSyncID) ||
		errors.Is(err, ErrInvalidSyncStatus) ||
		errors.Is(err, ErrInvalidSyncTransition) ||
		errors.Is(err, ErrSyncDisabled) ||
		errors.Is(err, ErrSyncAlreadyEnabled) ||
		errors.Is(err, ErrConflictNotFound) ||
		errors.Is(err, ErrConflictAlreadyResolved) ||
		errors.Is(err, ErrInvalidResolution) ||
		errors.Is(err, ErrExternalSourceUnavailable) ||
		errors.Is(err, ErrExternalSourceError) ||
		errors.Is(err, ErrExternalSourceTimeout)
}
