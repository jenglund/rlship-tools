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

func TestActivityHistory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	testUser := testutil.CreateTestUser(t, db)

	t.Run("Visit History", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: models.JSONMap{
				"address": "123 Test St",
				"visits":  []interface{}{},
			},
		}

		err := repo.Create(activity)
		require.NoError(t, err)

		// Add visit
		visits := []interface{}{time.Now().UTC().Format(time.RFC3339)}
		activity.Metadata = models.JSONMap{
			"address": "123 Test St",
			"visits":  visits,
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify visit was recorded
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, "123 Test St", updated.Metadata["address"])
		assert.Equal(t, visits, updated.Metadata["visits"])
	})

	t.Run("Activity Metadata Updates", func(t *testing.T) {
		// Create test activity with initial metadata
		activity := &models.Activity{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: models.JSONMap{
				"rating": 4.5,
				"tags":   []interface{}{"restaurant", "italian"},
			},
		}

		err := repo.Create(activity)
		require.NoError(t, err)

		// Update metadata
		activity.Metadata = models.JSONMap{
			"rating": 4.8,
			"tags":   []interface{}{"restaurant", "italian", "pizza"},
			"price":  "$$",
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify metadata was updated
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, 4.8, updated.Metadata["rating"])
		assert.Equal(t, []interface{}{"restaurant", "italian", "pizza"}, updated.Metadata["tags"])
		assert.Equal(t, "$$", updated.Metadata["price"])
	})

	t.Run("Activity Type Specific Metadata", func(t *testing.T) {
		testCases := []struct {
			name     string
			activity *models.Activity
		}{
			{
				name: "Location Activity",
				activity: &models.Activity{
					UserID:      testUser.ID,
					Type:        models.ActivityTypeLocation,
					Name:        "Test Location",
					Description: "Test Description",
					Visibility:  models.VisibilityPublic,
					Metadata: models.JSONMap{
						"address":     "123 Test St",
						"latitude":    37.7749,
						"longitude":   -122.4194,
						"googleMapId": "abc123",
					},
				},
			},
			{
				name: "Interest Activity",
				activity: &models.Activity{
					UserID:      testUser.ID,
					Type:        models.ActivityTypeInterest,
					Name:        "Test Interest",
					Description: "Test Description",
					Visibility:  models.VisibilityPublic,
					Metadata: models.JSONMap{
						"category":   "Outdoor",
						"difficulty": "Easy",
						"duration":   "2 hours",
					},
				},
			},
			{
				name: "List Activity",
				activity: &models.Activity{
					UserID:      testUser.ID,
					Type:        models.ActivityTypeList,
					Name:        "Test List",
					Description: "Test Description",
					Visibility:  models.VisibilityPublic,
					Metadata: models.JSONMap{
						"items": []interface{}{"Item 1", "Item 2"},
						"type":  "Bucket List",
					},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create activity
				err := repo.Create(tc.activity)
				require.NoError(t, err)

				// Verify metadata was stored correctly
				created, err := repo.GetByID(tc.activity.ID)
				require.NoError(t, err)
				assert.Equal(t, tc.activity.Metadata, created.Metadata)
			})
		}
	})
}

func TestActivityDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	testUser := testutil.CreateTestUser(t, db)

	t.Run("Delete Existing Activity", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Activity to Delete",
			Description: "This activity will be deleted",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{"test": "data"},
		}

		err := repo.Create(activity)
		require.NoError(t, err)

		// Verify activity exists
		_, err = repo.GetByID(activity.ID)
		require.NoError(t, err)

		// Delete the activity
		err = repo.Delete(activity.ID)
		require.NoError(t, err)

		// Verify activity no longer appears in queries
		_, err = repo.GetByID(activity.ID)
		require.Error(t, err, "Activity should not be found after deletion")
		assert.Contains(t, err.Error(), "activity not found")
	})

	t.Run("Delete Non-existent Activity", func(t *testing.T) {
		// Try to delete an activity that doesn't exist
		nonExistentID := uuid.New()
		err := repo.Delete(nonExistentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "activity not found")
	})

	t.Run("Delete Already Deleted Activity", func(t *testing.T) {
		// Create and then delete an activity
		activity := &models.Activity{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Activity to Delete Twice",
			Description: "This activity will be deleted twice",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{"test": "data"},
		}

		err := repo.Create(activity)
		require.NoError(t, err)

		// Delete once
		err = repo.Delete(activity.ID)
		require.NoError(t, err)

		// Try to delete again
		err = repo.Delete(activity.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "activity not found")
	})
}

func TestActivityList(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	testUser := testutil.CreateTestUser(t, db)

	// Create a set of test activities
	activities := []*models.Activity{
		{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Activity 1",
			Description: "First test activity",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{"order": 1},
		},
		{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeInterest,
			Name:        "Activity 2",
			Description: "Second test activity",
			Visibility:  models.VisibilityPrivate,
			Metadata:    models.JSONMap{"order": 2},
		},
		{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeList,
			Name:        "Activity 3",
			Description: "Third test activity",
			Visibility:  models.VisibilityShared,
			Metadata:    models.JSONMap{"order": 3},
		},
	}

	// Insert all activities
	for _, activity := range activities {
		err := repo.Create(activity)
		require.NoError(t, err)
	}

	t.Run("List All Activities", func(t *testing.T) {
		// Get all activities (no pagination)
		result, err := repo.List(0, 100)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3, "Should have at least 3 activities")

		// Verify the created activities are in the result
		activityIDs := make(map[uuid.UUID]bool)
		for _, activity := range activities {
			activityIDs[activity.ID] = true
		}

		foundCount := 0
		for _, a := range result {
			if activityIDs[a.ID] {
				foundCount++
			}
		}
		assert.Equal(t, 3, foundCount, "All created activities should be found in the list")
	})

	t.Run("List With Pagination", func(t *testing.T) {
		// Test first page
		page1, err := repo.List(0, 2)
		require.NoError(t, err)
		assert.Len(t, page1, 2, "First page should have 2 activities")

		// Test second page
		page2, err := repo.List(2, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(page2), 1, "Second page should have at least 1 activity")

		// Verify different activities on different pages
		for _, activity1 := range page1 {
			for _, activity2 := range page2 {
				assert.NotEqual(t, activity1.ID, activity2.ID, "Activities on different pages should be different")
			}
		}
	})

	t.Run("List After Deletion", func(t *testing.T) {
		// Delete one of the activities
		activityToDelete := activities[1]
		err := repo.Delete(activityToDelete.ID)
		require.NoError(t, err)

		// List activities again
		result, err := repo.List(0, 100)
		require.NoError(t, err)

		// Verify deleted activity is not in the list
		for _, activity := range result {
			assert.NotEqual(t, activityToDelete.ID, activity.ID, "Deleted activity should not be in the list")
		}
	})

	t.Run("List With Zero Limit", func(t *testing.T) {
		// Test with zero limit (should return empty list)
		result, err := repo.List(0, 0)
		require.NoError(t, err)
		assert.Empty(t, result, "Zero limit should return empty list")
	})
}

func TestActivityOwnerManagement(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	testUser := testutil.CreateTestUser(t, db)
	secondUser := testutil.CreateTestUser(t, db)

	// Create a test tribe
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		Name:       "Test Tribe for Ownership",
		Type:       models.TribeTypeCouple,
		Visibility: models.VisibilityPrivate,
	}
	_, err := db.Exec(`
		INSERT INTO tribes (id, name, type, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, tribe.ID, tribe.Name, tribe.Type, tribe.Visibility)
	require.NoError(t, err)

	// Create test activity
	activity := &models.Activity{
		UserID:      testUser.ID,
		Type:        models.ActivityTypeLocation,
		Name:        "Activity for Owner Tests",
		Description: "Testing ownership management",
		Visibility:  models.VisibilityPublic,
		Metadata:    models.JSONMap{"test": "data"},
	}
	err = repo.Create(activity)
	require.NoError(t, err)

	t.Run("AddOwner - User", func(t *testing.T) {
		// Add second user as owner
		err := repo.AddOwner(activity.ID, secondUser.ID, models.OwnerTypeUser)
		require.NoError(t, err)

		// Check owner was added
		owners, err := repo.GetOwners(activity.ID)
		require.NoError(t, err)

		// Verify the ownership is correctly recorded
		var foundUser bool
		for _, owner := range owners {
			if owner.OwnerID == secondUser.ID && owner.OwnerType == models.OwnerTypeUser {
				foundUser = true
				break
			}
		}
		assert.True(t, foundUser, "Second user should be an owner")
	})

	t.Run("AddOwner - Tribe", func(t *testing.T) {
		// Add tribe as owner
		err := repo.AddOwner(activity.ID, tribe.ID, models.OwnerTypeTribe)
		require.NoError(t, err)

		// Check owner was added
		owners, err := repo.GetOwners(activity.ID)
		require.NoError(t, err)

		// Verify the tribe ownership is correctly recorded
		var foundTribe bool
		for _, owner := range owners {
			if owner.OwnerID == tribe.ID && owner.OwnerType == models.OwnerTypeTribe {
				foundTribe = true
				break
			}
		}
		assert.True(t, foundTribe, "Tribe should be an owner")
	})

	t.Run("GetOwners", func(t *testing.T) {
		// Get all owners
		owners, err := repo.GetOwners(activity.ID)
		require.NoError(t, err)

		// Should include the initial creator by default plus the 2 we added
		assert.GreaterOrEqual(t, len(owners), 2, "Should have at least 2 owners")

		// Verify both types of ownership
		var foundUser, foundTribe bool
		for _, owner := range owners {
			if owner.OwnerID == secondUser.ID && owner.OwnerType == models.OwnerTypeUser {
				foundUser = true
			}
			if owner.OwnerID == tribe.ID && owner.OwnerType == models.OwnerTypeTribe {
				foundTribe = true
			}
		}
		assert.True(t, foundUser, "Second user should be an owner")
		assert.True(t, foundTribe, "Tribe should be an owner")
	})

	t.Run("RemoveOwner - User", func(t *testing.T) {
		// Remove second user as owner
		err := repo.RemoveOwner(activity.ID, secondUser.ID)
		require.NoError(t, err)

		// Check owner was removed
		owners, err := repo.GetOwners(activity.ID)
		require.NoError(t, err)

		// Verify the ownership is removed
		var foundUser bool
		for _, owner := range owners {
			if owner.OwnerID == secondUser.ID {
				foundUser = true
				break
			}
		}
		assert.False(t, foundUser, "Second user should no longer be an owner")
	})

	t.Run("RemoveOwner - Tribe", func(t *testing.T) {
		// Remove tribe as owner
		err := repo.RemoveOwner(activity.ID, tribe.ID)
		require.NoError(t, err)

		// Check owner was removed
		owners, err := repo.GetOwners(activity.ID)
		require.NoError(t, err)

		// Verify the tribe ownership is removed
		var foundTribe bool
		for _, owner := range owners {
			if owner.OwnerID == tribe.ID {
				foundTribe = true
				break
			}
		}
		assert.False(t, foundTribe, "Tribe should no longer be an owner")
	})

	t.Run("RemoveOwner - Non-existent Owner", func(t *testing.T) {
		// Try to remove an owner that doesn't exist
		nonExistentID := uuid.New()
		err := repo.RemoveOwner(activity.ID, nonExistentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "activity owner not found")
	})

	t.Run("GetOwners - No Owners", func(t *testing.T) {
		// Create new activity with no additional owners
		newActivity := &models.Activity{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Activity with no additional owners",
			Description: "Testing with no additional owners",
			Visibility:  models.VisibilityPrivate,
			Metadata:    models.JSONMap{"test": "data"},
		}
		err := repo.Create(newActivity)
		require.NoError(t, err)

		// Since the Create method doesn't automatically add owners
		// (unlike in List repository), we need to manually add at least
		// one owner for this test
		err = repo.AddOwner(newActivity.ID, testUser.ID, models.OwnerTypeUser)
		require.NoError(t, err)

		// Get owners - should have the one we added
		owners, err := repo.GetOwners(newActivity.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, owners, "Should have at least one owner we added manually")

		// Check that the owner is the one we added
		var foundOwner bool
		for _, owner := range owners {
			if owner.OwnerID == testUser.ID && owner.OwnerType == models.OwnerTypeUser {
				foundOwner = true
				break
			}
		}
		assert.True(t, foundOwner, "The manually added owner should be found")
	})
}
