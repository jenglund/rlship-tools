package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockListRepository is a mock implementation of models.ListRepository
type MockListRepository struct {
	mock.Mock
}

func (m *MockListRepository) Create(list *models.List) error {
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
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) AddItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListRepository) UpdateItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListRepository) RemoveItem(listID, itemID uuid.UUID) error {
	args := m.Called(listID, itemID)
	return args.Error(0)
}

func (m *MockListRepository) GetItems(listID uuid.UUID) ([]*models.ListItem, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListRepository) GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
	args := m.Called(listIDs, filters)
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListRepository) UpdateItemStats(itemID uuid.UUID, chosen bool) error {
	args := m.Called(itemID, chosen)
	return args.Error(0)
}

func (m *MockListRepository) UpdateSyncStatus(listID uuid.UUID, status models.ListSyncStatus) error {
	args := m.Called(listID, status)
	return args.Error(0)
}

func (m *MockListRepository) GetListsBySource(source string) ([]*models.List, error) {
	args := m.Called(source)
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) CreateConflict(conflict *models.ListConflict) error {
	args := m.Called(conflict)
	return args.Error(0)
}

func (m *MockListRepository) GetConflicts(listID uuid.UUID) ([]*models.ListConflict, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListConflict), args.Error(1)
}

func (m *MockListRepository) ResolveConflict(conflictID uuid.UUID) error {
	args := m.Called(conflictID)
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
	return args.Get(0).([]*models.ListOwner), args.Error(1)
}

func (m *MockListRepository) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListRepository) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
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
	return args.Get(0).([]*models.List), args.Error(1)
}

func TestListService(t *testing.T) {
	t.Run("Basic CRUD Operations", func(t *testing.T) {
		mockRepo := new(MockListRepository)
		service := NewListService(mockRepo)
		defer mockRepo.AssertExpectations(t)

		list := &models.List{
			ID:            uuid.New(),
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
		}

		// Test CreateList - success
		mockRepo.On("Create", list).Return(nil)
		err := service.CreateList(list)
		assert.NoError(t, err)

		// Test CreateList - error
		expectedErr := fmt.Errorf("failed to create list: database error")
		mockRepo.On("Create", list).Return(expectedErr)
		err = service.CreateList(list)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create list")

		// Test GetList - success
		mockRepo.On("GetByID", list.ID).Return(list, nil)
		retrieved, err := service.GetList(list.ID)
		assert.NoError(t, err)
		assert.Equal(t, list, retrieved)

		// Test GetList - not found
		mockRepo.On("GetByID", list.ID).Return(nil, models.ErrNotFound)
		retrieved, err = service.GetList(list.ID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrNotFound))
		assert.Nil(t, retrieved)

		// Test UpdateList - success
		mockRepo.On("Update", list).Return(nil)
		err = service.UpdateList(list)
		assert.NoError(t, err)

		// Test UpdateList - error
		expectedErr = fmt.Errorf("failed to update list: database error")
		mockRepo.On("Update", list).Return(expectedErr)
		err = service.UpdateList(list)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update list")

		// Test DeleteList - success
		mockRepo.On("Delete", list.ID).Return(nil)
		err = service.DeleteList(list.ID)
		assert.NoError(t, err)

		// Test DeleteList - error
		expectedErr = fmt.Errorf("failed to delete list: database error")
		mockRepo.On("Delete", list.ID).Return(expectedErr)
		err = service.DeleteList(list.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete list")
	})

	t.Run("List Items Management", func(t *testing.T) {
		mockRepo := new(MockListRepository)
		service := NewListService(mockRepo)
		defer mockRepo.AssertExpectations(t)

		listID := uuid.New()
		item := &models.ListItem{
			ID:          uuid.New(),
			ListID:      listID,
			Name:        "Test Item",
			Description: "Test Description",
			Weight:      1.5,
		}

		// Test AddListItem
		mockRepo.On("AddItem", item).Return(nil)
		err := service.AddListItem(item)
		assert.NoError(t, err)

		// Test UpdateListItem
		mockRepo.On("UpdateItem", item).Return(nil)
		err = service.UpdateListItem(item)
		assert.NoError(t, err)

		// Test GetListItems
		items := []*models.ListItem{item}
		mockRepo.On("GetItems", listID).Return(items, nil)
		retrieved, err := service.GetListItems(listID)
		assert.NoError(t, err)
		assert.Equal(t, items, retrieved)

		// Test RemoveListItem
		mockRepo.On("RemoveItem", listID, item.ID).Return(nil)
		err = service.RemoveListItem(listID, item.ID)
		assert.NoError(t, err)
	})

	t.Run("Menu Generation", func(t *testing.T) {
		mockRepo := new(MockListRepository)
		service := NewListService(mockRepo)
		defer mockRepo.AssertExpectations(t)

		listID := uuid.New()
		cooldownDays := 7
		list := &models.List{
			ID:            listID,
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
			CooldownDays:  &cooldownDays,
		}

		items := []*models.ListItem{
			{
				ID:     uuid.New(),
				ListID: listID,
				Name:   "Item 1",
				Weight: 2.0,
			},
			{
				ID:     uuid.New(),
				ListID: listID,
				Name:   "Item 2",
				Weight: 1.5,
			},
		}

		params := &models.MenuParams{
			ListIDs: []uuid.UUID{listID},
			Filters: map[string]interface{}{
				"max_items": 1,
			},
		}

		mockRepo.On("GetByID", listID).Return(list, nil)
		mockRepo.On("GetEligibleItems", []uuid.UUID{listID}, params.Filters).Return(items, nil)

		result, err := service.GenerateMenu(params)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, listID, result[0].ID)
		assert.Len(t, result[0].Items, 1)
		assert.Equal(t, "Item 1", result[0].Items[0].Name)
	})

	t.Run("List Sync and Conflicts", func(t *testing.T) {
		mockRepo := new(MockListRepository)
		service := NewListService(mockRepo)
		defer mockRepo.AssertExpectations(t)

		listID := uuid.New()
		now := time.Now()

		// Test conflict management
		conflict := &models.ListConflict{
			BaseModel: models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: now,
				UpdatedAt: now,
			},
			ListID:       listID,
			ConflictType: "modified",
			LocalData: map[string]interface{}{
				"name": "Local Name",
			},
			ExternalData: map[string]interface{}{
				"name": "External Name",
			},
		}

		mockRepo.On("GetConflicts", listID).Return([]*models.ListConflict{conflict}, nil)
		conflicts, err := service.GetListConflicts(listID)
		require.NoError(t, err)
		assert.Len(t, conflicts, 1)
		assert.Equal(t, conflict, conflicts[0])

		mockRepo.On("ResolveConflict", conflict.ID).Return(nil)
		err = service.ResolveListConflict(listID, conflict.ID, "accept")
		require.NoError(t, err)
	})

	t.Run("List Sync", func(t *testing.T) {
		mockRepo := new(MockListRepository)
		service := NewListService(mockRepo)
		defer mockRepo.AssertExpectations(t)

		listID := uuid.New()
		list := &models.List{
			ID:            listID,
			Type:          models.ListTypeGoogleMap,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
			SyncSource:    "google_maps",
		}

		// Test SyncList - success case
		mockRepo.On("GetByID", listID).Return(list, nil)
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil)
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil)

		err := service.SyncList(listID)
		assert.NoError(t, err)

		// Test sync failure - list not found
		notFoundID := uuid.New()
		mockRepo.On("GetByID", notFoundID).Return(nil, models.ErrNotFound)

		err = service.SyncList(notFoundID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrNotFound))

		// Test sync failure - list not configured
		unconfiguredList := &models.List{
			ID:   uuid.New(),
			Type: models.ListTypeLocation,
		}
		mockRepo.On("GetByID", unconfiguredList.ID).Return(unconfiguredList, nil)

		err = service.SyncList(unconfiguredList.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured for syncing")
	})

	t.Run("List Ownership", func(t *testing.T) {
		mockRepo := new(MockListRepository)
		service := NewListService(mockRepo)
		defer mockRepo.AssertExpectations(t)

		listID := uuid.New()
		userID := uuid.New()
		tribeID := uuid.New()

		// Test GetListOwners
		expectedOwners := []*models.ListOwner{
			{OwnerID: userID, OwnerType: models.OwnerTypeUser},
			{OwnerID: tribeID, OwnerType: models.OwnerTypeTribe},
		}
		mockRepo.On("GetOwners", listID).Return(expectedOwners, nil)
		owners, err := service.GetListOwners(listID)
		require.NoError(t, err)
		assert.Equal(t, expectedOwners, owners)

		// Test AddListOwner - success case
		mockRepo.On("AddOwner", expectedOwners[0]).Return(nil)
		err = service.AddListOwner(listID, expectedOwners[0].OwnerID, string(expectedOwners[0].OwnerType))
		require.NoError(t, err)

		// Test AddListOwner - invalid owner type
		err = service.AddListOwner(listID, userID, "invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid owner type")

		// Test RemoveListOwner
		mockRepo.On("RemoveOwner", listID, expectedOwners[0].OwnerID).Return(nil)
		err = service.RemoveListOwner(listID, expectedOwners[0].OwnerID)
		require.NoError(t, err)

		// Test GetUserLists
		lists := []*models.List{
			{ID: listID, Name: "Test List"},
		}
		mockRepo.On("GetUserLists", userID).Return(lists, nil)
		userLists, err := service.GetUserLists(userID)
		require.NoError(t, err)
		assert.Equal(t, lists, userLists)

		// Test GetTribeLists
		mockRepo.On("GetTribeLists", tribeID).Return(lists, nil)
		tribeLists, err := service.GetTribeLists(tribeID)
		require.NoError(t, err)
		assert.Equal(t, lists, tribeLists)
	})

	t.Run("List Sharing", func(t *testing.T) {
		mockRepo := new(MockListRepository)
		service := NewListService(mockRepo)
		defer mockRepo.AssertExpectations(t)

		listID := uuid.New()
		userID := uuid.New()
		tribeID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)

		list := &models.List{
			ID:            listID,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
		}

		share := &models.ListShare{
			ListID:    listID,
			TribeID:   tribeID,
			UserID:    userID,
			ExpiresAt: &expiresAt,
		}

		// Test ShareListWithTribe - success case
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{OwnerID: userID, OwnerType: "user"},
		}, nil).Once()
		mockRepo.On("ShareWithTribe", share).Return(nil).Once()

		err := service.ShareListWithTribe(listID, tribeID, userID, &expiresAt)
		require.NoError(t, err)

		// Test ShareListWithTribe - list not found
		notFoundID := uuid.New()
		mockRepo.On("GetByID", notFoundID).Return(nil, models.ErrNotFound).Once()

		err = service.ShareListWithTribe(notFoundID, tribeID, userID, &expiresAt)
		require.Error(t, err)
		assert.Equal(t, models.ErrNotFound, err)

		// Test ShareListWithTribe - no permission
		otherUserID := uuid.New()
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{OwnerID: otherUserID, OwnerType: "user"},
		}, nil).Once()

		err = service.ShareListWithTribe(listID, tribeID, userID, &expiresAt)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not have permission")

		// Test UnshareListWithTribe - success case
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{OwnerID: userID, OwnerType: "user"},
		}, nil).Once()
		mockRepo.On("UnshareWithTribe", listID, tribeID).Return(nil).Once()

		err = service.UnshareListWithTribe(listID, tribeID, userID)
		require.NoError(t, err)

		// Test GetSharedLists
		sharedLists := []*models.List{list}
		mockRepo.On("GetSharedLists", tribeID).Return(sharedLists, nil).Once()

		retrieved, err := service.GetSharedLists(tribeID)
		require.NoError(t, err)
		assert.Equal(t, sharedLists, retrieved)
	})
}

func TestListService_List(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

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

	// Test List - success case
	mockRepo.On("List", 0, 10).Return(lists, nil)
	result, err := service.List(0, 10)
	require.NoError(t, err)
	assert.Equal(t, lists, result)

	// Test List - error case
	expectedErr := fmt.Errorf("database error")
	mockRepo.On("List", 10, 10).Return(nil, expectedErr)
	result, err = service.List(10, 10)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestListService_GenerateMenu(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	list1 := &models.List{
		ID:            uuid.New(),
		Type:          models.ListTypeLocation,
		Name:          "Test List 1",
		Description:   "Test Description 1",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1,
	}

	list2 := &models.List{
		ID:            uuid.New(),
		Type:          models.ListTypeActivity,
		Name:          "Test List 2",
		Description:   "Test Description 2",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1,
	}

	item1 := &models.ListItem{
		ID:          uuid.New(),
		ListID:      list1.ID,
		Name:        "Test Item 1",
		Description: "Test Description 1",
		Weight:      1,
	}

	item2 := &models.ListItem{
		ID:          uuid.New(),
		ListID:      list2.ID,
		Name:        "Test Item 2",
		Description: "Test Description 2",
		Weight:      1,
	}

	params := &models.MenuParams{
		ListIDs: []uuid.UUID{list1.ID, list2.ID},
		Filters: map[string]interface{}{
			"max_items": 1,
		},
	}

	// Test successful menu generation
	mockRepo.On("GetByID", list1.ID).Return(list1, nil)
	mockRepo.On("GetByID", list2.ID).Return(list2, nil)
	mockRepo.On("GetEligibleItems", []uuid.UUID{list1.ID}, params.Filters).Return([]*models.ListItem{item1}, nil)
	mockRepo.On("GetEligibleItems", []uuid.UUID{list2.ID}, params.Filters).Return([]*models.ListItem{item2}, nil)

	lists, err := service.GenerateMenu(params)
	assert.NoError(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, list1.ID, lists[0].ID)
	assert.Equal(t, list2.ID, lists[1].ID)
	assert.Len(t, lists[0].Items, 1)
	assert.Len(t, lists[1].Items, 1)
	assert.Equal(t, item1.ID, lists[0].Items[0].ID)
	assert.Equal(t, item2.ID, lists[1].Items[0].ID)
}

func TestListService_GenerateMenu_Errors(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	list := &models.List{
		ID:            uuid.New(),
		Type:          models.ListTypeLocation,
		Name:          "Test List",
		Description:   "Test Description",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1,
	}

	params := &models.MenuParams{
		ListIDs: []uuid.UUID{list.ID},
		Filters: map[string]interface{}{
			"max_items": 1,
		},
	}

	// Test error getting list
	mockRepo.On("GetByID", list.ID).Return(nil, fmt.Errorf("database error"))
	lists, err := service.GenerateMenu(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting list")
	assert.Nil(t, lists)

	// Test error getting eligible items
	mockRepo.On("GetByID", list.ID).Return(list, nil)
	mockRepo.On("GetEligibleItems", []uuid.UUID{list.ID}, params.Filters).Return(nil, fmt.Errorf("database error"))
	lists, err = service.GenerateMenu(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting eligible items")
	assert.Nil(t, lists)
}

func TestListService_SyncList(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	list := &models.List{
		ID:            uuid.New(),
		Type:          models.ListTypeGoogleMap,
		Name:          "Test List",
		Description:   "Test Description",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1.0,
		SyncSource:    "google_maps",
	}

	// Test successful sync
	mockRepo.On("GetByID", list.ID).Return(list, nil)
	mockRepo.On("UpdateSyncStatus", list.ID, models.ListSyncStatusPending).Return(nil)
	mockRepo.On("UpdateSyncStatus", list.ID, models.ListSyncStatusSynced).Return(nil)

	err := service.SyncList(list.ID)
	assert.NoError(t, err)

	// Test error getting list
	mockRepo.On("GetByID", list.ID).Return(nil, fmt.Errorf("database error"))
	err = service.SyncList(list.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Test error updating sync status
	mockRepo.On("GetByID", list.ID).Return(list, nil)
	mockRepo.On("UpdateSyncStatus", list.ID, models.ListSyncStatusPending).Return(fmt.Errorf("database error"))
	err = service.SyncList(list.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
