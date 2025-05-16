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
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(db)
	user := testutil.CreateTestUser(t, db)
	tribe := testutil.CreateTestTribe(t, db, []testutil.TestUser{
		{
			ID:          user.ID,
			FirebaseUID: user.FirebaseUID,
			Provider:    user.Provider,
			Email:       user.Email,
			Name:        user.Name,
			AvatarURL:   user.AvatarURL,
		},
	})

	t.Run("Basic CRUD Operations", func(t *testing.T) {
		// Create a test list
		maxItems := 10
		cooldownDays := 7
		ownerType := models.OwnerTypeUser
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
			MaxItems:      &maxItems,
			CooldownDays:  &cooldownDays,
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
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
		assert.Equal(t, list.OwnerID, retrieved.OwnerID)
		assert.Equal(t, list.OwnerType, retrieved.OwnerType)

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
		assert.ErrorIs(t, err, models.ErrNotFound)
	})

	t.Run("List Items Management", func(t *testing.T) {
		// Create a test list
		ownerType := models.OwnerTypeUser
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
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
		ownerType := models.OwnerTypeUser
		list1 := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "List 1",
			DefaultWeight: 1.0,
			CooldownDays:  testutil.IntPtr(cooldownDays),
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
		}
		err := repo.Create(list1)
		require.NoError(t, err)

		list2 := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "List 2",
			DefaultWeight: 2.0,
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
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
		assert.Len(t, eligible, 2) // Item 1 and Item 3 should be eligible

		// Test GetEligibleItems without cooldown filter
		eligible, err = repo.GetEligibleItems([]uuid.UUID{list1.ID, list2.ID}, nil)
		require.NoError(t, err)
		assert.Len(t, eligible, 3) // All items should be eligible

		// Test MarkItemChosen
		err = repo.MarkItemChosen(items[0].ID)
		require.NoError(t, err)

		chosen, err := repo.GetItems(list1.ID)
		require.NoError(t, err)
		require.Len(t, chosen, 2)
		for _, item := range chosen {
			if item.ID == items[0].ID {
				assert.Equal(t, 2, item.ChosenCount)
			}
		}
	})

	t.Run("Sync Management", func(t *testing.T) {
		ownerType := models.OwnerTypeUser
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPublic,
			DefaultWeight: 1.0,
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
			SyncStatus:    models.ListSyncStatusNone,
			SyncSource:    models.SyncSourceGoogleMaps,
		}
		err := repo.Create(list)
		require.NoError(t, err)

		// Test UpdateSyncStatus
		err = repo.UpdateSyncStatus(list.ID, models.ListSyncStatusSynced)
		require.NoError(t, err)

		updated, err := repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ListSyncStatusSynced, updated.SyncStatus)

		// Test AddConflict
		conflict := &models.SyncConflict{
			ID:         uuid.New(),
			ListID:     list.ID,
			Type:       "name_conflict",
			LocalData:  "Local Name",
			RemoteData: "Remote Name",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		err = repo.AddConflict(conflict)
		require.NoError(t, err)

		// Test GetConflicts
		conflicts, err := repo.GetConflicts(list.ID)
		require.NoError(t, err)
		require.Len(t, conflicts, 1)
		assert.Equal(t, conflict.Type, conflicts[0].Type)
	})

	t.Run("Owner Management", func(t *testing.T) {
		// Test GetListsByOwner for user
		userLists, err := repo.GetListsByOwner(user.ID, models.OwnerTypeUser)
		require.NoError(t, err)
		assert.NotEmpty(t, userLists)

		// Test GetListsByOwner for tribe
		tribeLists, err := repo.GetListsByOwner(tribe.ID, models.OwnerTypeTribe)
		require.NoError(t, err)
		assert.NotEmpty(t, tribeLists)
	})

	t.Run("Share Management", func(t *testing.T) {
		ownerType := models.OwnerTypeUser
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			Visibility:    models.VisibilityPrivate,
			DefaultWeight: 1.0,
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
		}
		err := repo.Create(list)
		require.NoError(t, err)

		// Test ShareWithTribe
		share := &models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = repo.ShareWithTribe(share)
		require.NoError(t, err)

		// Test GetSharedTribes
		tribes, err := repo.GetSharedTribes(list.ID)
		require.NoError(t, err)
		require.Len(t, tribes, 1)
		assert.Equal(t, tribe.ID, tribes[0].ID)
	})

	t.Run("List Sharing", func(t *testing.T) {
		// Create test users and tribe
		user2 := testutil.CreateTestUser(t, db)
		tribe := testutil.CreateTestTribe(t, db, []testutil.TestUser{
			{
				ID:          user.ID,
				FirebaseUID: user.FirebaseUID,
				Provider:    user.Provider,
				Email:       user.Email,
				Name:        user.Name,
				AvatarURL:   user.AvatarURL,
			},
			{
				ID:          user2.ID,
				FirebaseUID: user2.FirebaseUID,
				Provider:    user2.Provider,
				Email:       user2.Email,
				Name:        user2.Name,
				AvatarURL:   user2.AvatarURL,
			},
		})

		// Create a test list
		ownerType := models.OwnerTypeUser
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Shared List",
			Description:   "List to be shared",
			Visibility:    models.VisibilityPrivate,
			DefaultWeight: 1.0,
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
		}
		err := repo.Create(list)
		require.NoError(t, err)

		// Test ShareWithTribe
		share := &models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = repo.ShareWithTribe(share)
		require.NoError(t, err)

		// Test GetSharedTribes
		sharedTribes, err := repo.GetSharedTribes(list.ID)
		require.NoError(t, err)
		assert.Len(t, sharedTribes, 1)
		assert.Equal(t, tribe.ID, sharedTribes[0].ID)

		// Test UnshareWithTribe
		err = repo.UnshareWithTribe(list.ID, tribe.ID)
		require.NoError(t, err)

		sharedTribes, err = repo.GetSharedTribes(list.ID)
		require.NoError(t, err)
		assert.Empty(t, sharedTribes)
	})

	t.Run("List Sharing Expiration", func(t *testing.T) {
		// Create test users and tribe
		user2 := testutil.CreateTestUser(t, db)
		tribe := testutil.CreateTestTribe(t, db, []testutil.TestUser{
			{
				ID:          user.ID,
				FirebaseUID: user.FirebaseUID,
				Provider:    user.Provider,
				Email:       user.Email,
				Name:        user.Name,
				AvatarURL:   user.AvatarURL,
			},
			{
				ID:          user2.ID,
				FirebaseUID: user2.FirebaseUID,
				Provider:    user2.Provider,
				Email:       user2.Email,
				Name:        user2.Name,
				AvatarURL:   user2.AvatarURL,
			},
		})

		// Create a test list
		ownerType := models.OwnerTypeUser
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Expiring Shared List",
			Description:   "List with expiring share",
			Visibility:    models.VisibilityPrivate,
			DefaultWeight: 1.0,
			OwnerID:       &user.ID,
			OwnerType:     &ownerType,
		}
		err := repo.Create(list)
		require.NoError(t, err)

		// Share list with expiration
		expiresAt := time.Now().Add(24 * time.Hour)
		share := &models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: &expiresAt,
		}
		err = repo.ShareWithTribe(share)
		require.NoError(t, err)

		// Test GetSharedTribes before expiration
		sharedTribes, err := repo.GetSharedTribes(list.ID)
		require.NoError(t, err)
		assert.Len(t, sharedTribes, 1)
		assert.Equal(t, tribe.ID, sharedTribes[0].ID)

		// Test GetSharedTribes after expiration
		_, err = db.Exec(`
			UPDATE list_sharing 
			SET expires_at = $1 
			WHERE list_id = $2 AND shared_with_id = $3
		`, time.Now().Add(-1*time.Hour), list.ID, tribe.ID)
		require.NoError(t, err)

		sharedTribes, err = repo.GetSharedTribes(list.ID)
		require.NoError(t, err)
		assert.Empty(t, sharedTribes)
	})
}

func TestListRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewListRepository(db)

	t.Run("success", func(t *testing.T) {
		maxItems := 10
		cooldownDays := 7
		lastSyncAt := time.Now().UTC()
		list := &models.List{
			ID:            uuid.New(),
			Type:          models.ListTypeGeneral,
			Name:          "Test List",
			Description:   "A test list",
			Visibility:    models.VisibilityPrivate,
			SyncStatus:    models.ListSyncStatusNone,
			DefaultWeight: 1,
			MaxItems:      &maxItems,
			CooldownDays:  &cooldownDays,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			LastSyncAt:    &lastSyncAt,
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
			($1, $4, $5)`,
		listID, userID, models.OwnerTypeUser, tribeID, models.OwnerTypeTribe,
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
