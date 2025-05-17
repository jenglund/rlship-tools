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

func TestTribeRepository(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	repo := NewTribeRepository(db)
	userRepo := NewUserRepository(db)

	// Create test users that we'll use throughout the tests
	user1 := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err := userRepo.Create(user1)
	require.NoError(t, err)

	user2 := &models.User{
		FirebaseUID: "test-firebase-" + uuid.New().String()[:8],
		Provider:    "google",
		Email:       "test-" + uuid.New().String()[:8] + "@example.com",
		Name:        "Test User " + uuid.New().String()[:8],
	}
	err = userRepo.Create(user2)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		t.Run("valid tribe", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        "Test Tribe " + uuid.New().String()[:8],
				Type:        models.TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}

			err := repo.Create(tribe)
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, tribe.ID)

			found, err := repo.GetByID(tribe.ID)
			require.NoError(t, err)
			assert.Equal(t, tribe.Name, found.Name)
		})

		t.Run("duplicate name", func(t *testing.T) {
			now := time.Now()
			tribe1 := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        "Test Tribe " + uuid.New().String()[:8],
				Type:        models.TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}

			err := repo.Create(tribe1)
			require.NoError(t, err)

			tribe2 := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        tribe1.Name,
				Type:        models.TribeTypeCouple,
				Description: "Another test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}

			err = repo.Create(tribe2)
			assert.Error(t, err)
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Run("existing tribe", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        "Test Tribe " + uuid.New().String()[:8],
				Type:        models.TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}

			err := repo.Create(tribe)
			require.NoError(t, err)

			found, err := repo.GetByID(tribe.ID)
			require.NoError(t, err)
			assert.Equal(t, tribe.Name, found.Name)
			assert.Equal(t, tribe.Type, found.Type)
			assert.Equal(t, tribe.Description, found.Description)
			assert.Equal(t, tribe.Visibility, found.Visibility)
		})

		t.Run("non-existent tribe", func(t *testing.T) {
			found, err := repo.GetByID(uuid.New())
			assert.Error(t, err)
			assert.Nil(t, found)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("existing tribe", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        "Test Tribe " + uuid.New().String()[:8],
				Type:        models.TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}

			err := repo.Create(tribe)
			require.NoError(t, err)

			// Wait a moment to ensure UpdatedAt will be different
			time.Sleep(time.Millisecond * 100)

			tribe.Name = "Updated Name"

			err = repo.Update(tribe)
			require.NoError(t, err)

			found, err := repo.GetByID(tribe.ID)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", found.Name)
			assert.True(t, found.UpdatedAt.After(found.CreatedAt))
		})

		t.Run("non-existent tribe", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        "Test Tribe " + uuid.New().String()[:8],
				Type:        models.TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}

			err := repo.Update(tribe)
			assert.Error(t, err)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("existing tribe", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        "Test Tribe " + uuid.New().String()[:8],
				Type:        models.TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}

			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.Delete(tribe.ID)
			require.NoError(t, err)

			found, err := repo.GetByID(tribe.ID)
			assert.Error(t, err)
			assert.Nil(t, found)
		})

		t.Run("non-existent tribe", func(t *testing.T) {
			err := repo.Delete(uuid.New())
			assert.Error(t, err)
		})
	})

	t.Run("List", func(t *testing.T) {
		// Create multiple tribes
		tribes := make([]*models.Tribe, 3)
		for i := range tribes {
			now := time.Now()
			tribes[i] = &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name:        "Test Tribe " + uuid.New().String()[:8],
				Type:        models.TribeTypeCouple,
				Description: "A test tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    models.JSONMap{},
			}
			err := repo.Create(tribes[i])
			require.NoError(t, err)
		}

		found, err := repo.List(0, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(found), len(tribes))

		// Verify all created tribes are in the list
		foundIDs := make(map[uuid.UUID]bool)
		for _, tribe := range found {
			foundIDs[tribe.ID] = true
		}
		for _, tribe := range tribes {
			assert.True(t, foundIDs[tribe.ID])
		}
	})

	t.Run("AddMember", func(t *testing.T) {
		t.Run("valid member addition", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.AddMember(tribe.ID, user1.ID, models.MembershipFull, nil, nil)
			require.NoError(t, err)

			members, err := repo.GetMembers(tribe.ID)
			require.NoError(t, err)
			assert.Len(t, members, 1)
			assert.Equal(t, user1.ID, members[0].UserID)
			assert.Equal(t, models.MembershipFull, members[0].MembershipType)
		})

		t.Run("duplicate member", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.AddMember(tribe.ID, user1.ID, models.MembershipFull, nil, nil)
			require.NoError(t, err)

			err = repo.AddMember(tribe.ID, user1.ID, models.MembershipFull, nil, nil)
			assert.Error(t, err)
		})

		t.Run("non-existent tribe", func(t *testing.T) {
			err := repo.AddMember(uuid.New(), user1.ID, models.MembershipFull, nil, nil)
			assert.Error(t, err)
		})

		t.Run("non-existent user", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.AddMember(tribe.ID, uuid.New(), models.MembershipFull, nil, nil)
			assert.Error(t, err)
		})
	})

	t.Run("UpdateMember", func(t *testing.T) {
		t.Run("valid member update", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.AddMember(tribe.ID, user1.ID, models.MembershipFull, nil, nil)
			require.NoError(t, err)

			err = repo.UpdateMember(tribe.ID, user1.ID, models.MembershipGuest, nil)
			require.NoError(t, err)

			members, err := repo.GetMembers(tribe.ID)
			require.NoError(t, err)
			assert.Len(t, members, 1)
			assert.Equal(t, models.MembershipGuest, members[0].MembershipType)
		})

		t.Run("non-existent tribe", func(t *testing.T) {
			err := repo.UpdateMember(uuid.New(), user1.ID, models.MembershipGuest, nil)
			assert.Error(t, err)
		})

		t.Run("non-existent user", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.UpdateMember(tribe.ID, uuid.New(), models.MembershipGuest, nil)
			assert.Error(t, err)
		})
	})

	t.Run("RemoveMember", func(t *testing.T) {
		t.Run("valid member removal", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.AddMember(tribe.ID, user1.ID, models.MembershipFull, nil, nil)
			require.NoError(t, err)

			err = repo.RemoveMember(tribe.ID, user1.ID)
			require.NoError(t, err)

			members, err := repo.GetMembers(tribe.ID)
			require.NoError(t, err)
			assert.Len(t, members, 0)
		})

		t.Run("non-existent tribe", func(t *testing.T) {
			err := repo.RemoveMember(uuid.New(), user1.ID)
			assert.Error(t, err)
		})

		t.Run("non-existent user", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			err = repo.RemoveMember(tribe.ID, uuid.New())
			assert.Error(t, err)
		})
	})

	t.Run("GetMembers", func(t *testing.T) {
		t.Run("multiple members", func(t *testing.T) {
			now := time.Now()
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			// Add both test users as members
			err = repo.AddMember(tribe.ID, user1.ID, models.MembershipFull, nil, nil)
			require.NoError(t, err)
			err = repo.AddMember(tribe.ID, user2.ID, models.MembershipGuest, nil, nil)
			require.NoError(t, err)

			members, err := repo.GetMembers(tribe.ID)
			require.NoError(t, err)
			assert.Len(t, members, 2)

			// Verify member details
			for _, member := range members {
				assert.NotNil(t, member.User)
				if member.UserID == user1.ID {
					assert.Equal(t, models.MembershipFull, member.MembershipType)
				} else {
					assert.Equal(t, models.MembershipGuest, member.MembershipType)
				}
			}
		})

		t.Run("non-existent tribe", func(t *testing.T) {
			members, err := repo.GetMembers(uuid.New())
			assert.Error(t, err)
			assert.Nil(t, members)
		})
	})

	t.Run("GetUserTribes", func(t *testing.T) {
		t.Run("multiple tribes", func(t *testing.T) {
			// Clean up any existing tribes
			_, err := db.Exec("DELETE FROM tribe_members")
			require.NoError(t, err)
			_, err = db.Exec("DELETE FROM tribes")
			require.NoError(t, err)

			// Create multiple tribes and add the user to them
			userTribes := make([]*models.Tribe, 3)
			for i := range userTribes {
				userTribes[i] = &models.Tribe{
					BaseModel: models.BaseModel{
						ID:        uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name: "Test Tribe " + uuid.New().String()[:8],
				}
				err := repo.Create(userTribes[i])
				require.NoError(t, err)

				err = repo.AddMember(userTribes[i].ID, user1.ID, models.MembershipFull, nil, nil)
				require.NoError(t, err)
			}

			// Create a tribe that the user is not a member of
			otherTribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name: "Other Tribe " + uuid.New().String()[:8],
			}
			err = repo.Create(otherTribe)
			require.NoError(t, err)

			found, err := repo.GetUserTribes(user1.ID)
			require.NoError(t, err)
			assert.Len(t, found, len(userTribes))

			// Verify all user's tribes are in the list
			foundIDs := make(map[uuid.UUID]bool)
			for _, tribe := range found {
				foundIDs[tribe.ID] = true
			}
			for _, tribe := range userTribes {
				assert.True(t, foundIDs[tribe.ID])
			}
			// Verify the other tribe is not in the list
			assert.False(t, foundIDs[otherTribe.ID])
		})

		t.Run("non-existent user", func(t *testing.T) {
			tribes, err := repo.GetUserTribes(uuid.New())
			assert.NoError(t, err) // Should return empty list, not error
			assert.Empty(t, tribes)
		})
	})

	t.Run("GetExpiredGuestMemberships", func(t *testing.T) {
		t.Run("expired memberships exist", func(t *testing.T) {
			tribe := &models.Tribe{
				BaseModel: models.BaseModel{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name: "Test Tribe " + uuid.New().String()[:8],
			}
			err := repo.Create(tribe)
			require.NoError(t, err)

			// Add a guest member with an expiration time in the past
			expiredTime := time.Now().Add(-24 * time.Hour) // 1 day ago
			err = repo.AddMember(tribe.ID, user1.ID, models.MembershipGuest, &expiredTime, nil)
			require.NoError(t, err)

			// Add another member that shouldn't be expired
			err = repo.AddMember(tribe.ID, user2.ID, models.MembershipGuest, nil, nil)
			require.NoError(t, err)

			expired, err := repo.GetExpiredGuestMemberships()
			require.NoError(t, err)
			assert.Len(t, expired, 1)
			assert.Equal(t, user1.ID, expired[0].UserID)
			assert.Equal(t, tribe.ID, expired[0].TribeID)
		})
	})
}
