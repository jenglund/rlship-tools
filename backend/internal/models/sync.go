package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SyncSource represents a supported external sync source
type SyncSource string

const (
	SyncSourceNone       SyncSource = "none"
	SyncSourceGoogleMaps SyncSource = "google_maps"
	SyncSourceManual     SyncSource = "manual"
	SyncSourceImported   SyncSource = "imported"
)

// Validate checks if the sync source is valid
func (s SyncSource) Validate() error {
	switch s {
	case SyncSourceNone, SyncSourceGoogleMaps, SyncSourceManual, SyncSourceImported:
		return nil
	default:
		return fmt.Errorf("%w: invalid sync source: %s", ErrInvalidInput, s)
	}
}

// SyncTransition represents a valid state transition in the sync state machine
type SyncTransition struct {
	From   ListSyncStatus
	To     ListSyncStatus
	Action string
}

// ValidSyncTransitions defines all valid state transitions
var ValidSyncTransitions = []SyncTransition{
	{From: ListSyncStatusNone, To: ListSyncStatusPending, Action: "configure_sync"},
	{From: ListSyncStatusPending, To: ListSyncStatusSynced, Action: "sync_complete"},
	{From: ListSyncStatusPending, To: ListSyncStatusConflict, Action: "conflict_detected"},
	{From: ListSyncStatusSynced, To: ListSyncStatusPending, Action: "local_change"},
	{From: ListSyncStatusSynced, To: ListSyncStatusConflict, Action: "remote_change_conflict"},
	{From: ListSyncStatusConflict, To: ListSyncStatusPending, Action: "resolve_conflict"},
	{From: ListSyncStatusConflict, To: ListSyncStatusSynced, Action: "auto_resolve"},
	{From: ListSyncStatusPending, To: ListSyncStatusNone, Action: "disable_sync"},
	{From: ListSyncStatusSynced, To: ListSyncStatusNone, Action: "disable_sync"},
	{From: ListSyncStatusConflict, To: ListSyncStatusNone, Action: "disable_sync"},
}

// SyncConfig represents the sync configuration for a list
type SyncConfig struct {
	Source     SyncSource     `json:"source" db:"sync_source"`
	ID         string         `json:"id" db:"sync_id"`
	Status     ListSyncStatus `json:"status" db:"sync_status"`
	LastSyncAt *time.Time     `json:"last_sync_at" db:"last_sync_at"`
}

// Validate performs validation on the sync configuration
func (sc *SyncConfig) Validate() error {
	if err := sc.Source.Validate(); err != nil {
		return err
	}

	if err := sc.Status.Validate(); err != nil {
		return err
	}

	// Validate sync source specific rules
	switch sc.Source {
	case SyncSourceGoogleMaps:
		if sc.ID == "" {
			return fmt.Errorf("%w: Google Maps sync requires a valid place ID", ErrInvalidInput)
		}
	case SyncSourceNone:
		if sc.Status != ListSyncStatusNone {
			return fmt.Errorf("%w: sync status must be none when no source is configured", ErrInvalidInput)
		}
		if sc.ID != "" {
			return fmt.Errorf("%w: sync ID must be empty when no source is configured", ErrInvalidInput)
		}
		if sc.LastSyncAt != nil {
			return fmt.Errorf("%w: last sync time must be empty when no source is configured", ErrInvalidInput)
		}
	}

	return nil
}

// ValidateTransition checks if a sync status transition is valid
func ValidateTransition(from, to ListSyncStatus, action string) error {
	for _, t := range ValidSyncTransitions {
		if t.From == from && t.To == to && t.Action == action {
			return nil
		}
	}
	return fmt.Errorf("invalid transition from '%s' to '%s' with action '%s'",
		from, to, action)
}

// SyncConflict represents a sync conflict that needs resolution
type SyncConflict struct {
	ID         uuid.UUID   `json:"id" db:"id"`
	ListID     uuid.UUID   `json:"list_id" db:"list_id"`
	ItemID     *uuid.UUID  `json:"item_id,omitempty" db:"item_id"`
	Type       string      `json:"type" db:"type"`
	LocalData  interface{} `json:"local_data" db:"local_data"`
	RemoteData interface{} `json:"remote_data" db:"remote_data"`
	Resolution string      `json:"resolution" db:"resolution"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
	ResolvedAt *time.Time  `json:"resolved_at,omitempty" db:"resolved_at"`
}

// Validate performs validation on the sync conflict
func (sc *SyncConflict) Validate() error {
	if sc.ID == uuid.Nil {
		return fmt.Errorf("%w: conflict ID is required", ErrInvalidInput)
	}
	if sc.ListID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", ErrInvalidInput)
	}
	if sc.Type == "" {
		return fmt.Errorf("%w: conflict type is required", ErrInvalidInput)
	}
	if sc.LocalData == nil {
		return fmt.Errorf("%w: local data is required", ErrInvalidInput)
	}
	if sc.RemoteData == nil {
		return fmt.Errorf("%w: remote data is required", ErrInvalidInput)
	}
	return nil
}
