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

// List represents a collection of items that can be used to generate menus
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
	Items []*ListItem `json:"items,omitempty" db:"-"`
}

// ListItem represents an item in a list
type ListItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ListID      uuid.UUID `json:"list_id" db:"list_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`

	// Item metadata
	Metadata   interface{} `json:"metadata,omitempty" db:"metadata"`       // Type-specific metadata (location data, etc.)
	ExternalID string      `json:"external_id,omitempty" db:"external_id"` // ID in external system

	// Menu generation parameters
	Weight      float64    `json:"weight" db:"weight"` // Custom weight for this item
	LastChosen  *time.Time `json:"last_chosen,omitempty" db:"last_chosen"`
	ChosenCount int        `json:"chosen_count" db:"chosen_count"`

	// Location-specific fields (null for non-location items)
	Latitude  *float64 `json:"latitude,omitempty" db:"latitude"`
	Longitude *float64 `json:"longitude,omitempty" db:"longitude"`
	Address   string   `json:"address,omitempty" db:"address"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ListConflict represents a sync conflict between local and external lists
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

// ListRepository defines the interface for list operations
type ListRepository interface {
	// Basic CRUD
	Create(list *List) error
	GetByID(id uuid.UUID) (*List, error)
	Update(list *List) error
	Delete(id uuid.UUID) error
	List(offset, limit int) ([]*List, error)

	// Item management
	AddItem(item *ListItem) error
	UpdateItem(item *ListItem) error
	RemoveItem(listID, itemID uuid.UUID) error
	GetItems(listID uuid.UUID) ([]*ListItem, error)

	// Menu generation
	GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*ListItem, error)
	UpdateItemStats(itemID uuid.UUID, chosen bool) error

	// Sync management
	UpdateSyncStatus(listID uuid.UUID, status ListSyncStatus) error
	GetListsBySource(source string) ([]*List, error)

	// Conflict management
	CreateConflict(conflict *ListConflict) error
	GetConflicts(listID uuid.UUID) ([]*ListConflict, error)
	ResolveConflict(conflictID uuid.UUID) error
}
