package testutil

import (
	"database/sql"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// TestSetup holds all test dependencies
type TestSetup struct {
	Router       *gin.Engine
	Repositories *MockRepositories
	Auth         *MockFirebaseAuth
	TestUser     *models.User
}

// SetupTest creates a new test environment with mocked dependencies
func SetupTest(t *testing.T) *TestSetup {
	t.Helper()

	// Create test user
	testUser := &models.User{
		ID:          uuid.New(),
		FirebaseUID: "test-firebase-uid",
		Provider:    models.AuthProviderGoogle,
		Email:       "test@example.com",
		Name:        "Test User",
		AvatarURL:   "https://example.com/avatar.jpg",
	}

	// Create mock repositories
	repos := NewMockRepositories()
	repos.Users.GetByFirebaseUIDFunc = func(firebaseUID string) (*models.User, error) {
		if firebaseUID == testUser.FirebaseUID {
			return testUser, nil
		}
		return nil, sql.ErrNoRows
	}
	repos.Users.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		if id == testUser.ID {
			return testUser, nil
		}
		return nil, sql.ErrNoRows
	}
	repos.Users.CreateFunc = func(user *models.User) error {
		return nil
	}
	repos.Users.UpdateFunc = func(user *models.User) error {
		return nil
	}

	// Create mock auth client
	auth := &MockFirebaseAuth{}

	// Set up router with mock auth middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("firebase_uid", testUser.FirebaseUID)
		c.Set("user_id", testUser.ID)
		c.Next()
	})

	return &TestSetup{
		Router:       router,
		Repositories: repos,
		Auth:         auth,
		TestUser:     testUser,
	}
}
