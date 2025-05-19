package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TribeType represents the type of tribe
type TribeType string

const (
	TribeTypeCouple    TribeType = "couple"    // Two-person romantic relationship
	TribeTypePolyCule  TribeType = "polycule"  // Polyamorous relationship group
	TribeTypeFriends   TribeType = "friends"   // Friend group
	TribeTypeFamily    TribeType = "family"    // Family group
	TribeTypeRoommates TribeType = "roommates" // Roommate group
	TribeTypeCoworkers TribeType = "coworkers" // Work team or group
	TribeTypeCustom    TribeType = "custom"    // Custom group type
)

func (t TribeType) Validate() error {
	switch t {
	case TribeTypeCouple, TribeTypePolyCule, TribeTypeFriends, TribeTypeFamily,
		TribeTypeRoommates, TribeTypeCoworkers, TribeTypeCustom:
		return nil
	default:
		return fmt.Errorf("%w: invalid tribe type: %s", ErrInvalidInput, t)
	}
}

// MembershipType represents the type of membership in a tribe
type MembershipType string

const (
	MembershipFull    MembershipType = "full"    // Full member with all privileges
	MembershipLimited MembershipType = "limited" // Limited access member
	MembershipGuest   MembershipType = "guest"   // Temporary guest access
	MembershipPending MembershipType = "pending" // Invited but not yet accepted
)

func (m MembershipType) Validate() error {
	switch m {
	case MembershipFull, MembershipLimited, MembershipGuest, MembershipPending:
		return nil
	default:
		return fmt.Errorf("%w: invalid membership type: %s", ErrInvalidInput, m)
	}
}

// Tribe represents a group of users who share activities and plans
type Tribe struct {
	BaseModel
	Name                      string         `json:"name" db:"name"`
	Type                      TribeType      `json:"type" db:"type"`
	Description               string         `json:"description" db:"description"`
	Visibility                VisibilityType `json:"visibility" db:"visibility"`
	Metadata                  JSONMap        `json:"metadata,omitempty" db:"metadata"`
	Members                   []*TribeMember `json:"members,omitempty" db:"-"`
	CurrentUserMembershipType MembershipType `json:"current_user_membership_type,omitempty" db:"-"`
}

// Validate performs validation on the tribe
func (t *Tribe) Validate() error {
	if err := t.BaseModel.Validate(); err != nil {
		return err
	}
	if t.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if len(t.Name) > 100 {
		return fmt.Errorf("%w: name cannot be longer than 100 characters", ErrInvalidInput)
	}
	if err := t.Type.Validate(); err != nil {
		return err
	}
	if err := t.Visibility.Validate(); err != nil {
		return err
	}

	// Ensure metadata is never nil
	if t.Metadata == nil {
		t.Metadata = JSONMap{}
	}

	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Tribe to ensure Metadata is never nil
func (t *Tribe) UnmarshalJSON(data []byte) error {
	// Define an alias type to avoid infinite recursion
	type AliasType Tribe

	// Use a temporary structure with the same fields
	aux := struct {
		*AliasType
		Metadata *JSONMap `json:"metadata,omitempty"`
	}{
		AliasType: (*AliasType)(t),
	}

	// Unmarshal JSON into the temporary structure
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Initialize empty metadata if it's nil
	if aux.Metadata == nil {
		t.Metadata = JSONMap{}
	} else {
		t.Metadata = *aux.Metadata
	}

	return nil
}

// TribeMember represents a user's membership in a tribe
type TribeMember struct {
	BaseModel
	TribeID        uuid.UUID      `json:"tribe_id" db:"tribe_id"`
	UserID         uuid.UUID      `json:"user_id" db:"user_id"`
	MembershipType MembershipType `json:"membership_type" db:"membership_type"`
	DisplayName    string         `json:"display_name" db:"display_name"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty" db:"expires_at"`
	InvitedBy      *uuid.UUID     `json:"invited_by,omitempty" db:"invited_by"`
	InvitedAt      *time.Time     `json:"invited_at,omitempty" db:"invited_at"`
	Metadata       JSONMap        `json:"metadata,omitempty" db:"metadata"`
	User           *User          `json:"user,omitempty" db:"-"`
}

// Validate performs validation on the tribe member
func (tm *TribeMember) Validate() error {
	if err := tm.BaseModel.Validate(); err != nil {
		return err
	}
	if tm.TribeID == uuid.Nil {
		return fmt.Errorf("%w: tribe_id is required", ErrInvalidInput)
	}
	if tm.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if err := tm.MembershipType.Validate(); err != nil {
		return err
	}
	if tm.DisplayName == "" {
		return fmt.Errorf("%w: display_name is required", ErrInvalidInput)
	}
	if tm.MembershipType == MembershipGuest {
		if tm.ExpiresAt == nil {
			return fmt.Errorf("%w: guest membership requires expiration date", ErrInvalidInput)
		}
		if tm.ExpiresAt.Before(time.Now()) {
			return fmt.Errorf("%w: expiration date cannot be in the past", ErrInvalidInput)
		}
	}
	return nil
}

// TribeRepository defines the interface for tribe data operations
type TribeRepository interface {
	// Basic CRUD operations
	Create(tribe *Tribe) error
	GetByID(id uuid.UUID) (*Tribe, error)
	Update(tribe *Tribe) error
	Delete(id uuid.UUID) error
	List(offset, limit int) ([]*Tribe, error)

	// Member management
	AddMember(tribeID, userID uuid.UUID, memberType MembershipType, expiresAt *time.Time, invitedBy *uuid.UUID) error
	UpdateMember(tribeID, userID uuid.UUID, memberType MembershipType, expiresAt *time.Time) error
	RemoveMember(tribeID, userID uuid.UUID) error
	GetMembers(tribeID uuid.UUID) ([]*TribeMember, error)
	GetUserTribes(userID uuid.UUID) ([]*Tribe, error)

	// Queries
	GetByType(tribeType TribeType, offset, limit int) ([]*Tribe, error)
	Search(query string, offset, limit int) ([]*Tribe, error)
	GetExpiredGuestMemberships() ([]*TribeMember, error)
}
