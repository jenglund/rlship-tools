package service

import (
	"context"
	"errors"
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
		ID:         listID,
		Type:       models.ListTypeGoogleMap,
		Name:       "Test List",
		SyncStatus: models.ListSyncStatusNone,
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
					ID:         listID,
					Type:       models.ListTypeGoogleMap,
					Name:       "Test List",
					SyncStatus: models.ListSyncStatusSynced,
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
				ID:         listID,
				SyncStatus: models.ListSyncStatusSynced,
				SyncSource: models.SyncSourceGoogleMaps,
				SyncID:     "place123",
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
				ID:         listID,
				SyncStatus: models.ListSyncStatusPending,
				SyncSource: models.SyncSourceGoogleMaps,
				SyncID:     "place123",
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusPending,
					SyncSource: models.SyncSourceGoogleMaps,
					SyncID:     "place123",
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
				ID:         listID,
				SyncStatus: models.ListSyncStatusSynced,
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
				ID:         listID,
				SyncStatus: models.ListSyncStatusNone,
			},
			setup: func() {
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					SyncStatus: models.ListSyncStatusNone,
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
