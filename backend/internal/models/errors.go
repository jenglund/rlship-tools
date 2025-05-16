package models

import "errors"

// Common errors
var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
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
