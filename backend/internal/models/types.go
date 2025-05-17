package models

import (
	"encoding/json"
	"fmt"
	"time"

	"database/sql/driver"

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
		return fmt.Errorf("%w: invalid visibility type: %s", ErrInvalidInput, v)
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
		return fmt.Errorf("%w: invalid owner type: %s", ErrInvalidInput, o)
	}
}

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Version   int        `json:"version" db:"version"`
}

// Validate performs basic validation on the base model
func (b *BaseModel) Validate() error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	if b.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created_at is required", ErrInvalidInput)
	}
	if b.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: updated_at is required", ErrInvalidInput)
	}
	if b.DeletedAt != nil && b.DeletedAt.Before(b.CreatedAt) {
		return fmt.Errorf("%w: deleted_at cannot be before created_at", ErrInvalidInput)
	}
	if b.Version < 1 {
		return fmt.Errorf("%w: version must be at least 1", ErrInvalidInput)
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
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if l.Address == "" {
		return fmt.Errorf("%w: address is required", ErrInvalidInput)
	}
	if l.Latitude < -90 || l.Latitude > 90 {
		return fmt.Errorf("%w: latitude must be between -90 and 90", ErrInvalidInput)
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("%w: longitude must be between -180 and 180", ErrInvalidInput)
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
		return fmt.Errorf("%w: at least one list ID is required", ErrInvalidInput)
	}
	if m.Count <= 0 {
		return fmt.Errorf("%w: count must be positive", ErrInvalidInput)
	}
	return nil
}

// JSONMap is a type alias for map[string]interface{} that implements database scanning
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap) // Always initialize to empty map, never nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("%w: unsupported JSONMap scan type: %T", ErrInvalidInput, value)
	}

	// Handle empty string/bytes as empty map
	if len(bytes) == 0 {
		*m = make(JSONMap)
		return nil
	}

	// Parse JSON into map
	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return fmt.Errorf("%w: invalid JSON data: %v", ErrInvalidInput, err)
	}

	// Convert any []interface{} to []string where appropriate
	for key, value := range result {
		if arr, ok := value.([]interface{}); ok {
			strArr := make([]string, len(arr))
			for i, v := range arr {
				strArr[i] = fmt.Sprint(v)
			}
			result[key] = strArr
		}
	}

	*m = result
	return nil
}

// Value implements the driver.Valuer interface
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}

	// Convert any []string back to []interface{} for consistent JSON encoding
	converted := make(map[string]interface{})
	for key, value := range m {
		if strArr, ok := value.([]string); ok {
			interfaceArr := make([]interface{}, len(strArr))
			for i, v := range strArr {
				interfaceArr[i] = v
			}
			converted[key] = interfaceArr
		} else {
			converted[key] = value
		}
	}

	return json.Marshal(converted)
}

// MarshalJSON implements the json.Marshaler interface
func (m JSONMap) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}(m))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (m *JSONMap) UnmarshalJSON(data []byte) error {
	if m == nil {
		return fmt.Errorf("%w: JSONMap is nil", ErrInvalidInput)
	}

	// Handle null JSON value
	if string(data) == "null" {
		*m = make(JSONMap)
		return nil
	}

	// Parse JSON into temporary map
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: invalid JSON data: %v", ErrInvalidInput, err)
	}

	// Convert any []interface{} to []string where appropriate
	for key, value := range raw {
		if arr, ok := value.([]interface{}); ok {
			strArr := make([]string, len(arr))
			for i, v := range arr {
				strArr[i] = fmt.Sprint(v)
			}
			raw[key] = strArr
		}
	}

	*m = raw
	return nil
}
