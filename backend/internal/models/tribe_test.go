package models

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
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
