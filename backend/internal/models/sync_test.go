package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSyncSourceValidation(t *testing.T) {
	tests := []struct {
		name      string
		source    SyncSource
		expectErr bool
	}{
		{
			name:      "none source",
			source:    SyncSourceNone,
			expectErr: false,
		},
		{
			name:      "google maps source",
			source:    SyncSourceGoogleMaps,
			expectErr: false,
		},
		{
			name:      "manual source",
			source:    SyncSourceManual,
			expectErr: false,
		},
		{
			name:      "imported source",
			source:    SyncSourceImported,
			expectErr: false,
		},
		{
			name:      "invalid source",
			source:    SyncSource("invalid"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.source.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncConfigValidation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		config    *SyncConfig
		expectErr bool
	}{
		{
			name: "valid google maps config",
			config: &SyncConfig{
				Source:     SyncSourceGoogleMaps,
				ID:         "ChIJN1t_tDeuEmsRUsoyG83frY4", // Example Google Maps place ID
				Status:     ListSyncStatusPending,
				LastSyncAt: &now,
			},
			expectErr: false,
		},
		{
			name: "google maps without ID",
			config: &SyncConfig{
				Source:     SyncSourceGoogleMaps,
				ID:         "",
				Status:     ListSyncStatusPending,
				LastSyncAt: &now,
			},
			expectErr: true,
		},
		{
			name: "valid manual config",
			config: &SyncConfig{
				Source:     SyncSourceManual,
				ID:         "manual-1",
				Status:     ListSyncStatusPending,
				LastSyncAt: &now,
			},
			expectErr: false,
		},
		{
			name: "valid none config",
			config: &SyncConfig{
				Source:     SyncSourceNone,
				ID:         "",
				Status:     ListSyncStatusNone,
				LastSyncAt: nil,
			},
			expectErr: false,
		},
		{
			name: "none with invalid status",
			config: &SyncConfig{
				Source:     SyncSourceNone,
				ID:         "",
				Status:     ListSyncStatusPending,
				LastSyncAt: nil,
			},
			expectErr: true,
		},
		{
			name: "none with non-empty ID",
			config: &SyncConfig{
				Source:     SyncSourceNone,
				ID:         "something",
				Status:     ListSyncStatusNone,
				LastSyncAt: nil,
			},
			expectErr: true,
		},
		{
			name: "none with last sync time",
			config: &SyncConfig{
				Source:     SyncSourceNone,
				ID:         "",
				Status:     ListSyncStatusNone,
				LastSyncAt: &now,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name      string
		from      ListSyncStatus
		to        ListSyncStatus
		action    string
		expectErr bool
	}{
		{
			name:      "configure sync",
			from:      ListSyncStatusNone,
			to:        ListSyncStatusPending,
			action:    "configure_sync",
			expectErr: false,
		},
		{
			name:      "sync complete",
			from:      ListSyncStatusPending,
			to:        ListSyncStatusSynced,
			action:    "sync_complete",
			expectErr: false,
		},
		{
			name:      "conflict detected",
			from:      ListSyncStatusPending,
			to:        ListSyncStatusConflict,
			action:    "conflict_detected",
			expectErr: false,
		},
		{
			name:      "local change",
			from:      ListSyncStatusSynced,
			to:        ListSyncStatusPending,
			action:    "local_change",
			expectErr: false,
		},
		{
			name:      "remote conflict",
			from:      ListSyncStatusSynced,
			to:        ListSyncStatusConflict,
			action:    "remote_change_conflict",
			expectErr: false,
		},
		{
			name:      "resolve conflict",
			from:      ListSyncStatusConflict,
			to:        ListSyncStatusPending,
			action:    "resolve_conflict",
			expectErr: false,
		},
		{
			name:      "auto resolve",
			from:      ListSyncStatusConflict,
			to:        ListSyncStatusSynced,
			action:    "auto_resolve",
			expectErr: false,
		},
		{
			name:      "disable sync from pending",
			from:      ListSyncStatusPending,
			to:        ListSyncStatusNone,
			action:    "disable_sync",
			expectErr: false,
		},
		{
			name:      "disable sync from synced",
			from:      ListSyncStatusSynced,
			to:        ListSyncStatusNone,
			action:    "disable_sync",
			expectErr: false,
		},
		{
			name:      "disable sync from conflict",
			from:      ListSyncStatusConflict,
			to:        ListSyncStatusNone,
			action:    "disable_sync",
			expectErr: false,
		},
		{
			name:      "invalid transition",
			from:      ListSyncStatusNone,
			to:        ListSyncStatusSynced,
			action:    "direct_sync",
			expectErr: true,
		},
		{
			name:      "invalid action",
			from:      ListSyncStatusPending,
			to:        ListSyncStatusSynced,
			action:    "invalid_action",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransition(tt.from, tt.to, tt.action)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncConflictValidation(t *testing.T) {
	validID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name      string
		conflict  *SyncConflict
		expectErr bool
	}{
		{
			name: "valid conflict",
			conflict: &SyncConflict{
				ID:         validID,
				ListID:     validID,
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
			},
			expectErr: false,
		},
		{
			name: "valid conflict with item ID",
			conflict: &SyncConflict{
				ID:         validID,
				ListID:     validID,
				ItemID:     &itemID,
				Type:       "item_delete",
				LocalData:  map[string]interface{}{"deleted": true},
				RemoteData: map[string]interface{}{"deleted": false},
			},
			expectErr: false,
		},
		{
			name: "nil ID",
			conflict: &SyncConflict{
				ID:         uuid.Nil,
				ListID:     validID,
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
			},
			expectErr: true,
		},
		{
			name: "nil list ID",
			conflict: &SyncConflict{
				ID:         validID,
				ListID:     uuid.Nil,
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
			},
			expectErr: true,
		},
		{
			name: "empty type",
			conflict: &SyncConflict{
				ID:         validID,
				ListID:     validID,
				Type:       "",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
			},
			expectErr: true,
		},
		{
			name: "nil local data",
			conflict: &SyncConflict{
				ID:         validID,
				ListID:     validID,
				Type:       "item_update",
				LocalData:  nil,
				RemoteData: map[string]interface{}{"name": "Remote Name"},
			},
			expectErr: true,
		},
		{
			name: "nil remote data",
			conflict: &SyncConflict{
				ID:         validID,
				ListID:     validID,
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: nil,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conflict.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				if tt.conflict.ID == uuid.Nil {
					assert.Contains(t, err.Error(), "conflict ID is required")
				} else if tt.conflict.ListID == uuid.Nil {
					assert.Contains(t, err.Error(), "list ID is required")
				} else if tt.conflict.Type == "" {
					assert.Contains(t, err.Error(), "conflict type is required")
				} else if tt.conflict.LocalData == nil {
					assert.Contains(t, err.Error(), "local data is required")
				} else if tt.conflict.RemoteData == nil {
					assert.Contains(t, err.Error(), "remote data is required")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
