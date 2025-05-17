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

func TestActivitySharing(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	user1 := testutil.CreateTestUser(t, db)
	// Create second user for future test cases if needed
	_ = testutil.CreateTestUser(t, db)

	// Create test tribe
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		Name:       "Test Tribe",
		Type:       models.TribeTypeCouple,
		Visibility: models.VisibilityPrivate,
	}
	_, err := db.Exec(`
		INSERT INTO tribes (id, name, type, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, tribe.ID, tribe.Name, tribe.Type, tribe.Visibility)
	require.NoError(t, err)

	t.Run("Share_Activity_with_Tribe", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			ID:          uuid.New(),
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Activity",
			Description: "Test Description",
			Visibility:  models.VisibilityShared,
			Metadata:    models.JSONMap{"test": "data"},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Share activity with tribe
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, nil)
		assert.NoError(t, err)

		// Get shared activities
		shared, err := repo.GetSharedActivities(tribe.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, shared, "Should have shared activities")

		// Test that the activity is in the shared list
		var found bool
		for _, a := range shared {
			if a.ID == activity.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "The shared activity should be found in the list")
	})

	t.Run("Share_Activity_with_Expiration", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			ID:          uuid.New(),
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Activity with Expiration",
			Description: "Test Description",
			Visibility:  models.VisibilityShared,
			Metadata:    models.JSONMap{"test": "data"},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Share activity with tribe with expiration
		expires := time.Now().Add(24 * time.Hour)
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, &expires)
		assert.NoError(t, err)

		// Get shared activities
		shared, err := repo.GetSharedActivities(tribe.ID)
		assert.NoError(t, err)

		// Test that the activity is in the shared list
		var found bool
		for _, a := range shared {
			if a.ID == activity.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "The shared activity with expiration should be found in the list")
	})

	t.Run("Unshare_Activity", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			ID:          uuid.New(),
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Activity to Unshare",
			Description: "Test Description",
			Visibility:  models.VisibilityShared,
			Metadata:    models.JSONMap{"test": "data"},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Share activity with tribe
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, nil)
		assert.NoError(t, err)

		// Unshare activity
		err = repo.UnshareWithTribe(activity.ID, tribe.ID)
		assert.NoError(t, err)

		// Verify the activity is no longer shared
		shared, err := repo.GetSharedActivities(tribe.ID)
		assert.NoError(t, err)

		// Activity should not be in the shared list
		var found bool
		for _, a := range shared {
			if a.ID == activity.ID {
				found = true
				break
			}
		}
		assert.False(t, found, "The unshared activity should not be found in the list")
	})

	t.Run("Share_Activity_with_Multiple_Tribes", func(t *testing.T) {
		// Create another test tribe
		tribe2 := &models.Tribe{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			Name:       "Test Tribe 2",
			Type:       models.TribeTypeCouple,
			Visibility: models.VisibilityPrivate,
		}
		_, err := db.Exec(`
			INSERT INTO tribes (id, name, type, visibility, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`, tribe2.ID, tribe2.Name, tribe2.Type, tribe2.Visibility)
		require.NoError(t, err)

		// Create test activity
		activity := &models.Activity{
			ID:          uuid.New(),
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Multi-Tribe Activity",
			Description: "Test Description",
			Visibility:  models.VisibilityShared,
			Metadata:    models.JSONMap{"test": "data"},
		}
		err = repo.Create(activity)
		require.NoError(t, err)

		// Share with both tribes
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, nil)
		assert.NoError(t, err)
		err = repo.ShareWithTribe(activity.ID, tribe2.ID, user1.ID, nil)
		assert.NoError(t, err)

		// Check both tribes can see the activity
		shared1, err := repo.GetSharedActivities(tribe.ID)
		assert.NoError(t, err)
		var found1 bool
		for _, a := range shared1 {
			if a.ID == activity.ID {
				found1 = true
				break
			}
		}
		assert.True(t, found1, "The activity should be shared with the first tribe")

		shared2, err := repo.GetSharedActivities(tribe2.ID)
		assert.NoError(t, err)
		var found2 bool
		for _, a := range shared2 {
			if a.ID == activity.ID {
				found2 = true
				break
			}
		}
		assert.True(t, found2, "The activity should be shared with the second tribe")
	})

	t.Run("Share_Private_Activity", func(t *testing.T) {
		// Create private activity
		privateActivity := &models.Activity{
			ID:          uuid.New(),
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Private Activity",
			Description: "Private Description",
			Visibility:  models.VisibilityPrivate, // Private visibility
			Metadata:    models.JSONMap{"test": "private"},
		}
		err := repo.Create(privateActivity)
		require.NoError(t, err)

		// Attempt to share private activity
		err = repo.ShareWithTribe(privateActivity.ID, tribe.ID, user1.ID, nil)
		assert.NoError(t, err)

		// Verify sharing was successful by checking that visibility was updated
		activity, err := repo.GetByID(privateActivity.ID)
		assert.NoError(t, err)
		assert.Equal(t, models.VisibilityShared, activity.Visibility, "Activity visibility should be updated to shared")

		// Verify activity appears in shared list
		shared, err := repo.GetSharedActivities(tribe.ID)
		assert.NoError(t, err)
		var found bool
		for _, a := range shared {
			if a.ID == privateActivity.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "The private activity should now be in the shared list")
	})
}
