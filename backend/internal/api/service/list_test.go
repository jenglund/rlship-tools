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
	conflict := &models.SyncConflict{
		ID:         conflictID,
		ListID:     listID,
		Type:       "name_conflict",
		LocalData:  "Local Name",
		RemoteData: "Remote Name",
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
		err := service.ResolveListConflict(listID, conflictID, "use_local")
		assert.NoError(t, err)
	})
}

func TestListService_Ownership(t *testing.T) {
	mockRepo := new(testutil.MockListRepository)
	service := NewListService(mockRepo)
	defer mockRepo.AssertExpectations(t)

	listID := uuid.New()
	ownerID := uuid.New()

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
