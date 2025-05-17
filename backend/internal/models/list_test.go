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
