package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestSyncList tests the SyncList function with various scenarios
func TestSyncList(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(*testutil.MockListRepository, uuid.UUID) *models.List
		wantErr error
	}{
		{
			name: "Happy path - sync successful",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "map_id_123",
					SyncStatus: models.ListSyncStatusNone,
					LastSyncAt: nil,
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil).Once()
				return list
			},
			wantErr: nil,
		},
		{
			name: "List not found",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound).Once()
				return nil
			},
			wantErr: models.ErrNotFound,
		},
		{
			name: "No sync source configured",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: "", // No sync source
					SyncID:     "",
					SyncStatus: models.ListSyncStatusNone,
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				return list
			},
			wantErr: models.ErrInvalidInput,
		},
		{
			name: "Error updating to pending status",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "map_id_123",
					SyncStatus: models.ListSyncStatusNone,
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(errors.New("db error")).Once()
				return list
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "Error updating to synced status",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "map_id_123",
					SyncStatus: models.ListSyncStatusNone,
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(errors.New("db error")).Once()
				return list
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "List already in pending state",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "map_id_123",
					SyncStatus: models.ListSyncStatusPending, // Already pending
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil).Once()
				return list
			},
			wantErr: nil,
		},
		{
			name: "List with conflict status",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "map_id_123",
					SyncStatus: models.ListSyncStatusConflict, // In conflict state
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil).Once()
				return list
			},
			wantErr: nil,
		},
		{
			name: "External source unavailable",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "map_id_123",
					SyncStatus: models.ListSyncStatusNone,
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(models.ErrExternalSourceUnavailable).Once()
				return list
			},
			wantErr: models.ErrExternalSourceUnavailable,
		},
		{
			name: "External source timeout",
			setup: func(mockRepo *testutil.MockListRepository, listID uuid.UUID) *models.List {
				list := &models.List{
					ID:         listID,
					Type:       models.ListTypeGeneral,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "map_id_123",
					SyncStatus: models.ListSyncStatusNone,
				}
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
				mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(models.ErrExternalSourceTimeout).Once()
				return list
			},
			wantErr: models.ErrExternalSourceTimeout,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(testutil.MockListRepository)
			service := NewListService(mockRepo)
			listID := uuid.New()

			_ = tc.setup(mockRepo, listID)

			err := service.SyncList(listID)

			if tc.wantErr != nil {
				if tc.wantErr.Error() == "db error" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "db error")
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetListConflicts tests the GetListConflicts function with various scenarios
func TestGetListConflicts(t *testing.T) {
	listID := uuid.New()
	now := time.Now()
	itemID := uuid.New()

	conflicts := []*models.SyncConflict{
		{
			ID:         uuid.New(),
			ListID:     listID,
			ItemID:     &itemID,
			Type:       "name_conflict",
			LocalData:  "Local Name",
			RemoteData: "Remote Name",
			CreatedAt:  now,
			UpdatedAt:  now,
			ResolvedAt: nil,
		},
		{
			ID:         uuid.New(),
			ListID:     listID,
			ItemID:     nil, // No item ID for list-level conflict
			Type:       "list_deletion_conflict",
			LocalData:  "Deleted locally",
			RemoteData: "Still exists remotely",
			CreatedAt:  now,
			UpdatedAt:  now,
			ResolvedAt: nil,
		},
	}

	testCases := []struct {
		name      string
		setup     func(*testutil.MockListRepository)
		wantErr   error
		wantCount int
	}{
		{
			name: "Happy path - conflicts found",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetConflicts", listID).Return(conflicts, nil).Once()
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name: "No conflicts found",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetConflicts", listID).Return([]*models.SyncConflict{}, nil).Once()
			},
			wantErr:   nil,
			wantCount: 0,
		},
		{
			name: "List not found",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetConflicts", listID).Return(nil, models.ErrNotFound).Once()
			},
			wantErr:   models.ErrNotFound,
			wantCount: 0,
		},
		{
			name: "Sync disabled",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetConflicts", listID).Return(nil, models.ErrSyncDisabled).Once()
			},
			wantErr:   models.ErrSyncDisabled,
			wantCount: 0,
		},
		{
			name: "Database error",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetConflicts", listID).Return(nil, errors.New("db error")).Once()
			},
			wantErr:   errors.New("db error"),
			wantCount: 0,
		},
		{
			name: "External source unavailable",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetConflicts", listID).Return(nil, models.ErrExternalSourceUnavailable).Once()
			},
			wantErr:   models.ErrExternalSourceUnavailable,
			wantCount: 0,
		},
		{
			name: "External source timeout",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetConflicts", listID).Return(nil, models.ErrExternalSourceTimeout).Once()
			},
			wantErr:   models.ErrExternalSourceTimeout,
			wantCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(testutil.MockListRepository)
			service := NewListService(mockRepo)

			tc.setup(mockRepo)

			result, err := service.GetListConflicts(listID)

			if tc.wantErr != nil {
				if tc.wantErr.Error() == "db error" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "db error")
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tc.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestResolveListConflict tests the ResolveListConflict function with various scenarios
func TestResolveListConflict(t *testing.T) {
	listID := uuid.New()
	conflictID := uuid.New()

	testCases := []struct {
		name       string
		resolution string
		setup      func(*testutil.MockListRepository)
		wantErr    error
	}{
		{
			name:       "Happy path - accept local",
			resolution: "accept_local",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", conflictID).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:       "Happy path - accept remote",
			resolution: "accept_remote",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", conflictID).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:       "Invalid conflict ID",
			resolution: "accept_local",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", uuid.Nil).Return(models.ErrInvalidInput).Once()
			},
			wantErr: models.ErrInvalidInput,
		},
		{
			name:       "Conflict not found",
			resolution: "accept_local",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", conflictID).Return(models.ErrNotFound).Once()
			},
			wantErr: models.ErrNotFound,
		},
		{
			name:       "Conflict already resolved",
			resolution: "accept_local",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", conflictID).Return(models.ErrConflictAlreadyResolved).Once()
			},
			wantErr: models.ErrConflictAlreadyResolved,
		},
		{
			name:       "Invalid resolution",
			resolution: "invalid_resolution",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", conflictID).Return(models.ErrInvalidResolution).Once()
			},
			wantErr: models.ErrInvalidResolution,
		},
		{
			name:       "Sync disabled",
			resolution: "accept_local",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", conflictID).Return(models.ErrSyncDisabled).Once()
			},
			wantErr: models.ErrSyncDisabled,
		},
		{
			name:       "Database error",
			resolution: "accept_local",
			setup: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("ResolveConflict", conflictID).Return(errors.New("db error")).Once()
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(testutil.MockListRepository)
			service := NewListService(mockRepo)

			if tc.name == "Invalid conflict ID" {
				conflictID = uuid.Nil
			} else {
				conflictID = uuid.New()
			}

			tc.setup(mockRepo)

			err := service.ResolveListConflict(listID, conflictID, tc.resolution)

			if tc.wantErr != nil {
				if tc.wantErr.Error() == "db error" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "db error")
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestSyncEdgeCases tests edge cases in the sync functionality
func TestSyncEdgeCases(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()

	t.Run("Concurrent modification handling", func(t *testing.T) {
		list := &models.List{
			ID:         listID,
			Type:       models.ListTypeGeneral,
			SyncSource: models.SyncSourceGoogleMaps,
			SyncID:     "map_id_123",
			SyncStatus: models.ListSyncStatusNone,
		}

		// First attempt gets a concurrent modification error
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(models.ErrConcurrentModification).Once()

		// Second attempt after retrying gets the updated list
		updatedList := *list
		mockRepo.On("GetByID", listID).Return(&updatedList, nil).Once()
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil).Once()

		err := service.SyncList(listID)
		assert.NoError(t, err)
	})

	t.Run("External source integration", func(t *testing.T) {
		// This test would simulate an actual integration with an external source
		// For now, we'll just test the basic flow with mocks
		list := &models.List{
			ID:         listID,
			Type:       models.ListTypeGeneral,
			SyncSource: models.SyncSourceGoogleMaps,
			SyncID:     "map_id_123",
			SyncStatus: models.ListSyncStatusNone,
		}

		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()

		// In a real implementation, we would call the external API here
		// We're simulating a successful sync
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil).Once()

		err := service.SyncList(listID)
		assert.NoError(t, err)
	})

	// Add additional edge cases as needed
}

func TestCreateListConflict(t *testing.T) {
	// Create a mock repository
	mockRepo := new(testutil.MockListRepository)

	// Create a list service with the mock repository
	service := NewListService(mockRepo)

	// Set up test data
	listID := uuid.New()
	conflictID := uuid.New()

	testCases := []struct {
		name      string
		conflict  *models.SyncConflict
		mockSetup func()
		expectErr bool
		errCheck  func(error) bool
	}{
		{
			name: "valid conflict",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockSetup: func() {
				mockRepo.On("CreateConflict", mock.MatchedBy(func(c *models.SyncConflict) bool {
					return c.ID == conflictID && c.ListID == listID
				})).Return(nil).Once()
			},
			expectErr: false,
		},
		{
			name: "repository error",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockSetup: func() {
				mockRepo.On("CreateConflict", mock.Anything).Return(errors.New("database error")).Once()
			},
			expectErr: true,
			errCheck: func(err error) bool {
				return err != nil && err.Error() == "database error"
			},
		},
		{
			name:     "nil conflict",
			conflict: nil,
			mockSetup: func() {
				// The implementation now validates nil conflicts
				// No repository call expected
			},
			expectErr: true,
			errCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "conflict cannot be nil")
			},
		},
		{
			name: "invalid conflict - missing ID",
			conflict: &models.SyncConflict{
				// Missing ID
				ListID:     listID,
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockSetup: func() {
				// The service passes all validation to the repository
				mockRepo.On("CreateConflict", mock.MatchedBy(func(c *models.SyncConflict) bool {
					return c.ID == uuid.Nil && c.ListID == listID
				})).Return(errors.New("invalid conflict: ID is required")).Once()
			},
			expectErr: true,
			errCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "invalid conflict: ID is required")
			},
		},
		{
			name: "invalid conflict - missing ListID",
			conflict: &models.SyncConflict{
				ID: conflictID,
				// Missing ListID
				Type:       "item_update",
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockSetup: func() {
				// The service passes all validation to the repository
				mockRepo.On("CreateConflict", mock.MatchedBy(func(c *models.SyncConflict) bool {
					return c.ID == conflictID && c.ListID == uuid.Nil
				})).Return(errors.New("invalid conflict: list ID is required")).Once()
			},
			expectErr: true,
			errCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "invalid conflict: list ID is required")
			},
		},
		{
			name: "invalid conflict - missing Type",
			conflict: &models.SyncConflict{
				ID:     conflictID,
				ListID: listID,
				// Missing Type
				LocalData:  map[string]interface{}{"name": "Local Name"},
				RemoteData: map[string]interface{}{"name": "Remote Name"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockSetup: func() {
				// The service passes all validation to the repository
				mockRepo.On("CreateConflict", mock.MatchedBy(func(c *models.SyncConflict) bool {
					return c.ID == conflictID && c.ListID == listID && c.Type == ""
				})).Return(errors.New("invalid conflict: conflict type is required")).Once()
			},
			expectErr: true,
			errCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "invalid conflict: conflict type is required")
			},
		},
		{
			name: "invalid conflict - missing LocalData",
			conflict: &models.SyncConflict{
				ID:     conflictID,
				ListID: listID,
				Type:   "item_update",
				// Missing LocalData
				RemoteData: map[string]interface{}{"name": "Remote Name"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockSetup: func() {
				// The service passes all validation to the repository
				mockRepo.On("CreateConflict", mock.MatchedBy(func(c *models.SyncConflict) bool {
					return c.ID == conflictID && c.ListID == listID && c.LocalData == nil
				})).Return(errors.New("invalid conflict: local data is required")).Once()
			},
			expectErr: true,
			errCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "invalid conflict: local data is required")
			},
		},
		{
			name: "invalid conflict - missing RemoteData",
			conflict: &models.SyncConflict{
				ID:        conflictID,
				ListID:    listID,
				Type:      "item_update",
				LocalData: map[string]interface{}{"name": "Local Name"},
				// Missing RemoteData
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockSetup: func() {
				// The service passes all validation to the repository
				mockRepo.On("CreateConflict", mock.MatchedBy(func(c *models.SyncConflict) bool {
					return c.ID == conflictID && c.ListID == listID && c.RemoteData == nil
				})).Return(errors.New("invalid conflict: remote data is required")).Once()
			},
			expectErr: true,
			errCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "invalid conflict: remote data is required")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mock expectations
			if tc.mockSetup != nil {
				tc.mockSetup()
			}

			// Call the service function
			var err error
			if tc.conflict != nil {
				err = service.CreateListConflict(tc.conflict)
			} else {
				err = service.CreateListConflict(nil)
			}

			// Verify expectations
			if tc.expectErr {
				assert.Error(t, err)
				if tc.errCheck != nil {
					assert.True(t, tc.errCheck(err), "Error did not match the expected error condition: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mock was called as expected
			mockRepo.AssertExpectations(t)
		})
	}
}
