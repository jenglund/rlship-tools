package postgres

import (
	"testing"
	"time"

	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivitySharing(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)

	// Create test users and tribe
	user1 := testutil.CreateTestUser(t, db)
	user2 := testutil.CreateTestUser(t, db)
	tribe := testutil.CreateTestTribe(t, db, []testutil.TestUser{user1, user2})

	t.Run("Share Activity with Tribe", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Share activity with tribe
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, nil)
		require.NoError(t, err)

		// Verify activity is shared
		shared, err := repo.GetSharedActivities(tribe.ID)
		require.NoError(t, err)
		require.Len(t, shared, 1)
		assert.Equal(t, activity.ID, shared[0].ID)
	})

	t.Run("Share Activity with Expiration", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Share activity with expiration
		expiresAt := time.Now().Add(24 * time.Hour)
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, &expiresAt)
		require.NoError(t, err)

		// Verify activity is shared
		shared, err := repo.GetSharedActivities(tribe.ID)
		require.NoError(t, err)
		require.Len(t, shared, 1)
		assert.Equal(t, activity.ID, shared[0].ID)

		// Fast forward past expiration
		expiresAt = time.Now().Add(-1 * time.Hour)
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, &expiresAt)
		require.NoError(t, err)

		// Verify activity is no longer shared
		shared, err = repo.GetSharedActivities(tribe.ID)
		require.NoError(t, err)
		assert.Empty(t, shared)
	})

	t.Run("Unshare Activity", func(t *testing.T) {
		// Create and share test activity
		activity := &models.Activity{
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, nil)
		require.NoError(t, err)

		// Unshare activity
		err = repo.UnshareWithTribe(activity.ID, tribe.ID)
		require.NoError(t, err)

		// Verify activity is no longer shared
		shared, err := repo.GetSharedActivities(tribe.ID)
		require.NoError(t, err)
		assert.Empty(t, shared)
	})

	t.Run("Share Activity with Multiple Tribes", func(t *testing.T) {
		// Create another tribe
		tribe2 := testutil.CreateTestTribe(t, db, []testutil.TestUser{user1})

		// Create test activity
		activity := &models.Activity{
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Share with both tribes
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, nil)
		require.NoError(t, err)
		err = repo.ShareWithTribe(activity.ID, tribe2.ID, user1.ID, nil)
		require.NoError(t, err)

		// Verify activity is shared with both tribes
		shared1, err := repo.GetSharedActivities(tribe.ID)
		require.NoError(t, err)
		require.Len(t, shared1, 1)
		assert.Equal(t, activity.ID, shared1[0].ID)

		shared2, err := repo.GetSharedActivities(tribe2.ID)
		require.NoError(t, err)
		require.Len(t, shared2, 1)
		assert.Equal(t, activity.ID, shared2[0].ID)

		// Unshare from one tribe
		err = repo.UnshareWithTribe(activity.ID, tribe.ID)
		require.NoError(t, err)

		// Verify activity is only shared with second tribe
		shared1, err = repo.GetSharedActivities(tribe.ID)
		require.NoError(t, err)
		assert.Empty(t, shared1)

		shared2, err = repo.GetSharedActivities(tribe2.ID)
		require.NoError(t, err)
		require.Len(t, shared2, 1)
		assert.Equal(t, activity.ID, shared2[0].ID)
	})

	t.Run("Share Private Activity", func(t *testing.T) {
		// Create private activity
		activity := &models.Activity{
			UserID:      user1.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Private Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPrivate,
			Metadata:    models.JSONMap{},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Share private activity
		err = repo.ShareWithTribe(activity.ID, tribe.ID, user1.ID, nil)
		require.NoError(t, err)

		// Verify activity is shared
		shared, err := repo.GetSharedActivities(tribe.ID)
		require.NoError(t, err)
		require.Len(t, shared, 1)
		assert.Equal(t, activity.ID, shared[0].ID)
		assert.Equal(t, models.VisibilityPrivate, shared[0].Visibility)
	})
}
