package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ListType represents the type of list
type ListType string

const (
	ListTypeGeneral   ListType = "general"
	ListTypeLocation  ListType = "location"
	ListTypeActivity  ListType = "activity"
	ListTypeInterest  ListType = "interest"
	ListTypeGoogleMap ListType = "google_map"
)

func (lt ListType) Validate() error {
	switch lt {
	case ListTypeGeneral, ListTypeLocation, ListTypeActivity, ListTypeInterest, ListTypeGoogleMap:
		return nil
	default:
		return fmt.Errorf("%w: invalid list type: %s", ErrInvalidInput, lt)
	}
}

// ListSyncStatus represents the sync status of a list
type ListSyncStatus string

const (
	ListSyncStatusNone     ListSyncStatus = "none"     // No sync configured
	ListSyncStatusSynced   ListSyncStatus = "synced"   // In sync with external source
	ListSyncStatusPending  ListSyncStatus = "pending"  // Changes pending sync
	ListSyncStatusConflict ListSyncStatus = "conflict" // Conflicts detected
)

func (ls ListSyncStatus) Validate() error {
	switch ls {
	case ListSyncStatusNone, ListSyncStatusSynced, ListSyncStatusPending, ListSyncStatusConflict:
		return nil
	default:
		return fmt.Errorf("%w: invalid sync status: %s", ErrInvalidInput, ls)
	}
}

// List represents a list in the system
type List struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	Type          ListType       `json:"type" db:"type"`
	Name          string         `json:"name" db:"name"`
	Description   string         `json:"description" db:"description"`
	Visibility    VisibilityType `json:"visibility" db:"visibility"`
	SyncStatus    ListSyncStatus `json:"sync_status" db:"sync_status"`
	SyncSource    string         `json:"sync_source" db:"sync_source"`
	SyncID        string         `json:"sync_id" db:"sync_id"`
	LastSyncAt    *time.Time     `json:"last_sync_at" db:"last_sync_at"`
	DefaultWeight float64        `json:"default_weight" db:"default_weight"`
	MaxItems      *int           `json:"max_items" db:"max_items"`
	CooldownDays  *int           `json:"cooldown_days" db:"cooldown_days"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
	Items         []*ListItem    `json:"items,omitempty" db:"-"`
	Owners        []*ListOwner   `json:"owners,omitempty" db:"-"`
	Shares        []*ListShare   `json:"shares,omitempty" db:"-"`

	// Owner fields for creation/update operations
	OwnerID   *uuid.UUID `json:"owner_id,omitempty" db:"-"`
	OwnerType *OwnerType `json:"owner_type,omitempty" db:"-"`
}

// Validate performs validation on the List
func (l *List) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", ErrInvalidInput)
	}
	if err := l.Type.Validate(); err != nil {
		return err
	}
	if err := l.Visibility.Validate(); err != nil {
		return err
	}
	if err := l.SyncStatus.Validate(); err != nil {
		return err
	}
	if l.Name == "" {
		return fmt.Errorf("%w: list name is required", ErrInvalidInput)
	}
	if l.DefaultWeight <= 0 {
		return fmt.Errorf("%w: default weight must be positive", ErrInvalidInput)
	}
	if l.MaxItems != nil && *l.MaxItems <= 0 {
		return fmt.Errorf("%w: max items must be positive", ErrInvalidInput)
	}
	if l.CooldownDays != nil && *l.CooldownDays < 0 {
		return fmt.Errorf("%w: cooldown days cannot be negative", ErrInvalidInput)
	}

	// Validate owner fields if provided
	if l.OwnerID != nil && l.OwnerType == nil {
		return fmt.Errorf("%w: owner type is required when owner ID is provided", ErrInvalidInput)
	}
	if l.OwnerID == nil && l.OwnerType != nil {
		return fmt.Errorf("%w: owner ID is required when owner type is provided", ErrInvalidInput)
	}
	if l.OwnerType != nil {
		if err := l.OwnerType.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ListItem represents an item in a list
type ListItem struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	ListID      uuid.UUID    `json:"list_id" db:"list_id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Metadata    Metadata     `json:"metadata,omitempty" db:"metadata"`
	ExternalID  string       `json:"external_id,omitempty" db:"external_id"`
	Latitude    *float64     `json:"latitude,omitempty" db:"latitude"`
	Longitude   *float64     `json:"longitude,omitempty" db:"longitude"`
	Address     *string      `json:"address,omitempty" db:"address"`
	Location    *LocationRef `json:"location,omitempty" db:"-"`
	Weight      float64      `json:"weight" db:"weight"`
	LastChosen  *time.Time   `json:"last_chosen" db:"last_chosen"`
	ChosenCount int          `json:"chosen_count" db:"chosen_count"`
	LastUsed    *time.Time   `json:"last_used" db:"last_used"`
	UseCount    int          `json:"use_count" db:"use_count"`
	Cooldown    *int         `json:"cooldown" db:"cooldown"`
	Available   bool         `json:"available" db:"available"`
	Seasonal    bool         `json:"seasonal" db:"seasonal"`
	StartDate   *time.Time   `json:"start_date" db:"start_date"`
	EndDate     *time.Time   `json:"end_date" db:"end_date"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Validate performs validation on the ListItem
func (li *ListItem) Validate() error {
	if li.ID == uuid.Nil {
		return fmt.Errorf("%w: item ID is required", ErrInvalidInput)
	}
	if li.ListID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", ErrInvalidInput)
	}
	if li.Name == "" {
		return fmt.Errorf("%w: item name is required", ErrInvalidInput)
	}
	if li.Weight <= 0 {
		return fmt.Errorf("%w: weight must be positive", ErrInvalidInput)
	}
	if li.Cooldown != nil && *li.Cooldown < 0 {
		return fmt.Errorf("%w: cooldown cannot be negative", ErrInvalidInput)
	}
	if li.Seasonal {
		if li.StartDate == nil || li.EndDate == nil {
			return fmt.Errorf("%w: seasonal items must have start and end dates", ErrInvalidInput)
		}
		if li.StartDate.After(*li.EndDate) {
			return fmt.Errorf("%w: start date must be before end date", ErrInvalidInput)
		}
	}
	if li.Latitude != nil && (li.Longitude == nil || li.Address == nil) {
		return fmt.Errorf("%w: location requires latitude, longitude, and address", ErrInvalidInput)
	}
	return nil
}

// LocationRef represents a reference to a location
type LocationRef struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
	PlaceID   string  `json:"place_id,omitempty"`
}

// Validate performs validation on the LocationRef
func (l *LocationRef) Validate() error {
	if l.Latitude < -90 || l.Latitude > 90 {
		return fmt.Errorf("%w: latitude must be between -90 and 90", ErrInvalidInput)
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("%w: longitude must be between -180 and 180", ErrInvalidInput)
	}
	if l.Address == "" {
		return fmt.Errorf("%w: address is required", ErrInvalidInput)
	}
	return nil
}

// Metadata represents additional data for a list item
type Metadata = JSONMap

// Validate performs validation on the Metadata
func (m Metadata) Validate() error {
	if m == nil {
		return nil
	}
	// Ensure metadata can be marshaled to JSON
	_, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("%w: invalid metadata: %v", ErrInvalidInput, err)
	}
	return nil
}

// ListOwner represents an owner of a list
type ListOwner struct {
	ListID    uuid.UUID  `json:"list_id" db:"list_id"`
	OwnerID   uuid.UUID  `json:"owner_id" db:"owner_id"`
	OwnerType OwnerType  `json:"owner_type" db:"owner_type"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Validate performs validation on the ListOwner
func (lo *ListOwner) Validate() error {
	if lo.ListID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", ErrInvalidInput)
	}
	if lo.OwnerID == uuid.Nil {
		return fmt.Errorf("%w: owner ID is required", ErrInvalidInput)
	}
	if err := lo.OwnerType.Validate(); err != nil {
		return err
	}
	return nil
}

// ListShare represents a shared list with a tribe
type ListShare struct {
	ListID    uuid.UUID  `json:"list_id" db:"list_id"`
	TribeID   uuid.UUID  `json:"tribe_id" db:"tribe_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Validate performs validation on the ListShare
func (ls *ListShare) Validate() error {
	if ls.ListID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", ErrInvalidInput)
	}
	if ls.TribeID == uuid.Nil {
		return fmt.Errorf("%w: tribe ID is required", ErrInvalidInput)
	}
	if ls.UserID == uuid.Nil {
		return fmt.Errorf("%w: user ID is required", ErrInvalidInput)
	}
	if ls.ExpiresAt != nil && ls.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("%w: expiration date must be in the future", ErrInvalidInput)
	}
	return nil
}

// ListRepository defines the interface for list storage operations
type ListRepository interface {
	// Basic CRUD operations
	Create(list *List) error
	GetByID(id uuid.UUID) (*List, error)
	Update(list *List) error
	Delete(id uuid.UUID) error
	List(offset, limit int) ([]*List, error)

	// List item operations
	AddItem(item *ListItem) error
	UpdateItem(item *ListItem) error
	RemoveItem(listID, itemID uuid.UUID) error
	GetItems(listID uuid.UUID) ([]*ListItem, error)
	GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*ListItem, error)
	UpdateItemStats(itemID uuid.UUID, chosen bool) error

	// Owner management
	AddOwner(owner *ListOwner) error
	RemoveOwner(listID, ownerID uuid.UUID) error
	GetOwners(listID uuid.UUID) ([]*ListOwner, error)
	GetUserLists(userID uuid.UUID) ([]*List, error)
	GetTribeLists(tribeID uuid.UUID) ([]*List, error)

	// Share management
	ShareWithTribe(share *ListShare) error
	UnshareWithTribe(listID, tribeID uuid.UUID) error
	GetSharedLists(tribeID uuid.UUID) ([]*List, error)

	// Sync management
	UpdateSyncStatus(listID uuid.UUID, status ListSyncStatus) error
	GetConflicts(listID uuid.UUID) ([]*ListConflict, error)
	CreateConflict(conflict *ListConflict) error
	ResolveConflict(conflictID uuid.UUID) error
	GetListsBySource(source string) ([]*List, error)
}
