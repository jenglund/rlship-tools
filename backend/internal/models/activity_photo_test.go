package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestActivityPhotoValidation(t *testing.T) {
	validID := uuid.New()
	activityID := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		photo     *ActivityPhoto
		expectErr bool
		errType   error
	}{
		{
			name: "valid photo",
			photo: &ActivityPhoto{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				ActivityID: activityID,
				URL:        "https://example.com/photo.jpg",
				Caption:    "Test photo",
				Metadata:   JSONMap{"location": "Test Location"},
			},
			expectErr: false,
		},
		{
			name: "nil ID",
			photo: &ActivityPhoto{
				BaseModel: BaseModel{
					ID:        uuid.Nil,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				ActivityID: activityID,
				URL:        "https://example.com/photo.jpg",
				Caption:    "Test photo",
				Metadata:   JSONMap{"location": "Test Location"},
			},
			expectErr: true,
			errType:   ErrInvalidInput,
		},
		{
			name: "nil activity ID",
			photo: &ActivityPhoto{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				ActivityID: uuid.Nil,
				URL:        "https://example.com/photo.jpg",
				Caption:    "Test photo",
				Metadata:   JSONMap{"location": "Test Location"},
			},
			expectErr: true,
			errType:   ErrInvalidActivityID,
		},
		{
			name: "empty URL",
			photo: &ActivityPhoto{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				ActivityID: activityID,
				URL:        "",
				Caption:    "Test photo",
				Metadata:   JSONMap{"location": "Test Location"},
			},
			expectErr: true,
			errType:   ErrInvalidURL,
		},
		{
			name: "empty caption is valid",
			photo: &ActivityPhoto{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				ActivityID: activityID,
				URL:        "https://example.com/photo.jpg",
				Caption:    "",
				Metadata:   JSONMap{"location": "Test Location"},
			},
			expectErr: false,
		},
		{
			name: "nil metadata is valid",
			photo: &ActivityPhoto{
				BaseModel: BaseModel{
					ID:        validID,
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
				ActivityID: activityID,
				URL:        "https://example.com/photo.jpg",
				Caption:    "Test photo",
				Metadata:   nil,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.photo.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
