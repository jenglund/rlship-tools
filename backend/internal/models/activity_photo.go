package models

import (
	"time"

	"github.com/google/uuid"
)

// ActivityPhoto represents a photo associated with an activity
type ActivityPhoto struct {
	ID         uuid.UUID  `json:"id"`
	ActivityID uuid.UUID  `json:"activity_id"`
	URL        string     `json:"url"`
	Caption    string     `json:"caption"`
	Metadata   JSONMap    `json:"metadata"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// Validate ensures the activity photo data is valid
func (p *ActivityPhoto) Validate() error {
	if p.ID == uuid.Nil {
		return ErrInvalidID
	}
	if p.ActivityID == uuid.Nil {
		return ErrInvalidActivityID
	}
	if p.URL == "" {
		return ErrInvalidURL
	}
	if p.CreatedAt.IsZero() {
		return ErrInvalidCreatedAt
	}
	if p.UpdatedAt.IsZero() {
		return ErrInvalidUpdatedAt
	}
	if p.DeletedAt != nil && p.DeletedAt.Before(p.CreatedAt) {
		return ErrInvalidDeletedAt
	}
	return nil
}
