package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestWithSchema sets up a test database with the schema path properly configured
func setupTestWithSchema(t *testing.T) (*testutil.SchemaDB, func()) {
	t.Helper()

	// Setup test database with schema
	schemaDB := testutil.SetupTestDB(t)
	schemaName := schemaDB.GetSchemaName()

	// Ensure the schema context exists for this schema
	if ctx := testutil.GetSchemaContext(schemaName); ctx == nil {
		// Create a new schema context and register it
		newCtx := testutil.NewSchemaContext(schemaName)
		testutil.SchemaContextMutex.Lock()
		testutil.SchemaContexts[schemaName] = newCtx
		testutil.SchemaContextMutex.Unlock()
	}

	// Log for debugging
	t.Logf("setupTestWithSchema: Using schema %s with context", schemaName)

	// Create a teardown function
	teardown := func() {
		testutil.TeardownTestDB(t, schemaDB)
	}

	return schemaDB, teardown
}

func TestListRepository_ShareWithTribe(t *testing.T) {
	// Use our custom setup function
	schemaDB, teardown := setupTestWithSchema(t)
	defer teardown()

	// Get schema context for testing
	schemaName := schemaDB.GetSchemaName()
	schemaCtx := testutil.GetSchemaContext(schemaName)
	if schemaCtx == nil {
		t.Fatalf("Schema context not found for schema %s", schemaName)
	}

	// Use the context with schema information
	testCtx := schemaCtx.WithContext(context.Background())

	// Get raw DB for repository constructors
	rawDB := testutil.UnwrapDB(schemaDB)

	// Create repositories using the raw DB with schema path set
	listRepo := NewListRepository(rawDB)
	userRepo := NewUserRepository(rawDB)
	tribeRepo := NewTribeRepository(rawDB)

	// Create a test user
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create a test tribe
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Tribe " + uuid.New().String()[:8],
		Type:        models.TribeTypeCouple,
		Description: "A test tribe",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}
	err = tribeRepo.Create(tribe)
	require.NoError(t, err)

	// Create a list
	maxItems := 100
	cooldownDays := 7
	list := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test Share List " + uuid.New().String()[:8],
		Description:   "Test Description",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}
	err = listRepo.Create(list)
	require.NoError(t, err)

	t.Run("with_expiry", func(t *testing.T) {
		// Create a share with expiration date
		expiresAt := time.Now().Add(24 * time.Hour)
		share := &models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: &expiresAt,
		}

		err := listRepo.ShareWithTribe(share)
		require.NoError(t, err)

		// Verify share was created
		var shareExists bool
		err = schemaDB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM list_sharing 
				WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
			)`,
			list.ID, tribe.ID,
		).Scan(&shareExists)
		require.NoError(t, err)
		assert.True(t, shareExists, "share record should exist")

		// Verify expiration date was set
		var storedExpiresAt time.Time
		err = schemaDB.QueryRow(`
			SELECT expires_at FROM list_sharing
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
		`, list.ID, tribe.ID).Scan(&storedExpiresAt)
		require.NoError(t, err)
		assert.WithinDuration(t, expiresAt, storedExpiresAt, time.Second, "expires_at should match")
	})

	t.Run("update_existing", func(t *testing.T) {
		// Create a new list for this test to avoid duplicate key issues
		maxItems := 100
		cooldownDays := 7
		list2 := &models.List{
			Type:          models.ListTypeActivity,
			Name:          "Test Update Share List " + uuid.New().String()[:8],
			Description:   "Test Description for update",
			Visibility:    models.VisibilityPrivate,
			DefaultWeight: 1.0,
			MaxItems:      &maxItems,
			CooldownDays:  &cooldownDays,
			SyncStatus:    models.ListSyncStatusNone,
			SyncSource:    models.SyncSourceNone,
			SyncID:        "",
			Owners: []*models.ListOwner{
				{
					OwnerID:   user.ID,
					OwnerType: models.OwnerTypeUser,
				},
			},
		}
		err := listRepo.Create(list2)
		require.NoError(t, err)

		// Create initial share
		initialExpiresAt := time.Now().Add(24 * time.Hour)
		initialShare := &models.ListShare{
			ListID:    list2.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: &initialExpiresAt,
		}

		err = listRepo.ShareWithTribe(initialShare)
		require.NoError(t, err)

		// Verify share was created
		var shareExists bool
		err = schemaDB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM list_sharing 
				WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
			)`,
			list2.ID, tribe.ID,
		).Scan(&shareExists)
		require.NoError(t, err)
		assert.True(t, shareExists, "share record should exist")

		// Get the initial version
		var initialVersion int
		err = schemaDB.QueryRow(`
			SELECT version FROM list_sharing
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
		`, list2.ID, tribe.ID).Scan(&initialVersion)
		require.NoError(t, err)

		// Update the share with a new expiration
		newExpiresAt := time.Now().Add(48 * time.Hour)
		updatedShare := &models.ListShare{
			ListID:    list2.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: &newExpiresAt,
		}

		// Now we can do the share update in a single call
		err = listRepo.ShareWithTribe(updatedShare)
		require.NoError(t, err)

		// Verify share still exists and was updated (not soft-deleted)
		err = schemaDB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM list_sharing 
				WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
			)`,
			list2.ID, tribe.ID,
		).Scan(&shareExists)
		require.NoError(t, err)
		assert.True(t, shareExists, "share record should still exist and not be soft-deleted")

		// Verify no soft-deleted shares exist
		var deletedShareExists bool
		err = schemaDB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM list_sharing 
				WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NOT NULL
			)`,
			list2.ID, tribe.ID,
		).Scan(&deletedShareExists)
		require.NoError(t, err)
		assert.False(t, deletedShareExists, "no soft-deleted share records should exist")

		// Verify share has been updated with new expiration date
		var storedExpiresAt time.Time
		var version int
		err = schemaDB.QueryRow(`
			SELECT expires_at, version FROM list_sharing
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
		`, list2.ID, tribe.ID).Scan(&storedExpiresAt, &version)
		require.NoError(t, err)
		assert.WithinDuration(t, newExpiresAt, storedExpiresAt, time.Second, "expires_at should be updated")
		assert.Equal(t, initialVersion+1, version, "version should be incremented")
	})

	t.Run("null_expiry_date", func(t *testing.T) {
		// Create a new list for this test
		maxItems := 100
		cooldownDays := 7
		list3 := &models.List{
			Type:          models.ListTypeActivity,
			Name:          "Test Null Expiry List " + uuid.New().String()[:8],
			Description:   "Test Description for null expiry",
			Visibility:    models.VisibilityPrivate,
			DefaultWeight: 1.0,
			MaxItems:      &maxItems,
			CooldownDays:  &cooldownDays,
			SyncStatus:    models.ListSyncStatusNone,
			SyncSource:    models.SyncSourceNone,
			SyncID:        "",
			Owners: []*models.ListOwner{
				{
					OwnerID:   user.ID,
					OwnerType: models.OwnerTypeUser,
				},
			},
		}
		err := listRepo.Create(list3)
		require.NoError(t, err)

		// Create share with null expiry date
		share := &models.ListShare{
			ListID:    list3.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: nil, // No expiration date
		}

		err = listRepo.ShareWithTribe(share)
		require.NoError(t, err)

		// Verify share was created with null expiry
		var expiresAt *time.Time
		err = schemaDB.QueryRow(`
			SELECT expires_at FROM list_sharing
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
		`, list3.ID, tribe.ID).Scan(&expiresAt)
		require.NoError(t, err)
		assert.Nil(t, expiresAt, "expires_at should be NULL")
	})

	t.Run("list_not_found", func(t *testing.T) {
		share := &models.ListShare{
			ListID:    uuid.New(), // Non-existent list ID
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: nil,
		}

		err := listRepo.ShareWithTribe(share)
		assert.ErrorIs(t, err, models.ErrNotFound)
	})

	t.Run("tribe_not_found", func(t *testing.T) {
		share := &models.ListShare{
			ListID:    list.ID,
			TribeID:   uuid.New(), // Non-existent tribe ID
			UserID:    user.ID,
			ExpiresAt: nil,
		}

		err := listRepo.ShareWithTribe(share)
		assert.ErrorIs(t, err, models.ErrNotFound)
	})

	t.Run("user_not_found", func(t *testing.T) {
		share := &models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    uuid.New(), // Non-existent user ID
			ExpiresAt: nil,
		}

		err := listRepo.ShareWithTribe(share)
		assert.ErrorIs(t, err, models.ErrNotFound)
	})

	t.Run("validation_error", func(t *testing.T) {
		// Create a share with nil list ID
		share := &models.ListShare{
			ListID:    uuid.Nil, // Invalid list ID
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: nil,
		}

		err := listRepo.ShareWithTribe(share)
		assert.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("expired_date", func(t *testing.T) {
		// Create a share with a past expiration date
		expiredDate := time.Now().Add(-24 * time.Hour)
		share := &models.ListShare{
			ListID:    list.ID,
			TribeID:   tribe.ID,
			UserID:    user.ID,
			ExpiresAt: &expiredDate,
		}

		err := listRepo.ShareWithTribe(share)
		assert.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	// Verify context usage works
	t.Run("verify context with schema", func(t *testing.T) {
		// Perform a simple read operation using the context
		_, err := listRepo.GetTribeListsWithContext(testCtx, tribe.ID)
		require.NoError(t, err, "Should be able to call repo methods with schema context")
	})
}

func TestListRepository_GetListShares(t *testing.T) {
	// Use our custom setup function
	schemaDB, teardown := setupTestWithSchema(t)
	defer teardown()

	// Get raw DB for repository constructors
	rawDB := testutil.UnwrapDB(schemaDB)

	// Create repositories
	listRepo := NewListRepository(rawDB)
	userRepo := NewUserRepository(rawDB)
	tribeRepo := NewTribeRepository(rawDB)

	// Create a test user
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create test tribes
	tribe1 := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Tribe 1 " + uuid.New().String()[:8],
		Type:        models.TribeTypeCouple,
		Description: "A test tribe",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}
	err = tribeRepo.Create(tribe1)
	require.NoError(t, err)

	tribe2 := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Tribe 2 " + uuid.New().String()[:8],
		Type:        models.TribeTypeCouple,
		Description: "Another test tribe",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}
	err = tribeRepo.Create(tribe2)
	require.NoError(t, err)

	tribe3 := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Tribe 3 " + uuid.New().String()[:8],
		Type:        models.TribeTypeCouple,
		Description: "A third test tribe",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}
	err = tribeRepo.Create(tribe3)
	require.NoError(t, err)

	// Create a list
	maxItems := 100
	cooldownDays := 7
	list := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test List Shares " + uuid.New().String()[:8],
		Description:   "Test Description",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}
	err = listRepo.Create(list)
	require.NoError(t, err)

	// Create a second list to test filtering
	list2 := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Test List 2 " + uuid.New().String()[:8],
		Description:   "Another list",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}
	err = listRepo.Create(list2)
	require.NoError(t, err)

	// Add shares
	now := time.Now()
	expiresAt1 := now.Add(24 * time.Hour)
	expiresAt2 := now.Add(48 * time.Hour)
	expiresAt3 := now.Add(72 * time.Hour)

	// Share with first tribe (created first but with an earlier createdAt timestamp)
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list.ID, tribe1.ID, user.ID, now, now, expiresAt1, 1,
	)
	require.NoError(t, err)

	// Share with second tribe (created second with a later createdAt timestamp)
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list.ID, tribe2.ID, user.ID, now.Add(1*time.Hour), now.Add(1*time.Hour), expiresAt2, 1,
	)
	require.NoError(t, err)

	// Share third tribe with a different version
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list.ID, tribe3.ID, user.ID, now.Add(2*time.Hour), now.Add(2*time.Hour), expiresAt3, 3,
	)
	require.NoError(t, err)

	// Share second list with first tribe
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list2.ID, tribe1.ID, user.ID, now, now, expiresAt1, 1,
	)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		shares, err := listRepo.GetListShares(list.ID)
		require.NoError(t, err)
		require.Len(t, shares, 3, "should return 3 shares")

		// Shares should be ordered by created_at DESC
		assert.Equal(t, tribe3.ID, shares[0].TribeID, "first share should be for tribe3")
		assert.Equal(t, tribe2.ID, shares[1].TribeID, "second share should be for tribe2")
		assert.Equal(t, tribe1.ID, shares[2].TribeID, "third share should be for tribe1")

		// Verify share details
		for _, share := range shares {
			assert.Equal(t, list.ID, share.ListID)
			assert.Equal(t, user.ID, share.UserID)
			assert.NotZero(t, share.CreatedAt)
			assert.NotZero(t, share.UpdatedAt)
			assert.NotNil(t, share.ExpiresAt)
			assert.Nil(t, share.DeletedAt)
			assert.NotZero(t, share.Version, "Version should be set")
		}

		// Verify version values
		assert.Equal(t, 3, shares[0].Version, "tribe3 share should have version 3")
		assert.Equal(t, 1, shares[1].Version, "tribe2 share should have version 1")
		assert.Equal(t, 1, shares[2].Version, "tribe1 share should have version 1")
	})

	t.Run("ordering", func(t *testing.T) {
		// Update the timestamp for tribe1's share to be the most recent
		_, err := schemaDB.Exec(`
			UPDATE list_sharing
			SET created_at = $1, updated_at = $1
			WHERE list_id = $2 AND tribe_id = $3`,
			now.Add(3*time.Hour), list.ID, tribe1.ID,
		)
		require.NoError(t, err)

		shares, err := listRepo.GetListShares(list.ID)
		require.NoError(t, err)
		require.Len(t, shares, 3, "should return 3 shares")

		// Verify new ordering by updated created_at
		assert.Equal(t, tribe1.ID, shares[0].TribeID, "first share should now be for tribe1")
		assert.Equal(t, tribe3.ID, shares[1].TribeID, "second share should be for tribe3")
		assert.Equal(t, tribe2.ID, shares[2].TribeID, "third share should be for tribe2")
	})

	t.Run("no_shares", func(t *testing.T) {
		// Test with a non-existent list ID
		shares, err := listRepo.GetListShares(uuid.New())
		require.NoError(t, err)
		assert.Empty(t, shares, "should return empty slice for list with no shares")
	})

	t.Run("deleted_shares", func(t *testing.T) {
		// Delete one of the shares
		_, err := schemaDB.Exec(`
			UPDATE list_sharing
			SET deleted_at = NOW()
			WHERE list_id = $1 AND tribe_id = $2`,
			list.ID, tribe2.ID,
		)
		require.NoError(t, err)

		// Verify only non-deleted shares are returned
		shares, err := listRepo.GetListShares(list.ID)
		require.NoError(t, err)
		require.Len(t, shares, 2, "should return 2 shares after deletion")

		// The shares for tribe2 should no longer be in the result
		for _, share := range shares {
			assert.NotEqual(t, tribe2.ID, share.TribeID, "tribe2 share should no longer be in the results")
		}
	})

	t.Run("multiple_versions", func(t *testing.T) {
		t.Skip("Skipping due to database constraint issues - needs further investigation")

		// Create a new list for testing multiple versions
		list4 := &models.List{
			Type:          models.ListTypeActivity,
			Name:          "Test Multiple Versions List " + uuid.New().String()[:8],
			Description:   "List for testing multiple versions",
			Visibility:    models.VisibilityPrivate,
			DefaultWeight: 1.0,
			MaxItems:      &maxItems,
			CooldownDays:  &cooldownDays,
			SyncStatus:    models.ListSyncStatusNone,
			SyncSource:    models.SyncSourceNone,
			SyncID:        "",
			Owners: []*models.ListOwner{
				{
					OwnerID:   user.ID,
					OwnerType: models.OwnerTypeUser,
				},
			},
		}
		err := listRepo.Create(list4)
		require.NoError(t, err)

		// Create multiple versions of the same share by directly inserting and soft deleting
		// Each with increasing version numbers

		// Version 1 - soft deleted
		_, err = schemaDB.Exec(`
			INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, deleted_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			list4.ID, tribe1.ID, user.ID, now, now, expiresAt1, now.Add(30*time.Minute), 1,
		)
		require.NoError(t, err)

		// Version 2 - soft deleted
		_, err = schemaDB.Exec(`
			INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, deleted_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			list4.ID, tribe1.ID, user.ID, now.Add(1*time.Hour), now.Add(1*time.Hour),
			expiresAt2, now.Add(2*time.Hour), 2,
		)
		require.NoError(t, err)

		// Version 3 - current active version
		_, err = schemaDB.Exec(`
			INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			list4.ID, tribe1.ID, user.ID, now.Add(3*time.Hour), now.Add(3*time.Hour),
			expiresAt3, 3,
		)
		require.NoError(t, err)

		// Get shares and verify only the active one is returned
		shares, err := listRepo.GetListShares(list4.ID)
		require.NoError(t, err)
		require.Len(t, shares, 1, "should return only 1 active share")
		assert.Equal(t, tribe1.ID, shares[0].TribeID)
		assert.Equal(t, 3, shares[0].Version, "should return version 3")
	})
}

func TestListRepository_GetSharedLists(t *testing.T) {
	// We've fixed the database connection issues with batch loading, so we can now run this test
	// t.Skip("Skipping due to database constraint issues - needs further investigation")

	schemaDB, teardown := setupTestWithSchema(t)
	defer teardown()

	// Get raw DB for repository constructors
	rawDB := testutil.UnwrapDB(schemaDB)

	// Create repositories
	listRepo := NewListRepository(rawDB)
	userRepo := NewUserRepository(rawDB)
	tribeRepo := NewTribeRepository(rawDB)

	// Create a test user
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create another test user for ownership testing
	user2 := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User 2 " + uuid.New().String()[:8],
	}
	err = userRepo.Create(user2)
	require.NoError(t, err)

	// Create test tribes
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Tribe " + uuid.New().String()[:8],
		Type:        models.TribeTypeCouple,
		Description: "A test tribe",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}
	err = tribeRepo.Create(tribe)
	require.NoError(t, err)

	tribe2 := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Second Test Tribe " + uuid.New().String()[:8],
		Type:        models.TribeTypeFamily,
		Description: "Another test tribe",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}
	err = tribeRepo.Create(tribe2)
	require.NoError(t, err)

	// Create lists by inserting directly to avoid validation issues
	now := time.Now()
	list1ID := uuid.New()
	list2ID := uuid.New()
	list3ID := uuid.New()

	// Insert first list
	_, err = schemaDB.Exec(`
		INSERT INTO lists (
			id, type, name, description, visibility,
			sync_status, sync_source, sync_id,
			default_weight, created_at, updated_at,
			owner_id, owner_type
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13
		)`,
		list1ID, models.ListTypeActivity, "Test List 1", "Test Description 1", models.VisibilityPrivate,
		models.ListSyncStatusNone, models.SyncSourceNone, "",
		1.0, now, now,
		user.ID, models.OwnerTypeUser,
	)
	require.NoError(t, err)

	// Add user as owner for list1
	_, err = schemaDB.Exec(`
		INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
		list1ID, user.ID, models.OwnerTypeUser, now, now,
	)
	require.NoError(t, err)

	// Insert second list
	_, err = schemaDB.Exec(`
		INSERT INTO lists (
			id, type, name, description, visibility,
			sync_status, sync_source, sync_id,
			default_weight, created_at, updated_at,
			owner_id, owner_type
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13
		)`,
		list2ID, models.ListTypeActivity, "Test List 2", "Test Description 2", models.VisibilityPrivate,
		models.ListSyncStatusNone, models.SyncSourceNone, "",
		1.0, now, now,
		user.ID, models.OwnerTypeUser,
	)
	require.NoError(t, err)

	// Insert third list with more complex setup (different type, multiple owners)
	_, err = schemaDB.Exec(`
		INSERT INTO lists (
			id, type, name, description, visibility,
			sync_status, sync_source, sync_id,
			default_weight, created_at, updated_at,
			owner_id, owner_type
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13
		)`,
		list3ID, models.ListTypeLocation, "Test Places List", "A list of places", models.VisibilityPrivate,
		models.ListSyncStatusNone, models.SyncSourceNone, "",
		1.0, now, now,
		user.ID, models.OwnerTypeUser,
	)
	require.NoError(t, err)

	// Add user as primary owner for list3
	_, err = schemaDB.Exec(`
		INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
		list3ID, user.ID, models.OwnerTypeUser, now, now,
	)
	require.NoError(t, err)

	// Add user2 as secondary owner for list3
	_, err = schemaDB.Exec(`
		INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
		list3ID, user2.ID, models.OwnerTypeUser, now, now,
	)
	require.NoError(t, err)

	// Add items to list3 for testing item retrieval
	item1ID := uuid.New()
	item2ID := uuid.New()

	// Add first item
	_, err = schemaDB.Exec(`
		INSERT INTO list_items (
			id, list_id, name, description, weight,
			external_id, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9
		)`,
		item1ID, list3ID, "Test Item 1", "First test item", 1.0,
		"", "{}",
		now, now,
	)
	require.NoError(t, err)

	// Add second item with location data
	lat := 37.7749
	long := -122.4194
	address := "San Francisco, CA"

	_, err = schemaDB.Exec(`
		INSERT INTO list_items (
			id, list_id, name, description, weight,
			latitude, longitude, address,
			external_id, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10,
			$11, $12
		)`,
		item2ID, list3ID, "Test Item 2", "Second test item with location", 1.0,
		lat, long, address,
		"", "{}",
		now, now,
	)
	require.NoError(t, err)

	// Add shares with different expiration dates
	futureDate := now.Add(24 * time.Hour)
	pastDate := now.Add(-24 * time.Hour)

	// Share first list with future expiration
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list1ID, tribe.ID, user.ID, now, now, futureDate, 1,
	)
	require.NoError(t, err)

	// Share second list with past expiration
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list2ID, tribe.ID, user.ID, now, now, pastDate, 1,
	)
	require.NoError(t, err)

	// Share third list with tribe2, no expiration
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		list3ID, tribe2.ID, user.ID, now, now, 1,
	)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		sharedLists, err := listRepo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		require.Len(t, sharedLists, 1, "should return only lists with valid expiration")

		// Verify only the first list is returned (future expiration)
		assert.Equal(t, list1ID, sharedLists[0].ID)
		assert.Equal(t, "Test List 1", sharedLists[0].Name)
	})

	t.Run("no_expiration", func(t *testing.T) {
		// Add a new list with no expiration
		list4ID := uuid.New()

		// Insert fourth list
		_, err = schemaDB.Exec(`
			INSERT INTO lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id,
				default_weight, created_at, updated_at,
				owner_id, owner_type
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11,
				$12, $13
			)`,
			list4ID, models.ListTypeActivity, "Test List 4", "Test Description 4", models.VisibilityPrivate,
			models.ListSyncStatusNone, models.SyncSourceNone, "",
			1.0, now, now,
			user.ID, models.OwnerTypeUser,
		)
		require.NoError(t, err)

		// Add user as owner for list4
		_, err = schemaDB.Exec(`
			INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`,
			list4ID, user.ID, models.OwnerTypeUser, now, now,
		)
		require.NoError(t, err)

		// Share with no expiration
		_, err = schemaDB.Exec(`
			INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, version)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			list4ID, tribe.ID, user.ID, now, now, 1,
		)
		require.NoError(t, err)

		sharedLists, err := listRepo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		require.Len(t, sharedLists, 2, "should return lists with future or no expiration")

		// Verify both lists are returned (list1 and list4)
		foundList1, foundList4 := false, false
		for _, list := range sharedLists {
			if list.ID == list1ID {
				foundList1 = true
			}
			if list.ID == list4ID {
				foundList4 = true
			}
		}
		assert.True(t, foundList1, "list1 should be in results")
		assert.True(t, foundList4, "list4 should be in results")
	})

	t.Run("with_items_and_owners", func(t *testing.T) {
		// Test retrieving the complex list with items and multiple owners
		sharedLists, err := listRepo.GetSharedLists(tribe2.ID)
		require.NoError(t, err)
		require.Len(t, sharedLists, 1, "should return 1 list for tribe2")

		list := sharedLists[0]
		assert.Equal(t, list3ID, list.ID, "should be list3")
		assert.Equal(t, models.ListTypeLocation, list.Type, "should have correct type")

		// Verify items are populated
		require.Len(t, list.Items, 2, "should have 2 items")

		// Check item details
		var foundItem1, foundItem2 bool
		for _, item := range list.Items {
			if item.ID == item1ID {
				foundItem1 = true
				assert.Equal(t, "Test Item 1", item.Name)
				assert.Nil(t, item.Latitude, "item1 should have no latitude")
			}
			if item.ID == item2ID {
				foundItem2 = true
				assert.Equal(t, "Test Item 2", item.Name)
				assert.NotNil(t, item.Latitude, "item2 should have latitude")
				assert.Equal(t, lat, *item.Latitude)
				assert.Equal(t, long, *item.Longitude)
				assert.Equal(t, address, *item.Address)
			}
		}
		assert.True(t, foundItem1, "item1 should be found")
		assert.True(t, foundItem2, "item2 should be found")

		// Verify owners are populated
		require.Len(t, list.Owners, 2, "should have 2 owners")

		// Check owner details
		var foundUser1, foundUser2 bool
		for _, owner := range list.Owners {
			if owner.OwnerID == user.ID {
				foundUser1 = true
				assert.Equal(t, models.OwnerTypeUser, owner.OwnerType)
			}
			if owner.OwnerID == user2.ID {
				foundUser2 = true
				assert.Equal(t, models.OwnerTypeUser, owner.OwnerType)
			}
		}
		assert.True(t, foundUser1, "user1 should be an owner")
		assert.True(t, foundUser2, "user2 should be an owner")
	})

	t.Run("deleted_list", func(t *testing.T) {
		// Clean up any existing list sharing records to start with a clean state
		_, err := schemaDB.Exec(`
			UPDATE list_sharing
			SET deleted_at = NOW()
			WHERE tribe_id = $1 AND deleted_at IS NULL`,
			tribe.ID,
		)
		require.NoError(t, err)

		// Delete the first list and make sure it's fully soft-deleted
		_, err = schemaDB.Exec(`
			UPDATE lists
			SET deleted_at = NOW()
			WHERE id = $1`,
			list1ID,
		)
		require.NoError(t, err)

		// Also make sure any list_sharing records are soft-deleted
		_, err = schemaDB.Exec(`
			UPDATE list_sharing
			SET deleted_at = NOW()
			WHERE list_id = $1`,
			list1ID,
		)
		require.NoError(t, err)

		// Add a list with no expiration for this test
		list4ID := uuid.New()

		// Insert fourth list
		_, err = schemaDB.Exec(`
			INSERT INTO lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id,
				default_weight, created_at, updated_at,
				owner_id, owner_type
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11,
				$12, $13
			)`,
			list4ID, models.ListTypeActivity, "Test List 4", "Test Description 4", models.VisibilityPrivate,
			models.ListSyncStatusNone, models.SyncSourceNone, "",
			1.0, now, now,
			user.ID, models.OwnerTypeUser,
		)
		require.NoError(t, err)

		// Add user as owner for list4
		_, err = schemaDB.Exec(`
			INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`,
			list4ID, user.ID, models.OwnerTypeUser, now, now,
		)
		require.NoError(t, err)

		// Share with no expiration
		_, err = schemaDB.Exec(`
			INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, version)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			list4ID, tribe.ID, user.ID, now, now, 1,
		)
		require.NoError(t, err)

		sharedLists, err := listRepo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		require.Len(t, sharedLists, 1, "should only return non-deleted lists")

		// The only list should be list4 (list1 deleted, list2 expired, list3 not shared with this tribe)
		assert.Equal(t, list4ID, sharedLists[0].ID, "list4 should be the only returned list")
	})

	t.Run("no_shared_lists", func(t *testing.T) {
		// Test with a tribe that has no shared lists
		newTribeID := uuid.New()
		sharedLists, err := listRepo.GetSharedLists(newTribeID)
		require.NoError(t, err)
		assert.Empty(t, sharedLists, "should return empty slice for tribe with no shared lists")
	})

	t.Run("deleted_sharing", func(t *testing.T) {
		// Create a new list
		list5ID := uuid.New()

		// Insert fifth list
		_, err = schemaDB.Exec(`
			INSERT INTO lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id,
				default_weight, created_at, updated_at,
				owner_id, owner_type
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11,
				$12, $13
			)`,
			list5ID, models.ListTypeActivity, "Test List 5", "Test Description 5", models.VisibilityPrivate,
			models.ListSyncStatusNone, models.SyncSourceNone, "",
			1.0, now, now,
			user.ID, models.OwnerTypeUser,
		)
		require.NoError(t, err)

		// Add user as owner for list5
		_, err = schemaDB.Exec(`
			INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`,
			list5ID, user.ID, models.OwnerTypeUser, now, now,
		)
		require.NoError(t, err)

		// Add a share
		_, err = schemaDB.Exec(`
			INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, version)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			list5ID, tribe.ID, user.ID, now, now, 1,
		)
		require.NoError(t, err)

		// Delete the share
		_, err = schemaDB.Exec(`
			UPDATE list_sharing 
			SET deleted_at = NOW() 
			WHERE list_id = $1 AND tribe_id = $2`,
			list5ID, tribe.ID,
		)
		require.NoError(t, err)

		// Verify the list is not returned as shared
		sharedLists, err := listRepo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		for _, list := range sharedLists {
			assert.NotEqual(t, list5ID, list.ID, "list with deleted sharing should not be returned")
		}
	})

	t.Run("ordering", func(t *testing.T) {
		// Create three new lists to test ordering
		list6ID := uuid.New()
		list7ID := uuid.New()
		list8ID := uuid.New()

		// Create tribe for this test
		orderingTribe := &models.Tribe{
			BaseModel: models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:        "Ordering Test Tribe " + uuid.New().String()[:8],
			Type:        models.TribeTypeCouple,
			Description: "Tribe for testing ordering",
			Visibility:  models.VisibilityPrivate,
			Metadata:    models.JSONMap{},
		}
		err := tribeRepo.Create(orderingTribe)
		require.NoError(t, err)

		// Create lists with different creation dates
		twoHoursAgo := now.Add(-2 * time.Hour)
		oneHourAgo := now.Add(-1 * time.Hour)
		oneHourLater := now.Add(1 * time.Hour)

		// Insert lists with different timestamps
		for id, timestamp := range map[uuid.UUID]time.Time{
			list6ID: twoHoursAgo,
			list7ID: oneHourAgo,
			list8ID: oneHourLater,
		} {
			// Insert list
			_, err = schemaDB.Exec(`
				INSERT INTO lists (
					id, type, name, description, visibility,
					sync_status, sync_source, sync_id,
					default_weight, created_at, updated_at,
					owner_id, owner_type
				) VALUES (
					$1, $2, $3, $4, $5,
					$6, $7, $8,
					$9, $10, $11,
					$12, $13
				)`,
				id, models.ListTypeActivity, "Ordering Test List "+id.String()[:8], "For testing ordering", models.VisibilityPrivate,
				models.ListSyncStatusNone, models.SyncSourceNone, "",
				1.0, timestamp, timestamp,
				user.ID, models.OwnerTypeUser,
			)
			require.NoError(t, err)

			// Add owner
			_, err = schemaDB.Exec(`
				INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)`,
				id, user.ID, models.OwnerTypeUser, timestamp, timestamp,
			)
			require.NoError(t, err)

			// Share with tribe
			_, err = schemaDB.Exec(`
				INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, version)
				VALUES ($1, $2, $3, $4, $5, $6)`,
				id, orderingTribe.ID, user.ID, timestamp, timestamp, 1,
			)
			require.NoError(t, err)
		}

		// Retrieve shared lists and verify order (should be descending by created_at)
		sharedLists, err := listRepo.GetSharedLists(orderingTribe.ID)
		require.NoError(t, err)
		require.Len(t, sharedLists, 3, "should return all 3 lists")

		// Lists should be ordered newest first
		assert.Equal(t, list8ID, sharedLists[0].ID, "newest list should be first")
		assert.Equal(t, list7ID, sharedLists[1].ID, "middle list should be second")
		assert.Equal(t, list6ID, sharedLists[2].ID, "oldest list should be last")
	})

	t.Run("expired_date_boundary", func(t *testing.T) {
		// Test boundary case with exactly current time expiration
		boundaryList := uuid.New()

		// Insert boundary test list
		_, err = schemaDB.Exec(`
			INSERT INTO lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id,
				default_weight, created_at, updated_at,
				owner_id, owner_type
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11,
				$12, $13
			)`,
			boundaryList, models.ListTypeActivity, "Boundary Test List", "For testing expiration boundary", models.VisibilityPrivate,
			models.ListSyncStatusNone, models.SyncSourceNone, "",
			1.0, now, now,
			user.ID, models.OwnerTypeUser,
		)
		require.NoError(t, err)

		// Add owner
		_, err = schemaDB.Exec(`
			INSERT INTO list_owners (list_id, owner_id, owner_type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`,
			boundaryList, user.ID, models.OwnerTypeUser, now, now,
		)
		require.NoError(t, err)

		// Share with tribe using exact now() as expiration
		expiresNow := time.Now()
		_, err = schemaDB.Exec(`
			INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			boundaryList, tribe.ID, user.ID, now, now, expiresNow, 1,
		)
		require.NoError(t, err)

		// The list should not be returned as it's expired
		sharedLists, err := listRepo.GetSharedLists(tribe.ID)
		require.NoError(t, err)

		for _, list := range sharedLists {
			assert.NotEqual(t, boundaryList, list.ID, "boundary list should not be returned (expired)")
		}
	})
}

func TestListRepository_CleanupExpiredShares(t *testing.T) {
	// Use custom setup function
	schemaDB, teardown := setupTestWithSchema(t)
	defer teardown()

	// Get schema context for testing
	schemaName := schemaDB.GetSchemaName()
	schemaCtx := testutil.GetSchemaContext(schemaName)
	if schemaCtx == nil {
		t.Fatalf("Schema context not found for schema %s", schemaName)
	}

	// Use the context with schema information
	testCtx := schemaCtx.WithContext(context.Background())

	// Get raw DB for repository constructors
	rawDB := testutil.UnwrapDB(schemaDB)

	// Create repositories
	listRepo := NewListRepository(rawDB)
	userRepo := NewUserRepository(rawDB)
	tribeRepo := NewTribeRepository(rawDB)

	// Create test user
	user := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User for Cleanup " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	// Create test tribe
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Tribe for Cleanup " + uuid.New().String()[:8],
		Type:        models.TribeTypeCouple,
		Description: "A test tribe for cleanup tests",
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{},
	}
	err = tribeRepo.Create(tribe)
	require.NoError(t, err)

	// Create lists
	maxItems := 100
	cooldownDays := 7
	now := time.Now()

	// Create first list
	list1 := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Cleanup Test List 1 " + uuid.New().String()[:8],
		Description:   "List with expired share",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}
	err = listRepo.Create(list1)
	require.NoError(t, err)

	// Create second list
	list2 := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Cleanup Test List 2 " + uuid.New().String()[:8],
		Description:   "List with future expiry",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}
	err = listRepo.Create(list2)
	require.NoError(t, err)

	// Create shares with different expiry dates
	// List 1: Share with past expiry date
	pastDate := now.Add(-24 * time.Hour)
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list1.ID, tribe.ID, user.ID, now, now, pastDate, 1,
	)
	require.NoError(t, err)

	// List 2: Share with future expiry date
	futureDate := now.Add(24 * time.Hour)
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, expires_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		list2.ID, tribe.ID, user.ID, now, now, futureDate, 1,
	)
	require.NoError(t, err)

	// Create a third list with no expiry
	list3 := &models.List{
		Type:          models.ListTypeActivity,
		Name:          "Cleanup Test List 3 " + uuid.New().String()[:8],
		Description:   "List with no expiry date",
		Visibility:    models.VisibilityPrivate,
		DefaultWeight: 1.0,
		MaxItems:      &maxItems,
		CooldownDays:  &cooldownDays,
		SyncStatus:    models.ListSyncStatusNone,
		SyncSource:    models.SyncSourceNone,
		SyncID:        "",
		Owners: []*models.ListOwner{
			{
				OwnerID:   user.ID,
				OwnerType: models.OwnerTypeUser,
			},
		},
	}
	err = listRepo.Create(list3)
	require.NoError(t, err)

	// Share with no expiry date
	_, err = schemaDB.Exec(`
		INSERT INTO list_sharing (list_id, tribe_id, user_id, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		list3.ID, tribe.ID, user.ID, now, now, 1,
	)
	require.NoError(t, err)

	t.Run("cleanup_and_verify_lists", func(t *testing.T) {
		// Verify we have one expired share, one valid share, and one with no expiry
		var countBefore int
		err = schemaDB.QueryRow(`
			SELECT COUNT(*) FROM list_sharing 
			WHERE deleted_at IS NULL
		`).Scan(&countBefore)
		require.NoError(t, err)
		assert.Equal(t, 3, countBefore, "should have 3 active shares before cleanup")

		// Run the cleanup function
		err = listRepo.CleanupExpiredShares(testCtx)
		require.NoError(t, err)

		// Verify that the expired share is now soft-deleted
		var isPastShareDeleted bool
		err = schemaDB.QueryRow(`
			SELECT deleted_at IS NOT NULL 
			FROM list_sharing 
			WHERE list_id = $1 AND tribe_id = $2
		`, list1.ID, tribe.ID).Scan(&isPastShareDeleted)
		require.NoError(t, err)
		assert.True(t, isPastShareDeleted, "expired share should be soft-deleted")

		// Verify that future and null expiry shares are not deleted
		var isFutureShareActive, isNoExpiryShareActive bool
		err = schemaDB.QueryRow(`
			SELECT deleted_at IS NULL 
			FROM list_sharing 
			WHERE list_id = $1 AND tribe_id = $2
		`, list2.ID, tribe.ID).Scan(&isFutureShareActive)
		require.NoError(t, err)
		assert.True(t, isFutureShareActive, "future-dated share should still be active")

		err = schemaDB.QueryRow(`
			SELECT deleted_at IS NULL 
			FROM list_sharing 
			WHERE list_id = $1 AND tribe_id = $2
		`, list3.ID, tribe.ID).Scan(&isNoExpiryShareActive)
		require.NoError(t, err)
		assert.True(t, isNoExpiryShareActive, "share with no expiry should still be active")

		// Check total active shares after cleanup
		var countAfter int
		err = schemaDB.QueryRow(`
			SELECT COUNT(*) FROM list_sharing 
			WHERE deleted_at IS NULL
		`).Scan(&countAfter)
		require.NoError(t, err)
		assert.Equal(t, 2, countAfter, "should have 2 active shares after cleanup")

		// Check if version was incremented for the expired share
		var version int
		err = schemaDB.QueryRow(`
			SELECT version FROM list_sharing 
			WHERE list_id = $1 AND tribe_id = $2
		`, list1.ID, tribe.ID).Scan(&version)
		require.NoError(t, err)
		assert.Equal(t, 2, version, "version should be incremented for soft-deleted share")

		// PART 2: Test GetSharedLists after cleanup
		// After cleanup, GetSharedLists should only return the non-expired shares
		sharedLists, err := listRepo.GetSharedLists(tribe.ID)
		require.NoError(t, err)
		require.Len(t, sharedLists, 2, "should return only lists with valid or no expiration")

		// Verify we only have list2 and list3
		foundList2, foundList3 := false, false
		for _, list := range sharedLists {
			if list.ID == list2.ID {
				foundList2 = true
			}
			if list.ID == list3.ID {
				foundList3 = true
			}
			// Should never find list1 as it's expired
			assert.NotEqual(t, list1.ID, list.ID, "list with expired share should not be returned")
		}
		assert.True(t, foundList2, "list with future expiry should be returned")
		assert.True(t, foundList3, "list with no expiry should be returned")
	})
}
