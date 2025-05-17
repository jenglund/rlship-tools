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

func TestActivityPhotos(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)
	testUser := testutil.CreateTestUser(t, db)

	// Create test activity
	activity := &models.Activity{
		ID:          uuid.New(),
		UserID:      testUser.ID,
		Type:        models.ActivityTypeLocation,
		Name:        "Test Location",
		Description: "Test Description",
		Visibility:  models.VisibilityPublic,
		Metadata:    models.JSONMap{"photos": []interface{}{}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(activity)
	require.NoError(t, err)

	t.Run("Add Photo", func(t *testing.T) {
		// Add photo metadata
		photos := []interface{}{
			map[string]interface{}{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 1",
				"tags":        []interface{}{"food", "interior"},
			},
		}
		activity.Metadata = models.JSONMap{"photos": photos}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photo was added
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, activity.Metadata, updated.Metadata)

		// Verify the structure of the metadata
		updatedPhotos, ok := updated.Metadata["photos"].([]interface{})
		require.True(t, ok, "photos should be a slice of interface{}")
		require.Len(t, updatedPhotos, 1)

		photo, ok := updatedPhotos[0].(map[string]interface{})
		require.True(t, ok, "photo should be a map[string]interface{}")
		assert.Equal(t, "https://example.com/photos/1.jpg", photo["url"])
	})

	t.Run("Add Multiple Photos", func(t *testing.T) {
		// Add multiple photos
		photos := []interface{}{
			map[string]interface{}{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 1",
				"tags":        []interface{}{"food", "interior"},
			},
			map[string]interface{}{
				"url":         "https://example.com/photos/2.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 2",
				"tags":        []interface{}{"exterior", "building"},
			},
		}
		activity.Metadata = models.JSONMap{"photos": photos}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photos were added
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, activity.Metadata, updated.Metadata)

		// Verify the structure of the metadata
		updatedPhotos, ok := updated.Metadata["photos"].([]interface{})
		require.True(t, ok, "photos should be a slice of interface{}")
		require.Len(t, updatedPhotos, 2)

		for i, p := range updatedPhotos {
			photo, ok := p.(map[string]interface{})
			require.True(t, ok, "photo should be a map[string]interface{}")
			assert.Equal(t, photos[i].(map[string]interface{})["url"], photo["url"])
		}
	})

	t.Run("Remove Photo", func(t *testing.T) {
		// Create test activity with photos
		photos := []interface{}{
			map[string]interface{}{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 1",
				"tags":        []interface{}{"food", "interior"},
			},
			map[string]interface{}{
				"url":         "https://example.com/photos/2.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 2",
				"tags":        []interface{}{"exterior", "building"},
			},
		}

		activity.Metadata = models.JSONMap{"photos": photos}
		err = repo.Update(activity)
		require.NoError(t, err)

		// Remove first photo
		photos = photos[1:]
		activity.Metadata = models.JSONMap{"photos": photos}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photo was removed
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, activity.Metadata, updated.Metadata)
	})

	t.Run("Update Photo Metadata", func(t *testing.T) {
		// Create test activity with photo
		photos := []interface{}{
			map[string]interface{}{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Original description",
				"tags":        []interface{}{"food"},
			},
		}

		activity.Metadata = models.JSONMap{"photos": photos}
		err = repo.Update(activity)
		require.NoError(t, err)

		// Update photo metadata
		photos[0].(map[string]interface{})["description"] = "Updated description"
		photos[0].(map[string]interface{})["tags"] = []interface{}{"food", "drinks"}
		activity.Metadata = models.JSONMap{"photos": photos}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photo metadata was updated
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, activity.Metadata, updated.Metadata)
	})
}

func TestActivityPhotosRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	repo := NewActivityPhotosRepository(db)
	testUser := testutil.CreateTestUser(t, db)

	// Create test activity
	activity := &models.Activity{
		ID:          uuid.New(),
		UserID:      testUser.ID,
		Name:        "Test Activity",
		Description: "Test Description",
		Type:        models.ActivityTypeEvent,
		Visibility:  models.VisibilityPrivate,
		Metadata:    models.JSONMap{"test": "data"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create activity first
	activityRepo := NewActivityRepository(db)
	err := activityRepo.Create(activity)
	require.NoError(t, err)

	// Create test photo
	now := time.Now()
	photo := &models.ActivityPhoto{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
		ActivityID: activity.ID,
		URL:        "https://example.com/photo.jpg",
		Caption:    "Test Caption",
		Metadata:   models.JSONMap{"test": "data"},
	}

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(photo)
		assert.NoError(t, err)
		assert.Equal(t, 1, photo.Version)
	})

	t.Run("GetByID", func(t *testing.T) {
		retrieved, err := repo.GetByID(photo.ID)
		assert.NoError(t, err)
		assert.Equal(t, photo.ID, retrieved.ID)
		assert.Equal(t, photo.ActivityID, retrieved.ActivityID)
		assert.Equal(t, photo.URL, retrieved.URL)
		assert.Equal(t, photo.Caption, retrieved.Caption)
		assert.Equal(t, photo.Metadata, retrieved.Metadata)
		assert.Equal(t, photo.Version, retrieved.Version)
	})

	t.Run("Update", func(t *testing.T) {
		updated := &models.ActivityPhoto{
			BaseModel: models.BaseModel{
				ID:        photo.ID,
				CreatedAt: photo.CreatedAt,
				UpdatedAt: time.Now(),
				Version:   photo.Version,
			},
			ActivityID: photo.ActivityID,
			URL:        "https://example.com/updated.jpg",
			Caption:    "Updated Caption",
			Metadata:   models.JSONMap{"updated": "data"},
		}

		err := repo.Update(updated)
		assert.NoError(t, err)

		retrieved, err := repo.GetByID(photo.ID)
		assert.NoError(t, err)
		assert.Equal(t, updated.URL, retrieved.URL)
		assert.Equal(t, updated.Caption, retrieved.Caption)
		assert.Equal(t, updated.Metadata, retrieved.Metadata)
		assert.Equal(t, photo.Version+1, retrieved.Version)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(photo.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(photo.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, models.ErrNotFound)
	})
}
