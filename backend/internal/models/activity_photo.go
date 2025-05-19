package models

import (
	"github.com/google/uuid"
)

// ActivityPhoto represents a photo associated with an activity
type ActivityPhoto struct {
	BaseModel
	ActivityID uuid.UUID `json:"activity_id"`
	URL        string    `json:"url"`
	Caption    string    `json:"caption"`
	Metadata   JSONMap   `json:"metadata"`
}

// Validate ensures the activity photo data is valid
func (p *ActivityPhoto) Validate() error {
	if err := p.BaseModel.Validate(); err != nil {
		return err
	}
	if p.ActivityID == uuid.Nil {
		return ErrInvalidActivityID
	}
	if p.URL == "" {
		return ErrInvalidURL
	}
	return nil
}

// ActivityPhotosRepository defines the interface for activity photo data operations
type ActivityPhotosRepository interface {
	Create(photo *ActivityPhoto) error
	GetByID(id uuid.UUID) (*ActivityPhoto, error)
	Update(photo *ActivityPhoto) error
	Delete(id uuid.UUID) error
}
