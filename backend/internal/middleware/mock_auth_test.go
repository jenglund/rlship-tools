package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMockAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		userID      string
		firebaseUID string
	}{
		{
			name:        "valid user",
			userID:      "user-123",
			firebaseUID: "firebase-456",
		},
		{
			name:        "empty user",
			userID:      "",
			firebaseUID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Use the mock middleware
			router.Use(MockAuthMiddleware(tt.userID, tt.firebaseUID))

			// Add a test endpoint that checks the context values
			router.GET("/test", func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, tt.userID, userID)

				firebaseUID, exists := c.Get("firebase_uid")
				assert.True(t, exists)
				assert.Equal(t, tt.firebaseUID, firebaseUID)

				c.Status(http.StatusOK)
			})

			// Make a test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestMockFirebaseAuth(t *testing.T) {
	mockAuth := NewMockFirebaseAuth()

	// Test the default implementation
	token, err := mockAuth.VerifyIDToken(context.Background(), "dummy-token")
	assert.Error(t, err)
	assert.Nil(t, token)

	// Test with custom implementation
	mockAuth.SetVerifyIDTokenFunc(func(ctx context.Context, token string) (*auth.Token, error) {
		return &auth.Token{
			UID: "custom-uid",
		}, nil
	})

	token, err = mockAuth.VerifyIDToken(context.Background(), "dummy-token")
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "custom-uid", token.UID)
}
