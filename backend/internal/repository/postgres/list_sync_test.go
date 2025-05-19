package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRepository_UpdateSyncStatus(t *testing.T) {
	// Use the test utility with timeouts to prevent test hangs
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))
	userRepo := NewUserRepository(testutil.UnwrapDB(db))

	// Create a test user first
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create a test list
	ownerType := models.OwnerTypeUser
	list := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test Sync List " + uuid.New().String()[:8],
		Description:   "Test Description",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &user.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err = repo.Create(list)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		// Update status to pending
		err := repo.UpdateSyncStatus(list.ID, models.ListSyncStatusPending)
		require.NoError(t, err)

		// Verify updated status
		updatedList, err := repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ListSyncStatusPending, updatedList.SyncStatus)

		// Update status to synced
		err = repo.UpdateSyncStatus(list.ID, models.ListSyncStatusSynced)
		require.NoError(t, err)

		// Verify updated status
		updatedList, err = repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ListSyncStatusSynced, updatedList.SyncStatus)

		// Update status to conflict
		err = repo.UpdateSyncStatus(list.ID, models.ListSyncStatusConflict)
		require.NoError(t, err)

		// Verify updated status
		updatedList, err = repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ListSyncStatusConflict, updatedList.SyncStatus)

		// Reset status to none
		err = repo.UpdateSyncStatus(list.ID, models.ListSyncStatusNone)
		require.NoError(t, err)

		// Verify updated status
		updatedList, err = repo.GetByID(list.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ListSyncStatusNone, updatedList.SyncStatus)
	})

	t.Run("non-existent list", func(t *testing.T) {
		err := repo.UpdateSyncStatus(uuid.New(), models.ListSyncStatusPending)
		assert.ErrorIs(t, err, models.ErrNotFound)
	})

	t.Run("invalid status", func(t *testing.T) {
		err := repo.UpdateSyncStatus(list.ID, models.ListSyncStatus("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sync status")
	})
}

func TestListRepository_CreateConflict(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))
	userRepo := NewUserRepository(testutil.UnwrapDB(db))

	// Create a test user
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create a test list
	ownerType := models.OwnerTypeUser
	list := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test Conflict List " + uuid.New().String()[:8],
		Description:   "Test Description",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &user.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err = repo.Create(list)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		// Create a list conflict
		localData := map[string]interface{}{
			"name": "Local Name",
			"items": []map[string]interface{}{
				{"id": "1", "name": "Local Item 1"},
				{"id": "2", "name": "Local Item 2"},
			},
		}

		remoteData := map[string]interface{}{
			"name": "Remote Name",
			"items": []map[string]interface{}{
				{"id": "1", "name": "Remote Item 1"},
				{"id": "3", "name": "Remote Item 3"},
			},
		}

		conflict := &models.SyncConflict{
			ID:         uuid.New(), // Set a valid ID
			ListID:     list.ID,
			Type:       "list_metadata",
			LocalData:  localData,
			RemoteData: remoteData,
			CreatedAt:  time.Now(), // Set creation time
			UpdatedAt:  time.Now(), // Set update time
		}

		// Create conflict
		err = repo.CreateConflict(conflict)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, conflict.ID, "Conflict ID should be set")

		// Verify conflict created successfully
		conflicts, err := repo.GetConflicts(list.ID)
		require.NoError(t, err)
		require.Len(t, conflicts, 1, "Should have one conflict")

		// Check conflict data
		assert.Equal(t, list.ID, conflicts[0].ListID)
		assert.Equal(t, "list_metadata", conflicts[0].Type)

		// Convert to JSON for comparing complex data
		expectedLocalJSON, err := json.Marshal(localData)
		require.NoError(t, err)
		actualLocalJSON, err := json.Marshal(conflicts[0].LocalData)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedLocalJSON), string(actualLocalJSON))

		expectedRemoteJSON, err := json.Marshal(remoteData)
		require.NoError(t, err)
		actualRemoteJSON, err := json.Marshal(conflicts[0].RemoteData)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedRemoteJSON), string(actualRemoteJSON))
	})

	t.Run("missing data", func(t *testing.T) {
		conflict := &models.SyncConflict{
			ID:     uuid.New(), // Set a valid ID
			ListID: list.ID,
			Type:   "list_metadata",
			// Missing LocalData and RemoteData
			CreatedAt: time.Now(), // Set creation time
			UpdatedAt: time.Now(), // Set update time
		}

		err = repo.CreateConflict(conflict)
		assert.Error(t, err)
	})
}

func TestListRepository_GetConflicts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))
	userRepo := NewUserRepository(testutil.UnwrapDB(db))

	// Create a test user
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create two test lists
	ownerType := models.OwnerTypeUser
	list1 := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test List 1 " + uuid.New().String()[:8],
		Description:   "Test Description 1",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &user.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err = repo.Create(list1)
	require.NoError(t, err)

	list2 := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test List 2 " + uuid.New().String()[:8],
		Description:   "Test Description 2",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &user.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err = repo.Create(list2)
	require.NoError(t, err)

	// Create multiple conflicts for list1
	createTestConflict := func(listID uuid.UUID, conflictType string, index int) error {
		localData := map[string]interface{}{
			"name":  "Local Name " + conflictType + " " + string(rune('A'+index)),
			"index": index,
		}

		remoteData := map[string]interface{}{
			"name":  "Remote Name " + conflictType + " " + string(rune('A'+index)),
			"index": index + 100,
		}

		conflict := &models.SyncConflict{
			ID:         uuid.New(), // Set a valid ID
			ListID:     listID,
			Type:       conflictType,
			LocalData:  localData,
			RemoteData: remoteData,
			CreatedAt:  time.Now(), // Set creation time
			UpdatedAt:  time.Now(), // Set update time
		}

		return repo.CreateConflict(conflict)
	}

	// Create 3 conflicts for list1
	err = createTestConflict(list1.ID, "list_metadata", 0)
	require.NoError(t, err)
	err = createTestConflict(list1.ID, "list_items", 1)
	require.NoError(t, err)
	err = createTestConflict(list1.ID, "item_details", 2)
	require.NoError(t, err)

	// Create 1 conflict for list2
	err = createTestConflict(list2.ID, "list_metadata", 0)
	require.NoError(t, err)

	t.Run("get list1 conflicts", func(t *testing.T) {
		conflicts, err := repo.GetConflicts(list1.ID)
		require.NoError(t, err)
		require.Len(t, conflicts, 3, "Should have three conflicts for list1")

		// Check that all conflicts belong to list1
		for _, conflict := range conflicts {
			assert.Equal(t, list1.ID, conflict.ListID)
			assert.NotNil(t, conflict.LocalData)
			assert.NotNil(t, conflict.RemoteData)
		}

		// Check conflict types - should be sorted by created_at DESC
		assert.Equal(t, "item_details", conflicts[0].Type)
		assert.Equal(t, "list_items", conflicts[1].Type)
		assert.Equal(t, "list_metadata", conflicts[2].Type)
	})

	t.Run("get list2 conflicts", func(t *testing.T) {
		conflicts, err := repo.GetConflicts(list2.ID)
		require.NoError(t, err)
		require.Len(t, conflicts, 1, "Should have one conflict for list2")

		assert.Equal(t, list2.ID, conflicts[0].ListID)
		assert.Equal(t, "list_metadata", conflicts[0].Type)
	})

	t.Run("no conflicts", func(t *testing.T) {
		nonExistentID := uuid.New()
		conflicts, err := repo.GetConflicts(nonExistentID)
		require.NoError(t, err)
		assert.Empty(t, conflicts, "Should return empty slice for list with no conflicts")
	})
}

func TestListRepository_ResolveConflict(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))
	userRepo := NewUserRepository(testutil.UnwrapDB(db))

	// Create a test user
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create a test list
	ownerType := models.OwnerTypeUser
	list := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test Resolve List " + uuid.New().String()[:8],
		Description:   "Test Description",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		OwnerID:       &user.ID,
		OwnerType:     &ownerType,
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}

	err = repo.Create(list)
	require.NoError(t, err)

	// Create a conflict
	localData := map[string]interface{}{
		"name": "Local Name",
	}

	remoteData := map[string]interface{}{
		"name": "Remote Name",
	}

	conflict := &models.SyncConflict{
		ID:         uuid.New(), // Set a valid ID
		ListID:     list.ID,
		Type:       "name_conflict",
		LocalData:  localData,
		RemoteData: remoteData,
		CreatedAt:  time.Now(), // Set creation time
		UpdatedAt:  time.Now(), // Set update time
	}

	err = repo.CreateConflict(conflict)
	require.NoError(t, err)

	t.Run("resolve conflict", func(t *testing.T) {
		// Resolve the conflict
		err = repo.ResolveConflict(conflict.ID)
		require.NoError(t, err)

		// Verify conflict is resolved
		conflicts, err := repo.GetConflicts(list.ID)
		require.NoError(t, err)
		assert.Empty(t, conflicts, "Should have no active conflicts after resolution")
	})

	t.Run("non-existent conflict", func(t *testing.T) {
		err := repo.ResolveConflict(uuid.New())
		assert.ErrorIs(t, err, models.ErrNotFound)
	})

	t.Run("already resolved conflict", func(t *testing.T) {
		// Try to resolve the already resolved conflict
		err = repo.ResolveConflict(conflict.ID)
		assert.ErrorIs(t, err, models.ErrNotFound)
	})
}

// setupTestListSync creates tables directly with hard-coded schema name for tests
func setupTestListSync(t *testing.T, schemaName string) (*sql.DB, func()) {
	t.Helper()

	// Use testutil's SetupTestDB instead of direct DB creation
	db := testutil.SetupTestDB(t)
	sqlDB := testutil.UnwrapDB(db)

	// Return a cleanup function that uses proper teardown
	cleanup := func() {
		testutil.TeardownTestDB(t, db)
	}

	return sqlDB, cleanup
}

func TestListRepository_GetListsBySyncSource(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	// Use the repository for testing GetListsBySource
	listRepo := NewListRepository(testutil.UnwrapDB(db))

	// Get the schema name directly from the database
	var schemaName string
	err := db.QueryRow("SELECT current_schema()").Scan(&schemaName)
	require.NoError(t, err)
	t.Logf("Current schema: %s", schemaName)

	// Create a test user directly with SQL
	userID := uuid.New()
	firebaseUID := "test-firebase-" + uuid.New().String()[:8]
	email := "test-" + uuid.New().String()[:8] + "@example.com"
	name := "Test User " + uuid.New().String()[:8]
	now := time.Now()

	// Use fully qualified table names
	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO %s.users (
			id, firebase_uid, provider, email, name, 
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7
		)`, schemaName),
		userID, firebaseUID, "google", email, name,
		now, now,
	)
	require.NoError(t, err)

	// Create lists with different sync sources using direct SQL with schema
	insertTestList := func(name, syncSource, syncID string, syncStatus models.ListSyncStatus) uuid.UUID {
		listID := uuid.New()
		now := time.Now()

		_, err := db.Exec(fmt.Sprintf(`
			INSERT INTO %s.lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id,
				default_weight, owner_id, owner_type,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11,
				$12, $13
			)`, schemaName),
			listID, models.ListTypeActivity,
			name+" "+uuid.New().String()[:8],
			"Description for "+name,
			models.VisibilityPrivate,
			syncStatus, syncSource, syncID,
			1.0, userID, models.OwnerTypeUser,
			now, now,
		)
		require.NoError(t, err)

		// Add owner
		_, err = db.Exec(fmt.Sprintf(`
			INSERT INTO %s.list_owners (
				list_id, owner_id, owner_type,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5)`, schemaName),
			listID, userID, models.OwnerTypeUser,
			now, now,
		)
		require.NoError(t, err)

		return listID
	}

	// Create lists with different sync sources
	googleList1ID := insertTestList("Google List 1", string(models.SyncSourceGoogleMaps), "google_place_1", models.ListSyncStatusSynced)
	googleList2ID := insertTestList("Google List 2", string(models.SyncSourceGoogleMaps), "google_place_2", models.ListSyncStatusSynced)
	manualListID := insertTestList("Manual List", string(models.SyncSourceManual), "manual_1", models.ListSyncStatusSynced)
	importedListID := insertTestList("Imported List", string(models.SyncSourceImported), "import_1", models.ListSyncStatusSynced)
	noSyncListID := insertTestList("No Sync List", string(models.SyncSourceNone), "", models.ListSyncStatusNone)

	t.Run("filter by google maps source", func(t *testing.T) {
		lists, err := listRepo.GetListsBySource(string(models.SyncSourceGoogleMaps))
		require.NoError(t, err)
		require.Len(t, lists, 2, "Should have two Google Maps lists")

		// Check IDs match what we expect
		foundList1 := false
		foundList2 := false
		for _, list := range lists {
			if list.ID == googleList1ID {
				foundList1 = true
			}
			if list.ID == googleList2ID {
				foundList2 = true
			}
		}

		assert.True(t, foundList1, "Google List 1 not found")
		assert.True(t, foundList2, "Google List 2 not found")
	})

	t.Run("filter by manual source", func(t *testing.T) {
		lists, err := listRepo.GetListsBySource(string(models.SyncSourceManual))
		require.NoError(t, err)
		require.Len(t, lists, 1, "Should have one Manual list")
		assert.Equal(t, manualListID, lists[0].ID)
	})

	t.Run("filter by imported source", func(t *testing.T) {
		lists, err := listRepo.GetListsBySource(string(models.SyncSourceImported))
		require.NoError(t, err)
		require.Len(t, lists, 1, "Should have one Imported list")
		assert.Equal(t, importedListID, lists[0].ID)
	})

	t.Run("filter by none source", func(t *testing.T) {
		lists, err := listRepo.GetListsBySource(string(models.SyncSourceNone))
		require.NoError(t, err)
		require.Len(t, lists, 1, "Should have one unsynchronized list")
		assert.Equal(t, noSyncListID, lists[0].ID)
	})

	t.Run("filter by unknown source", func(t *testing.T) {
		lists, err := listRepo.GetListsBySource("unknown_source")
		require.NoError(t, err)
		assert.Empty(t, lists, "Should return empty slice for unknown source")
	})

	// Keep the original repository method test
	t.Run("get google maps lists via repository", func(t *testing.T) {
		lists, err := listRepo.GetListsBySource(string(models.SyncSourceGoogleMaps))
		require.NoError(t, err)
		require.Len(t, lists, 2, "Should have two Google Maps lists")

		// Additional checks can be added here to verify the lists
	})
}

func TestListRepository_CreateConflict_Extra(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewListRepository(testutil.UnwrapDB(db))

	// Create a test user directly with SQL
	userID := uuid.New()
	firebaseUID := "test-firebase-" + uuid.New().String()[:8]
	email := "test-" + uuid.New().String()[:8] + "@example.com"
	name := "Test User " + uuid.New().String()[:8]
	now := time.Now()

	_, err := db.Exec(`
		INSERT INTO users (
			id, firebase_uid, provider, email, name, 
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7
		)`,
		userID, firebaseUID, "google", email, name,
		now, now,
	)
	require.NoError(t, err)

	// Create a test list directly with SQL
	listID := uuid.New()
	listName := "Test AddConflict List " + uuid.New().String()[:8]

	_, err = db.Exec(`
		INSERT INTO lists (
			id, type, name, description, visibility,
			sync_status, sync_source, sync_id,
			default_weight, owner_id, owner_type,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13
		)`,
		listID, models.ListTypeActivity,
		listName,
		"Test Description",
		models.VisibilityPrivate,
		models.ListSyncStatusNone, models.SyncSourceNone, "",
		1.0, userID, models.OwnerTypeUser,
		now, now,
	)
	require.NoError(t, err)

	// Add owner
	_, err = db.Exec(`
		INSERT INTO list_owners (
			list_id, owner_id, owner_type,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)`,
		listID, userID, models.OwnerTypeUser,
		now, now,
	)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		// Create a list conflict
		localData := map[string]interface{}{
			"name": "Local Name for AddConflict",
		}

		remoteData := map[string]interface{}{
			"name": "Remote Name for AddConflict",
		}

		conflict := &models.SyncConflict{
			ID:         uuid.New(), // Set a valid ID
			ListID:     listID,
			Type:       "name_conflict_add",
			LocalData:  localData,
			RemoteData: remoteData,
			CreatedAt:  time.Now(), // Set creation time
			UpdatedAt:  time.Now(), // Set update time
		}

		// Create conflict using CreateConflict
		err = repo.CreateConflict(conflict)
		require.NoError(t, err)

		// Verify conflict created successfully
		conflicts, err := repo.GetConflicts(listID)
		require.NoError(t, err)
		require.Len(t, conflicts, 1, "Should have one conflict")

		// Check conflict data
		assert.Equal(t, listID, conflicts[0].ListID)
		assert.Equal(t, "name_conflict_add", conflicts[0].Type)

		// Convert to JSON for comparing complex data
		expectedLocalJSON, err := json.Marshal(localData)
		require.NoError(t, err)
		actualLocalJSON, err := json.Marshal(conflicts[0].LocalData)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedLocalJSON), string(actualLocalJSON))

		expectedRemoteJSON, err := json.Marshal(remoteData)
		require.NoError(t, err)
		actualRemoteJSON, err := json.Marshal(conflicts[0].RemoteData)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedRemoteJSON), string(actualRemoteJSON))
	})

	t.Run("missing data", func(t *testing.T) {
		conflict := &models.SyncConflict{
			ID:     uuid.New(), // Set a valid ID
			ListID: listID,
			Type:   "invalid_conflict",
			// Missing LocalData and RemoteData
			CreatedAt: time.Now(), // Set creation time
			UpdatedAt: time.Now(), // Set update time
		}

		err = repo.CreateConflict(conflict)
		assert.Error(t, err)
	})
}
