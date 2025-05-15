package postgres

import (
	"testing"
	"time"

	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityPhotos(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)

	t.Run("Add Photo to Activity", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: map[string]interface{}{
				"photos": []map[string]interface{}{},
			},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Add photo metadata
		photos := []map[string]interface{}{
			{
				"url":         "https://example.com/photos/1.jpg",
				"uploadedAt":  time.Now().UTC().Format(time.RFC3339),
				"description": "Test photo 1",
				"tags":        []string{"food", "interior"},
			},
		}
		activity.Metadata = map[string]interface{}{
			"photos": photos,
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photo was added
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		metadata, ok := updated.Metadata.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, photos, metadata["photos"])
	})

	t.Run("Add Multiple Photos", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: map[string]interface{}{
				"photos": []map[string]interface{}{},
			},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

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
		activity.Metadata = map[string]interface{}{
			"photos": photos,
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photos were added
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		metadata, ok := updated.Metadata.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, photos, metadata["photos"])
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

		activity := &models.Activity{
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: map[string]interface{}{
				"photos": photos,
			},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Remove first photo
		photos = photos[1:]
		activity.Metadata = map[string]interface{}{
			"photos": photos,
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photo was removed
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		metadata, ok := updated.Metadata.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, photos, metadata["photos"])
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

		activity := &models.Activity{
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: map[string]interface{}{
				"photos": photos,
			},
		}
		err := repo.Create(activity)
		require.NoError(t, err)

		// Update photo metadata
		photos[0]["description"] = "Updated description"
		photos[0]["tags"] = []string{"food", "drinks"}
		activity.Metadata = map[string]interface{}{
			"photos": photos,
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify photo metadata was updated
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		metadata, ok := updated.Metadata.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, photos, metadata["photos"])
	})
}
