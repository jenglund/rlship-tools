package models

import (
	"time"

	"github.com/google/uuid"
)

// ActivityType represents the type of activity (location, interest, list, etc.)
type ActivityType string

const (
	ActivityTypeLocation ActivityType = "location"
	ActivityTypeInterest ActivityType = "interest"
	ActivityTypeList     ActivityType = "list"
	ActivityTypeActivity ActivityType = "activity"
	ActivityTypeEvent    ActivityType = "event"
)

// Activity represents any type of activity, interest, location, or list
type Activity struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Type        ActivityType   `json:"type" db:"type"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Visibility  VisibilityType `json:"visibility" db:"visibility"`
	Metadata    JSONMap        `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ActivityOwner represents an owner (user or tribe) of an activity
type ActivityOwner struct {
	ActivityID uuid.UUID  `json:"activity_id" db:"activity_id"`
	OwnerID    uuid.UUID  `json:"owner_id" db:"owner_id"`
	OwnerType  OwnerType  `json:"owner_type" db:"owner_type"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ActivityShare represents a shared activity with a tribe
type ActivityShare struct {
	ActivityID uuid.UUID  `json:"activity_id" db:"activity_id"`
	TribeID    uuid.UUID  `json:"tribe_id" db:"tribe_id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ActivityRepository defines the interface for activity data operations
type ActivityRepository interface {
	Create(activity *Activity) error
	GetByID(id uuid.UUID) (*Activity, error)
	Update(activity *Activity) error
	Delete(id uuid.UUID) error
	List(offset, limit int) ([]*Activity, error)

	// Owner management
	AddOwner(activityID, ownerID uuid.UUID, ownerType OwnerType) error
	RemoveOwner(activityID, ownerID uuid.UUID) error
	GetOwners(activityID uuid.UUID) ([]*ActivityOwner, error)
	GetUserActivities(userID uuid.UUID) ([]*Activity, error)
	GetTribeActivities(tribeID uuid.UUID) ([]*Activity, error)

	// Sharing management
	ShareWithTribe(activityID, tribeID, userID uuid.UUID, expiresAt *time.Time) error
	UnshareWithTribe(activityID, tribeID uuid.UUID) error
	GetSharedActivities(tribeID uuid.UUID) ([]*Activity, error)

	// Cleanup
	MarkForDeletion(activityID uuid.UUID) error
	CleanupOrphanedActivities() error
}
