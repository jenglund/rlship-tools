package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestMockListRepository_BasicCRUD(t *testing.T) {
	repo := &MockListRepository{}
	testID := uuid.New()
	testList := &models.List{
		ID:   testID,
		Name: "Test List",
	}

	t.Run("Create", func(t *testing.T) {
		repo.On("Create", testList).Return(nil)
		err := repo.Create(testList)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetByID", func(t *testing.T) {
		repo.On("GetByID", testID).Return(testList, nil)
		list, err := repo.GetByID(testID)
		assert.NoError(t, err)
		assert.Equal(t, testList, list)
		repo.AssertExpectations(t)
	})

	t.Run("Update", func(t *testing.T) {
		repo.On("Update", testList).Return(nil)
		err := repo.Update(testList)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Delete", func(t *testing.T) {
		repo.On("Delete", testID).Return(nil)
		err := repo.Delete(testID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("List", func(t *testing.T) {
		lists := []*models.List{testList}
		repo.On("List", 0, 10).Return(lists, nil)
		result, err := repo.List(0, 10)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
		repo.AssertExpectations(t)
	})
}

func TestMockListRepository_Items(t *testing.T) {
	repo := &MockListRepository{}
	listID := uuid.New()
	itemID := uuid.New()
	testItem := &models.ListItem{
		ID:     itemID,
		ListID: listID,
		Name:   "Test Item",
	}

	t.Run("AddItem", func(t *testing.T) {
		repo.On("AddItem", testItem).Return(nil)
		err := repo.AddItem(testItem)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("UpdateItem", func(t *testing.T) {
		repo.On("UpdateItem", testItem).Return(nil)
		err := repo.UpdateItem(testItem)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("RemoveItem", func(t *testing.T) {
		repo.On("RemoveItem", listID, itemID).Return(nil)
		err := repo.RemoveItem(listID, itemID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetItems", func(t *testing.T) {
		items := []*models.ListItem{testItem}
		repo.On("GetItems", listID).Return(items, nil)
		result, err := repo.GetItems(listID)
		assert.NoError(t, err)
		assert.Equal(t, items, result)
		repo.AssertExpectations(t)
	})

	t.Run("GetEligibleItems", func(t *testing.T) {
		listIDs := []uuid.UUID{listID}
		filters := map[string]interface{}{"key": "value"}
		items := []*models.ListItem{testItem}
		repo.On("GetEligibleItems", listIDs, filters).Return(items, nil)
		result, err := repo.GetEligibleItems(listIDs, filters)
		assert.NoError(t, err)
		assert.Equal(t, items, result)
		repo.AssertExpectations(t)
	})

	t.Run("UpdateItemStats", func(t *testing.T) {
		repo.On("UpdateItemStats", itemID, true).Return(nil)
		err := repo.UpdateItemStats(itemID, true)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("MarkItemChosen", func(t *testing.T) {
		repo.On("MarkItemChosen", itemID).Return(nil)
		err := repo.MarkItemChosen(itemID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestMockListRepository_Owners(t *testing.T) {
	repo := &MockListRepository{}
	listID := uuid.New()
	ownerID := uuid.New()
	testOwner := &models.ListOwner{
		ListID:    listID,
		OwnerID:   ownerID,
		OwnerType: models.OwnerTypeUser,
	}

	t.Run("AddOwner", func(t *testing.T) {
		repo.On("AddOwner", testOwner).Return(nil)
		err := repo.AddOwner(testOwner)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("RemoveOwner", func(t *testing.T) {
		repo.On("RemoveOwner", listID, ownerID).Return(nil)
		err := repo.RemoveOwner(listID, ownerID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetOwners", func(t *testing.T) {
		owners := []*models.ListOwner{testOwner}
		repo.On("GetOwners", listID).Return(owners, nil)
		result, err := repo.GetOwners(listID)
		assert.NoError(t, err)
		assert.Equal(t, owners, result)
		repo.AssertExpectations(t)
	})

	t.Run("GetUserLists", func(t *testing.T) {
		lists := []*models.List{{ID: listID}}
		repo.On("GetUserLists", ownerID).Return(lists, nil)
		result, err := repo.GetUserLists(ownerID)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
		repo.AssertExpectations(t)
	})

	t.Run("GetTribeLists", func(t *testing.T) {
		tribeID := uuid.New()
		lists := []*models.List{{ID: listID}}
		repo.On("GetTribeLists", tribeID).Return(lists, nil)
		result, err := repo.GetTribeLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
		repo.AssertExpectations(t)
	})

	t.Run("GetListsByOwner", func(t *testing.T) {
		lists := []*models.List{{ID: listID}}
		repo.On("GetListsByOwner", ownerID, models.OwnerTypeUser).Return(lists, nil)
		result, err := repo.GetListsByOwner(ownerID, models.OwnerTypeUser)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
		repo.AssertExpectations(t)
	})
}

func TestMockListRepository_Sharing(t *testing.T) {
	repo := &MockListRepository{}
	listID := uuid.New()
	tribeID := uuid.New()
	testShare := &models.ListShare{
		ListID:  listID,
		TribeID: tribeID,
	}

	t.Run("ShareWithTribe", func(t *testing.T) {
		repo.On("ShareWithTribe", testShare).Return(nil)
		err := repo.ShareWithTribe(testShare)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("UnshareWithTribe", func(t *testing.T) {
		repo.On("UnshareWithTribe", listID, tribeID).Return(nil)
		err := repo.UnshareWithTribe(listID, tribeID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetSharedLists", func(t *testing.T) {
		lists := []*models.List{{ID: listID}}
		repo.On("GetSharedLists", tribeID).Return(lists, nil)
		result, err := repo.GetSharedLists(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
		repo.AssertExpectations(t)
	})

	t.Run("GetSharedTribes", func(t *testing.T) {
		now := time.Now()
		tribes := []*models.Tribe{{
			BaseModel: models.BaseModel{
				ID:        tribeID,
				CreatedAt: now,
				UpdatedAt: now,
				Version:   1,
			},
			Name:       "Test Tribe",
			Type:       models.TribeTypeCouple,
			Visibility: models.VisibilityPrivate,
		}}
		repo.On("GetSharedTribes", listID).Return(tribes, nil)
		result, err := repo.GetSharedTribes(listID)
		assert.NoError(t, err)
		assert.Equal(t, tribes, result)
		repo.AssertExpectations(t)
	})

	t.Run("GetListShares", func(t *testing.T) {
		shares := []*models.ListShare{testShare}
		repo.On("GetListShares", listID).Return(shares, nil)
		result, err := repo.GetListShares(listID)
		assert.NoError(t, err)
		assert.Equal(t, shares, result)
		repo.AssertExpectations(t)
	})
}

func TestMockListRepository_SyncOperations(t *testing.T) {
	listID := uuid.New()
	testConflict := &models.SyncConflict{
		ID:         uuid.New(),
		ListID:     listID,
		Type:       "item",
		LocalData:  map[string]interface{}{"name": "local"},
		RemoteData: map[string]interface{}{"name": "remote"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t.Run("UpdateSyncStatus", func(t *testing.T) {
		repo := &MockListRepository{}
		repo.On("UpdateSyncStatus", listID, models.ListSyncStatusPending).Return(nil)
		err := repo.UpdateSyncStatus(listID, models.ListSyncStatusPending)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetConflicts", func(t *testing.T) {
		repo := &MockListRepository{}
		conflicts := []*models.SyncConflict{testConflict}
		repo.On("GetConflicts", listID).Return(conflicts, nil)
		result, err := repo.GetConflicts(listID)
		assert.NoError(t, err)
		assert.Equal(t, conflicts, result)
		repo.AssertExpectations(t)
	})

	t.Run("CreateConflict", func(t *testing.T) {
		repo := &MockListRepository{}
		repo.On("CreateConflict", testConflict).Return(nil)
		err := repo.CreateConflict(testConflict)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("ResolveConflict", func(t *testing.T) {
		repo := &MockListRepository{}
		repo.On("ResolveConflict", testConflict.ID).Return(nil)
		err := repo.ResolveConflict(testConflict.ID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetListsBySource", func(t *testing.T) {
		repo := &MockListRepository{}
		source := string(models.SyncSourceGoogleMaps)
		lists := []*models.List{{ID: listID}}
		repo.On("GetListsBySource", source).Return(lists, nil)
		result, err := repo.GetListsBySource(source)
		assert.NoError(t, err)
		assert.Equal(t, lists, result)
		repo.AssertExpectations(t)
	})

	t.Run("AddConflict", func(t *testing.T) {
		repo := &MockListRepository{}
		repo.On("AddConflict", testConflict).Return(nil)
		err := repo.AddConflict(testConflict)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// Test error cases
	t.Run("UpdateSyncStatus error", func(t *testing.T) {
		repo := &MockListRepository{}
		expectedErr := fmt.Errorf("sync status update failed")
		repo.On("UpdateSyncStatus", listID, models.ListSyncStatusConflict).Return(expectedErr)
		err := repo.UpdateSyncStatus(listID, models.ListSyncStatusConflict)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetConflicts error", func(t *testing.T) {
		repo := &MockListRepository{}
		expectedErr := fmt.Errorf("failed to get conflicts")
		repo.On("GetConflicts", listID).Return([]*models.SyncConflict(nil), expectedErr)
		result, err := repo.GetConflicts(listID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("CreateConflict error", func(t *testing.T) {
		repo := &MockListRepository{}
		expectedErr := fmt.Errorf("failed to create conflict")
		repo.On("CreateConflict", testConflict).Return(expectedErr)
		err := repo.CreateConflict(testConflict)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})

	t.Run("ResolveConflict error", func(t *testing.T) {
		repo := &MockListRepository{}
		expectedErr := fmt.Errorf("failed to resolve conflict")
		repo.On("ResolveConflict", testConflict.ID).Return(expectedErr)
		err := repo.ResolveConflict(testConflict.ID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetListsBySource error", func(t *testing.T) {
		repo := &MockListRepository{}
		source := string(models.SyncSourceGoogleMaps)
		expectedErr := fmt.Errorf("failed to get lists by source")
		repo.On("GetListsBySource", source).Return([]*models.List(nil), expectedErr)
		result, err := repo.GetListsBySource(source)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("AddConflict error", func(t *testing.T) {
		repo := &MockListRepository{}
		expectedErr := fmt.Errorf("failed to add conflict")
		repo.On("AddConflict", testConflict).Return(expectedErr)
		err := repo.AddConflict(testConflict)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestMockListRepository_ItemStats(t *testing.T) {
	itemID := uuid.New()

	t.Run("UpdateItemStats", func(t *testing.T) {
		repo := &MockListRepository{}
		repo.On("UpdateItemStats", itemID, true).Return(nil)
		err := repo.UpdateItemStats(itemID, true)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("MarkItemChosen", func(t *testing.T) {
		repo := &MockListRepository{}
		repo.On("MarkItemChosen", itemID).Return(nil)
		err := repo.MarkItemChosen(itemID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// Test error cases
	t.Run("UpdateItemStats error", func(t *testing.T) {
		repo := &MockListRepository{}
		expectedErr := fmt.Errorf("failed to update item stats")
		repo.On("UpdateItemStats", itemID, false).Return(expectedErr)
		err := repo.UpdateItemStats(itemID, false)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})

	t.Run("MarkItemChosen error", func(t *testing.T) {
		repo := &MockListRepository{}
		expectedErr := fmt.Errorf("failed to mark item as chosen")
		repo.On("MarkItemChosen", itemID).Return(expectedErr)
		err := repo.MarkItemChosen(itemID)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}
