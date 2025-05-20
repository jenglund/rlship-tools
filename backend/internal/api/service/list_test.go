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

// TestListService_BasicCRUD tests the basic CRUD operations of the list service
func TestListService_BasicCRUD(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	list := &models.List{
		ID:            uuid.New(),
		Type:          models.ListTypeLocation,
		Name:          "Test List",
		Description:   "Test Description",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		LastSyncAt:    nil,
	}

	t.Run("Create", func(t *testing.T) {
		mockRepo.On("Create", list).Return(nil).Once()
		err := service.CreateList(list)
		assert.NoError(t, err)
	})

	t.Run("Create Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("database error")
		mockRepo.On("Create", list).Return(expectedErr).Once()
		err := service.CreateList(list)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Get", func(t *testing.T) {
		mockRepo.On("GetByID", list.ID).Return(list, nil).Once()
		retrieved, err := service.GetList(list.ID)
		assert.NoError(t, err)
		assert.Equal(t, list, retrieved)
	})

	t.Run("Get Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", list.ID).Return(nil, models.ErrNotFound).Once()
		retrieved, err := service.GetList(list.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrNotFound))
		assert.Nil(t, retrieved)
	})

	t.Run("Update", func(t *testing.T) {
		mockRepo.On("Update", list).Return(nil).Once()
		err := service.UpdateList(list)
		assert.NoError(t, err)
	})

	t.Run("Update Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("database error")
		mockRepo.On("Update", list).Return(expectedErr).Once()
		err := service.UpdateList(list)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Delete", func(t *testing.T) {
		mockRepo.On("Delete", list.ID).Return(nil).Once()
		err := service.DeleteList(list.ID)
		assert.NoError(t, err)
	})
}

func TestListService_ItemManagement(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	itemID := uuid.New()
	item := &models.ListItem{
		ID:     itemID,
		ListID: listID,
		Name:   "Test Item",
		Weight: 1.0,
	}

	t.Run("Add Item", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(&models.List{ID: listID}, nil).Once()
		mockRepo.On("AddItem", item).Return(nil).Once()
		err := service.AddListItem(item)
		assert.NoError(t, err)
	})

	t.Run("Update Item", func(t *testing.T) {
		mockRepo.On("UpdateItem", item).Return(nil).Once()
		err := service.UpdateListItem(item)
		assert.NoError(t, err)
	})

	t.Run("Remove Item", func(t *testing.T) {
		mockRepo.On("RemoveItem", listID, itemID).Return(nil).Once()
		err := service.RemoveListItem(listID, itemID)
		assert.NoError(t, err)
	})

	t.Run("Get Items", func(t *testing.T) {
		items := []*models.ListItem{item}
		mockRepo.On("GetItems", listID).Return(items, nil).Once()
		retrieved, err := service.GetListItems(listID)
		assert.NoError(t, err)
		assert.Equal(t, items, retrieved)
	})
}

func TestListService_MenuGeneration(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	params := &models.MenuParams{
		ListIDs: []uuid.UUID{listID},
		Filters: map[string]interface{}{
			"min_weight": 0.5,
		},
	}

	items := []*models.ListItem{
		{
			ID:     uuid.New(),
			ListID: listID,
			Name:   "Test Item",
			Weight: 1.0,
		},
	}

	t.Run("Generate Menu", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(&models.List{ID: listID}, nil).Once()
		mockRepo.On("GetEligibleItems", params.ListIDs, params.Filters).Return(items, nil).Once()
		lists, err := service.GenerateMenu(params)
		assert.NoError(t, err)
		assert.Len(t, lists, 1)
		assert.Equal(t, items, lists[0].Items)
	})
}

func TestListService_Sync(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	list := &models.List{
		ID:         listID,
		Type:       models.ListTypeGoogleMap,
		SyncSource: models.SyncSourceGoogleMaps,
	}

	t.Run("Sync List", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil).Once()
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil).Once()
		err := service.SyncList(listID)
		assert.NoError(t, err)
	})
}

func TestListService_Conflicts(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	conflictID := uuid.New()
	itemID := uuid.New()
	now := time.Now()

	conflict := &models.SyncConflict{
		ID:         conflictID,
		ListID:     listID,
		ItemID:     &itemID,
		Type:       "name_conflict",
		LocalData:  "Local Name",
		RemoteData: "Remote Name",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	t.Run("Get Conflicts", func(t *testing.T) {
		conflicts := []*models.SyncConflict{conflict}
		mockRepo.On("GetConflicts", listID).Return(conflicts, nil).Once()
		retrieved, err := service.GetListConflicts(listID)
		assert.NoError(t, err)
		assert.Equal(t, conflicts, retrieved)
	})

	t.Run("Resolve Conflict", func(t *testing.T) {
		mockRepo.On("ResolveConflict", conflictID).Return(nil).Once()
		err := service.ResolveListConflict(listID, conflictID, "accept_local")
		assert.NoError(t, err)
	})
}

func TestListService_Ownership(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	ownerID := uuid.New()
	tribeID := uuid.New()

	t.Run("Add Owner", func(t *testing.T) {
		mockRepo.On("AddOwner", mock.MatchedBy(func(owner *models.ListOwner) bool {
			return owner.ListID == listID && owner.OwnerID == ownerID
		})).Return(nil).Once()
		err := service.AddListOwner(listID, ownerID, string(models.OwnerTypeUser))
		assert.NoError(t, err)
	})

	t.Run("Remove Owner", func(t *testing.T) {
		mockRepo.On("RemoveOwner", listID, ownerID).Return(nil).Once()
		err := service.RemoveListOwner(listID, ownerID)
		assert.NoError(t, err)
	})

	t.Run("Get Owners", func(t *testing.T) {
		owners := []*models.ListOwner{{ListID: listID, OwnerID: ownerID}}
		mockRepo.On("GetOwners", listID).Return(owners, nil).Once()
		retrieved, err := service.GetListOwners(listID)
		assert.NoError(t, err)
		assert.Equal(t, owners, retrieved)
	})

	t.Run("Get User Lists", func(t *testing.T) {
		lists := []*models.List{{ID: listID}}
		mockRepo.On("GetUserLists", ownerID).Return(lists, nil).Once()
		retrieved, err := service.GetUserLists(ownerID)
		assert.NoError(t, err)
		assert.Equal(t, lists, retrieved)
	})

	t.Run("Get Tribe Lists", func(t *testing.T) {
		lists := []*models.List{{ID: listID}}
		mockRepo.On("GetTribeLists", tribeID).Return(lists, nil).Once()
		retrieved, err := service.GetTribeLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, lists, retrieved)
	})

	t.Run("Get Tribe Lists Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("database error")
		mockRepo.On("GetTribeLists", tribeID).Return(nil, expectedErr).Once()
		retrieved, err := service.GetTribeLists(tribeID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, retrieved)
	})
}

func TestListService_Sharing(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	tribeID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)

	t.Run("Share With Tribe", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(&models.List{ID: listID}, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{{OwnerID: userID, OwnerType: models.OwnerTypeUser}}, nil).Once()
		mockRepo.On("ShareWithTribe", mock.MatchedBy(func(share *models.ListShare) bool {
			return share.ListID == listID && share.TribeID == tribeID && share.UserID == userID
		})).Return(nil).Once()
		err := service.ShareListWithTribe(listID, tribeID, userID, &expiresAt)
		assert.NoError(t, err)
	})

	t.Run("Unshare With Tribe", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(&models.List{ID: listID}, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{{OwnerID: userID, OwnerType: models.OwnerTypeUser}}, nil).Once()
		mockRepo.On("UnshareWithTribe", listID, tribeID).Return(nil).Once()
		err := service.UnshareListWithTribe(listID, tribeID, userID)
		assert.NoError(t, err)
	})

	t.Run("Get Shared Lists", func(t *testing.T) {
		lists := []*models.List{{ID: listID}}
		mockRepo.On("GetSharedLists", tribeID).Return(lists, nil).Once()
		retrieved, err := service.GetSharedLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, lists, retrieved)
	})
}

func TestCreateList(t *testing.T) {
	// ... rest of the test cases ...
}

func TestListService_Delete(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()

	t.Run("Delete Success", func(t *testing.T) {
		mockRepo.On("Delete", listID).Return(nil).Once()
		err := service.DeleteList(listID)
		assert.NoError(t, err)
	})

	t.Run("Delete Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("database error")
		mockRepo.On("Delete", listID).Return(expectedErr).Once()
		err := service.DeleteList(listID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error deleting list")
		assert.Contains(t, err.Error(), expectedErr.Error())
	})

	t.Run("Delete With Nil ID", func(t *testing.T) {
		err := service.DeleteList(uuid.Nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list ID is required")
		assert.ErrorIs(t, errors.Unwrap(err), models.ErrInvalidInput)
	})
}

func TestListService_List(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	t.Run("List Success", func(t *testing.T) {
		offset, limit := 0, 10
		lists := []*models.List{
			{
				ID:   uuid.New(),
				Name: "List 1",
			},
			{
				ID:   uuid.New(),
				Name: "List 2",
			},
		}

		mockRepo.On("List", offset, limit).Return(lists, nil).Once()
		result, err := service.List(offset, limit)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
	})

	t.Run("List Error", func(t *testing.T) {
		offset, limit := 0, 10
		expectedErr := fmt.Errorf("database error")
		mockRepo.On("List", offset, limit).Return(nil, expectedErr).Once()
		result, err := service.List(offset, limit)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "error listing lists")
		assert.Contains(t, err.Error(), expectedErr.Error())
	})

	t.Run("List With Negative Offset", func(t *testing.T) {
		offset, limit := -1, 10
		result, err := service.List(offset, limit)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "offset cannot be negative")
		assert.ErrorIs(t, errors.Unwrap(err), models.ErrInvalidInput)
	})

	t.Run("List With Zero Limit", func(t *testing.T) {
		offset, limit := 0, 0
		result, err := service.List(offset, limit)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "limit must be positive")
		assert.ErrorIs(t, errors.Unwrap(err), models.ErrInvalidInput)
	})

	t.Run("List With Negative Limit", func(t *testing.T) {
		offset, limit := 0, -10
		result, err := service.List(offset, limit)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "limit must be positive")
		assert.ErrorIs(t, errors.Unwrap(err), models.ErrInvalidInput)
	})
}

func TestListService_GetListShares(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	tribeID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	expiry := now.Add(7 * 24 * time.Hour)

	t.Run("Get List Shares Success", func(t *testing.T) {
		shares := []*models.ListShare{
			{
				ListID:    listID,
				TribeID:   tribeID,
				UserID:    userID,
				CreatedAt: now,
				UpdatedAt: now,
				ExpiresAt: &expiry,
			},
		}

		// First verify list exists
		mockRepo.On("GetByID", listID).Return(&models.List{ID: listID}, nil).Once()
		// Then get list shares
		mockRepo.On("GetListShares", listID).Return(shares, nil).Once()

		result, err := service.GetListShares(listID)
		assert.NoError(t, err)
		assert.Equal(t, shares, result)
	})

	t.Run("List Not Found", func(t *testing.T) {
		expectedErr := models.ErrNotFound
		mockRepo.On("GetByID", listID).Return(nil, expectedErr).Once()

		result, err := service.GetListShares(listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Get Shares Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("database error")

		// First verify list exists
		mockRepo.On("GetByID", listID).Return(&models.List{ID: listID}, nil).Once()
		// Then get list shares with error
		mockRepo.On("GetListShares", listID).Return(nil, expectedErr).Once()

		result, err := service.GetListShares(listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}

// MockListRepository is a mock implementation of the ListRepository interface
type MockListRepository struct {
	mock.Mock
}

func (m *MockListRepository) Create(list *models.List) error {
	args := m.Called(list)
	return args.Error(0)
}

func (m *MockListRepository) CreateList(list *models.List) error {
	args := m.Called(list)
	return args.Error(0)
}

func (m *MockListRepository) GetByID(id uuid.UUID) (*models.List, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.List), args.Error(1)
}

func (m *MockListRepository) GetBySlug(slug string) (*models.List, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.List), args.Error(1)
}

func (m *MockListRepository) Update(list *models.List) error {
	args := m.Called(list)
	return args.Error(0)
}

func (m *MockListRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockListRepository) List(offset, limit int) ([]*models.List, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) AddItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListRepository) GetItems(listID uuid.UUID) ([]*models.ListItem, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListRepository) UpdateItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListRepository) RemoveItem(listID, itemID uuid.UUID) error {
	args := m.Called(listID, itemID)
	return args.Error(0)
}

func (m *MockListRepository) UpdateSyncStatus(listID uuid.UUID, status models.ListSyncStatus) error {
	args := m.Called(listID, status)
	return args.Error(0)
}

func (m *MockListRepository) GetConflicts(listID uuid.UUID) ([]*models.SyncConflict, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SyncConflict), args.Error(1)
}

func (m *MockListRepository) ResolveConflict(conflictID uuid.UUID) error {
	args := m.Called(conflictID)
	return args.Error(0)
}

func (m *MockListRepository) CreateConflict(conflict *models.SyncConflict) error {
	args := m.Called(conflict)
	return args.Error(0)
}

func (m *MockListRepository) AddOwner(owner *models.ListOwner) error {
	args := m.Called(owner)
	return args.Error(0)
}

func (m *MockListRepository) RemoveOwner(listID, ownerID uuid.UUID) error {
	args := m.Called(listID, ownerID)
	return args.Error(0)
}

func (m *MockListRepository) GetOwners(listID uuid.UUID) ([]*models.ListOwner, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListOwner), args.Error(1)
}

func (m *MockListRepository) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) ShareWithTribe(share *models.ListShare) error {
	args := m.Called(share)
	return args.Error(0)
}

func (m *MockListRepository) UnshareWithTribe(listID, tribeID uuid.UUID) error {
	args := m.Called(listID, tribeID)
	return args.Error(0)
}

func (m *MockListRepository) GetSharedLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetListShares(listID uuid.UUID) ([]*models.ListShare, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListShare), args.Error(1)
}

func (m *MockListRepository) GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
	args := m.Called(listIDs, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListRepository) UpdateItemStats(itemID uuid.UUID, chosen bool) error {
	args := m.Called(itemID, chosen)
	return args.Error(0)
}

func (m *MockListRepository) MarkItemChosen(itemID uuid.UUID) error {
	args := m.Called(itemID)
	return args.Error(0)
}

func (m *MockListRepository) GetListsByOwner(ownerID uuid.UUID, ownerType models.OwnerType) ([]*models.List, error) {
	args := m.Called(ownerID, ownerType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetSharedTribes(listID uuid.UUID) ([]*models.Tribe, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tribe), args.Error(1)
}

func (m *MockListRepository) GetListsBySource(source string) ([]*models.List, error) {
	args := m.Called(source)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetUserListsWithContext(ctx context.Context, userID uuid.UUID) ([]*models.List, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetTribeListsWithContext(ctx context.Context, tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(ctx, tribeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

// Helper function to check if an error is a common error type
func isCommonError(err error) bool {
	return errors.Is(err, models.ErrNotFound) ||
		errors.Is(err, models.ErrForbidden) ||
		errors.Is(err, models.ErrInvalidInput) ||
		errors.Is(err, models.ErrUnauthorized)
}

// Using a simple error variable for testing
var errInternalServer = fmt.Errorf("internal server error")

func TestListService_UnshareListWithTribe(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)

	listID := uuid.New()
	tribeID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		listID        uuid.UUID
		tribeID       uuid.UUID
		userID        uuid.UUID
		setupMocks    func(*MockListRepository)
		expectedError error
	}{
		{
			name:    "success",
			listID:  listID,
			tribeID: tribeID,
			userID:  userID,
			setupMocks: func(mockRepo *MockListRepository) {
				// Mock GetByID to return a valid list
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Visibility: models.VisibilityShared,
				}, nil)

				// Mock GetOwners to return list with user as owner
				mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
					{
						ListID:    listID,
						OwnerID:   userID,
						OwnerType: "user",
					},
				}, nil)

				// Mock UnshareWithTribe to succeed
				mockRepo.On("UnshareWithTribe", listID, tribeID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "list not found",
			listID:  listID,
			tribeID: tribeID,
			userID:  userID,
			setupMocks: func(mockRepo *MockListRepository) {
				// Mock GetByID to return list not found error
				mockRepo.On("GetByID", listID).Return(nil, models.ErrNotFound)
			},
			expectedError: models.ErrNotFound,
		},
		{
			name:    "error getting list",
			listID:  listID,
			tribeID: tribeID,
			userID:  userID,
			setupMocks: func(mockRepo *MockListRepository) {
				// Mock GetByID to return an unexpected error
				mockRepo.On("GetByID", listID).Return(nil, errInternalServer)
			},
			expectedError: errInternalServer,
		},
		{
			name:    "error getting owners",
			listID:  listID,
			tribeID: tribeID,
			userID:  userID,
			setupMocks: func(mockRepo *MockListRepository) {
				// Mock GetByID to return a valid list
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Visibility: models.VisibilityShared,
				}, nil)

				// Mock GetOwners to return an error
				mockRepo.On("GetOwners", listID).Return(nil, errInternalServer)
			},
			expectedError: errInternalServer,
		},
		{
			name:    "user not an owner",
			listID:  listID,
			tribeID: tribeID,
			userID:  userID,
			setupMocks: func(mockRepo *MockListRepository) {
				// Mock GetByID to return a valid list
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Visibility: models.VisibilityShared,
				}, nil)

				// Mock GetOwners to return list with different user as owner
				mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
					{
						ListID:    listID,
						OwnerID:   uuid.New(), // Different user ID
						OwnerType: "user",
					},
				}, nil)
			},
			expectedError: models.ErrForbidden,
		},
		{
			name:    "tribe is owner but not user",
			listID:  listID,
			tribeID: tribeID,
			userID:  userID,
			setupMocks: func(mockRepo *MockListRepository) {
				// Mock GetByID to return a valid list
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Visibility: models.VisibilityShared,
				}, nil)

				// Mock GetOwners to return list with tribe as owner but not user
				mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
					{
						ListID:    listID,
						OwnerID:   tribeID,
						OwnerType: "tribe", // Tribe, not user
					},
				}, nil)
			},
			expectedError: models.ErrForbidden,
		},
		{
			name:    "error unsharing",
			listID:  listID,
			tribeID: tribeID,
			userID:  userID,
			setupMocks: func(mockRepo *MockListRepository) {
				// Mock GetByID to return a valid list
				mockRepo.On("GetByID", listID).Return(&models.List{
					ID:         listID,
					Name:       "Test List",
					Visibility: models.VisibilityShared,
				}, nil)

				// Mock GetOwners to return list with user as owner
				mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
					{
						ListID:    listID,
						OwnerID:   userID,
						OwnerType: "user",
					},
				}, nil)

				// Mock UnshareWithTribe to fail
				mockRepo.On("UnshareWithTribe", listID, tribeID).Return(errInternalServer)
			},
			expectedError: errInternalServer,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock expectations
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil

			// Setup mocks
			tc.setupMocks(mockRepo)

			// Call the service method
			err := service.UnshareListWithTribe(tc.listID, tc.tribeID, tc.userID)

			// Check error
			if tc.expectedError != nil {
				assert.Error(t, err)
				// For common errors like ErrNotFound, check exact error
				if isCommonError(tc.expectedError) {
					assert.ErrorIs(t, err, tc.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all mocks were called as expected
			mockRepo.AssertExpectations(t)
		})
	}
}
