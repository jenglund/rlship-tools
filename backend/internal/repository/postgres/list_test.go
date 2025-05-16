package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(db)

	t.Run("Basic CRUD Operations", func(t *testing.T) {
		// Create a test list
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
			MaxItems:      testutil.IntPtr(10),
			CooldownDays:  testutil.IntPtr(7),
		}

		// Test Create
		err := repo.Create(list)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, list.ID)

		// Test GetByID
		retrieved, err := repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, list.Name, retrieved.Name)
		assert.Equal(t, list.Type, retrieved.Type)
		assert.Equal(t, list.Description, retrieved.Description)
		assert.Equal(t, list.Visibility, retrieved.Visibility)
		assert.Equal(t, list.DefaultWeight, retrieved.DefaultWeight)
		assert.Equal(t, list.MaxItems, retrieved.MaxItems)
		assert.Equal(t, list.CooldownDays, retrieved.CooldownDays)

		// Test Update
		list.Name = "Updated List"
		list.Description = "Updated Description"
		err = repo.Update(list)
		require.NoError(t, err)

		updated, err := repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated List", updated.Name)
		assert.Equal(t, "Updated Description", updated.Description)

		// Test Delete
		err = repo.Delete(list.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(list.ID)
		assert.Error(t, err)
	})

	t.Run("List Items Management", func(t *testing.T) {
		// Create a test list
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
		}
		err := repo.Create(list)
		require.NoError(t, err)

		// Test AddItem
		item := &models.ListItem{
			ListID:      list.ID,
			Name:        "Test Item",
			Description: "Test Item Description",
			Weight:      1.5,
			Metadata: map[string]interface{}{
				"test_key": "test_value",
			},
			ExternalID: "ext123",
			Latitude:   testutil.Float64Ptr(37.7749),
			Longitude:  testutil.Float64Ptr(-122.4194),
			Address:    "123 Test St",
		}

		err = repo.AddItem(item)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, item.ID)

		// Test GetItems
		items, err := repo.GetItems(list.ID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, item.Name, items[0].Name)
		assert.Equal(t, item.Description, items[0].Description)
		assert.Equal(t, item.Weight, items[0].Weight)
		assert.Equal(t, item.ExternalID, items[0].ExternalID)
		assert.Equal(t, item.Latitude, items[0].Latitude)
		assert.Equal(t, item.Longitude, items[0].Longitude)
		assert.Equal(t, item.Address, items[0].Address)

		// Test UpdateItem
		item.Name = "Updated Item"
		item.Description = "Updated Description"
		err = repo.UpdateItem(item)
		require.NoError(t, err)

		updated, err := repo.GetItems(list.ID)
		require.NoError(t, err)
		require.Len(t, updated, 1)
		assert.Equal(t, "Updated Item", updated[0].Name)
		assert.Equal(t, "Updated Description", updated[0].Description)

		// Test RemoveItem
		err = repo.RemoveItem(list.ID, item.ID)
		require.NoError(t, err)

		items, err = repo.GetItems(list.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("Menu Generation", func(t *testing.T) {
		// Create test lists
		list1 := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "List 1",
			DefaultWeight: 1.0,
			CooldownDays:  testutil.IntPtr(7),
		}
		err := repo.Create(list1)
		require.NoError(t, err)

		list2 := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "List 2",
			DefaultWeight: 2.0,
		}
		err = repo.Create(list2)
		require.NoError(t, err)

		// Add items to lists
		items := []*models.ListItem{
			{
				ListID:      list1.ID,
				Name:        "Item 1",
				Weight:      1.5,
				LastChosen:  testutil.TimePtr(time.Now().Add(-10 * 24 * time.Hour)), // 10 days ago
				ChosenCount: 1,
			},
			{
				ListID:      list1.ID,
				Name:        "Item 2",
				Weight:      1.0,
				LastChosen:  testutil.TimePtr(time.Now().Add(-3 * 24 * time.Hour)), // 3 days ago
				ChosenCount: 2,
			},
			{
				ListID:      list2.ID,
				Name:        "Item 3",
				Weight:      0.0, // Should use list's default weight
				LastChosen:  nil, // Never chosen
				ChosenCount: 0,
			},
		}

		for _, item := range items {
			err := repo.AddItem(item)
			require.NoError(t, err)
		}

		// Test GetEligibleItems with cooldown filter
		filters := map[string]interface{}{
			"cooldown_days": 7,
		}
		eligible, err := repo.GetEligibleItems([]uuid.UUID{list1.ID, list2.ID}, filters)
		require.NoError(t, err)
		assert.Len(t, eligible, 2) // Should exclude Item 2 (chosen 3 days ago)

		// Test UpdateItemStats
		err = repo.UpdateItemStats(items[2].ID, true)
		require.NoError(t, err)

		updated, err := repo.GetItems(list2.ID)
		require.NoError(t, err)
		require.Len(t, updated, 1)
		assert.Equal(t, 1, updated[0].ChosenCount)
		assert.NotNil(t, updated[0].LastChosen)
	})

	t.Run("List Sync and Conflicts", func(t *testing.T) {
		// Create a test list with sync configuration
		list := &models.List{
			Type:       models.ListTypeGoogleMap,
			Name:       "Google Maps List",
			SyncStatus: models.ListSyncStatusNone,
			SyncSource: "google_maps:want_to_go",
			SyncID:     "gm123",
		}
		err := repo.Create(list)
		require.NoError(t, err)

		// Test UpdateSyncStatus
		err = repo.UpdateSyncStatus(list.ID, models.ListSyncStatusSynced)
		require.NoError(t, err)

		updated, err := repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ListSyncStatusSynced, updated.SyncStatus)
		assert.NotNil(t, updated.LastSyncAt)

		// Test GetListsBySource
		lists, err := repo.GetListsBySource("google_maps:want_to_go")
		require.NoError(t, err)
		require.Len(t, lists, 1)
		assert.Equal(t, list.ID, lists[0].ID)

		// Test conflict management
		conflict := &models.ListConflict{
			ListID:       list.ID,
			ConflictType: "modified",
			LocalData: map[string]interface{}{
				"name": "Local Name",
			},
			ExternalData: map[string]interface{}{
				"name": "External Name",
			},
		}

		err = repo.CreateConflict(conflict)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, conflict.ID)

		conflicts, err := repo.GetConflicts(list.ID)
		require.NoError(t, err)
		require.Len(t, conflicts, 1)
		assert.Equal(t, conflict.ConflictType, conflicts[0].ConflictType)

		err = repo.ResolveConflict(conflict.ID)
		require.NoError(t, err)

		conflicts, err = repo.GetConflicts(list.ID)
		require.NoError(t, err)
		assert.Empty(t, conflicts)
	})
}
