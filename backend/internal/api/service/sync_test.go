package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncService_ConfigureSync(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()
	list := &models.List{
		ID:            listID,
		Type:          models.ListTypeGoogleMap,
		Name:          "Test List",
		Description:   "Test Description",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
	}

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
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
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
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidSyncSource)
			},
		},
		{
			name:   "empty sync ID for google maps",
			source: models.SyncSourceGoogleMaps,
			syncID: "",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrMissingSyncID)
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
				syncedList := &models.List{
					ID:            listID,
					Type:          models.ListTypeGoogleMap,
					Name:          "Test List",
					Description:   "Test Description",
					Visibility:    models.VisibilityPrivate,
					DefaultWeight: 1.0,
					SyncStatus:    models.ListSyncStatusSynced,
					SyncSource:    models.SyncSourceGoogleMaps,
					SyncID:        "place123",
				}
				mockRepo.On("GetByID", listID).Return(syncedList, nil).Once()
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
					assert.True(t, tt.errCheck(err))
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
			name: "pending to synced",
			list: &models.List{
				ID:         listID,
				SyncStatus: models.ListSyncStatusPending,
				SyncSource: models.SyncSourceGoogleMaps,
				SyncID:     "place123",
			},
			newStatus: models.ListSyncStatusSynced,
			action:    "sync_complete",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusPending,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place123",
				}, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusSynced &&
						l.LastSyncAt != nil
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "synced to conflict",
			list: &models.List{
				ID:         listID,
				SyncStatus: models.ListSyncStatusSynced,
				SyncSource: models.SyncSourceGoogleMaps,
				SyncID:     "place123",
			},
			newStatus: models.ListSyncStatusConflict,
			action:    "remote_change_conflict",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusSynced,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place123",
				}, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusConflict
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "invalid transition",
			list: &models.List{
				ID:         listID,
				SyncStatus: models.ListSyncStatusNone,
				SyncSource: models.SyncSourceGoogleMaps,
				SyncID:     "place123",
			},
			newStatus: models.ListSyncStatusSynced,
			action:    "invalid_action",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusNone,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place123",
				}, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidSyncTransition)
			},
		},
		{
			name: "sync disabled",
			list: &models.List{
				ID:         listID,
				SyncStatus: models.ListSyncStatusNone,
				SyncSource: "",
			},
			newStatus: models.ListSyncStatusPending,
			action:    "configure_sync",
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusNone,
					SyncSource: "",
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
				ID:         uuid.New(),
				ListID:     listID,
				ItemID:     &itemID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusSynced,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place123",
				}, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusConflict
				})).Return(nil).Once()
				mockRepo.On("CreateConflict", mock.AnythingOfType("*models.SyncConflict")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "invalid conflict data",
			conflict: &models.SyncConflict{
				ID:        uuid.New(),
				ListID:    listID,
				Type:      "",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			setup:   func() {},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidSyncConfig)
			},
		},
		{
			name: "list not found",
			conflict: &models.SyncConflict{
				ID:         uuid.New(),
				ListID:     listID,
				Type:       "item_update",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.CreateConflict(ctx, tt.conflict)
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

func TestSyncService_ResolveConflict(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewSyncService(mockRepo)
	ctx := context.Background()

	listID := uuid.New()
	conflictID := uuid.New()
	now := time.Now()

	tests := []struct {
		name       string
		resolution string
		setup      func()
		wantErr    bool
		errCheck   func(err error) bool
	}{
		{
			name:       "resolve valid conflict",
			resolution: "use_local",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{
					{
						ID:         conflictID,
						ListID:     listID,
						Type:       "item_update",
						LocalData:  "Local Name",
						RemoteData: "Remote Name",
						CreatedAt:  now,
						UpdatedAt:  now,
					},
				}, nil).Once()
				mockRepo.On("ResolveConflict", conflictID).Return(nil).Once()
				mockRepo.On("GetConflicts", listID).Return([]*models.SyncConflict{}, nil).Once()
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusConflict,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place123",
				}, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID &&
						l.SyncStatus == models.ListSyncStatusPending
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:       "conflict not found",
			resolution: "use_local",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{}, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrConflictNotFound)
			},
		},
		{
			name:       "already resolved",
			resolution: "use_local",
			setup: func() {
				resolvedAt := now
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{
					{
						ID:         conflictID,
						ListID:     listID,
						Type:       "item_update",
						ResolvedAt: &resolvedAt,
					},
				}, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrConflictAlreadyResolved)
			},
		},
		{
			name:       "empty resolution",
			resolution: "",
			setup: func() {
				mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{
					{
						ID:     conflictID,
						ListID: listID,
						Type:   "item_update",
					},
				}, nil).Once()
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, models.ErrInvalidResolution)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.ResolveConflict(ctx, conflictID, tt.resolution)
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
	list := &models.List{
		ID:         listID,
		SyncStatus: models.ListSyncStatusPending,
		SyncSource: models.SyncSourceGoogleMaps,
		SyncID:     "place123",
	}

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
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID
				})).Return(models.ErrExternalSourceUnavailable).Once()
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
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID
				})).Return(models.ErrExternalSourceTimeout).Once()
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
				mockRepo.On("GetByID", listID).Return(list, nil).Once()
				mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
					return l.ID == listID
				})).Return(models.ErrExternalSourceError).Once()
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
			assert.Error(t, err)
			if tt.errCheck != nil {
				assert.True(t, tt.errCheck(err))
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
	list := &models.List{
		ID:         listID,
		SyncStatus: models.ListSyncStatusPending,
		SyncSource: models.SyncSourceGoogleMaps,
		SyncID:     "place123",
	}

	t.Run("concurrent sync status updates", func(t *testing.T) {
		// First call setup
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("Update", mock.MatchedBy(func(l *models.List) bool {
			return l.ID == listID && l.SyncStatus == models.ListSyncStatusSynced
		})).Return(models.ErrConcurrentModification).Once()

		// Second call setup (after retry)
		updatedList := &models.List{
			ID:         listID,
			SyncStatus: models.ListSyncStatusConflict, // Changed by another operation
			SyncSource: models.SyncSourceGoogleMaps,
			SyncID:     "place123",
		}
		mockRepo.On("GetByID", listID).Return(updatedList, nil).Once()

		err := service.UpdateSyncStatus(ctx, listID, models.ListSyncStatusSynced, "sync_complete")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrInvalidSyncTransition))
	})

	t.Run("concurrent conflict resolution", func(t *testing.T) {
		conflictID := uuid.New()
		conflict := &models.SyncConflict{
			ID:         conflictID,
			ListID:     listID,
			Type:       "item_update",
			LocalData:  "Local Name",
			RemoteData: "Remote Name",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// First attempt
		mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{conflict}, nil).Once()
		mockRepo.On("ResolveConflict", conflictID).Return(models.ErrConcurrentModification).Once()

		// Second attempt - conflict already resolved
		resolvedConflict := *conflict
		resolvedAt := time.Now()
		resolvedConflict.ResolvedAt = &resolvedAt
		mockRepo.On("GetConflicts", conflictID).Return([]*models.SyncConflict{&resolvedConflict}, nil).Once()

		err := service.ResolveConflict(ctx, conflictID, "use_local")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrConflictAlreadyResolved))
	})
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
