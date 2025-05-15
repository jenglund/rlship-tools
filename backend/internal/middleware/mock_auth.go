package middleware

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

// MockFirebaseAuth is a mock implementation of the Firebase auth client
type MockFirebaseAuth struct {
	verifyIDTokenFunc func(context.Context, string) (*auth.Token, error)
}

// NewMockFirebaseAuth creates a new mock Firebase auth client
func NewMockFirebaseAuth() *MockFirebaseAuth {
	return &MockFirebaseAuth{
		verifyIDTokenFunc: func(ctx context.Context, token string) (*auth.Token, error) {
			return nil, errors.New("not implemented")
		},
	}
}

// VerifyIDToken implements the auth.Client interface
func (m *MockFirebaseAuth) VerifyIDToken(ctx context.Context, token string) (*auth.Token, error) {
	return m.verifyIDTokenFunc(ctx, token)
}

// SetVerifyIDTokenFunc sets the mock implementation for VerifyIDToken
func (m *MockFirebaseAuth) SetVerifyIDTokenFunc(fn func(context.Context, string) (*auth.Token, error)) {
	m.verifyIDTokenFunc = fn
}

// MockAuthMiddleware creates a mock auth middleware for testing
func MockAuthMiddleware(userID, firebaseUID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("firebase_uid", firebaseUID)
		c.Next()
	}
}
