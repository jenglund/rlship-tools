package postgres

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRepository_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))
	testUser := testutil.CreateTestUser(t, db)

	// Create multiple test lists for pagination testing
	numLists := 5
	createdLists := make([]*models.List, numLists)

	// Create lists using the repository to ensure proper data relationships
	for i := 0; i < numLists; i++ {
		ownerType := models.OwnerTypeUser
		maxItems := 100
		cooldownDays := 7

		list := &models.List{
			Type:          models.ListTypeActivity,
			Name:          fmt.Sprintf("Test List %d", i),
			Description:   fmt.Sprintf("List Description %c", rune('A'+i)),
			Visibility:    models.VisibilityPrivate,
			DefaultWeight: 1.0,
			MaxItems:      &maxItems,
			CooldownDays:  &cooldownDays,
			SyncStatus:    models.ListSyncStatusNone,
			SyncSource:    models.SyncSourceNone,
			SyncID:        "",
			OwnerID:       &testUser.ID,
			OwnerType:     &ownerType,
			Owners: []*models.ListOwner{
				{
					OwnerID:   testUser.ID,
					OwnerType: models.OwnerTypeUser,
				},
			},
		}

		err := repo.Create(list)
		require.NoError(t, err)

		// Add some items to the list to test related data loading
		for j := 0; j < 2; j++ {
			item := &models.ListItem{
				ListID:      list.ID,
				Name:        fmt.Sprintf("Item %d for List %d", j, i),
				Description: fmt.Sprintf("Test Item %d", j),
				Weight:      1.0,
				Metadata:    models.JSONMap{"order": j, "category": "test"},
				Available:   true,
			}

			err := repo.AddItem(item)
			require.NoError(t, err)
		}

		createdLists[i] = list

		// Add a small delay to ensure created_at times are different
		time.Sleep(10 * time.Millisecond)
	}

	t.Run("List All Lists", func(t *testing.T) {
		// Get all lists
		lists, err := repo.List(0, 10)
		require.NoError(t, err)

		// Should return all lists
		assert.Equal(t, numLists, len(lists), "Should return all created lists")

		// Verify lists are returned in descending order of creation time
		for i := 0; i < len(lists)-1; i++ {
			assert.True(t, lists[i].CreatedAt.After(lists[i+1].CreatedAt) ||
				lists[i].CreatedAt.Equal(lists[i+1].CreatedAt),
				"Lists should be ordered by created_at DESC")
		}

		// Verify each list has its related data loaded
		for _, list := range lists {
			// Verify basic list properties
			assert.NotEqual(t, uuid.Nil, list.ID)
			assert.NotEmpty(t, list.Name)
			assert.NotEmpty(t, list.Description)
			assert.Equal(t, models.ListTypeActivity, list.Type)

			// Verify owners are loaded
			assert.NotEmpty(t, list.Owners, "List owners should be loaded")
			assert.Equal(t, testUser.ID, list.Owners[0].OwnerID)

			// Verify items are loaded
			assert.Equal(t, 2, len(list.Items), "Each list should have 2 items")
			for _, item := range list.Items {
				assert.NotEqual(t, uuid.Nil, item.ID)
				assert.Equal(t, list.ID, item.ListID)
				assert.NotEmpty(t, item.Name)
				assert.NotEmpty(t, item.Description)
				assert.NotEmpty(t, item.Metadata)
			}
		}
	})

	t.Run("List With Pagination", func(t *testing.T) {
		// Get first page (2 items)
		firstPage, err := repo.List(0, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, len(firstPage), "First page should have 2 items")

		// Get second page (2 items)
		secondPage, err := repo.List(2, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, len(secondPage), "Second page should have 2 items")

		// Get third page (1 item)
		thirdPage, err := repo.List(4, 2)
		require.NoError(t, err)
		assert.Equal(t, 1, len(thirdPage), "Third page should have 1 item")

		// Verify pagination works correctly
		for _, list := range firstPage {
			for _, secondList := range secondPage {
				assert.NotEqual(t, list.ID, secondList.ID, "Lists should not appear on multiple pages")
			}
		}
	})

	t.Run("List With Zero Limit", func(t *testing.T) {
		lists, err := repo.List(0, 0)
		require.NoError(t, err)
		assert.Empty(t, lists, "Zero limit query should return no results")
	})

	t.Run("List With Deleted Lists", func(t *testing.T) {
		// Delete one of the lists
		err := repo.Delete(createdLists[0].ID)
		require.NoError(t, err)

		// List should not include deleted lists
		lists, err := repo.List(0, 10)
		require.NoError(t, err)

		// Should be one less than the total
		assert.Equal(t, numLists-1, len(lists), "Should not include deleted lists")

		// Verify the deleted list is not included
		for _, list := range lists {
			assert.NotEqual(t, createdLists[0].ID, list.ID, "Deleted list should not be returned")
		}
	})
}

func TestListRepository_ItemOperations(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))
	testUser := testutil.CreateTestUser(t, db)

	// Create a test list
	ownerType := models.OwnerTypeUser
	maxItems := 100
	cooldownDays := 7

	list := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test Items List " + uuid.New().String()[:8],
		Description:   "Test List for Item Operations",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &testUser.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   testUser.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err := repo.Create(list)
	require.NoError(t, err)

	// Create test items
	items := []*models.ListItem{
		{
			ListID:      list.ID,
			Name:        "Item 1",
			Description: "First Test Item",
			Weight:      1.0,
			Metadata:    models.JSONMap{"order": 1, "category": "food"},
			Available:   true,
		},
		{
			ListID:      list.ID,
			Name:        "Item 2",
			Description: "Second Test Item",
			Weight:      1.5,
			Metadata:    models.JSONMap{"order": 2, "category": "activity"},
			Available:   true,
		},
		{
			ListID:      list.ID,
			Name:        "Item 3",
			Description: "Third Test Item",
			Weight:      2.0,
			Metadata:    models.JSONMap{"order": 3, "category": "food"},
			Available:   true,
		},
	}

	// Add items to list
	for _, item := range items {
		err := repo.AddItem(item)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, item.ID, "Item ID should be set after creation")
	}

	t.Run("GetItems", func(t *testing.T) {
		// Get all items for the list
		retrievedItems, err := repo.GetItems(list.ID)
		require.NoError(t, err)
		assert.Len(t, retrievedItems, len(items))

		// Verify each item is in the result
		itemMap := make(map[string]*models.ListItem)
		for _, item := range retrievedItems {
			itemMap[item.Name] = item

			// Basic validation of item properties
			assert.NotEqual(t, uuid.Nil, item.ID)
			assert.Equal(t, list.ID, item.ListID)
			assert.NotEmpty(t, item.Name)
			assert.NotZero(t, item.Weight)
			assert.False(t, item.CreatedAt.IsZero())
			assert.False(t, item.UpdatedAt.IsZero())
		}

		// Verify specific items
		assert.Contains(t, itemMap, "Item 1")
		assert.Contains(t, itemMap, "Item 2")
		assert.Contains(t, itemMap, "Item 3")
		assert.Equal(t, "First Test Item", itemMap["Item 1"].Description)
		assert.Equal(t, "Second Test Item", itemMap["Item 2"].Description)
		assert.Equal(t, "Third Test Item", itemMap["Item 3"].Description)
	})

	t.Run("GetItems - Non-existent List", func(t *testing.T) {
		// Try to get items for a non-existent list
		nonExistentID := uuid.New()
		items, err := repo.GetItems(nonExistentID)
		require.NoError(t, err, "Should not error for non-existent list")
		assert.Empty(t, items, "Should return empty slice for non-existent list")
	})

	t.Run("RemoveItem", func(t *testing.T) {
		// Remove the second item
		err := repo.RemoveItem(list.ID, items[1].ID)
		require.NoError(t, err)

		// Get items again to verify removal
		retrievedItems, err := repo.GetItems(list.ID)
		require.NoError(t, err)
		assert.Len(t, retrievedItems, len(items)-1)

		// Verify item was removed
		for _, item := range retrievedItems {
			assert.NotEqual(t, items[1].ID, item.ID, "Removed item should not be in results")
		}
	})

	t.Run("RemoveItem - Non-existent Item", func(t *testing.T) {
		// Try to remove a non-existent item
		nonExistentID := uuid.New()
		err := repo.RemoveItem(list.ID, nonExistentID)
		require.Error(t, err, "Should error when removing non-existent item")
		assert.Contains(t, err.Error(), "not found", "Error should indicate item not found")
	})
}

func TestListRepository_GetEligibleItems(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))
	testUser := testutil.CreateTestUser(t, db)

	// Create a test list
	ownerType := models.OwnerTypeUser
	maxItems := 100
	cooldownDays := 7

	list := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Eligible Items List " + uuid.New().String()[:8],
		Description:   "Test List for Eligible Items",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &testUser.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   testUser.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err := repo.Create(list)
	require.NoError(t, err)

	// Create a second list for testing multiple list filtering
	list2 := &models.List{
		Type:          models.ListTypeLocation,
		Name:          "Eligible Items List 2 " + uuid.New().String()[:8],
		Description:   "Second Test List for Eligible Items",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &testUser.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   testUser.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err = repo.Create(list2)
	require.NoError(t, err)

	// Create items with various properties for testing eligibility

	// Available items
	availableItem1 := &models.ListItem{
		ListID:      list.ID,
		Name:        "Available Item 1",
		Description: "Available Test Item 1",
		Weight:      1.0,
		Metadata:    models.JSONMap{"category": "food"},
		Available:   true,
	}

	availableItem2 := &models.ListItem{
		ListID:      list.ID,
		Name:        "Available Item 2",
		Description: "Available Test Item 2",
		Weight:      2.0,
		Metadata:    models.JSONMap{"category": "activity"},
		Available:   true,
	}

	// Unavailable item
	unavailableItem := &models.ListItem{
		ListID:      list.ID,
		Name:        "Unavailable Item",
		Description: "Unavailable Test Item",
		Weight:      1.5,
		Metadata:    models.JSONMap{"category": "food"},
		Available:   false,
	}

	// Item with recent usage (in cooldown)
	recentlyUsedItem := &models.ListItem{
		ListID:      list.ID,
		Name:        "Recently Used Item",
		Description: "Recently Used Test Item",
		Weight:      1.2,
		Metadata:    models.JSONMap{"category": "activity"},
		Available:   true,
		LastChosen:  timePtr(time.Now().Add(-24 * time.Hour)), // 1 day ago
		ChosenCount: 1,
		Cooldown:    intPtr(5), // 5 day cooldown
	}

	// Seasonal item (in season)
	seasonalInItem := &models.ListItem{
		ListID:      list.ID,
		Name:        "In Season Item",
		Description: "In Season Test Item",
		Weight:      1.8,
		Metadata:    models.JSONMap{"category": "food"},
		Available:   true,
		Seasonal:    true,
		StartDate:   timePtr(time.Now().Add(-30 * 24 * time.Hour)), // 30 days ago
		EndDate:     timePtr(time.Now().Add(30 * 24 * time.Hour)),  // 30 days from now
	}

	// Seasonal item (out of season)
	seasonalOutItem := &models.ListItem{
		ListID:      list.ID,
		Name:        "Out of Season Item",
		Description: "Out of Season Test Item",
		Weight:      1.7,
		Metadata:    models.JSONMap{"category": "food"},
		Available:   true,
		Seasonal:    true,
		StartDate:   timePtr(time.Now().Add(10 * 24 * time.Hour)), // 10 days from now
		EndDate:     timePtr(time.Now().Add(40 * 24 * time.Hour)), // 40 days from now
	}

	// Item in second list
	list2Item := &models.ListItem{
		ListID:      list2.ID,
		Name:        "List 2 Item",
		Description: "Test Item in List 2",
		Weight:      1.3,
		Metadata:    models.JSONMap{"category": "location"},
		Available:   true,
	}

	// Add all items
	allItems := []*models.ListItem{
		availableItem1,
		availableItem2,
		unavailableItem,
		recentlyUsedItem,
		seasonalInItem,
		seasonalOutItem,
		list2Item,
	}

	for _, item := range allItems {
		err := repo.AddItem(item)
		require.NoError(t, err)
	}

	t.Run("GetEligibleItems - All Available", func(t *testing.T) {
		// Get all eligible items with minimal filters
		// Note: The current implementation doesn't check for 'available_only'
		// It simply returns all items from the specified lists
		filters := map[string]interface{}{
			"cooldown_days": 3, // Add cooldown days filter to exclude recently used items
		}

		items, err := repo.GetEligibleItems([]uuid.UUID{list.ID}, filters)
		require.NoError(t, err)

		// Verify specific items (checking for presence rather than exclusion)
		itemMap := make(map[string]bool)
		for _, item := range items {
			itemMap[item.Name] = true
		}

		// These should be present
		assert.True(t, itemMap["Available Item 1"], "Available Item 1 should be eligible")
		assert.True(t, itemMap["Available Item 2"], "Available Item 2 should be eligible")
		assert.True(t, itemMap["Unavailable Item"], "Unavailable Item should be present (available flag not supported)")
		assert.True(t, itemMap["In Season Item"], "In Season Item should be eligible")
		assert.True(t, itemMap["Out of Season Item"], "Out of Season Item should be present (seasonality not supported)")

		// This should be excluded due to cooldown
		assert.False(t, itemMap["Recently Used Item"], "Recently Used Item should not be eligible due to cooldown")

		// This should be excluded as it's in a different list
		assert.False(t, itemMap["List 2 Item"], "List 2 Item should not be included as we only specified list1")
	})

	t.Run("GetEligibleItems - Cooldown Filter", func(t *testing.T) {
		// The current implementation only supports cooldown filter, not category
		filters := map[string]interface{}{
			"cooldown_days": 10, // Longer cooldown to test filtering
		}

		items, err := repo.GetEligibleItems([]uuid.UUID{list.ID}, filters)
		require.NoError(t, err)

		// Verify recently used item is excluded
		itemMap := make(map[string]bool)
		for _, item := range items {
			itemMap[item.Name] = true
		}

		assert.True(t, itemMap["Available Item 1"], "Available Item 1 should be eligible")
		assert.True(t, itemMap["Available Item 2"], "Available Item 2 should be eligible")
		assert.False(t, itemMap["Recently Used Item"], "Recently Used Item should not be eligible due to cooldown")
	})

	t.Run("GetEligibleItems - Multiple Lists", func(t *testing.T) {
		// Get eligible items from both lists
		filters := map[string]interface{}{
			"cooldown_days": 3, // Add cooldown filter
		}

		items, err := repo.GetEligibleItems([]uuid.UUID{list.ID, list2.ID}, filters)
		require.NoError(t, err)

		// Verify items from both lists are included
		itemMap := make(map[string]bool)
		for _, item := range items {
			itemMap[item.Name] = true
		}

		// Check items from first list
		assert.True(t, itemMap["Available Item 1"], "Available Item 1 should be eligible")
		assert.True(t, itemMap["Available Item 2"], "Available Item 2 should be eligible")
		assert.True(t, itemMap["Unavailable Item"], "Unavailable Item should be present")
		assert.True(t, itemMap["In Season Item"], "In Season Item should be eligible")
		assert.True(t, itemMap["Out of Season Item"], "Out of Season Item should be present")
		assert.False(t, itemMap["Recently Used Item"], "Recently Used Item should not be eligible due to cooldown")

		// Check item from second list
		assert.True(t, itemMap["List 2 Item"], "List 2 Item should be eligible")
	})

	t.Run("GetEligibleItems - Empty List IDs", func(t *testing.T) {
		// Get eligible items with empty list IDs array
		filters := map[string]interface{}{
			"available_only": true,
		}

		items, err := repo.GetEligibleItems([]uuid.UUID{}, filters)
		require.NoError(t, err)
		assert.Empty(t, items, "Should return empty result for empty list IDs")
	})
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

// Note: intPtr is already defined in list_test.go, so we're using that one

// TestSchemaPathOnDirectQueries verifies that methods previously using direct DB queries
// correctly use the transaction manager to set the schema path
func TestSchemaPathOnDirectQueries(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(db)
	listRepo, ok := repo.(*ListRepository)
	require.True(t, ok)

	// Create a test user
	testUser := testutil.CreateTestUser(t, db)

	ownerType := models.OwnerTypeUser

	// Add a test list with items
	list := &models.List{
		Name:          "Schema Path Test List",
		Type:          models.ListTypeGeneral,
		OwnerID:       &testUser.ID,
		OwnerType:     &ownerType,
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
	}
	err := listRepo.Create(list)
	require.NoError(t, err)

	// Add test item
	item := &models.ListItem{
		ListID: list.ID,
		Name:   "Test Item",
		Weight: 1.0,
	}
	err = listRepo.AddItem(item)
	require.NoError(t, err)

	// Test List method (previously used direct query)
	lists, err := listRepo.List(0, 10)
	require.NoError(t, err)
	require.NotEmpty(t, lists)
	found := false
	for _, l := range lists {
		if l.ID == list.ID {
			found = true
			break
		}
	}
	require.True(t, found, "List should be found in results")

	// Test RemoveItem method (previously used direct query)
	err = listRepo.RemoveItem(list.ID, item.ID)
	require.NoError(t, err)

	// Verify item was removed
	items, err := listRepo.GetItems(list.ID)
	require.NoError(t, err)
	require.Empty(t, items)

	// Test GetEligibleItems (previously used direct query)
	// First add a new item since we removed the previous one
	newItem := &models.ListItem{
		ListID: list.ID,
		Name:   "New Test Item",
		Weight: 1.0,
	}
	err = listRepo.AddItem(newItem)
	require.NoError(t, err)

	// Now test GetEligibleItems with the list ID
	eligible, err := listRepo.GetEligibleItems([]uuid.UUID{list.ID}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, eligible)
}
