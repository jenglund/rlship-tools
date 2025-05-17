package testutil

import (
	"context"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/assert"
)

func TestMockFirebaseAuth_VerifyIDToken(t *testing.T) {
	ctx := context.Background()

	t.Run("default implementation", func(t *testing.T) {
		mock := &MockFirebaseAuth{}
		token, err := mock.VerifyIDToken(ctx, "test-token")
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "test-uid", token.UID)
		assert.Equal(t, "test@example.com", token.Claims["email"])
		assert.Equal(t, "Test User", token.Claims["name"])
	})

	t.Run("custom implementation", func(t *testing.T) {
		customToken := &auth.Token{
			UID: "custom-uid",
			Claims: map[string]interface{}{
				"custom": "claim",
			},
		}

		mock := &MockFirebaseAuth{
			VerifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
				assert.Equal(t, "custom-token", idToken)
				return customToken, nil
			},
		}

		token, err := mock.VerifyIDToken(ctx, "custom-token")
		assert.NoError(t, err)
		assert.Equal(t, customToken, token)
	})
}

func TestMockFirebaseAuth_GetUser(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("get existing user", func(t *testing.T) {
		user, err := mock.GetUser(ctx, "test-uid")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test-uid", user.UID)
		assert.Equal(t, "test@example.com", user.Email)
	})
}

func TestMockFirebaseAuth_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("get user by email", func(t *testing.T) {
		email := "test@example.com"
		user, err := mock.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test-uid", user.UID)
		assert.Equal(t, email, user.Email)
	})
}

func TestMockFirebaseAuth_GetUserByPhoneNumber(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("get user by phone", func(t *testing.T) {
		phone := "+1234567890"
		user, err := mock.GetUserByPhoneNumber(ctx, phone)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test-uid", user.UID)
		assert.Equal(t, phone, user.PhoneNumber)
	})
}

func TestMockFirebaseAuth_CreateUser(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("create new user", func(t *testing.T) {
		params := &auth.UserToCreate{}
		user, err := mock.CreateUser(ctx, params)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "new-test-uid", user.UID)
		assert.Equal(t, "new@example.com", user.Email)
	})
}

func TestMockFirebaseAuth_UpdateUser(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("update existing user", func(t *testing.T) {
		params := &auth.UserToUpdate{}
		user, err := mock.UpdateUser(ctx, "test-uid", params)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test-uid", user.UID)
		assert.Equal(t, "updated@example.com", user.Email)
	})
}

func TestMockFirebaseAuth_DeleteUsers(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("delete multiple users", func(t *testing.T) {
		uids := []string{"uid1", "uid2", "uid3"}
		result, err := mock.DeleteUsers(ctx, uids)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, len(uids), result.SuccessCount)
		assert.Equal(t, 0, result.FailureCount)
	})
}

func TestMockFirebaseAuth_ImportUsers(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("import users", func(t *testing.T) {
		users := []*auth.UserToImport{
			{},
			{},
		}
		result, err := mock.ImportUsers(ctx, users)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, len(users), result.SuccessCount)
		assert.Equal(t, 0, result.FailureCount)
	})
}

func TestMockFirebaseAuth_SessionCookie(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("create session cookie", func(t *testing.T) {
		cookie, err := mock.SessionCookie(ctx, "test-token", 3600)
		assert.NoError(t, err)
		assert.Equal(t, "mock-session-cookie", cookie)
	})

	t.Run("verify session cookie", func(t *testing.T) {
		token, err := mock.VerifySessionCookie(ctx, "test-session-cookie")
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "test-uid", token.UID)
		assert.Equal(t, "test@example.com", token.Claims["email"])
		assert.Equal(t, "Test User", token.Claims["name"])
	})
}

func TestMockFirebaseAuth_GenerateLinks(t *testing.T) {
	ctx := context.Background()
	mock := &MockFirebaseAuth{}

	t.Run("generate email verification link", func(t *testing.T) {
		link, err := mock.GenerateEmailVerificationLink(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "mock-verification-link", link)
	})

	t.Run("generate password reset link", func(t *testing.T) {
		link, err := mock.GeneratePasswordResetLink(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "mock-reset-link", link)
	})

	t.Run("generate sign in with email link", func(t *testing.T) {
		link, err := mock.GenerateSignInWithEmailLink(ctx, "test@example.com", nil)
		assert.NoError(t, err)
		assert.Equal(t, "mock-signin-link", link)
	})
}
