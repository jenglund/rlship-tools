package postgres

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityRepository_GetUserActivities(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	testUser := testutil.CreateTestUser(t, db)
	anotherUser := testutil.CreateTestUser(t, db)

	// Create test activities for the main test user
	userActivities := []*models.Activity{
		{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "User Test Activity 1",
			Description: "First user activity",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{"order": 1},
		},
		{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeInterest,
			Name:        "User Test Activity 2",
			Description: "Second user activity",
			Visibility:  models.VisibilityPrivate,
			Metadata:    models.JSONMap{"order": 2},
		},
	}

	// Create activities and add the user as owner
	for _, activity := range userActivities {
		err := repo.Create(activity)
		require.NoError(t, err)

		// Add the user as owner (this is usually handled at service layer)
		err = repo.AddOwner(activity.ID, testUser.ID, models.OwnerTypeUser)
		require.NoError(t, err)
	}

	// Create an activity owned by another user
	otherActivity := &models.Activity{
		UserID:      anotherUser.ID,
		Type:        models.ActivityTypeLocation,
		Name:        "Another User's Activity",
		Description: "Activity for another user",
		Visibility:  models.VisibilityPublic,
		Metadata:    models.JSONMap{"order": 3},
	}
	err := repo.Create(otherActivity)
	require.NoError(t, err)

	// Add the other user as owner
	err = repo.AddOwner(otherActivity.ID, anotherUser.ID, models.OwnerTypeUser)
	require.NoError(t, err)

	t.Run("Get User Activities - Success", func(t *testing.T) {
		// Get activities for test user
		activities, err := repo.GetUserActivities(testUser.ID)
		require.NoError(t, err)

		// Should return exactly the 2 activities we created for this user
		assert.Len(t, activities, 2, "Should return the 2 activities owned by the user")

		// Verify correct activities are returned
		var names []string
		for _, a := range activities {
			names = append(names, a.Name)
		}
		assert.Contains(t, names, "User Test Activity 1")
		assert.Contains(t, names, "User Test Activity 2")
		assert.NotContains(t, names, "Another User's Activity")
	})

	t.Run("Get User Activities - Non-existent User", func(t *testing.T) {
		// Try to get activities for non-existent user
		nonExistentID := uuid.New()
		activities, err := repo.GetUserActivities(nonExistentID)
		require.NoError(t, err)
		assert.Empty(t, activities, "Non-existent user should have no activities")
	})

	t.Run("Get User Activities - Deleted Activity", func(t *testing.T) {
		// Delete one of the test user's activities
		err := repo.Delete(userActivities[0].ID)
		require.NoError(t, err)

		// Get activities for test user
		activities, err := repo.GetUserActivities(testUser.ID)
		require.NoError(t, err)

		// Should return only 1 activity now (the non-deleted one)
		assert.Len(t, activities, 1, "Should return only the non-deleted activity")
		assert.Equal(t, "User Test Activity 2", activities[0].Name)
	})
}

func TestActivityRepository_GetTribeActivities(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	testUser := testutil.CreateTestUser(t, db)

	// Create test tribes
	testTribe := &models.Tribe{
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
	`, testTribe.ID, testTribe.Name, testTribe.Type, testTribe.Visibility)
	require.NoError(t, err)

	anotherTribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		Name:       "Another Tribe",
		Type:       models.TribeTypeCouple,
		Visibility: models.VisibilityPrivate,
	}
	_, err = db.Exec(`
		INSERT INTO tribes (id, name, type, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, anotherTribe.ID, anotherTribe.Name, anotherTribe.Type, anotherTribe.Visibility)
	require.NoError(t, err)

	// Create test activities for the main test tribe
	tribeActivities := []*models.Activity{
		{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Tribe Test Activity 1",
			Description: "First tribe activity",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{"order": 1},
		},
		{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeInterest,
			Name:        "Tribe Test Activity 2",
			Description: "Second tribe activity",
			Visibility:  models.VisibilityPrivate,
			Metadata:    models.JSONMap{"order": 2},
		},
	}

	// Create activities and add the tribe as owner
	for _, activity := range tribeActivities {
		err := repo.Create(activity)
		require.NoError(t, err)

		// Add the tribe as owner
		err = repo.AddOwner(activity.ID, testTribe.ID, models.OwnerTypeTribe)
		require.NoError(t, err)
	}

	// Create an activity owned by another tribe
	otherActivity := &models.Activity{
		UserID:      testUser.ID,
		Type:        models.ActivityTypeLocation,
		Name:        "Another Tribe's Activity",
		Description: "Activity for another tribe",
		Visibility:  models.VisibilityPublic,
		Metadata:    models.JSONMap{"order": 3},
	}
	err = repo.Create(otherActivity)
	require.NoError(t, err)

	// Add the other tribe as owner
	err = repo.AddOwner(otherActivity.ID, anotherTribe.ID, models.OwnerTypeTribe)
	require.NoError(t, err)

	t.Run("Get Tribe Activities - Success", func(t *testing.T) {
		// Get activities for test tribe
		activities, err := repo.GetTribeActivities(testTribe.ID)
		require.NoError(t, err)

		// Should return exactly the 2 activities we created for this tribe
		assert.Len(t, activities, 2, "Should return the 2 activities owned by the tribe")

		// Verify correct activities are returned
		var names []string
		for _, a := range activities {
			names = append(names, a.Name)
		}
		assert.Contains(t, names, "Tribe Test Activity 1")
		assert.Contains(t, names, "Tribe Test Activity 2")
		assert.NotContains(t, names, "Another Tribe's Activity")
	})

	t.Run("Get Tribe Activities - Non-existent Tribe", func(t *testing.T) {
		// Try to get activities for non-existent tribe
		nonExistentID := uuid.New()
		activities, err := repo.GetTribeActivities(nonExistentID)
		require.NoError(t, err)
		assert.Empty(t, activities, "Non-existent tribe should have no activities")
	})

	t.Run("Get Tribe Activities - Deleted Activity", func(t *testing.T) {
		// Delete one of the test tribe's activities
		err := repo.Delete(tribeActivities[0].ID)
		require.NoError(t, err)

		// Get activities for test tribe
		activities, err := repo.GetTribeActivities(testTribe.ID)
		require.NoError(t, err)

		// Should return only 1 activity now (the non-deleted one)
		assert.Len(t, activities, 1, "Should return only the non-deleted activity")
		assert.Equal(t, "Tribe Test Activity 2", activities[0].Name)
	})

	t.Run("Get Tribe Activities - Deleted Owner", func(t *testing.T) {
		// Create a new activity
		newActivity := &models.Activity{
			UserID:      testUser.ID,
			Type:        models.ActivityTypeLocation,
			Name:        "Activity with Deleted Owner",
			Description: "This activity will have its owner deleted",
			Visibility:  models.VisibilityPublic,
			Metadata:    models.JSONMap{"test": "data"},
		}
		err := repo.Create(newActivity)
		require.NoError(t, err)

		// Add the tribe as owner
		err = repo.AddOwner(newActivity.ID, testTribe.ID, models.OwnerTypeTribe)
		require.NoError(t, err)

		// Verify activity appears in tribe activities
		activities, err := repo.GetTribeActivities(testTribe.ID)
		require.NoError(t, err)
		assert.Contains(t, extractNames(activities), "Activity with Deleted Owner")

		// Remove owner
		err = repo.RemoveOwner(newActivity.ID, testTribe.ID)
		require.NoError(t, err)

		// Verify activity no longer appears in tribe activities
		activities, err = repo.GetTribeActivities(testTribe.ID)
		require.NoError(t, err)
		assert.NotContains(t, extractNames(activities), "Activity with Deleted Owner")
	})
}

// Helper function to extract activity names from a slice of activities
func extractNames(activities []*models.Activity) []string {
	names := make([]string, len(activities))
	for i, activity := range activities {
		names[i] = activity.Name
	}
	return names
}
