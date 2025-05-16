package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/rlship_test?sslmode=disable")
	require.NoError(t, err)
	return db
}

func TestListRepository(t *testing.T) {
	db := setupTestDB(t)
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
			Address:    testutil.StringPtr("123 Test St"),
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
		const (
			cooldownDays = 7
			daysAgo10    = -10 * 24 * time.Hour
			daysAgo3     = -3 * 24 * time.Hour
		)

		// Create test lists
		list1 := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "List 1",
			DefaultWeight: 1.0,
			CooldownDays:  testutil.IntPtr(cooldownDays),
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
				LastChosen:  testutil.TimePtr(time.Now().Add(daysAgo10)), // 10 days ago
				ChosenCount: 1,
			},
			{
				ListID:      list1.ID,
				Name:        "Item 2",
				Weight:      1.0,
				LastChosen:  testutil.TimePtr(time.Now().Add(daysAgo3)), // 3 days ago
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
			"cooldown_days": cooldownDays,
		}
		eligible, err := repo.GetEligibleItems([]uuid.UUID{list1.ID, list2.ID}, filters)
		require.NoError(t, err)
		assert.Len(t, eligible, 2) // Should exclude Item 2 (chosen 3 days ago)

		// Verify the correct items are returned
		var foundItem1, foundItem3 bool
		for _, item := range eligible {
			switch item.Name {
			case "Item 1":
				foundItem1 = true
			case "Item 3":
				foundItem3 = true
			case "Item 2":
				t.Error("Item 2 should not be eligible due to cooldown")
			}
		}
		assert.True(t, foundItem1, "Item 1 should be eligible")
		assert.True(t, foundItem3, "Item 3 should be eligible")

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
		lists, err := repo.GetListsBySource("google_maps")
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

	t.Run("List Ownership", func(t *testing.T) {
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

		// Test AddOwner
		userID := uuid.New()
		userOwner := &models.ListOwner{
			ListID:    list.ID,
			OwnerID:   userID,
			OwnerType: models.OwnerTypeUser,
		}
		err = repo.AddOwner(userOwner)
		require.NoError(t, err)

		tribeID := uuid.New()
		tribeOwner := &models.ListOwner{
			ListID:    list.ID,
			OwnerID:   tribeID,
			OwnerType: models.OwnerTypeTribe,
		}
		err = repo.AddOwner(tribeOwner)
		require.NoError(t, err)

		// Test GetOwners
		owners, err := repo.GetOwners(list.ID)
		require.NoError(t, err)
		require.Len(t, owners, 2)

		// Verify owner details
		var foundUser, foundTribe bool
		for _, owner := range owners {
			if owner.OwnerID == userID && owner.OwnerType == models.OwnerTypeUser {
				foundUser = true
			}
			if owner.OwnerID == tribeID && owner.OwnerType == models.OwnerTypeTribe {
				foundTribe = true
			}
		}
		assert.True(t, foundUser, "User owner not found")
		assert.True(t, foundTribe, "Tribe owner not found")

		// Test GetUserLists
		userLists, err := repo.GetUserLists(userID)
		require.NoError(t, err)
		require.Len(t, userLists, 1)
		assert.Equal(t, list.ID, userLists[0].ID)

		// Test GetTribeLists
		tribeLists, err := repo.GetTribeLists(tribeID)
		require.NoError(t, err)
		require.Len(t, tribeLists, 1)
		assert.Equal(t, list.ID, tribeLists[0].ID)

		// Test RemoveOwner
		err = repo.RemoveOwner(list.ID, userID)
		require.NoError(t, err)

		owners, err = repo.GetOwners(list.ID)
		require.NoError(t, err)
		require.Len(t, owners, 1)
		assert.Equal(t, tribeID, owners[0].OwnerID)
	})

	t.Run("List Sharing", func(t *testing.T) {
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

		// Create test user
		user := testutil.CreateTestUser(t, db)

		// Create test tribe
		tribe := testutil.CreateTestTribe(t, db, []testutil.TestUser{user})

		// Test AddOwner
		err = repo.AddOwner(&models.ListOwner{
			ListID:    list.ID,
			OwnerID:   user.ID,
			OwnerType: models.OwnerTypeUser,
		})
		require.NoError(t, err)

		// Test ShareWithTribe
		expiresAt := testutil.TimePtr(time.Now().Add(24 * time.Hour))
		err = repo.ShareWithTribe(&models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: expiresAt,
		})
		require.NoError(t, err)

		// Test GetSharedLists
		sharedLists, err := repo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		require.Len(t, sharedLists, 1)
		assert.Equal(t, list.ID, sharedLists[0].ID)

		// Verify share details
		shares := sharedLists[0].Shares
		require.Len(t, shares, 1)
		assert.Equal(t, tribe.ID, shares[0].TribeID)
		assert.Equal(t, user.ID, shares[0].UserID)
		assert.Equal(t, expiresAt.Unix(), shares[0].ExpiresAt.Unix())

		// Test UnshareWithTribe
		err = repo.UnshareWithTribe(list.ID, tribe.ID)
		require.NoError(t, err)

		sharedLists, err = repo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		assert.Empty(t, sharedLists)
	})

	t.Run("List Sharing Expiration", func(t *testing.T) {
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

		// Create test user
		user := testutil.CreateTestUser(t, db)

		// Create test tribe
		tribe := testutil.CreateTestTribe(t, db, []testutil.TestUser{user})

		// Add the user as an owner
		err = repo.AddOwner(&models.ListOwner{
			ListID:    list.ID,
			OwnerID:   user.ID,
			OwnerType: models.OwnerTypeUser,
		})
		require.NoError(t, err)

		// Share with expired time
		expiresAt := testutil.TimePtr(time.Now().Add(-1 * time.Hour)) // Expired 1 hour ago
		err = repo.ShareWithTribe(&models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: expiresAt,
		})
		require.NoError(t, err)

		// Verify expired share is not returned
		sharedLists, err := repo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		assert.Empty(t, sharedLists)
	})
}

func TestListRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)

	t.Run("success", func(t *testing.T) {
		list := &models.List{
			ID:            uuid.New(),
			Type:          models.ListTypeGeneral,
			Name:          "Test List",
			Description:   "A test list",
			Visibility:    models.VisibilityPrivate,
			SyncStatus:    models.ListSyncStatusNone,
			DefaultWeight: 1,
			MaxItems:      10,
			CooldownDays:  7,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			LastSyncAt:    time.Now().UTC(),
		}

		// Insert test data
		_, err := db.Exec(`
			INSERT INTO lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, $11, $12,
				$13, $14
			)`,
			list.ID, list.Type, list.Name, list.Description, list.Visibility,
			list.SyncStatus, list.SyncSource, list.SyncID, list.LastSyncAt,
			list.DefaultWeight, list.MaxItems, list.CooldownDays,
			list.CreatedAt, list.UpdatedAt,
		)
		require.NoError(t, err)

		// Test retrieval
		result, err := repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, list.ID, result.ID)
		assert.Equal(t, list.Name, result.Name)
		assert.Equal(t, list.Description, result.Description)
		assert.Equal(t, list.Type, result.Type)
		assert.Equal(t, list.Visibility, result.Visibility)
		assert.Equal(t, list.SyncStatus, result.SyncStatus)
		assert.Equal(t, list.DefaultWeight, result.DefaultWeight)
		assert.Equal(t, list.MaxItems, result.MaxItems)
		assert.Equal(t, list.CooldownDays, result.CooldownDays)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(uuid.New())
		assert.ErrorIs(t, err, models.ErrNotFound)
	})
}

func TestListRepository_GetOwners(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)

	listID := uuid.New()
	userID := uuid.New()
	tribeID := uuid.New()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO lists (id, type, name, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		listID, models.ListTypeGeneral, "Test List", models.VisibilityPrivate,
	)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO list_owners (list_id, owner_id, owner_type)
		VALUES 
			($1, $2, $3),
			($1, $3, $4)`,
		listID, userID, models.OwnerTypeUser, models.OwnerTypeTribe,
	)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		owners, err := repo.GetOwners(listID)
		require.NoError(t, err)
		assert.Len(t, owners, 2)

		// Verify user owner
		var foundUser, foundTribe bool
		for _, owner := range owners {
			switch owner.OwnerType {
			case models.OwnerTypeUser:
				assert.Equal(t, userID, owner.OwnerID)
				foundUser = true
			case models.OwnerTypeTribe:
				assert.Equal(t, tribeID, owner.OwnerID)
				foundTribe = true
			}
		}
		assert.True(t, foundUser, "user owner not found")
		assert.True(t, foundTribe, "tribe owner not found")
	})

	t.Run("no owners", func(t *testing.T) {
		owners, err := repo.GetOwners(uuid.New())
		require.NoError(t, err)
		assert.Empty(t, owners)
	})
}

func TestListRepository_GetListsBySource(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)

	const (
		googleMapsSource = "google_maps"
		sourceID1        = "source_id_1"
		sourceID2        = "source_id_2"
		unknownSource    = "unknown_source"
	)

	// Create test lists
	list1 := &models.List{
		Type:          models.ListTypeGoogleMap,
		Name:          "Test List 1",
		Description:   "Test Description 1",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1.0,
		SyncSource:    googleMapsSource,
		SyncID:        sourceID1,
	}
	err := repo.Create(list1)
	require.NoError(t, err)

	list2 := &models.List{
		Type:          models.ListTypeGoogleMap,
		Name:          "Test List 2",
		Description:   "Test Description 2",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1.0,
		SyncSource:    googleMapsSource,
		SyncID:        sourceID2,
	}
	err = repo.Create(list2)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		lists, err := repo.GetListsBySource(googleMapsSource)
		require.NoError(t, err)
		assert.Len(t, lists, 2)

		// Verify the lists are returned in the correct order
		foundList1, foundList2 := false, false
		for _, list := range lists {
			switch list.SyncID {
			case sourceID1:
				foundList1 = true
				assert.Equal(t, list1.Name, list.Name)
			case sourceID2:
				foundList2 = true
				assert.Equal(t, list2.Name, list.Name)
			}
		}
		assert.True(t, foundList1, "List 1 not found")
		assert.True(t, foundList2, "List 2 not found")
	})

	t.Run("no lists", func(t *testing.T) {
		lists, err := repo.GetListsBySource(unknownSource)
		require.NoError(t, err)
		assert.Empty(t, lists)
	})
}

func TestListRepository_ShareWithTribe(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)

	// Create test list
	list := &models.List{
		Type:          models.ListTypeLocation,
		Name:          "Test List",
		Description:   "Test Description",
		Visibility:    models.VisibilityPublic,
		DefaultWeight: 1.0,
	}
	err := repo.Create(list)
	require.NoError(t, err)

	// Create test user
	user := testutil.CreateTestUser(t, db)

	// Create test tribe
	tribe := testutil.CreateTestTribe(t, db, []testutil.TestUser{user})

	// Test sharing
	share := &models.ListShare{
		ListID:    list.ID,
		TribeID:   tribe.ID,
		UserID:    user.ID,
		ExpiresAt: nil,
	}
	err = repo.ShareWithTribe(share)
	require.NoError(t, err)

	// Test GetSharedLists
	sharedLists, err := repo.GetSharedLists(tribe.ID)
	require.NoError(t, err)
	require.Len(t, sharedLists, 1)
	assert.Equal(t, list.ID, sharedLists[0].ID)
}

func TestListRepository_UnshareWithTribe(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)

	listID := uuid.New()
	tribeID := uuid.New()
	userID := uuid.New()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO lists (id, type, name, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		listID, models.ListTypeGeneral, "Test List", models.VisibilityPrivate,
	)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO list_owners (list_id, owner_id, owner_type)
		VALUES ($1, $2, $3)`,
		listID, tribeID, models.OwnerTypeTribe,
	)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO list_shares (list_id, tribe_id, shared_by)
		VALUES ($1, $2, $3)`,
		listID, tribeID, userID,
	)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err := repo.UnshareWithTribe(listID, tribeID)
		require.NoError(t, err)

		// Verify tribe was removed as owner
		var ownerExists bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM list_owners 
				WHERE list_id = $1 AND owner_id = $2 AND owner_type = $3
			)`,
			listID, tribeID, models.OwnerTypeTribe,
		).Scan(&ownerExists)
		require.NoError(t, err)
		assert.False(t, ownerExists)

		// Verify share record was removed
		var shareExists bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM list_shares 
				WHERE list_id = $1 AND tribe_id = $2
			)`,
			listID, tribeID,
		).Scan(&shareExists)
		require.NoError(t, err)
		assert.False(t, shareExists)
	})

	t.Run("non-existent share", func(t *testing.T) {
		err := repo.UnshareWithTribe(uuid.New(), uuid.New())
		require.NoError(t, err)
	})
}
