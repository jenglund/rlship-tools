package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTribeType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tt      TribeType
		wantErr bool
	}{
		{"couple", TribeTypeCouple, false},
		{"polycule", TribeTypePolyCule, false},
		{"friends", TribeTypeFriends, false},
		{"family", TribeTypeFamily, false},
		{"roommates", TribeTypeRoommates, false},
		{"coworkers", TribeTypeCoworkers, false},
		{"custom", TribeTypeCustom, false},
		{"invalid", TribeType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tt.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMembershipType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mt      MembershipType
		wantErr bool
	}{
		{"full", MembershipFull, false},
		{"limited", MembershipLimited, false},
		{"guest", MembershipGuest, false},
		{"invalid", MembershipType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mt.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTribe_Validate(t *testing.T) {
	now := time.Now()
	validID := uuid.New()
	validBase := BaseModel{
		ID:        validID,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	tests := []struct {
		name    string
		t       Tribe
		wantErr bool
	}{
		{
			name: "valid",
			t: Tribe{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				Name:        "Test Tribe",
				Type:        TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  VisibilityPrivate,
			},
			wantErr: false,
		},
		{
			name: "invalid base model",
			t: Tribe{
				BaseModel:   BaseModel{},
				Name:        "Test Tribe",
				Type:        TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  VisibilityPrivate,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			t: Tribe{
				BaseModel:   validBase,
				Name:        "",
				Type:        TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  VisibilityPrivate,
			},
			wantErr: true,
		},
		{
			name: "name too long (over 100 characters)",
			t: Tribe{
				BaseModel:   validBase,
				Name:        string(make([]byte, 101)), // 101 character string
				Type:        TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  VisibilityPrivate,
			},
			wantErr: true,
		},
		{
			name: "name exactly 100 characters",
			t: Tribe{
				BaseModel:   validBase,
				Name:        string(make([]byte, 100)), // 100 character string
				Type:        TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  VisibilityPrivate,
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			t: Tribe{
				BaseModel:   validBase,
				Name:        "Test Tribe",
				Type:        TribeType("invalid"),
				Description: "A test tribe",
				Visibility:  VisibilityPrivate,
			},
			wantErr: true,
		},
		{
			name: "invalid visibility",
			t: Tribe{
				BaseModel:   validBase,
				Name:        "Test Tribe",
				Type:        TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  VisibilityType("invalid"),
			},
			wantErr: true,
		},
		{
			name: "nil metadata (should be initialized)",
			t: Tribe{
				BaseModel:   validBase,
				Name:        "Test Tribe",
				Type:        TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  VisibilityPrivate,
				Metadata:    nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
				// Check if metadata is initialized when previously nil
				if tt.name == "nil metadata (should be initialized)" {
					assert.NotNil(t, tt.t.Metadata)
					assert.Empty(t, tt.t.Metadata)
				}
			}
		})
	}
}

func TestTribeMember_Validate(t *testing.T) {
	now := time.Now()
	validID := uuid.New()
	validBase := BaseModel{
		ID:        validID,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		tm      TribeMember
		wantErr bool
	}{
		{
			name: "valid full member",
			tm: TribeMember{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipFull,
				DisplayName:    "Test Member",
			},
			wantErr: false,
		},
		{
			name: "valid limited member",
			tm: TribeMember{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipLimited,
				DisplayName:    "Test Limited Member",
			},
			wantErr: false,
		},
		{
			name: "valid guest member with future expiry",
			tm: TribeMember{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipGuest,
				DisplayName:    "Test Guest",
				ExpiresAt:      &future,
			},
			wantErr: false,
		},
		{
			name: "valid pending member",
			tm: TribeMember{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipPending,
				DisplayName:    "Test Pending Member",
				InvitedBy:      &validID,
				InvitedAt:      &now,
			},
			wantErr: false,
		},
		{
			name: "pending member with expiry (not required but allowed)",
			tm: TribeMember{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipPending,
				DisplayName:    "Test Pending Member with Expiry",
				ExpiresAt:      &future, // Not required for pending members, but allowed
				InvitedBy:      &validID,
				InvitedAt:      &now,
			},
			wantErr: false,
		},
		{
			name: "full member with expiry (not required but allowed)",
			tm: TribeMember{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipFull,
				DisplayName:    "Test Member",
				ExpiresAt:      &future, // Not required for full members, but allowed
			},
			wantErr: false,
		},
		{
			name: "limited member with expiry (not required but allowed)",
			tm: TribeMember{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipLimited,
				DisplayName:    "Test Limited Member",
				ExpiresAt:      &future, // Not required for limited members, but allowed
			},
			wantErr: false,
		},
		{
			name: "invalid base model",
			tm: TribeMember{
				BaseModel:      BaseModel{},
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipFull,
				DisplayName:    "Test Member",
			},
			wantErr: true,
		},
		{
			name: "nil tribe id",
			tm: TribeMember{
				BaseModel:      validBase,
				TribeID:        uuid.Nil,
				UserID:         uuid.New(),
				MembershipType: MembershipFull,
				DisplayName:    "Test Member",
			},
			wantErr: true,
		},
		{
			name: "nil user id",
			tm: TribeMember{
				BaseModel:      validBase,
				TribeID:        uuid.New(),
				UserID:         uuid.Nil,
				MembershipType: MembershipFull,
				DisplayName:    "Test Member",
			},
			wantErr: true,
		},
		{
			name: "invalid membership type",
			tm: TribeMember{
				BaseModel:      validBase,
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipType("invalid"),
				DisplayName:    "Test Member",
			},
			wantErr: true,
		},
		{
			name: "empty display name",
			tm: TribeMember{
				BaseModel:      validBase,
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipFull,
				DisplayName:    "",
			},
			wantErr: true,
		},
		{
			name: "guest without expiry",
			tm: TribeMember{
				BaseModel:      validBase,
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipGuest,
				DisplayName:    "Test Guest",
			},
			wantErr: true,
		},
		{
			name: "guest with past expiry",
			tm: TribeMember{
				BaseModel:      validBase,
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipGuest,
				DisplayName:    "Test Guest",
				ExpiresAt:      &past,
			},
			wantErr: true,
		},
		{
			name: "guest with expiry exactly at current time",
			tm: TribeMember{
				BaseModel:      validBase,
				TribeID:        uuid.New(),
				UserID:         uuid.New(),
				MembershipType: MembershipGuest,
				DisplayName:    "Test Guest",
				ExpiresAt:      &now, // Exactly at current time - should be in the future
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tm.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTribe_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		expectErr bool
		validate  func(t *testing.T, tribe *Tribe)
	}{
		{
			name:      "basic tribe",
			json:      `{"id":"123e4567-e89b-12d3-a456-426614174000","name":"Test Tribe","type":"couple","visibility":"private","created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T12:00:00Z","version":1}`,
			expectErr: false,
			validate: func(t *testing.T, tribe *Tribe) {
				assert.Equal(t, "Test Tribe", tribe.Name)
				assert.Equal(t, TribeTypeCouple, tribe.Type)
				assert.Equal(t, VisibilityPrivate, tribe.Visibility)
				assert.NotNil(t, tribe.Metadata)
				assert.Empty(t, tribe.Metadata)
			},
		},
		{
			name:      "tribe with metadata",
			json:      `{"id":"123e4567-e89b-12d3-a456-426614174000","name":"Family","type":"family","visibility":"private","metadata":{"color":"blue","icon":"home"},"created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T12:00:00Z","version":1}`,
			expectErr: false,
			validate: func(t *testing.T, tribe *Tribe) {
				assert.Equal(t, "Family", tribe.Name)
				assert.Equal(t, TribeTypeFamily, tribe.Type)
				assert.Equal(t, VisibilityPrivate, tribe.Visibility)
				assert.NotNil(t, tribe.Metadata)
				assert.Equal(t, "blue", tribe.Metadata["color"])
				assert.Equal(t, "home", tribe.Metadata["icon"])
			},
		},
		{
			name:      "tribe with null metadata",
			json:      `{"id":"123e4567-e89b-12d3-a456-426614174000","name":"Friends","type":"friends","visibility":"private","metadata":null,"created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T12:00:00Z","version":1}`,
			expectErr: false,
			validate: func(t *testing.T, tribe *Tribe) {
				assert.Equal(t, "Friends", tribe.Name)
				assert.Equal(t, TribeTypeFriends, tribe.Type)
				assert.NotNil(t, tribe.Metadata)
				assert.Empty(t, tribe.Metadata)
			},
		},
		{
			name:      "tribe without metadata field",
			json:      `{"id":"123e4567-e89b-12d3-a456-426614174000","name":"Roommates","type":"roommates","visibility":"private","created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T12:00:00Z","version":1}`,
			expectErr: false,
			validate: func(t *testing.T, tribe *Tribe) {
				assert.Equal(t, "Roommates", tribe.Name)
				assert.Equal(t, TribeTypeRoommates, tribe.Type)
				assert.NotNil(t, tribe.Metadata)
				assert.Empty(t, tribe.Metadata)
			},
		},
		{
			name:      "tribe with empty object metadata",
			json:      `{"id":"123e4567-e89b-12d3-a456-426614174000","name":"Coworkers","type":"coworkers","visibility":"private","metadata":{},"created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T12:00:00Z","version":1}`,
			expectErr: false,
			validate: func(t *testing.T, tribe *Tribe) {
				assert.Equal(t, "Coworkers", tribe.Name)
				assert.Equal(t, TribeTypeCoworkers, tribe.Type)
				assert.NotNil(t, tribe.Metadata)
				assert.Empty(t, tribe.Metadata)
			},
		},
		{
			name:      "invalid json",
			json:      `{"name":"Invalid`,
			expectErr: true,
			validate:  func(t *testing.T, tribe *Tribe) {},
		},
		{
			name:      "invalid metadata type",
			json:      `{"id":"123e4567-e89b-12d3-a456-426614174000","name":"Invalid","type":"couple","visibility":"private","metadata":"not an object","created_at":"2023-01-01T12:00:00Z","updated_at":"2023-01-01T12:00:00Z","version":1}`,
			expectErr: true,
			validate:  func(t *testing.T, tribe *Tribe) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tribe Tribe
			err := json.Unmarshal([]byte(tt.json), &tribe)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, &tribe)
			}
		})
	}
}
