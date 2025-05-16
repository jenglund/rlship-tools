package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVisibilityType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		v       VisibilityType
		wantErr bool
	}{
		{"private", VisibilityPrivate, false},
		{"shared", VisibilityShared, false},
		{"public", VisibilityPublic, false},
		{"invalid", VisibilityType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.v.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOwnerType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		o       OwnerType
		wantErr bool
	}{
		{"user", OwnerTypeUser, false},
		{"tribe", OwnerTypeTribe, false},
		{"invalid", OwnerType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.o.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseModel_Validate(t *testing.T) {
	now := time.Now()
	validID := uuid.New()

	tests := []struct {
		name    string
		b       BaseModel
		wantErr bool
	}{
		{
			name: "valid",
			b: BaseModel{
				ID:        validID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: false,
		},
		{
			name: "nil id",
			b: BaseModel{
				ID:        uuid.Nil,
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "zero created_at",
			b: BaseModel{
				ID:        validID,
				CreatedAt: time.Time{},
				UpdatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "zero updated_at",
			b: BaseModel{
				ID:        validID,
				CreatedAt: now,
				UpdatedAt: time.Time{},
			},
			wantErr: true,
		},
		{
			name: "deleted_at before created_at",
			b: BaseModel{
				ID:        validID,
				CreatedAt: now,
				UpdatedAt: now,
				DeletedAt: &time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.b.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLocationRef_Validate(t *testing.T) {
	tests := []struct {
		name    string
		l       LocationRef
		wantErr bool
	}{
		{
			name: "valid",
			l: LocationRef{
				Latitude:  45.0,
				Longitude: -122.0,
				Address:   "123 Test St",
				PlaceID:   "test_place_id",
			},
			wantErr: false,
		},
		{
			name: "invalid latitude high",
			l: LocationRef{
				Latitude:  91.0,
				Longitude: -122.0,
				Address:   "123 Test St",
			},
			wantErr: true,
		},
		{
			name: "invalid latitude low",
			l: LocationRef{
				Latitude:  -91.0,
				Longitude: -122.0,
				Address:   "123 Test St",
			},
			wantErr: true,
		},
		{
			name: "invalid longitude high",
			l: LocationRef{
				Latitude:  45.0,
				Longitude: 181.0,
				Address:   "123 Test St",
			},
			wantErr: true,
		},
		{
			name: "invalid longitude low",
			l: LocationRef{
				Latitude:  45.0,
				Longitude: -181.0,
				Address:   "123 Test St",
			},
			wantErr: true,
		},
		{
			name: "empty address",
			l: LocationRef{
				Latitude:  45.0,
				Longitude: -122.0,
				Address:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.l.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLocation_Validate(t *testing.T) {
	tests := []struct {
		name    string
		loc     *Location
		wantErr bool
	}{
		{
			name: "valid location",
			loc: &Location{
				BaseModel: BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Location",
				Address:   "123 Test St",
				Latitude:  45.0,
				Longitude: -122.0,
			},
			wantErr: false,
		},
		{
			name: "invalid latitude low",
			loc: &Location{
				BaseModel: BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Location",
				Address:   "123 Test St",
				Latitude:  -91.0,
				Longitude: -122.0,
			},
			wantErr: true,
		},
		{
			name: "invalid latitude high",
			loc: &Location{
				BaseModel: BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Location",
				Address:   "123 Test St",
				Latitude:  91.0,
				Longitude: -122.0,
			},
			wantErr: true,
		},
		{
			name: "invalid longitude low",
			loc: &Location{
				BaseModel: BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Location",
				Address:   "123 Test St",
				Latitude:  45.0,
				Longitude: -181.0,
			},
			wantErr: true,
		},
		{
			name: "invalid longitude high",
			loc: &Location{
				BaseModel: BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Location",
				Address:   "123 Test St",
				Latitude:  45.0,
				Longitude: 181.0,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			loc: &Location{
				BaseModel: BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Address:   "123 Test St",
				Latitude:  45.0,
				Longitude: -122.0,
			},
			wantErr: true,
		},
		{
			name: "empty address",
			loc: &Location{
				BaseModel: BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Location",
				Latitude:  45.0,
				Longitude: -122.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loc.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMenuParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		m       MenuParams
		wantErr bool
	}{
		{
			name: "valid",
			m: MenuParams{
				ListIDs: []uuid.UUID{uuid.New()},
				Count:   1,
			},
			wantErr: false,
		},
		{
			name: "no list IDs",
			m: MenuParams{
				Count: 1,
			},
			wantErr: true,
		},
		{
			name: "zero count",
			m: MenuParams{
				ListIDs: []uuid.UUID{uuid.New()},
				Count:   0,
			},
			wantErr: true,
		},
		{
			name: "negative count",
			m: MenuParams{
				ListIDs: []uuid.UUID{uuid.New()},
				Count:   -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListConflict_Validate(t *testing.T) {
	now := time.Now()
	validID := uuid.New()
	validBase := BaseModel{
		ID:        validID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name    string
		c       ListConflict
		wantErr bool
	}{
		{
			name: "valid",
			c: ListConflict{
				BaseModel:    validBase,
				ListID:       uuid.New(),
				ConflictType: "modified",
			},
			wantErr: false,
		},
		{
			name: "invalid base model",
			c: ListConflict{
				BaseModel:    BaseModel{},
				ListID:       uuid.New(),
				ConflictType: "modified",
			},
			wantErr: true,
		},
		{
			name: "nil list ID",
			c: ListConflict{
				BaseModel:    validBase,
				ConflictType: "modified",
			},
			wantErr: true,
		},
		{
			name: "empty conflict type",
			c: ListConflict{
				BaseModel: validBase,
				ListID:    uuid.New(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
