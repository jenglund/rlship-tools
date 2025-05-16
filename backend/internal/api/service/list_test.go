package service

import (
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

// TestListService_BasicCRUD tests the basic CRUD operations of the list service
func TestListService_BasicCRUD(t *testing.T) {
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

	t.Run("Delete Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("database error")
		mockRepo.On("Delete", list.ID).Return(expectedErr).Once()
		err := service.DeleteList(list.ID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

// TestListService_ItemManagement tests the list item management functionality
func TestListService_ItemManagement(t *testing.T) {
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
		Latitude:    testutil.Float64Ptr(37.7749),
		Longitude:   testutil.Float64Ptr(-122.4194),
		Address:     testutil.StringPtr("123 Test St"),
	}

	t.Run("Add Item", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(&models.List{ID: listID}, nil).Once()
		mockRepo.On("AddItem", item).Return(nil).Once()
		err := service.AddListItem(item)
		assert.NoError(t, err)
	})

	t.Run("Add Item with Invalid Location", func(t *testing.T) {
		invalidItem := &models.ListItem{
			ListID:    listID,
			Name:      "Invalid Location",
			Weight:    1.0,
			Latitude:  testutil.Float64Ptr(37.7749),
			Longitude: nil,
			Address:   testutil.StringPtr("123 Test St"),
		}
		err := service.AddListItem(invalidItem)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "location requires latitude, longitude, and address")
	})

	t.Run("Update Item", func(t *testing.T) {
		mockRepo.On("UpdateItem", item).Return(nil).Once()
		err := service.UpdateListItem(item)
		assert.NoError(t, err)
	})

	t.Run("Get Items", func(t *testing.T) {
		items := []*models.ListItem{item}
		mockRepo.On("GetItems", listID).Return(items, nil).Once()
		retrieved, err := service.GetListItems(listID)
		assert.NoError(t, err)
		assert.Equal(t, items, retrieved)
	})

	t.Run("Remove Item", func(t *testing.T) {
		mockRepo.On("RemoveItem", listID, item.ID).Return(nil).Once()
		err := service.RemoveListItem(listID, item.ID)
		assert.NoError(t, err)
	})
}

// TestListService_MenuGeneration tests the menu generation functionality
func TestListService_MenuGeneration(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	list1 := &models.List{
		ID:            uuid.New(),
		Type:          models.ListTypeLocation,
		Name:          "Test List 1",
		Description:   "Test Description 1",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1.0,
	}

	list2 := &models.List{
		ID:            uuid.New(),
		Type:          models.ListTypeActivity,
		Name:          "Test List 2",
		Description:   "Test Description 2",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1.0,
	}

	item1 := &models.ListItem{
		ID:          uuid.New(),
		ListID:      list1.ID,
		Name:        "Test Item 1",
		Description: "Test Description 1",
		Weight:      1.0,
	}

	item2 := &models.ListItem{
		ID:          uuid.New(),
		ListID:      list2.ID,
		Name:        "Test Item 2",
		Description: "Test Description 2",
		Weight:      1.0,
	}

	params := &models.MenuParams{
		ListIDs: []uuid.UUID{list1.ID, list2.ID},
		Filters: map[string]interface{}{
			"max_items": 1,
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", list1.ID).Return(list1, nil).Once()
		mockRepo.On("GetByID", list2.ID).Return(list2, nil).Once()
		mockRepo.On("GetEligibleItems", []uuid.UUID{list1.ID}, params.Filters).Return([]*models.ListItem{item1}, nil).Once()
		mockRepo.On("GetEligibleItems", []uuid.UUID{list2.ID}, params.Filters).Return([]*models.ListItem{item2}, nil).Once()

		lists, err := service.GenerateMenu(params)
		assert.NoError(t, err)
		assert.Len(t, lists, 2)
		assert.Equal(t, list1.ID, lists[0].ID)
		assert.Equal(t, list2.ID, lists[1].ID)
		assert.Len(t, lists[0].Items, 1)
		assert.Len(t, lists[1].Items, 1)
		assert.Equal(t, item1.ID, lists[0].Items[0].ID)
		assert.Equal(t, item2.ID, lists[1].Items[0].ID)
	})

	t.Run("List Not Found", func(t *testing.T) {
		mockRepo.On("GetByID", list1.ID).Return(nil, fmt.Errorf("database error")).Once()
		lists, err := service.GenerateMenu(params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error getting list")
		assert.Nil(t, lists)
	})

	t.Run("Error Getting Items", func(t *testing.T) {
		mockRepo.On("GetByID", list1.ID).Return(list1, nil).Once()
		mockRepo.On("GetEligibleItems", []uuid.UUID{list1.ID}, params.Filters).Return(nil, fmt.Errorf("database error")).Once()
		lists, err := service.GenerateMenu(params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error getting eligible items")
		assert.Nil(t, lists)
	})
}

// TestListService_Sync tests the list synchronization functionality
func TestListService_Sync(t *testing.T) {
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

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", list.ID).Return(list, nil).Once()
		mockRepo.On("UpdateSyncStatus", list.ID, models.ListSyncStatusPending).Return(nil).Once()
		mockRepo.On("UpdateSyncStatus", list.ID, models.ListSyncStatusSynced).Return(nil).Once()

		err := service.SyncList(list.ID)
		assert.NoError(t, err)
	})

	t.Run("List Not Found", func(t *testing.T) {
		notFoundID := uuid.New()
		mockRepo.On("GetByID", notFoundID).Return(nil, models.ErrNotFound).Once()

		err := service.SyncList(notFoundID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrNotFound))
	})

	t.Run("List Not Configured", func(t *testing.T) {
		unconfiguredList := &models.List{
			ID:   uuid.New(),
			Type: models.ListTypeLocation,
		}
		mockRepo.On("GetByID", unconfiguredList.ID).Return(unconfiguredList, nil).Once()

		err := service.SyncList(unconfiguredList.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured for syncing")
	})
}

// TestListService_Conflicts tests the list conflict management functionality
func TestListService_Conflicts(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	now := time.Now()
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

	t.Run("Get Conflicts", func(t *testing.T) {
		mockRepo.On("GetConflicts", listID).Return([]*models.ListConflict{conflict}, nil).Once()
		conflicts, err := service.GetListConflicts(listID)
		assert.NoError(t, err)
		assert.Len(t, conflicts, 1)
		assert.Equal(t, conflict, conflicts[0])
	})

	t.Run("Resolve Conflict", func(t *testing.T) {
		mockRepo.On("ResolveConflict", conflict.ID).Return(nil).Once()
		err := service.ResolveListConflict(listID, conflict.ID, "accept")
		assert.NoError(t, err)
	})
}

// TestListService_Ownership tests the list ownership functionality
func TestListService_Ownership(t *testing.T) {
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	userID := uuid.New()
	tribeID := uuid.New()

	t.Run("Get Owners", func(t *testing.T) {
		expectedOwners := []*models.ListOwner{
			{OwnerID: userID, OwnerType: models.OwnerTypeUser},
			{OwnerID: tribeID, OwnerType: models.OwnerTypeTribe},
		}
		mockRepo.On("GetOwners", listID).Return(expectedOwners, nil).Once()
		owners, err := service.GetListOwners(listID)
		assert.NoError(t, err)
		assert.Equal(t, expectedOwners, owners)
	})

	t.Run("Add Owner Success", func(t *testing.T) {
		owner := &models.ListOwner{
			ListID:    listID,
			OwnerID:   userID,
			OwnerType: models.OwnerTypeUser,
		}
		mockRepo.On("AddOwner", owner).Return(nil).Once()
		err := service.AddListOwner(listID, userID, string(models.OwnerTypeUser))
		assert.NoError(t, err)
	})

	t.Run("Add Owner Invalid Type", func(t *testing.T) {
		err := service.AddListOwner(listID, userID, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid owner type")
	})

	t.Run("Remove Owner", func(t *testing.T) {
		mockRepo.On("RemoveOwner", listID, userID).Return(nil).Once()
		err := service.RemoveListOwner(listID, userID)
		assert.NoError(t, err)
	})

	t.Run("Get User Lists", func(t *testing.T) {
		lists := []*models.List{{ID: listID, Name: "Test List"}}
		mockRepo.On("GetUserLists", userID).Return(lists, nil).Once()
		userLists, err := service.GetUserLists(userID)
		assert.NoError(t, err)
		assert.Equal(t, lists, userLists)
	})

	t.Run("Get Tribe Lists", func(t *testing.T) {
		lists := []*models.List{{ID: listID, Name: "Test List"}}
		mockRepo.On("GetTribeLists", tribeID).Return(lists, nil).Once()
		tribeLists, err := service.GetTribeLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, lists, tribeLists)
	})
}

// TestListService_Sharing tests the list sharing functionality
func TestListService_Sharing(t *testing.T) {
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

	t.Run("Share Success", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{OwnerID: userID, OwnerType: "user"},
		}, nil).Once()
		mockRepo.On("ShareWithTribe", &models.ListShare{
			ListID:    listID,
			TribeID:   tribeID,
			UserID:    userID,
			ExpiresAt: &expiresAt,
		}).Return(nil).Once()

		err := service.ShareListWithTribe(listID, tribeID, userID, &expiresAt)
		assert.NoError(t, err)
	})

	t.Run("Share List Not Found", func(t *testing.T) {
		notFoundID := uuid.New()
		mockRepo.On("GetByID", notFoundID).Return(nil, models.ErrNotFound).Once()

		err := service.ShareListWithTribe(notFoundID, tribeID, userID, &expiresAt)
		assert.Error(t, err)
		assert.Equal(t, models.ErrNotFound, err)
	})

	t.Run("Share No Permission", func(t *testing.T) {
		otherUserID := uuid.New()
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{OwnerID: otherUserID, OwnerType: "user"},
		}, nil).Once()

		err := service.ShareListWithTribe(listID, tribeID, userID, &expiresAt)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not have permission")
	})

	t.Run("Unshare Success", func(t *testing.T) {
		mockRepo.On("GetByID", listID).Return(list, nil).Once()
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{OwnerID: userID, OwnerType: "user"},
		}, nil).Once()
		mockRepo.On("UnshareWithTribe", listID, tribeID).Return(nil).Once()

		err := service.UnshareListWithTribe(listID, tribeID, userID)
		assert.NoError(t, err)
	})

	t.Run("Get Shared Lists", func(t *testing.T) {
		sharedLists := []*models.List{list}
		mockRepo.On("GetSharedLists", tribeID).Return(sharedLists, nil).Once()

		retrieved, err := service.GetSharedLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, sharedLists, retrieved)
	})
}
