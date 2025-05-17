package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestActivity_Validate(t *testing.T) {
	validID := uuid.New()
	validUserID := uuid.New()
	now := time.Now()

	tests := []struct {
		name    string
		a       Activity
		wantErr bool
	}{
		{
			name: "valid activity",
			a: Activity{
				ID:          validID,
				UserID:      validUserID,
				Type:        ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  VisibilityPrivate,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			a: Activity{
				UserID:      validUserID,
				Type:        ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  VisibilityPrivate,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing user ID",
			a: Activity{
				ID:          validID,
				Type:        ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  VisibilityPrivate,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			a: Activity{
				ID:          validID,
				UserID:      validUserID,
				Type:        ActivityTypeLocation,
				Description: "Test Description",
				Visibility:  VisibilityPrivate,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid visibility",
			a: Activity{
				ID:          validID,
				UserID:      validUserID,
				Type:        ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  VisibilityType("invalid"),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing created_at",
			a: Activity{
				ID:          validID,
				UserID:      validUserID,
				Type:        ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  VisibilityPrivate,
				UpdatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing updated_at",
			a: Activity{
				ID:          validID,
				UserID:      validUserID,
				Type:        ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  VisibilityPrivate,
				CreatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "deleted_at before created_at",
			a: Activity{
				ID:          validID,
				UserID:      validUserID,
				Type:        ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  VisibilityPrivate,
				CreatedAt:   now,
				UpdatedAt:   now,
				DeletedAt:   &time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.a.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
