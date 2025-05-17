package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
)

func TestAddListItem(t *testing.T) {
	// Create a mock repository
	mockRepo := new(testutil.MockListRepository)

	// Create a list service with the mock repository
	service := NewListService(mockRepo)

	// Test data
	listID := uuid.New()
	itemID := uuid.New()
	now := time.Now()

	list := &models.List{
		ID:          listID,
		Name:        "Test List",
		Description: "A test list",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	validItem := &models.ListItem{
		ID:        itemID,
		ListID:    listID,
		Name:      "Test Item",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name      string
		item      *models.ListItem
		mockSetup func()
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name: "valid item",
			item: validItem,
			mockSetup: func() {
				mockRepo.On("GetByID", listID).Return(list, nil)
				mockRepo.On("AddItem", validItem).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil list ID",
			item: &models.ListItem{
				ID:        itemID,
				Name:      "Test Item",
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockSetup: func() {
				// No mock calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidInput)
			},
		},
		{
			name: "empty name",
			item: &models.ListItem{
				ID:        itemID,
				ListID:    listID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockSetup: func() {
				// No mock calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidInput)
			},
		},
		{
			name: "list not found",
			item: validItem,
			mockSetup: func() {
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound)
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrNotFound)
			},
		},
		{
			name: "repository error",
			item: validItem,
			mockSetup: func() {
				mockRepo.On("GetByID", listID).Return(list, nil)
				mockRepo.On("AddItem", validItem).Return(errors.New("database error"))
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil

			// Set up mock
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// Call the method
			err := service.AddListItem(tt.item)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err), "error check failed: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateListItem(t *testing.T) {
	// Create a mock repository
	mockRepo := new(testutil.MockListRepository)

	// Create a list service with the mock repository
	service := NewListService(mockRepo)

	// Test data
	listID := uuid.New()
	itemID := uuid.New()
	now := time.Now()

	validItem := &models.ListItem{
		ID:        itemID,
		ListID:    listID,
		Name:      "Test Item",
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name      string
		item      *models.ListItem
		mockSetup func()
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name: "valid item",
			item: validItem,
			mockSetup: func() {
				mockRepo.On("UpdateItem", validItem).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil item ID",
			item: &models.ListItem{
				ListID:    listID,
				Name:      "Test Item",
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockSetup: func() {
				// No mock calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidInput)
			},
		},
		{
			name: "nil list ID",
			item: &models.ListItem{
				ID:        itemID,
				Name:      "Test Item",
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockSetup: func() {
				// No mock calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidInput)
			},
		},
		{
			name: "empty name",
			item: &models.ListItem{
				ID:        itemID,
				ListID:    listID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockSetup: func() {
				// No mock calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidInput)
			},
		},
		{
			name: "repository error",
			item: validItem,
			mockSetup: func() {
				mockRepo.On("UpdateItem", validItem).Return(errors.New("database error"))
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil

			// Set up mock
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// Call the method
			err := service.UpdateListItem(tt.item)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err), "error check failed: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRemoveListItem(t *testing.T) {
	// Create a mock repository
	mockRepo := new(testutil.MockListRepository)

	// Create a list service with the mock repository
	service := NewListService(mockRepo)

	// Test data
	listID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name      string
		listID    uuid.UUID
		itemID    uuid.UUID
		mockSetup func()
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name:   "valid IDs",
			listID: listID,
			itemID: itemID,
			mockSetup: func() {
				mockRepo.On("RemoveItem", listID, itemID).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "nil list ID",
			listID: uuid.Nil,
			itemID: itemID,
			mockSetup: func() {
				// No mock calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidInput)
			},
		},
		{
			name:   "nil item ID",
			listID: listID,
			itemID: uuid.Nil,
			mockSetup: func() {
				// No mock calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidInput)
			},
		},
		{
			name:   "repository error",
			listID: listID,
			itemID: itemID,
			mockSetup: func() {
				mockRepo.On("RemoveItem", listID, itemID).Return(errors.New("database error"))
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil

			// Set up mock
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// Call the method
			err := service.RemoveListItem(tt.listID, tt.itemID)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err), "error check failed: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
		})
	}
}
