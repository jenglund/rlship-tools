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

func TestJSONMap_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    JSONMap
		wantErr bool
	}{
		{
			name:  "nil input",
			input: nil,
			want:  JSONMap{},
		},
		{
			name:  "empty bytes",
			input: []byte{},
			want:  JSONMap{},
		},
		{
			name:  "empty string",
			input: "",
			want:  JSONMap{},
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`{"key": "value", "nums": [1, 2, 3]}`),
			want:  JSONMap{"key": "value", "nums": []string{"1", "2", "3"}},
		},
		{
			name:  "valid JSON string",
			input: `{"key": "value", "nums": [1, 2, 3]}`,
			want:  JSONMap{"key": "value", "nums": []string{"1", "2", "3"}},
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"key": "value"`),
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m JSONMap
			err := (&m).Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, m)
			}
		})
	}
}

func TestJSONMap_Value(t *testing.T) {
	tests := []struct {
		name    string
		m       JSONMap
		want    []byte
		wantErr bool
	}{
		{
			name: "nil map",
			m:    nil,
			want: nil,
		},
		{
			name: "empty map",
			m:    JSONMap{},
			want: []byte("{}"),
		},
		{
			name: "simple map",
			m:    JSONMap{"key": "value"},
			want: []byte(`{"key":"value"}`),
		},
		{
			name: "map with string array",
			m:    JSONMap{"arr": []string{"1", "2", "3"}},
			want: []byte(`{"arr":["1","2","3"]}`),
		},
		{
			name: "complex map",
			m: JSONMap{
				"str":   "value",
				"num":   123,
				"bool":  true,
				"arr":   []string{"a", "b", "c"},
				"null":  nil,
				"float": 123.45,
			},
			want: []byte(`{"arr":["a","b","c"],"bool":true,"float":123.45,"null":null,"num":123,"str":"value"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.Value()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.JSONEq(t, string(tt.want), string(got.([]byte)))
				}
			}
		})
	}
}

func TestJSONMap_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       JSONMap
		want    string
		wantErr bool
	}{
		{
			name: "nil map",
			m:    nil,
			want: "null",
		},
		{
			name: "empty map",
			m:    JSONMap{},
			want: "{}",
		},
		{
			name: "simple map",
			m:    JSONMap{"key": "value"},
			want: `{"key":"value"}`,
		},
		{
			name: "map with array",
			m:    JSONMap{"arr": []string{"1", "2", "3"}},
			want: `{"arr":["1","2","3"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, tt.want, string(got))
			}
		})
	}
}

func TestJSONMap_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    JSONMap
		wantErr bool
	}{
		{
			name:  "null input",
			input: "null",
			want:  JSONMap{},
		},
		{
			name:  "empty object",
			input: "{}",
			want:  JSONMap{},
		},
		{
			name:  "simple object",
			input: `{"key":"value"}`,
			want:  JSONMap{"key": "value"},
		},
		{
			name:  "object with array",
			input: `{"arr":[1,2,3]}`,
			want:  JSONMap{"arr": []string{"1", "2", "3"}},
		},
		{
			name:  "complex object",
			input: `{"str":"value","num":123,"bool":true,"arr":[1,2,3],"null":null}`,
			want: JSONMap{
				"str":  "value",
				"num":  float64(123),
				"bool": true,
				"arr":  []string{"1", "2", "3"},
				"null": nil,
			},
		},
		{
			name:    "invalid JSON",
			input:   `{"key":"value"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m JSONMap
			err := m.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, m)
			}
		})
	}
}
