package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestListType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		lt      ListType
		wantErr bool
	}{
		{
			name:    "valid general type",
			lt:      ListTypeGeneral,
			wantErr: false,
		},
		{
			name:    "valid location type",
			lt:      ListTypeLocation,
			wantErr: false,
		},
		{
			name:    "valid activity type",
			lt:      ListTypeActivity,
			wantErr: false,
		},
		{
			name:    "valid interest type",
			lt:      ListTypeInterest,
			wantErr: false,
		},
		{
			name:    "valid google map type",
			lt:      ListTypeGoogleMap,
			wantErr: false,
		},
		{
			name:    "invalid type",
			lt:      ListType("invalid"),
			wantErr: true,
		},
		{
			name:    "empty type",
			lt:      ListType(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lt.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListSyncStatus_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ls      ListSyncStatus
		wantErr bool
	}{
		{
			name:    "valid none status",
			ls:      ListSyncStatusNone,
			wantErr: false,
		},
		{
			name:    "valid synced status",
			ls:      ListSyncStatusSynced,
			wantErr: false,
		},
		{
			name:    "valid pending status",
			ls:      ListSyncStatusPending,
			wantErr: false,
		},
		{
			name:    "valid conflict status",
			ls:      ListSyncStatusConflict,
			wantErr: false,
		},
		{
			name:    "invalid status",
			ls:      ListSyncStatus("invalid"),
			wantErr: true,
		},
		{
			name:    "empty status",
			ls:      ListSyncStatus(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ls.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestList_Validate(t *testing.T) {
	validID := uuid.New()
	validMaxItems := 10
	validCooldown := 7
	now := time.Now()

	tests := []struct {
		name    string
		list    *List
		wantErr bool
	}{
		{
			name: "valid list",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
			},
			wantErr: false,
		},
		{
			name: "nil ID",
			list: &List{
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			list: &List{
				ID:            validID,
				Type:          ListType("invalid"),
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid visibility",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityType("invalid"),
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
			},
			wantErr: true,
		},
		{
			name: "invalid sync status",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatus("invalid"),
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
			},
			wantErr: true,
		},
		{
			name: "zero default weight",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 0,
			},
			wantErr: true,
		},
		{
			name: "negative default weight",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: -1.0,
			},
			wantErr: true,
		},
		{
			name: "zero max items",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				MaxItems:      &validMaxItems,
			},
			wantErr: false,
		},
		{
			name: "negative max items",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				MaxItems:      func() *int { i := -1; return &i }(),
			},
			wantErr: true,
		},
		{
			name: "negative cooldown days",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				CooldownDays:  func() *int { i := -1; return &i }(),
			},
			wantErr: true,
		},
		{
			name: "valid cooldown days",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				CooldownDays:  &validCooldown,
			},
			wantErr: false,
		},
		// SyncConfig validation tests
		{
			name: "valid sync config",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusPending,
				SyncSource:    SyncSourceGoogleMaps,
				SyncID:        "ChIJN1t_tDeuEmsRUsoyG83frY4",
				LastSyncAt:    &now,
				DefaultWeight: 1.0,
				SyncConfig: &SyncConfig{
					Source:     SyncSourceGoogleMaps,
					ID:         "ChIJN1t_tDeuEmsRUsoyG83frY4",
					Status:     ListSyncStatusPending,
					LastSyncAt: &now,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid sync config - status mismatch",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone, // Mismatch
				SyncSource:    SyncSourceGoogleMaps,
				SyncID:        "ChIJN1t_tDeuEmsRUsoyG83frY4",
				LastSyncAt:    &now,
				DefaultWeight: 1.0,
				SyncConfig: &SyncConfig{
					Source:     SyncSourceGoogleMaps,
					ID:         "ChIJN1t_tDeuEmsRUsoyG83frY4",
					Status:     ListSyncStatusPending, // Doesn't match list
					LastSyncAt: &now,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sync config - source mismatch",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusPending,
				SyncSource:    SyncSourceNone, // Mismatch
				SyncID:        "ChIJN1t_tDeuEmsRUsoyG83frY4",
				LastSyncAt:    &now,
				DefaultWeight: 1.0,
				SyncConfig: &SyncConfig{
					Source:     SyncSourceGoogleMaps, // Doesn't match list
					ID:         "ChIJN1t_tDeuEmsRUsoyG83frY4",
					Status:     ListSyncStatusPending,
					LastSyncAt: &now,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sync config - ID mismatch",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusPending,
				SyncSource:    SyncSourceGoogleMaps,
				SyncID:        "different-id", // Mismatch
				LastSyncAt:    &now,
				DefaultWeight: 1.0,
				SyncConfig: &SyncConfig{
					Source:     SyncSourceGoogleMaps,
					ID:         "ChIJN1t_tDeuEmsRUsoyG83frY4", // Doesn't match list
					Status:     ListSyncStatusPending,
					LastSyncAt: &now,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sync config - LastSyncAt mismatch",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusPending,
				SyncSource:    SyncSourceGoogleMaps,
				SyncID:        "ChIJN1t_tDeuEmsRUsoyG83frY4",
				LastSyncAt:    nil, // Mismatch
				DefaultWeight: 1.0,
				SyncConfig: &SyncConfig{
					Source:     SyncSourceGoogleMaps,
					ID:         "ChIJN1t_tDeuEmsRUsoyG83frY4",
					Status:     ListSyncStatusPending,
					LastSyncAt: &now, // Doesn't match list
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sync config itself",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusPending,
				SyncSource:    SyncSourceGoogleMaps,
				SyncID:        "", // Invalid - Google Maps requires ID
				LastSyncAt:    &now,
				DefaultWeight: 1.0,
				SyncConfig: &SyncConfig{
					Source:     SyncSourceGoogleMaps,
					ID:         "", // Invalid - Google Maps requires ID
					Status:     ListSyncStatusPending,
					LastSyncAt: &now,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid sync fields without config - non-None status",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusPending, // Should be None when no config
				SyncSource:    SyncSourceNone,
				SyncID:        "",
				LastSyncAt:    nil,
				DefaultWeight: 1.0,
				SyncConfig:    nil, // No config
			},
			wantErr: true,
		},
		{
			name: "invalid sync fields without config - non-None source",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceGoogleMaps, // Should be None when no config
				SyncID:        "",
				LastSyncAt:    nil,
				DefaultWeight: 1.0,
				SyncConfig:    nil, // No config
			},
			wantErr: true,
		},
		{
			name: "invalid sync fields without config - non-empty ID",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				SyncID:        "some-id", // Should be empty when no config
				LastSyncAt:    nil,
				DefaultWeight: 1.0,
				SyncConfig:    nil, // No config
			},
			wantErr: true,
		},
		{
			name: "invalid sync fields without config - non-nil LastSyncAt",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				SyncID:        "",
				LastSyncAt:    &now, // Should be nil when no config
				DefaultWeight: 1.0,
				SyncConfig:    nil, // No config
			},
			wantErr: true,
		},
		{
			name: "valid owner fields",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				OwnerID:       func() *uuid.UUID { id := uuid.New(); return &id }(),
				OwnerType:     func() *OwnerType { t := OwnerTypeUser; return &t }(),
			},
			wantErr: false,
		},
		{
			name: "missing owner type with owner ID",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				OwnerID:       func() *uuid.UUID { id := uuid.New(); return &id }(),
				OwnerType:     nil, // Missing but required when OwnerID is set
			},
			wantErr: true,
		},
		{
			name: "missing owner ID with owner type",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				OwnerID:       nil, // Missing but required when OwnerType is set
				OwnerType:     func() *OwnerType { t := OwnerTypeUser; return &t }(),
			},
			wantErr: true,
		},
		{
			name: "invalid owner type",
			list: &List{
				ID:            validID,
				Type:          ListTypeGeneral,
				Name:          "Test List",
				Visibility:    VisibilityPrivate,
				SyncStatus:    ListSyncStatusNone,
				SyncSource:    SyncSourceNone,
				DefaultWeight: 1.0,
				OwnerID:       func() *uuid.UUID { id := uuid.New(); return &id }(),
				OwnerType:     func() *OwnerType { t := OwnerType("invalid"); return &t }(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.list.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListItem_Validate(t *testing.T) {
	validID := uuid.New()
	validListID := uuid.New()
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)
	validCooldown := 7
	validLat := 45.0
	validLong := -122.0
	validAddr := "123 Test St"

	tests := []struct {
		name    string
		item    *ListItem
		wantErr bool
	}{
		{
			name: "valid item",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: false,
		},
		{
			name: "nil ID",
			item: &ListItem{
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "nil list ID",
			item: &ListItem{
				ID:          validID,
				Name:        "Test Item",
				Weight:      1.0,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Weight:      1.0,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "zero weight",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      0,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "negative weight",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      -1.0,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "negative cooldown",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Cooldown:    func() *int { i := -1; return &i }(),
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "valid cooldown",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Cooldown:    &validCooldown,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: false,
		},
		{
			name: "seasonal without dates",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Seasonal:    true,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "seasonal with invalid dates",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Seasonal:    true,
				StartDate:   &future,
				EndDate:     &past,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "seasonal with valid dates",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Seasonal:    true,
				StartDate:   &past,
				EndDate:     &future,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: false,
		},
		{
			name: "partial location data",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Latitude:    &validLat,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: true,
		},
		{
			name: "complete location data",
			item: &ListItem{
				ID:          validID,
				ListID:      validListID,
				Name:        "Test Item",
				Weight:      1.0,
				Latitude:    &validLat,
				Longitude:   &validLong,
				Address:     &validAddr,
				Available:   true,
				ChosenCount: 0,
				UseCount:    0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetadata_Validate(t *testing.T) {
	tests := []struct {
		name    string
		md      Metadata
		wantErr bool
	}{
		{
			name:    "nil metadata",
			md:      nil,
			wantErr: false,
		},
		{
			name: "valid metadata",
			md: Metadata{
				"key":    "value",
				"number": 42,
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid metadata with channel",
			md: Metadata{
				"channel": make(chan int),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.md.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListOwner_Validate(t *testing.T) {
	validListID := uuid.New()
	validOwnerID := uuid.New()

	tests := []struct {
		name    string
		owner   *ListOwner
		wantErr bool
	}{
		{
			name: "valid user owner",
			owner: &ListOwner{
				ListID:    validListID,
				OwnerID:   validOwnerID,
				OwnerType: OwnerTypeUser,
			},
			wantErr: false,
		},
		{
			name: "valid tribe owner",
			owner: &ListOwner{
				ListID:    validListID,
				OwnerID:   validOwnerID,
				OwnerType: OwnerTypeTribe,
			},
			wantErr: false,
		},
		{
			name: "nil list ID",
			owner: &ListOwner{
				OwnerID:   validOwnerID,
				OwnerType: OwnerTypeUser,
			},
			wantErr: true,
		},
		{
			name: "nil owner ID",
			owner: &ListOwner{
				ListID:    validListID,
				OwnerType: OwnerTypeUser,
			},
			wantErr: true,
		},
		{
			name: "invalid owner type",
			owner: &ListOwner{
				ListID:    validListID,
				OwnerID:   validOwnerID,
				OwnerType: OwnerType("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.owner.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListShare_Validate(t *testing.T) {
	validListID := uuid.New()
	validTribeID := uuid.New()
	validUserID := uuid.New()
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name    string
		share   *ListShare
		wantErr bool
	}{
		{
			name: "valid share without expiry",
			share: &ListShare{
				ListID:  validListID,
				TribeID: validTribeID,
				UserID:  validUserID,
			},
			wantErr: false,
		},
		{
			name: "valid share with future expiry",
			share: &ListShare{
				ListID:    validListID,
				TribeID:   validTribeID,
				UserID:    validUserID,
				ExpiresAt: &future,
			},
			wantErr: false,
		},
		{
			name: "nil list ID",
			share: &ListShare{
				TribeID: validTribeID,
				UserID:  validUserID,
			},
			wantErr: true,
		},
		{
			name: "nil tribe ID",
			share: &ListShare{
				ListID: validListID,
				UserID: validUserID,
			},
			wantErr: true,
		},
		{
			name: "nil user ID",
			share: &ListShare{
				ListID:  validListID,
				TribeID: validTribeID,
			},
			wantErr: true,
		},
		{
			name: "past expiry",
			share: &ListShare{
				ListID:    validListID,
				TribeID:   validTribeID,
				UserID:    validUserID,
				ExpiresAt: &past,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.share.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
