package models

import (
	"time"

	"github.com/google/uuid"
)

// MembershipType represents the type of tribe membership
type MembershipType string

const (
	MembershipFull  MembershipType = "full"
	MembershipGuest MembershipType = "guest"
)

// Tribe represents a group of users who share activities and interests
type Tribe struct {
	ID        uuid.UUID      `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
	Members   []*TribeMember `json:"members,omitempty" db:"-"`
}

// TribeMember represents a user's membership in a tribe
type TribeMember struct {
	TribeID   uuid.UUID      `json:"tribe_id" db:"tribe_id"`
	UserID    uuid.UUID      `json:"user_id" db:"user_id"`
	Type      MembershipType `json:"type" db:"type"`
	JoinedAt  time.Time      `json:"joined_at" db:"joined_at"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty" db:"expires_at"`
	InvitedBy *uuid.UUID     `json:"invited_by,omitempty" db:"invited_by"`
	User      *User          `json:"user,omitempty" db:"-"`
}

// TribeRepository defines the interface for tribe data operations
type TribeRepository interface {
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
	GetExpiredGuestMemberships() ([]*TribeMember, error)
}
