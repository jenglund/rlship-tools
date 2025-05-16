package models

import (
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

// ListSyncStatus represents the sync status of a list
type ListSyncStatus string

const (
	ListSyncStatusNone     ListSyncStatus = "none"     // No sync configured
	ListSyncStatusSynced   ListSyncStatus = "synced"   // In sync with external source
	ListSyncStatusPending  ListSyncStatus = "pending"  // Changes pending sync
	ListSyncStatusConflict ListSyncStatus = "conflict" // Conflicts detected
)

// List represents a collection of items that can be used for menu generation
type List struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Type        ListType       `json:"type" db:"type"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Visibility  VisibilityType `json:"visibility" db:"visibility"`

	// Sync information
	SyncStatus ListSyncStatus `json:"sync_status" db:"sync_status"`
	SyncSource string         `json:"sync_source,omitempty" db:"sync_source"` // e.g., "google_maps:want_to_go"
	SyncID     string         `json:"sync_id,omitempty" db:"sync_id"`         // External source ID
	LastSyncAt *time.Time     `json:"last_sync_at,omitempty" db:"last_sync_at"`

	// Menu generation settings
	DefaultWeight float64 `json:"default_weight" db:"default_weight"`         // Base weight for items in this list
	MaxItems      *int    `json:"max_items,omitempty" db:"max_items"`         // Max items to include in menu
	CooldownDays  *int    `json:"cooldown_days,omitempty" db:"cooldown_days"` // Days before reappearing in menu

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relations
	Items  []*ListItem  `json:"items,omitempty" db:"-"`
	Owners []*ListOwner `json:"owners,omitempty" db:"-"`
	Shares []*ListShare `json:"shares,omitempty" db:"-"`
}

// ListItem represents an item in a list
type ListItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ListID      uuid.UUID `json:"list_id" db:"list_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`

	// Item metadata
	Metadata   Metadata  `json:"metadata,omitempty" db:"metadata"`       // Type-specific metadata (location data, etc.)
	ExternalID string    `json:"external_id,omitempty" db:"external_id"` // ID in external system
	Location   *Location `json:"location,omitempty" db:"location"`       // Location data for location-type lists

	// Menu generation parameters
	Weight    float64    `json:"weight" db:"weight"`         // Item-specific weight for menu generation
	LastUsed  *time.Time `json:"last_used" db:"last_used"`   // Last time this item was selected in a menu
	UseCount  int        `json:"use_count" db:"use_count"`   // Number of times this item has been selected
	Cooldown  *int       `json:"cooldown" db:"cooldown"`     // Item-specific cooldown period in days
	Available bool       `json:"available" db:"available"`   // Whether this item is available for selection
	Seasonal  bool       `json:"seasonal" db:"seasonal"`     // Whether this item is seasonal
	StartDate *time.Time `json:"start_date" db:"start_date"` // Start of seasonal availability
	EndDate   *time.Time `json:"end_date" db:"end_date"`     // End of seasonal availability

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Location represents a geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
	PlaceID   string  `json:"place_id,omitempty"` // Google Maps Place ID
}

// Metadata represents additional data for a list item
type Metadata map[string]interface{}

// ListOwner represents an owner (user or tribe) of a list
type ListOwner struct {
	ListID    uuid.UUID  `json:"list_id" db:"list_id"`
	OwnerID   uuid.UUID  `json:"owner_id" db:"owner_id"`
	OwnerType string     `json:"owner_type" db:"owner_type"` // "user" or "tribe"
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ListShare represents a shared list with a tribe
type ListShare struct {
	ListID    uuid.UUID  `json:"list_id" db:"list_id"`
	TribeID   uuid.UUID  `json:"tribe_id" db:"tribe_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ListConflict represents a conflict during list synchronization
type ListConflict struct {
	ID     uuid.UUID  `json:"id" db:"id"`
	ListID uuid.UUID  `json:"list_id" db:"list_id"`
	ItemID *uuid.UUID `json:"item_id,omitempty" db:"item_id"`

	// Conflict details
	ConflictType string      `json:"conflict_type" db:"conflict_type"` // "added", "removed", "modified"
	LocalData    interface{} `json:"local_data,omitempty" db:"local_data"`
	ExternalData interface{} `json:"external_data,omitempty" db:"external_data"`

	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

// MenuParams represents parameters for menu generation
type MenuParams struct {
	ListIDs []uuid.UUID            `json:"list_ids"`
	Filters map[string]interface{} `json:"filters"`
}

// ListRepository defines the interface for list operations
type ListRepository interface {
	// List management
	Create(list *List) error
	GetByID(id uuid.UUID) (*List, error)
	Update(list *List) error
	Delete(id uuid.UUID) error
	List(offset, limit int) ([]*List, error)

	// List items
	AddItem(item *ListItem) error
	GetItems(listID uuid.UUID) ([]*ListItem, error)
	UpdateItem(item *ListItem) error
	RemoveItem(listID, itemID uuid.UUID) error
	GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*ListItem, error)
	UpdateItemStats(itemID uuid.UUID, selected bool) error

	// Sync management
	GetConflicts(listID uuid.UUID) ([]*ListConflict, error)
	CreateConflict(conflict *ListConflict) error
	ResolveConflict(conflictID uuid.UUID) error

	// Owner management
	AddOwner(listID, ownerID uuid.UUID, ownerType string) error
	RemoveOwner(listID, ownerID uuid.UUID) error
	GetOwners(listID uuid.UUID) ([]*ListOwner, error)
	GetUserLists(userID uuid.UUID) ([]*List, error)
	GetTribeLists(tribeID uuid.UUID) ([]*List, error)

	// Share management
	ShareWithTribe(listID, tribeID, userID uuid.UUID, expiresAt *time.Time) error
	UnshareWithTribe(listID, tribeID uuid.UUID) error
	GetSharedLists(tribeID uuid.UUID) ([]*List, error)
}
