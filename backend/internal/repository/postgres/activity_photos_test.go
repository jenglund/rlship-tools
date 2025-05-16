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

	// Create test activity
	activity := &models.Activity{
		ID:          uuid.New(),
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
		photos := []map[string]interface{}{
			{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 1",
				"tags":        []string{"food", "interior"},
			},
		}
		activity.Metadata = models.JSONMap{"photos": photos}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photo was added
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, activity.Metadata, updated.Metadata)
	})

	t.Run("Add Multiple Photos", func(t *testing.T) {
		// Add multiple photos
		photos := []map[string]interface{}{
			{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 1",
				"tags":        []string{"food", "interior"},
			},
			{
				"url":         "https://example.com/photos/2.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 2",
				"tags":        []string{"exterior", "building"},
			},
		}
		activity.Metadata = models.JSONMap{"photos": photos}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photos were added
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		assert.Equal(t, activity.Metadata, updated.Metadata)
	})

	t.Run("Remove Photo", func(t *testing.T) {
		// Create test activity with photos
		photos := []map[string]interface{}{
			{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 1",
				"tags":        []string{"food", "interior"},
			},
			{
				"url":         "https://example.com/photos/2.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 2",
				"tags":        []string{"exterior", "building"},
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
		photos := []map[string]interface{}{
			{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Original description",
				"tags":        []string{"food"},
			},
		}

		activity.Metadata = models.JSONMap{"photos": photos}
		err = repo.Update(activity)
		require.NoError(t, err)

		// Update photo metadata
		photos[0]["description"] = "Updated description"
		photos[0]["tags"] = []string{"food", "drinks"}
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
	setup := testutil.SetupTest(t)
	repo := NewActivityPhotosRepository(setup.DB)

	// Create test activity
	activity := &models.Activity{
		ID:          uuid.New(),
		Name:        "Test Activity",
		Description: "Test Description",
		Type:        models.ActivityTypeEvent,
		Metadata:    models.JSONMap{"test": "data"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create test photo
	photo := &models.ActivityPhoto{
		ID:         uuid.New(),
		ActivityID: activity.ID,
		URL:        "https://example.com/photo.jpg",
		Caption:    "Test Caption",
		Metadata:   models.JSONMap{"test": "data"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(photo)
		assert.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		retrieved, err := repo.GetByID(photo.ID)
		assert.NoError(t, err)
		assert.Equal(t, photo.ID, retrieved.ID)
		assert.Equal(t, photo.ActivityID, retrieved.ActivityID)
		assert.Equal(t, photo.URL, retrieved.URL)
		assert.Equal(t, photo.Caption, retrieved.Caption)
		assert.Equal(t, photo.Metadata, retrieved.Metadata)
	})

	t.Run("Update", func(t *testing.T) {
		updated := &models.ActivityPhoto{
			ID:         photo.ID,
			ActivityID: photo.ActivityID,
			URL:        "https://example.com/updated.jpg",
			Caption:    "Updated Caption",
			Metadata:   models.JSONMap{"updated": "data"},
			CreatedAt:  photo.CreatedAt,
			UpdatedAt:  time.Now(),
		}

		err := repo.Update(updated)
		assert.NoError(t, err)

		retrieved, err := repo.GetByID(photo.ID)
		assert.NoError(t, err)
		assert.Equal(t, updated.URL, retrieved.URL)
		assert.Equal(t, updated.Caption, retrieved.Caption)
		assert.Equal(t, updated.Metadata, retrieved.Metadata)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(photo.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(photo.ID)
		assert.Error(t, err)
	})
}
