package testutil

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/mock"
)

// MockFirebaseAuth implements the Firebase auth.Client interface for testing
type MockFirebaseAuth struct {
	mock.Mock
	VerifyIDTokenFunc func(ctx context.Context, idToken string) (*auth.Token, error)
}

// On exposes the mock's On method for setting expectations
func (m *MockFirebaseAuth) On(methodName string, args ...interface{}) *mock.Call {
	return m.Mock.On(methodName, args...)
}

// Called exposes the mock's Called method for asserting expectations
func (m *MockFirebaseAuth) Called(args ...interface{}) mock.Arguments {
	return m.Mock.Called(args...)
}

// VerifyIDToken mocks the verification of a Firebase ID token
func (m *MockFirebaseAuth) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	if m.VerifyIDTokenFunc != nil {
		return m.VerifyIDTokenFunc(ctx, idToken)
	}
	return &auth.Token{
		UID: "test-uid",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}, nil
}

// Required methods to implement auth.Client interface

func (m *MockFirebaseAuth) CustomToken(ctx context.Context, uid string) (string, error) {
	return "mock-custom-token", nil
}

func (m *MockFirebaseAuth) CustomTokenWithClaims(ctx context.Context, uid string, claims map[string]interface{}) (string, error) {
	return "mock-custom-token", nil
}

func (m *MockFirebaseAuth) RevokeRefreshTokens(ctx context.Context, uid string) error {
	return nil
}

func (m *MockFirebaseAuth) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	return &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:   uid,
			Email: "test@example.com",
		},
	}, nil
}

func (m *MockFirebaseAuth) GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	return &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:   "test-uid",
			Email: email,
		},
	}, nil
}

func (m *MockFirebaseAuth) GetUserByPhoneNumber(ctx context.Context, phone string) (*auth.UserRecord, error) {
	return &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:         "test-uid",
			PhoneNumber: phone,
		},
	}, nil
}

func (m *MockFirebaseAuth) GetUserByProviderID(ctx context.Context, providerID string) (*auth.UserRecord, error) {
	return &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:        "test-uid",
			ProviderID: providerID,
		},
	}, nil
}

func (m *MockFirebaseAuth) Users(ctx context.Context, nextPageToken string) *auth.UserIterator {
	return &auth.UserIterator{}
}

func (m *MockFirebaseAuth) CreateUser(ctx context.Context, params *auth.UserToCreate) (*auth.UserRecord, error) {
	return &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:   "new-test-uid",
			Email: "new@example.com",
		},
	}, nil
}

func (m *MockFirebaseAuth) UpdateUser(ctx context.Context, uid string, params *auth.UserToUpdate) (*auth.UserRecord, error) {
	return &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			UID:   uid,
			Email: "updated@example.com",
		},
	}, nil
}

func (m *MockFirebaseAuth) DeleteUser(ctx context.Context, uid string) error {
	return nil
}

func (m *MockFirebaseAuth) DeleteUsers(ctx context.Context, uids []string) (*auth.DeleteUsersResult, error) {
	return &auth.DeleteUsersResult{
		SuccessCount: len(uids),
		FailureCount: 0,
	}, nil
}

func (m *MockFirebaseAuth) ImportUsers(ctx context.Context, users []*auth.UserToImport) (*auth.UserImportResult, error) {
	return &auth.UserImportResult{
		SuccessCount: len(users),
		FailureCount: 0,
	}, nil
}

func (m *MockFirebaseAuth) SessionCookie(ctx context.Context, idToken string, expiresIn int64) (string, error) {
	return "mock-session-cookie", nil
}

func (m *MockFirebaseAuth) VerifySessionCookie(ctx context.Context, sessionCookie string) (*auth.Token, error) {
	return &auth.Token{
		UID: "test-uid",
		Claims: map[string]interface{}{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}, nil
}

func (m *MockFirebaseAuth) SetCustomUserClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	return nil
}

func (m *MockFirebaseAuth) GenerateEmailVerificationLink(ctx context.Context, email string) (string, error) {
	return "mock-verification-link", nil
}

func (m *MockFirebaseAuth) GeneratePasswordResetLink(ctx context.Context, email string) (string, error) {
	return "mock-reset-link", nil
}

func (m *MockFirebaseAuth) GenerateSignInWithEmailLink(ctx context.Context, email string, settings *auth.ActionCodeSettings) (string, error) {
	return "mock-signin-link", nil
}

func (m *MockFirebaseAuth) CreateSessionCookie(ctx context.Context, idToken string, expiresIn int64) (string, error) {
	return "mock-session-cookie", nil
}
