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

func TestUserRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewUserRepository(testutil.UnwrapDB(db))

	t.Run("Create", func(t *testing.T) {
		t.Run("valid user", func(t *testing.T) {
			user := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
				AvatarURL:   "https://example.com/avatars/" + uuid.New().String() + ".jpg",
			}

			err := repo.Create(user)
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, user.ID)
			assert.False(t, user.CreatedAt.IsZero())
			assert.False(t, user.UpdatedAt.IsZero())
		})

		t.Run("duplicate firebase_uid", func(t *testing.T) {
			user1 := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Create(user1)
			require.NoError(t, err)

			user2 := &models.User{
				FirebaseUID: user1.FirebaseUID,
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err = repo.Create(user2)
			assert.Error(t, err)
		})

		t.Run("duplicate email", func(t *testing.T) {
			user1 := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Create(user1)
			require.NoError(t, err)

			user2 := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       user1.Email,
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err = repo.Create(user2)
			assert.Error(t, err)
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			user := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Create(user)
			require.NoError(t, err)

			found, err := repo.GetByID(user.ID)
			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.Equal(t, user.FirebaseUID, found.FirebaseUID)
			assert.Equal(t, user.Email, found.Email)
			assert.Equal(t, user.Name, found.Name)
		})

		t.Run("non-existent user", func(t *testing.T) {
			found, err := repo.GetByID(uuid.New())
			assert.Error(t, err)
			assert.Nil(t, found)
		})
	})

	t.Run("GetByFirebaseUID", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			user := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Create(user)
			require.NoError(t, err)

			found, err := repo.GetByFirebaseUID(user.FirebaseUID)
			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.Equal(t, user.FirebaseUID, found.FirebaseUID)
			assert.Equal(t, user.Email, found.Email)
			assert.Equal(t, user.Name, found.Name)
		})

		t.Run("non-existent user", func(t *testing.T) {
			found, err := repo.GetByFirebaseUID("non-existent")
			assert.Error(t, err)
			assert.Nil(t, found)
		})
	})

	t.Run("GetByEmail", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			user := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Create(user)
			require.NoError(t, err)

			found, err := repo.GetByEmail(user.Email)
			require.NoError(t, err)
			assert.Equal(t, user.ID, found.ID)
			assert.Equal(t, user.FirebaseUID, found.FirebaseUID)
			assert.Equal(t, user.Email, found.Email)
			assert.Equal(t, user.Name, found.Name)
		})

		t.Run("non-existent user", func(t *testing.T) {
			found, err := repo.GetByEmail("non-existent@example.com")
			assert.Error(t, err)
			assert.Nil(t, found)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			user := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Create(user)
			require.NoError(t, err)

			// Wait a moment to ensure UpdatedAt will be different
			time.Sleep(time.Millisecond * 100)

			user.Name = "Updated Name"
			user.AvatarURL = "https://example.com/avatars/updated.jpg"

			err = repo.Update(user)
			require.NoError(t, err)

			found, err := repo.GetByID(user.ID)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", found.Name)
			assert.Equal(t, "https://example.com/avatars/updated.jpg", found.AvatarURL)
			assert.True(t, found.UpdatedAt.After(found.CreatedAt))
		})

		t.Run("non-existent user", func(t *testing.T) {
			user := &models.User{
				ID:          uuid.New(),
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Update(user)
			assert.Error(t, err)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			user := &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}

			err := repo.Create(user)
			require.NoError(t, err)

			err = repo.Delete(user.ID)
			require.NoError(t, err)

			found, err := repo.GetByID(user.ID)
			assert.Error(t, err)
			assert.Nil(t, found)
		})

		t.Run("non-existent user", func(t *testing.T) {
			err := repo.Delete(uuid.New())
			assert.Error(t, err)
		})
	})

	t.Run("List", func(t *testing.T) {
		// Create multiple users
		users := make([]*models.User, 3)
		for i := range users {
			users[i] = &models.User{
				FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
				Provider:    "google",
				Email:       "test-" + uuid.New().String()[:8] + "@example.com",
				Name:        "Test User " + uuid.New().String()[:8],
			}
			err := repo.Create(users[i])
			require.NoError(t, err)
		}

		found, err := repo.List(0, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(found), len(users))

		// Verify all created users are in the list
		foundIDs := make(map[uuid.UUID]bool)
		for _, u := range found {
			foundIDs[u.ID] = true
		}
		for _, u := range users {
			assert.True(t, foundIDs[u.ID])
		}
	})
}
