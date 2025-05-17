package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// createValidTestList creates a valid list with all required fields for testing
func createValidTestList(id uuid.UUID, syncStatus models.ListSyncStatus, syncSource models.SyncSource, syncID string) *models.List {
	now := time.Now()
	list := &models.List{
		ID:            id,
		Type:          models.ListTypeGeneral,
		Name:          "Test List",
		Description:   "Test Description",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    syncStatus,
		SyncSource:    syncSource,
		SyncID:        syncID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Add sync config if needed
	if syncStatus != models.ListSyncStatusNone || syncSource != models.SyncSourceNone || syncID != "" {
		list.SyncConfig = &models.SyncConfig{
			Source: syncSource,
			ID:     syncID,
			Status: syncStatus,
		}
		if syncStatus == models.ListSyncStatusSynced {
			syncTime := now.Add(-time.Hour)
			list.LastSyncAt = &syncTime
			list.SyncConfig.LastSyncAt = &syncTime
		}
	}

	return list
}

func TestSyncService_ConfigureSync(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()

	tests := []struct {
		name     string
		source   models.SyncSource
		syncID   string
		setup    func()
		wantErr  bool
		errCheck func(err error) bool
	}{
		{
			name:   "valid google maps sync",
			source: models.SyncSourceGoogleMaps,
			syncID: "place123",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceNone, ""), nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusPending &&
						l.SyncSource == models.SyncSourceGoogleMaps &&
						l.SyncID == "place123"
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "invalid source",
			source: models.SyncSource("invalid"),
			syncID: "place123",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceNone, ""), nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "invalid sync source")
			},
		},
		{
			name:   "empty sync ID for google maps",
			source: models.SyncSourceGoogleMaps,
			syncID: "",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceNone, ""), nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "Google Maps sync requires a valid place ID")
			},
		},
		{
			name:   "list not found",
			source: models.SyncSourceGoogleMaps,
			syncID: "place123",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrNotFound)
			},
		},
		{
			name:   "sync already enabled",
			source: models.SyncSourceGoogleMaps,
			syncID: "place123",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusSynced, models.SyncSourceGoogleMaps, "place123"), nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrSyncAlreadyEnabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.ConfigureSync(ctx, listID, tt.source, tt.syncID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err), "Expected error matching condition, got %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_DisableSync(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()

	tests := []struct {
		name     string
		list     *models.List
		setup    func()
		wantErr  bool
		errCheck func(err error) bool
	}{
		{
			name: "disable synced list",
			list: &models.List{
				ID:            listID,
				Type:          models.ListTypeGoogleMap,
				Name:          "Test List",
				Description:   "Test Description",
				Visibility:    models.VisibilityPrivate,
				DefaultWeight: 1.0,
				SyncStatus:    models.ListSyncStatusSynced,
				SyncSource:    models.SyncSourceGoogleMaps,
				SyncID:        "place123",
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:            listID,
					Type:          models.ListTypeGoogleMap,
					Name:          "Test List",
					Description:   "Test Description",
					Visibility:    models.VisibilityPrivate,
					DefaultWeight: 1.0,
					SyncStatus:    models.ListSyncStatusSynced,
					SyncSource:    models.SyncSourceGoogleMaps,
					SyncID:        "place123",
				}, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusNone &&
						l.SyncSource == models.SyncSourceNone &&
						l.SyncID == ""
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "disable pending list",
			list: &models.List{
				ID:            listID,
				Type:          models.ListTypeGoogleMap,
				Name:          "Test List",
				Description:   "Test Description",
				Visibility:    models.VisibilityPrivate,
				DefaultWeight: 1.0,
				SyncStatus:    models.ListSyncStatusPending,
				SyncSource:    models.SyncSourceGoogleMaps,
				SyncID:        "place123",
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:            listID,
					Type:          models.ListTypeGoogleMap,
					Name:          "Test List",
					Description:   "Test Description",
					Visibility:    models.VisibilityPrivate,
					DefaultWeight: 1.0,
					SyncStatus:    models.ListSyncStatusPending,
					SyncSource:    models.SyncSourceGoogleMaps,
					SyncID:        "place123",
				}, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusNone &&
						l.SyncSource == models.SyncSourceNone &&
						l.SyncID == ""
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "list not found",
			list: &models.List{
				ID:            listID,
				Type:          models.ListTypeGoogleMap,
				Name:          "Test List",
				Description:   "Test Description",
				Visibility:    models.VisibilityPrivate,
				DefaultWeight: 1.0,
				SyncStatus:    models.ListSyncStatusSynced,
				SyncSource:    models.SyncSourceGoogleMaps,
				SyncID:        "place123",
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrNotFound)
			},
		},
		{
			name: "already disabled",
			list: &models.List{
				ID:            listID,
				Type:          models.ListTypeGoogleMap,
				Name:          "Test List",
				Description:   "Test Description",
				Visibility:    models.VisibilityPrivate,
				DefaultWeight: 1.0,
				SyncStatus:    models.ListSyncStatusNone,
				SyncSource:    models.SyncSourceGoogleMaps,
				SyncID:        "place123",
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:            listID,
					Type:          models.ListTypeGoogleMap,
					Name:          "Test List",
					Description:   "Test Description",
					Visibility:    models.VisibilityPrivate,
					DefaultWeight: 1.0,
					SyncStatus:    models.ListSyncStatusNone,
					SyncSource:    models.SyncSourceGoogleMaps,
					SyncID:        "place123",
				}, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrSyncDisabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.DisableSync(ctx, listID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_UpdateSyncStatus(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()

	tests := []struct {
		name      string
		list      *models.List
		newStatus models.ListSyncStatus
		action    string
		setup     func()
		wantErr   bool
		errCheck  func(err error) bool
	}{
		{
			name:      "pending to synced",
			list:      createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"),
			newStatus: models.ListSyncStatusSynced,
			action:    "sync_complete",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusSynced &&
						l.LastSyncAt != nil
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "synced to conflict",
			list:      createValidTestList(listID, models.ListSyncStatusSynced, models.SyncSourceGoogleMaps, "place123"),
			newStatus: models.ListSyncStatusConflict,
			action:    "remote_change_conflict",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusSynced, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusConflict
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "invalid transition",
			list:      createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceGoogleMaps, "place123"),
			newStatus: models.ListSyncStatusSynced,
			action:    "invalid_action",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceGoogleMaps, "place123"), nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidSyncTransition)
			},
		},
		{
			name:      "sync disabled",
			list:      createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceNone, ""),
			newStatus: models.ListSyncStatusPending,
			action:    "configure_sync",
			setup: func() {
				// Important: Return a list with empty source string to trigger the specific error
				list := createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceNone, "")
				list.SyncSource = "" // Empty source string to trigger the ErrSyncDisabled error
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "cannot update sync status")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.UpdateSyncStatus(ctx, listID, tt.newStatus, tt.action)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_CreateConflict(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()
	itemID := uuid.New()
	conflictID := uuid.New()

	tests := []struct {
		name     string
		conflict *models.SyncConflict
		setup    func()
		wantErr  bool
		errCheck func(err error) bool
	}{
		{
			name: "create valid conflict",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				ItemID:     &itemID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				// First the service should get the list for status update
				// Using pending status since it can transition to conflict with "conflict_detected"
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()

				// Then it should update the list's sync status to conflict
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID && l.SyncStatus == models.ListSyncStatusConflict
				})).Return(nil).Once()

				// Finally it should create the conflict
				mockRepo.On("CreateConflict", mock.MatchedBy(func(c *models.SyncConflict) bool {
					return c.ID == conflictID && c.ListID == listID
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "invalid conflict data",
			conflict: &models.SyncConflict{
				ID:     conflictID,
				ListID: listID,
				Type:   "item_update",
				// Missing required data (LocalData and RemoteData)
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			setup: func() {
				// The Validate call in CreateConflict should fail first
				// No repository calls expected
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "invalid") &&
					errors.Is(err, models.ErrInvalidSyncConfig)
			},
		},
		{
			name: "list not found",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				// The service should try to get the list but it's not found
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "list not found")
			},
		},
		{
			name: "sync disabled",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				// Return a list with sync disabled
				list := createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceNone, "")
				list.SyncSource = "" // This triggers ErrSyncDisabled in UpdateSyncStatus
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "cannot update sync status") &&
					strings.Contains(err.Error(), "disabled")
			},
		},
		{
			name: "invalid transition",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				// Return a list with a state that can't transition to conflict with "conflict_detected"
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusNone, models.SyncSourceGoogleMaps, "place123"), nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "invalid transition") &&
					strings.Contains(err.Error(), "conflict_detected")
			},
		},
		{
			name: "update status error",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				// Return a valid list that can transition to conflict
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()

				// But update fails
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(errors.New("database error")).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "failed to update sync status")
			},
		},
		{
			name: "create conflict error",
			conflict: &models.SyncConflict{
				ID:         conflictID,
				ListID:     listID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				// Return a valid list that can transition to conflict
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()

				// Update succeeds
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(nil).Once()

				// But create conflict fails
				mockRepo.On("CreateConflict", mock.AnythingOfType("*models.SyncConflict")).Return(errors.New("database error")).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "failed to create conflict")
			},
		},
		{
			name:     "nil conflict",
			conflict: nil,
			setup: func() {
				// No mocks needed, the validation should fail immediately
			},
			wantErr: true,
			errCheck: func(err error) bool {
				// For nil conflict reference error
				return err != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock expectations
			if tt.setup != nil {
				tt.setup()
			}

			// Call the service method
			var err error
			if tt.conflict != nil {
				err = service.CreateConflict(ctx, tt.conflict)
			} else {
				// Handle the nil conflict case separately to avoid panic
				err = func() (testErr error) {
					defer func() {
						if r := recover(); r != nil {
							testErr = fmt.Errorf("panic occurred: %v", r)
						}
					}()
					return service.CreateConflict(ctx, nil)
				}()
			}

			// Verify expectations
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err), "Expected error matching condition, got %v", err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_ResolveConflict(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()
	itemID := uuid.New()
	conflictID := uuid.New()

	now := time.Now()

	createValidConflict := func(resolved bool) *models.SyncConflict {
		conflict := &models.SyncConflict{
			ID:         conflictID,
			ListID:     listID,
			ItemID:     &itemID,
			Type:       "item_update",
			LocalData:  "Local Name",
			RemoteData: "Remote Name",
			CreatedAt:  now.Add(-time.Hour),
			UpdatedAt:  now,
		}
		if resolved {
			resolvedTime := now.Add(-30 * time.Minute)
			conflict.ResolvedAt = &resolvedTime
		}
		return conflict
	}

	tests := []struct {
		name       string
		conflictID uuid.UUID
		resolution string
		setup      func()
		wantErr    bool
		errCheck   func(err error) bool
	}{
		{
			name:       "resolve valid conflict",
			conflictID: conflictID,
			resolution: "local",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{createValidConflict(false)}, nil).Once()
				mockRepo.On("ResolveConflict", conflictID).Return(nil).Once()
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusConflict, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusPending
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:       "concurrent modification",
			conflictID: conflictID,
			resolution: "local",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{createValidConflict(false)}, nil).Once()
				mockRepo.On("ResolveConflict", conflictID).Return(models.ErrConcurrentModification).Once()
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{createValidConflict(false)}, nil).Once()
				mockRepo.On("ResolveConflict", conflictID).Return(nil).Once()
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusConflict, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusPending
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:       "conflict not found",
			conflictID: conflictID,
			resolution: "local",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{}, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrNotFound)
			},
		},
		{
			name:       "already resolved",
			conflictID: conflictID,
			resolution: "local",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{createValidConflict(true)}, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrConflictAlreadyResolved)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.ResolveConflict(ctx, tt.conflictID, tt.resolution)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_GetConflicts(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()
	now := time.Now()
	conflicts := []*models.SyncConflict{
		{
			ID:         uuid.New(),
			ListID:     listID,
			Type:       "item_update",
			LocalData:  "Local Name",
			RemoteData: "Remote Name",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	tests := []struct {
		name     string
		setup    func()
		wantErr  bool
		errCheck func(err error) bool
	}{
		{
			name: "get conflicts success",
			setup: func() {
				mockRepo.On("GetConflicts", listID).Return(conflicts, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.On("GetConflicts", listID).Return(nil, fmt.Errorf("database error")).Once()
			},
			wantErr: true,
		},
		{
			name: "external source unavailable",
			setup: func() {
				mockRepo.On("GetConflicts", listID).Return(nil, models.ErrExternalSourceUnavailable).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrExternalSourceUnavailable)
			},
		},
		{
			name: "external source timeout",
			setup: func() {
				mockRepo.On("GetConflicts", listID).Return(nil, models.ErrExternalSourceTimeout).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrExternalSourceTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := service.GetConflicts(ctx, listID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, conflicts, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_UpdateSyncStatus_ExternalErrors(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()

	tests := []struct {
		name      string
		newStatus models.ListSyncStatus
		action    string
		setup     func()
		wantErr   bool
		errCheck  func(err error) bool
	}{
		{
			name:      "external source unavailable",
			newStatus: models.ListSyncStatusSynced,
			action:    "sync_complete",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(models.ErrExternalSourceUnavailable).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrExternalSourceUnavailable)
			},
		},
		{
			name:      "external source timeout",
			newStatus: models.ListSyncStatusSynced,
			action:    "sync_complete",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(models.ErrExternalSourceTimeout).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrExternalSourceTimeout)
			},
		},
		{
			name:      "external source error",
			newStatus: models.ListSyncStatusSynced,
			action:    "sync_complete",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(models.ErrExternalSourceError).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrExternalSourceError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.UpdateSyncStatus(ctx, listID, tt.newStatus, tt.action)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_ConcurrentOperations(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()
	conflictID := uuid.New()

	now := time.Now()

	tests := []struct {
		name    string
		setup   func()
		action  func() error
		wantErr bool
		check   func(err error) bool
	}{
		{
			name: "concurrent sync status updates",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusPending, models.SyncSourceGoogleMaps, "place123"), nil).Once()
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(models.ErrConcurrentModification).Once()
				mockRepo.On("GetByID", listID).Return(createValidTestList(listID, models.ListSyncStatusSynced, models.SyncSourceGoogleMaps, "place123"), nil).Once()
			},
			action: func() error {
				return service.UpdateSyncStatus(ctx, listID, models.ListSyncStatusSynced, "sync_complete")
			},
			wantErr: true,
			check: func(err error) bool {
				return errors.Is(err, models.ErrInvalidSyncTransition)
			},
		},
		{
			name: "concurrent conflict resolution",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{
					{
						ID:         conflictID,
						ListID:     listID,
						Type:       "item_update",
						LocalData:  "Local Data",
						RemoteData: "Remote Data",
						CreatedAt:  now,
						UpdatedAt:  now,
					},
				}, nil).Once()
				mockRepo.On("ResolveConflict", conflictID).Return(models.ErrConcurrentModification).Once()
				resolvedAt := now
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{
					{
						ID:         conflictID,
						ListID:     listID,
						Type:       "item_update",
						LocalData:  "Local Data",
						RemoteData: "Remote Data",
						CreatedAt:  now,
						UpdatedAt:  now,
						ResolvedAt: &resolvedAt,
					},
				}, nil).Once()
			},
			action: func() error {
				return service.ResolveConflict(ctx, conflictID, "local")
			},
			wantErr: true,
			check: func(err error) bool {
				return errors.Is(err, models.ErrConflictAlreadyResolved)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := tt.action()
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr && tt.check != nil {
				assert.True(t, tt.check(err))
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncService_InvalidStateTransitions(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()

	tests := []struct {
		name          string
		currentStatus models.ListSyncStatus
		newStatus     models.ListSyncStatus
		action        string
		expectedError string
	}{
		{
			name:          "none to synced",
			currentStatus: models.ListSyncStatusNone,
			newStatus:     models.ListSyncStatusSynced,
			action:        "sync_complete",
			expectedError: "invalid transition from 'none' to 'synced' with action 'sync_complete'",
		},
		{
			name:          "conflict to synced",
			currentStatus: models.ListSyncStatusConflict,
			newStatus:     models.ListSyncStatusSynced,
			action:        "sync_complete",
			expectedError: "invalid transition from 'conflict' to 'synced' with action 'sync_complete'",
		},
		{
			name:          "synced to pending",
			currentStatus: models.ListSyncStatusSynced,
			newStatus:     models.ListSyncStatusPending,
			action:        "retry_sync",
			expectedError: "invalid transition from 'synced' to 'pending' with action 'retry_sync'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := &models.List{
				ID:         listID,
				SyncStatus: tt.currentStatus,
				SyncSource: models.SyncSourceGoogleMaps,
				SyncID:     "place123",
			}
			mockRepo.On("GetByID", listID).Return(list, nil).Once()

			err := service.UpdateSyncStatus(ctx, listID, tt.newStatus, tt.action)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, models.ErrInvalidSyncTransition))
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// setupSyncTest initializes a test environment for sync service tests
func setupSyncTest(t *testing.T) (*SyncService, *testutil.MockListRepository) {
	mockListRepo := new(testutil.MockListRepository)
	syncService := NewSyncService(mockListRepo)
	require.NotNil(t, syncService)
	return syncService, mockListRepo
}

func TestUpdateSyncStatus(t *testing.T) {
	listID := uuid.New()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		setupMocks  func(*testutil.MockListRepository)
		listID      uuid.UUID
		newStatus   models.ListSyncStatus
		action      string
		expectError bool
		errorType   error
	}{
		{
			name: "successful update to synced",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				mockRepo.On("Update", mock.MatchedBy(func(list *models.List) bool {
					return list.ID == listID &&
						list.SyncStatus == models.ListSyncStatusSynced &&
						list.SyncConfig.Status == models.ListSyncStatusSynced &&
						list.LastSyncAt != nil
				})).Return(nil).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: false,
		},
		{
			name: "successful update to conflict",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				mockRepo.On("Update", mock.MatchedBy(func(list *models.List) bool {
					return list.ID == listID &&
						list.SyncStatus == models.ListSyncStatusConflict &&
						list.SyncConfig.Status == models.ListSyncStatusConflict
				})).Return(nil).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusConflict,
			action:      "conflict_detected",
			expectError: false,
		},
		{
			name: "nil sync config initialization",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: nil, // Nil sync config should be initialized
				}, nil).Once()

				mockRepo.On("Update", mock.MatchedBy(func(list *models.List) bool {
					return list.ID == listID &&
						list.SyncStatus == models.ListSyncStatusSynced &&
						list.SyncConfig != nil &&
						list.SyncConfig.Status == models.ListSyncStatusSynced &&
						list.LastSyncAt != nil
				})).Return(nil).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: false,
		},
		{
			name: "concurrent modification retry success",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				// First call to GetByID
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				// First update attempt fails with concurrent modification
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(models.ErrConcurrentModification).Once()

				// Retry GetByID call
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				// Second update attempt succeeds
				mockRepo.On("Update", mock.MatchedBy(func(list *models.List) bool {
					return list.ID == listID &&
						list.SyncStatus == models.ListSyncStatusSynced &&
						list.SyncConfig.Status == models.ListSyncStatusSynced
				})).Return(nil).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: false,
		},
		{
			name: "list not found",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: true,
			errorType:   models.ErrNotFound,
		},
		{
			name: "sync not enabled",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: "", // Empty sync source
					SyncStatus: models.ListSyncStatusNone,
					SyncConfig: nil,
				}, nil).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: true,
			errorType:   models.ErrSyncDisabled,
		},
		{
			name: "invalid transition",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusNone,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusNone,
					},
				}, nil).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "invalid_action", // Invalid action for this transition
			expectError: true,
			errorType:   models.ErrInvalidSyncTransition,
		},
		{
			name: "update error",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(errors.New("database error")).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: true,
		},
		{
			name: "retry get list error",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				// First call to GetByID
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				// First update attempt fails with concurrent modification
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(models.ErrConcurrentModification).Once()

				// Retry GetByID call fails
				mockRepo.On("GetByID", listID).Return(nil, errors.New("database error")).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: true,
		},
		{
			name: "retry update error",
			setupMocks: func(mockRepo *testutil.MockListRepository) {
				// First call to GetByID
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				// First update attempt fails with concurrent modification
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(models.ErrConcurrentModification).Once()

				// Retry GetByID call succeeds
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Type:       models.ListTypeGeneral,
					CreatedAt:  now,
					UpdatedAt:  now,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place-123",
					SyncStatus: models.ListSyncStatusPending,
					SyncConfig: &models.SyncConfig{
						Source: models.SyncSourceGoogleMaps,
						ID:     "place-123",
						Status: models.ListSyncStatusPending,
					},
				}, nil).Once()

				// Second update attempt fails
				mockRepo.On("Update", mock.AnythingOfType("*models.List")).Return(errors.New("database error")).Once()
			},
			listID:      listID,
			newStatus:   models.ListSyncStatusSynced,
			action:      "sync_complete",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncService, mockRepo := setupSyncTest(t)
			tt.setupMocks(mockRepo)

			err := syncService.UpdateSyncStatus(ctx, tt.listID, tt.newStatus, tt.action)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.True(t, errors.Is(err, tt.errorType),
						"Expected error to be of type %v, got %v", tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
