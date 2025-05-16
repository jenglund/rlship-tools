package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// VisibilityType defines who can see a resource
type VisibilityType string

const (
	VisibilityPrivate VisibilityType = "private" // Only visible to owner
	VisibilityShared  VisibilityType = "shared"  // Visible to specific tribes/users
	VisibilityPublic  VisibilityType = "public"  // Visible to everyone
)

func (v VisibilityType) Validate() error {
	switch v {
	case VisibilityPrivate, VisibilityShared, VisibilityPublic:
		return nil
	default:
		return fmt.Errorf("invalid visibility type: %s", v)
	}
}

// OwnerType defines what type of entity owns a resource
type OwnerType string

const (
	OwnerTypeUser  OwnerType = "user"
	OwnerTypeTribe OwnerType = "tribe"
)

func (o OwnerType) Validate() error {
	switch o {
	case OwnerTypeUser, OwnerTypeTribe:
		return nil
	default:
		return fmt.Errorf("invalid owner type: %s", o)
	}
}

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Validate performs basic validation on the base model
func (b *BaseModel) Validate() error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("id is required")
	}
	if b.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}
	if b.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	if b.DeletedAt != nil && b.DeletedAt.Before(b.CreatedAt) {
		return fmt.Errorf("deleted_at cannot be before created_at")
	}
	return nil
}

// Location represents a physical location that can be referenced in activities
type Location struct {
	BaseModel
	Name        string  `json:"name" db:"name"`
	Address     string  `json:"address" db:"address"`
	Latitude    float64 `json:"latitude" db:"latitude"`
	Longitude   float64 `json:"longitude" db:"longitude"`
	GoogleMapID string  `json:"google_map_id,omitempty" db:"google_map_id"`
	Metadata    JSONMap `json:"metadata,omitempty" db:"metadata"`
}

// Validate performs validation on the location
func (l *Location) Validate() error {
	if err := l.BaseModel.Validate(); err != nil {
		return err
	}
	if l.Name == "" {
		return fmt.Errorf("name is required")
	}
	if l.Address == "" {
		return fmt.Errorf("address is required")
	}
	if l.Latitude < -90 || l.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

// MenuParams represents parameters for generating a menu from lists
type MenuParams struct {
	ListIDs      []uuid.UUID            `json:"list_ids"`
	Count        int                    `json:"count"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	ExcludeItems []uuid.UUID            `json:"exclude_items,omitempty"`
}

// Validate performs validation on the menu parameters
func (m *MenuParams) Validate() error {
	if len(m.ListIDs) == 0 {
		return fmt.Errorf("at least one list ID is required")
	}
	if m.Count <= 0 {
		return fmt.Errorf("count must be positive")
	}
	return nil
}

// ListConflict represents a conflict between local and external list data
type ListConflict struct {
	BaseModel
	ListID       uuid.UUID  `json:"list_id" db:"list_id"`
	ItemID       *uuid.UUID `json:"item_id,omitempty" db:"item_id"`
	ConflictType string     `json:"conflict_type" db:"conflict_type"`
	LocalData    JSONMap    `json:"local_data" db:"local_data"`
	ExternalData JSONMap    `json:"external_data" db:"external_data"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

// Validate performs validation on the list conflict
func (c *ListConflict) Validate() error {
	if err := c.BaseModel.Validate(); err != nil {
		return err
	}
	if c.ListID == uuid.Nil {
		return fmt.Errorf("list ID is required")
	}
	if c.ConflictType == "" {
		return fmt.Errorf("conflict type is required")
	}
	return nil
}

// JSONMap is a type alias for map[string]interface{} that implements database scanning
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface
func (m *JSONMap) Scan(value interface{}) error {
	// Implementation will depend on your database driver and JSON handling
	return nil
}

// Value implements the driver.Valuer interface
func (m JSONMap) Value() (interface{}, error) {
	// Implementation will depend on your database driver and JSON handling
	return nil, nil
}
