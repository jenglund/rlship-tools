package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
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

func (m *MockListRepository) AddOwner(listID, userID uuid.UUID, ownerType string) error {
	args := m.Called(listID, userID, ownerType)
	return args.Error(0)
}

func (m *MockListRepository) RemoveOwner(listID, userID uuid.UUID) error {
	args := m.Called(listID, userID)
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

func (m *MockListRepository) ShareWithTribe(listID, tribeID, userID uuid.UUID, expiresAt *time.Time) error {
	args := m.Called(listID, tribeID, userID, expiresAt)
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
	mockRepo := new(MockListRepository)
	service := NewListService(mockRepo)

	t.Run("Basic CRUD Operations", func(t *testing.T) {
		list := &models.List{
			ID:            uuid.New(),
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			DefaultWeight: 1.0,
		}

		// Test CreateList
		mockRepo.On("Create", list).Return(nil)
		err := service.CreateList(list)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test GetList
		mockRepo.On("GetByID", list.ID).Return(list, nil)
		retrieved, err := service.GetList(list.ID)
		assert.NoError(t, err)
		assert.Equal(t, list, retrieved)
		mockRepo.AssertExpectations(t)

		// Test UpdateList
		mockRepo.On("Update", list).Return(nil)
		err = service.UpdateList(list)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test DeleteList
		mockRepo.On("Delete", list.ID).Return(nil)
		err = service.DeleteList(list.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("List Items Management", func(t *testing.T) {
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
		mockRepo.AssertExpectations(t)

		// Test UpdateListItem
		mockRepo.On("UpdateItem", item).Return(nil)
		err = service.UpdateListItem(item)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test GetListItems
		items := []*models.ListItem{item}
		mockRepo.On("GetItems", listID).Return(items, nil)
		retrieved, err := service.GetListItems(listID)
		assert.NoError(t, err)
		assert.Equal(t, items, retrieved)
		mockRepo.AssertExpectations(t)

		// Test RemoveListItem
		mockRepo.On("RemoveItem", listID, item.ID).Return(nil)
		err = service.RemoveListItem(listID, item.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Menu Generation", func(t *testing.T) {
		listIDs := []uuid.UUID{uuid.New(), uuid.New()}
		filters := map[string]interface{}{
			"cooldown_days": 7,
			"max_items":     2,
		}

		items := []*models.ListItem{
			{
				ID:          uuid.New(),
				ListID:      listIDs[0],
				Name:        "Item 1",
				Weight:      1.5,
				LastChosen:  nil,
				ChosenCount: 0,
			},
			{
				ID:          uuid.New(),
				ListID:      listIDs[0],
				Name:        "Item 2",
				Weight:      1.0,
				LastChosen:  nil,
				ChosenCount: 0,
			},
			{
				ID:          uuid.New(),
				ListID:      listIDs[1],
				Name:        "Item 3",
				Weight:      2.0,
				LastChosen:  nil,
				ChosenCount: 0,
			},
		}

		// Test GenerateMenu
		mockRepo.On("GetEligibleItems", listIDs, filters).Return(items, nil)
		mockRepo.On("UpdateItemStats", mock.Anything, true).Return(nil).Times(2)

		selected, err := service.GenerateMenu(listIDs, filters)
		assert.NoError(t, err)
		assert.Len(t, selected, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("List Sync", func(t *testing.T) {
		listID := uuid.New()
		list := &models.List{
			ID:         listID,
			Type:       models.ListTypeGoogleMap,
			SyncSource: "google_maps:want_to_go",
		}

		// Test SyncList
		mockRepo.On("GetByID", listID).Return(list, nil)
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil)
		mockRepo.On("UpdateSyncStatus", listID, models.ListSyncStatusSynced).Return(nil)

		err := service.SyncList(listID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test sync failure for unconfigured list
		unconfiguredList := &models.List{
			ID:   uuid.New(),
			Type: models.ListTypeLocation,
		}
		mockRepo.On("GetByID", unconfiguredList.ID).Return(unconfiguredList, nil)

		err = service.SyncList(unconfiguredList.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured for syncing")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Conflict Management", func(t *testing.T) {
		listID := uuid.New()
		conflict := &models.ListConflict{
			ID:           uuid.New(),
			ListID:       listID,
			ConflictType: "modified",
			LocalData: map[string]interface{}{
				"name": "Local Name",
			},
			ExternalData: map[string]interface{}{
				"name": "External Name",
			},
		}

		// Test CreateListConflict
		mockRepo.On("CreateConflict", conflict).Return(nil)
		err := service.CreateListConflict(conflict)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test GetListConflicts
		conflicts := []*models.ListConflict{conflict}
		mockRepo.On("GetConflicts", listID).Return(conflicts, nil)
		retrieved, err := service.GetListConflicts(listID)
		assert.NoError(t, err)
		assert.Equal(t, conflicts, retrieved)
		mockRepo.AssertExpectations(t)

		// Test ResolveListConflict
		mockRepo.On("ResolveConflict", conflict.ID).Return(nil)
		err = service.ResolveListConflict(conflict.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("List Ownership", func(t *testing.T) {
		listID := uuid.New()
		userID := uuid.New()
		tribeID := uuid.New()

		// Test AddListOwner
		mockRepo.On("AddOwner", listID, userID, "user").Return(nil)
		err := service.AddListOwner(listID, userID, "user")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test invalid owner type
		err = service.AddListOwner(listID, userID, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid owner type")

		// Test RemoveListOwner
		mockRepo.On("RemoveOwner", listID, userID).Return(nil)
		err = service.RemoveListOwner(listID, userID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test GetListOwners
		owners := []*models.ListOwner{
			{ListID: listID, OwnerID: userID, OwnerType: "user"},
			{ListID: listID, OwnerID: tribeID, OwnerType: "tribe"},
		}
		mockRepo.On("GetOwners", listID).Return(owners, nil)
		retrieved, err := service.GetListOwners(listID)
		assert.NoError(t, err)
		assert.Equal(t, owners, retrieved)
		mockRepo.AssertExpectations(t)

		// Test GetUserLists
		lists := []*models.List{
			{ID: listID, Name: "Test List"},
		}
		mockRepo.On("GetUserLists", userID).Return(lists, nil)
		userLists, err := service.GetUserLists(userID)
		assert.NoError(t, err)
		assert.Equal(t, lists, userLists)
		mockRepo.AssertExpectations(t)

		// Test GetTribeLists
		mockRepo.On("GetTribeLists", tribeID).Return(lists, nil)
		tribeLists, err := service.GetTribeLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, lists, tribeLists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("List Sharing", func(t *testing.T) {
		listID := uuid.New()
		userID := uuid.New()
		tribeID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)

		list := &models.List{
			ID:   listID,
			Name: "Test List",
		}

		// Test ShareListWithTribe - successful case
		mockRepo.On("GetByID", listID).Return(list, nil)
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{ListID: listID, OwnerID: userID, OwnerType: "user"},
		}, nil)
		mockRepo.On("ShareWithTribe", listID, tribeID, userID, &expiresAt).Return(nil)

		err := service.ShareListWithTribe(listID, tribeID, userID, &expiresAt)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test ShareListWithTribe - no permission
		mockRepo.On("GetByID", listID).Return(list, nil)
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{ListID: listID, OwnerID: uuid.New(), OwnerType: "user"},
		}, nil)

		err = service.ShareListWithTribe(listID, tribeID, userID, &expiresAt)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not have permission")
		mockRepo.AssertExpectations(t)

		// Test UnshareListWithTribe - successful case
		mockRepo.On("GetByID", listID).Return(list, nil)
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{ListID: listID, OwnerID: userID, OwnerType: "user"},
		}, nil)
		mockRepo.On("UnshareWithTribe", listID, tribeID).Return(nil)

		err = service.UnshareListWithTribe(listID, tribeID, userID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)

		// Test UnshareListWithTribe - no permission
		mockRepo.On("GetByID", listID).Return(list, nil)
		mockRepo.On("GetOwners", listID).Return([]*models.ListOwner{
			{ListID: listID, OwnerID: uuid.New(), OwnerType: "user"},
		}, nil)

		err = service.UnshareListWithTribe(listID, tribeID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not have permission")
		mockRepo.AssertExpectations(t)

		// Test GetSharedLists
		sharedLists := []*models.List{list}
		mockRepo.On("GetSharedLists", tribeID).Return(sharedLists, nil)

		retrieved, err := service.GetSharedLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, sharedLists, retrieved)
		mockRepo.AssertExpectations(t)
	})
}
