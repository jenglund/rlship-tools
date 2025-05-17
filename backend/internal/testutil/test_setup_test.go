package testutil

import (
	"context"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSetupTest(t *testing.T) {
	setup := SetupTest(t)

	// Test that all components are initialized
	assert.NotNil(t, setup)
	assert.NotNil(t, setup.Router)
	assert.NotNil(t, setup.Repositories)
	assert.NotNil(t, setup.Auth)
	assert.NotNil(t, setup.TestUser)

	// Test that the test user is properly initialized
	assert.NotEqual(t, uuid.Nil, setup.TestUser.ID)
	assert.Equal(t, "test-firebase-uid", setup.TestUser.FirebaseUID)
	assert.Equal(t, models.AuthProviderGoogle, setup.TestUser.Provider)
	assert.Equal(t, "test@example.com", setup.TestUser.Email)
	assert.Equal(t, "Test User", setup.TestUser.Name)
	assert.Equal(t, "https://example.com/avatar.jpg", setup.TestUser.AvatarURL)

	// Test that the mock repositories are properly configured
	t.Run("test user repository configuration", func(t *testing.T) {
		// Test GetByFirebaseUID
		user, err := setup.Repositories.Users.GetByFirebaseUID(setup.TestUser.FirebaseUID)
		assert.NoError(t, err)
		assert.Equal(t, setup.TestUser, user)

		// Test with unknown Firebase UID
		user, err = setup.Repositories.Users.GetByFirebaseUID("unknown")
		assert.Error(t, err)
		assert.Nil(t, user)

		// Test GetByID
		user, err = setup.Repositories.Users.GetByID(setup.TestUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, setup.TestUser, user)

		// Test with unknown ID
		user, err = setup.Repositories.Users.GetByID(uuid.New())
		assert.Error(t, err)
		assert.Nil(t, user)

		// Test Create
		newUser := &models.User{
			ID:          uuid.New(),
			FirebaseUID: "new-firebase-uid",
			Provider:    models.AuthProviderGoogle,
			Email:       "new@example.com",
			Name:        "New User",
		}
		err = setup.Repositories.Users.Create(newUser)
		assert.NoError(t, err)

		// Test Update
		setup.TestUser.Name = "Updated Name"
		err = setup.Repositories.Users.Update(setup.TestUser)
		assert.NoError(t, err)
	})

	// Test that the router is properly configured with auth middleware
	t.Run("test router auth middleware", func(t *testing.T) {
		// Add a test endpoint
		setup.Router.GET("/test", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			assert.True(t, exists)
			assert.Equal(t, setup.TestUser.ID, userID)

			firebaseUID, exists := c.Get("firebase_uid")
			assert.True(t, exists)
			assert.Equal(t, setup.TestUser.FirebaseUID, firebaseUID)

			c.JSON(200, gin.H{"status": "success"})
		})

		// Test the endpoint
		w := ExecuteRequest(t, setup.Router, TestRequest{
			Method: "GET",
			Path:   "/test",
		})

		assert.Equal(t, 200, w.Code)
		CheckResponse(t, w, TestResponse{
			Code: 200,
			Body: gin.H{"status": "success"},
		})
	})

	// Test that the mock auth client is properly configured
	t.Run("test mock auth configuration", func(t *testing.T) {
		// Test default implementation
		token, err := setup.Auth.VerifyIDToken(context.TODO(), "test-token")
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "test-uid", token.UID)
		assert.Equal(t, "test@example.com", token.Claims["email"])
		assert.Equal(t, "Test User", token.Claims["name"])

		// Test custom implementation
		customToken := &auth.Token{
			UID: "custom-uid",
			Claims: map[string]interface{}{
				"custom": "claim",
			},
		}

		setup.Auth.VerifyIDTokenFunc = func(ctx context.Context, idToken string) (*auth.Token, error) {
			assert.Equal(t, "custom-token", idToken)
			return customToken, nil
		}

		token, err = setup.Auth.VerifyIDToken(context.TODO(), "custom-token")
		assert.NoError(t, err)
		assert.Equal(t, customToken, token)
	})
}
