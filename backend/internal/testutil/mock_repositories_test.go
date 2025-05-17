package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewMockRepositories(t *testing.T) {
	repos := NewMockRepositories()
	assert.NotNil(t, repos)
	assert.NotNil(t, repos.Users)
	assert.NotNil(t, repos.Tribes)
	assert.NotNil(t, repos.Activities)
}

func TestMockUserRepository(t *testing.T) {
	repo := NewMockUserRepository()
	testID := uuid.New()
	testUser := &models.User{
		ID:          testID,
		FirebaseUID: "test-firebase-uid",
		Provider:    models.AuthProviderGoogle,
		Email:       "test@example.com",
		Name:        "Test User",
	}

	t.Run("Create", func(t *testing.T) {
		var called bool
		repo.CreateFunc = func(user *models.User) error {
			called = true
			assert.Equal(t, testUser, user)
			return nil
		}

		err := repo.Create(testUser)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("GetByID", func(t *testing.T) {
		var called bool
		repo.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
			called = true
			assert.Equal(t, testID, id)
			return testUser, nil
		}

		user, err := repo.GetByID(testID)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		assert.True(t, called)
	})

	t.Run("GetByFirebaseUID", func(t *testing.T) {
		var called bool
		repo.GetByFirebaseUIDFunc = func(firebaseUID string) (*models.User, error) {
			called = true
			assert.Equal(t, testUser.FirebaseUID, firebaseUID)
			return testUser, nil
		}

		user, err := repo.GetByFirebaseUID(testUser.FirebaseUID)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		assert.True(t, called)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		var called bool
		repo.GetByEmailFunc = func(email string) (*models.User, error) {
			called = true
			assert.Equal(t, testUser.Email, email)
			return testUser, nil
		}

		user, err := repo.GetByEmail(testUser.Email)
		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		assert.True(t, called)
	})

	t.Run("Update", func(t *testing.T) {
		var called bool
		repo.UpdateFunc = func(user *models.User) error {
			called = true
			assert.Equal(t, testUser, user)
			return nil
		}

		err := repo.Update(testUser)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Delete", func(t *testing.T) {
		var called bool
		repo.DeleteFunc = func(id uuid.UUID) error {
			called = true
			assert.Equal(t, testID, id)
			return nil
		}

		err := repo.Delete(testID)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("List", func(t *testing.T) {
		var called bool
		repo.ListFunc = func(offset, limit int) ([]*models.User, error) {
			called = true
			assert.Equal(t, 0, offset)
			assert.Equal(t, 10, limit)
			return []*models.User{testUser}, nil
		}

		users, err := repo.List(0, 10)
		assert.NoError(t, err)
		assert.Equal(t, []*models.User{testUser}, users)
		assert.True(t, called)
	})
}

func TestMockTribeRepository(t *testing.T) {
	repo := NewMockTribeRepository()
	testID := uuid.New()
	now := time.Now()
	testTribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        testID,
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
		Name:       "Test Tribe",
		Type:       models.TribeTypeCouple,
		Visibility: models.VisibilityPrivate,
	}

	t.Run("Create", func(t *testing.T) {
		var called bool
		repo.CreateFunc = func(tribe *models.Tribe) error {
			called = true
			assert.Equal(t, testTribe, tribe)
			return nil
		}

		err := repo.Create(testTribe)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("GetByID", func(t *testing.T) {
		var called bool
		repo.GetByIDFunc = func(id uuid.UUID) (*models.Tribe, error) {
			called = true
			assert.Equal(t, testID, id)
			return testTribe, nil
		}

		tribe, err := repo.GetByID(testID)
		assert.NoError(t, err)
		assert.Equal(t, testTribe, tribe)
		assert.True(t, called)
	})

	t.Run("Update", func(t *testing.T) {
		var called bool
		repo.UpdateFunc = func(tribe *models.Tribe) error {
			called = true
			assert.Equal(t, testTribe, tribe)
			return nil
		}

		err := repo.Update(testTribe)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Delete", func(t *testing.T) {
		var called bool
		repo.DeleteFunc = func(id uuid.UUID) error {
			called = true
			assert.Equal(t, testID, id)
			return nil
		}

		err := repo.Delete(testID)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("List", func(t *testing.T) {
		var called bool
		repo.ListFunc = func(offset, limit int) ([]*models.Tribe, error) {
			called = true
			assert.Equal(t, 0, offset)
			assert.Equal(t, 10, limit)
			return []*models.Tribe{testTribe}, nil
		}

		tribes, err := repo.List(0, 10)
		assert.NoError(t, err)
		assert.Equal(t, []*models.Tribe{testTribe}, tribes)
		assert.True(t, called)
	})

	t.Run("GetByType", func(t *testing.T) {
		var called bool
		repo.GetByTypeFunc = func(tribeType models.TribeType, offset, limit int) ([]*models.Tribe, error) {
			called = true
			assert.Equal(t, models.TribeTypeCouple, tribeType)
			assert.Equal(t, 0, offset)
			assert.Equal(t, 10, limit)
			return []*models.Tribe{testTribe}, nil
		}

		tribes, err := repo.GetByType(models.TribeTypeCouple, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, []*models.Tribe{testTribe}, tribes)
		assert.True(t, called)
	})

	t.Run("Search", func(t *testing.T) {
		var called bool
		repo.SearchFunc = func(query string, offset, limit int) ([]*models.Tribe, error) {
			called = true
			assert.Equal(t, "test", query)
			assert.Equal(t, 0, offset)
			assert.Equal(t, 10, limit)
			return []*models.Tribe{testTribe}, nil
		}

		tribes, err := repo.Search("test", 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, []*models.Tribe{testTribe}, tribes)
		assert.True(t, called)
	})
}

func TestMockActivityRepository(t *testing.T) {
	repo := NewMockActivityRepository()
	testID := uuid.New()
	testActivity := &models.Activity{
		ID:          testID,
		UserID:      uuid.New(),
		Type:        models.ActivityTypeLocation,
		Name:        "Test Activity",
		Description: "Test Description",
		Visibility:  models.VisibilityPrivate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("Create", func(t *testing.T) {
		var called bool
		repo.CreateFunc = func(activity *models.Activity) error {
			called = true
			assert.Equal(t, testActivity, activity)
			return nil
		}

		err := repo.Create(testActivity)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("GetByID", func(t *testing.T) {
		var called bool
		repo.GetByIDFunc = func(id uuid.UUID) (*models.Activity, error) {
			called = true
			assert.Equal(t, testID, id)
			return testActivity, nil
		}

		activity, err := repo.GetByID(testID)
		assert.NoError(t, err)
		assert.Equal(t, testActivity, activity)
		assert.True(t, called)
	})

	t.Run("Update", func(t *testing.T) {
		var called bool
		repo.UpdateFunc = func(activity *models.Activity) error {
			called = true
			assert.Equal(t, testActivity, activity)
			return nil
		}

		err := repo.Update(testActivity)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Delete", func(t *testing.T) {
		var called bool
		repo.DeleteFunc = func(id uuid.UUID) error {
			called = true
			assert.Equal(t, testID, id)
			return nil
		}

		err := repo.Delete(testID)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("List", func(t *testing.T) {
		var called bool
		repo.ListFunc = func(offset, limit int) ([]*models.Activity, error) {
			called = true
			assert.Equal(t, 0, offset)
			assert.Equal(t, 10, limit)
			return []*models.Activity{testActivity}, nil
		}

		activities, err := repo.List(0, 10)
		assert.NoError(t, err)
		assert.Equal(t, []*models.Activity{testActivity}, activities)
		assert.True(t, called)
	})

	t.Run("GetUserActivities", func(t *testing.T) {
		var called bool
		repo.GetUserActivitiesFunc = func(userID uuid.UUID) ([]*models.Activity, error) {
			called = true
			assert.Equal(t, testActivity.UserID, userID)
			return []*models.Activity{testActivity}, nil
		}

		activities, err := repo.GetUserActivities(testActivity.UserID)
		assert.NoError(t, err)
		assert.Equal(t, []*models.Activity{testActivity}, activities)
		assert.True(t, called)
	})

	t.Run("GetTribeActivities", func(t *testing.T) {
		tribeID := uuid.New()
		var called bool
		repo.GetTribeActivitiesFunc = func(id uuid.UUID) ([]*models.Activity, error) {
			called = true
			assert.Equal(t, tribeID, id)
			return []*models.Activity{testActivity}, nil
		}

		activities, err := repo.GetTribeActivities(tribeID)
		assert.NoError(t, err)
		assert.Equal(t, []*models.Activity{testActivity}, activities)
		assert.True(t, called)
	})
}
