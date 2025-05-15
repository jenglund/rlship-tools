package postgres

import (
	"testing"
	"time"

	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityHistory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewActivityRepository(db)

	t.Run("Visit History", func(t *testing.T) {
		// Create test activity
		activity := &models.Activity{
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: map[string]interface{}{
				"address": "123 Test St",
				"visits":  []string{},
			},
		}

		err := repo.Create(activity)
		require.NoError(t, err)

		// Add visit
		visits := []string{time.Now().UTC().Format(time.RFC3339)}
		activity.Metadata = map[string]interface{}{
			"address": "123 Test St",
			"visits":  visits,
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify visit was recorded
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		metadata, ok := updated.Metadata.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "123 Test St", metadata["address"])
		assert.Equal(t, visits, metadata["visits"])
	})

	t.Run("Activity Metadata Updates", func(t *testing.T) {
		// Create test activity with initial metadata
		activity := &models.Activity{
			Type:        models.ActivityTypeLocation,
			Name:        "Test Location",
			Description: "Test Description",
			Visibility:  models.VisibilityPublic,
			Metadata: map[string]interface{}{
				"rating": 4.5,
				"tags":   []string{"restaurant", "italian"},
			},
		}

		err := repo.Create(activity)
		require.NoError(t, err)

		// Update metadata
		activity.Metadata = map[string]interface{}{
			"rating": 4.8,
			"tags":   []string{"restaurant", "italian", "pizza"},
			"price":  "$$",
		}

		err = repo.Update(activity)
		require.NoError(t, err)

		// Verify metadata was updated
		updated, err := repo.GetByID(activity.ID)
		require.NoError(t, err)
		metadata, ok := updated.Metadata.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 4.8, metadata["rating"])
		assert.Equal(t, []string{"restaurant", "italian", "pizza"}, metadata["tags"])
		assert.Equal(t, "$$", metadata["price"])
	})

	t.Run("Activity Type Specific Metadata", func(t *testing.T) {
		testCases := []struct {
			name     string
			activity *models.Activity
		}{
			{
				name: "Location Activity",
				activity: &models.Activity{
					Type:        models.ActivityTypeLocation,
					Name:        "Test Location",
					Description: "Test Description",
					Visibility:  models.VisibilityPublic,
					Metadata: map[string]interface{}{
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
					Type:        models.ActivityTypeInterest,
					Name:        "Test Interest",
					Description: "Test Description",
					Visibility:  models.VisibilityPublic,
					Metadata: map[string]interface{}{
						"category":   "Outdoor",
						"difficulty": "Easy",
						"duration":   "2 hours",
					},
				},
			},
			{
				name: "List Activity",
				activity: &models.Activity{
					Type:        models.ActivityTypeList,
					Name:        "Test List",
					Description: "Test Description",
					Visibility:  models.VisibilityPublic,
					Metadata: map[string]interface{}{
						"items": []string{"Item 1", "Item 2"},
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
