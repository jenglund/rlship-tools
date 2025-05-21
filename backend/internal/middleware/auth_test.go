package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of the user repository
type MockUserRepository struct {
	mock.Mock
	CreateFunc           func(user *models.User) error
	GetByIDFunc          func(id uuid.UUID) (*models.User, error)
	GetByFirebaseUIDFunc func(firebaseUID string) (*models.User, error)
	GetByEmailFunc       func(email string) (*models.User, error)
	UpdateFunc           func(user *models.User) error
	DeleteFunc           func(id uuid.UUID) error
	ListFunc             func(offset, limit int) ([]*models.User, error)
}

func (m *MockUserRepository) Create(user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByFirebaseUID(firebaseUID string) (*models.User, error) {
	if m.GetByFirebaseUIDFunc != nil {
		return m.GetByFirebaseUIDFunc(firebaseUID)
	}
	args := m.Called(firebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(offset, limit int) ([]*models.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(offset, limit)
	}
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

// MockRepositories is a mock implementation of the repositories struct
type MockRepositories struct {
	Users      *MockUserRepository
	Tribes     *postgres.TribeRepository
	Activities *postgres.ActivityRepository
	db         *sql.DB
}

func (r *MockRepositories) DB() *sql.DB {
	return r.db
}

func (r *MockRepositories) GetUserRepository() models.UserRepository {
	return r.Users
}

// setupTest creates a test environment with mocked repositories
func setupTest(t *testing.T) *MockRepositories {
	t.Helper()
	return &MockRepositories{
		Users: &MockUserRepository{},
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		headerValue   string
		expectedToken string
	}{
		{
			name:          "valid bearer token",
			headerValue:   "Bearer token123",
			expectedToken: "token123",
		},
		{
			name:          "missing bearer prefix",
			headerValue:   "token123",
			expectedToken: "",
		},
		{
			name:          "empty header",
			headerValue:   "",
			expectedToken: "",
		},
		{
			name:          "invalid format",
			headerValue:   "Bearer",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.headerValue != "" {
				req.Header.Set("Authorization", tt.headerValue)
			}

			token := extractToken(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
	}{
		{
			name: "authenticated request",
			setupContext: func(c *gin.Context) {
				c.Set("firebase_uid", "test-uid")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthenticated request",
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			router.Use(RequireAuth())

			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestFirebaseAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		token          string
		setupAuth      func() *MockFirebaseAuth
		expectedStatus int
		checkClaims    bool
	}{
		{
			name:  "valid token",
			token: "valid-token",
			setupAuth: func() *MockFirebaseAuth {
				mock := NewMockFirebaseAuth()
				mock.SetVerifyIDTokenFunc(func(ctx context.Context, token string) (*auth.Token, error) {
					return &auth.Token{
						UID: "test-uid",
						Claims: map[string]interface{}{
							"email": "test@example.com",
							"name":  "Test User",
						},
					}, nil
				})
				return mock
			},
			expectedStatus: http.StatusOK,
			checkClaims:    true,
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			setupAuth: func() *MockFirebaseAuth {
				mock := NewMockFirebaseAuth()
				mock.SetVerifyIDTokenFunc(func(ctx context.Context, token string) (*auth.Token, error) {
					return nil, errors.New("invalid token")
				})
				return mock
			},
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
		{
			name:  "missing token",
			token: "",
			setupAuth: func() *MockFirebaseAuth {
				return NewMockFirebaseAuth()
			},
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			fa := &FirebaseAuth{client: tt.setupAuth()}
			router.Use(fa.AuthMiddleware())

			router.GET("/test", func(c *gin.Context) {
				if tt.checkClaims {
					assert.Equal(t, "test-uid", GetFirebaseUID(c))
					assert.Equal(t, "test@example.com", GetUserEmail(c))
					assert.Equal(t, "Test User", GetUserName(c))
				}
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

var errNotFound = errors.New("not found")

func TestUserIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set up test environment
	setup := setupTest(t)

	// Configure mock user repository behavior
	setup.Users.GetByFirebaseUIDFunc = func(firebaseUID string) (*models.User, error) {
		if firebaseUID == "test-uid" {
			return &models.User{
				ID:          uuid.New(),
				FirebaseUID: firebaseUID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		}
		return nil, errNotFound
	}

	tests := []struct {
		name           string
		firebaseUID    string
		expectedStatus int
	}{
		{
			name:           "valid user",
			firebaseUID:    "test-uid",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not found",
			firebaseUID:    "unknown-uid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing firebase uid",
			firebaseUID:    "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.firebaseUID != "" {
					c.Set("firebase_uid", tt.firebaseUID)
				}
				c.Next()
			})
			router.Use(UserIDMiddleware(setup))

			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserIDMiddleware_RepositoryErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		firebaseUID    string
		mockRepo       *MockRepositories
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "database error",
			firebaseUID: "test-uid",
			mockRepo: func() *MockRepositories {
				repos := setupTest(t)
				repos.Users.GetByFirebaseUIDFunc = func(firebaseUID string) (*models.User, error) {
					return nil, errors.New("database error")
				}
				return repos
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("firebase_uid", tt.firebaseUID)
				c.Next()
			})
			router.Use(UserIDMiddleware(tt.mockRepo))

			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestUserIDMiddleware_NilRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("firebase_uid", "test-uid")
		c.Next()
	})

	// Use nil directly for the repository
	router.Use(UserIDMiddleware(nil))

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "user repository not configured")
}

func TestMiddlewareChaining(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupAuth      func() *MockFirebaseAuth
		setupRepo      func(*MockUserRepository)
		token          string
		expectedStatus int
	}{
		{
			name: "successful chain",
			setupAuth: func() *MockFirebaseAuth {
				mock := NewMockFirebaseAuth()
				mock.SetVerifyIDTokenFunc(func(ctx context.Context, token string) (*auth.Token, error) {
					return &auth.Token{
						UID: "test-uid",
						Claims: map[string]interface{}{
							"email": "test@example.com",
						},
					}, nil
				})
				return mock
			},
			setupRepo: func(repo *MockUserRepository) {
				repo.GetByFirebaseUIDFunc = func(firebaseUID string) (*models.User, error) {
					return &models.User{
						ID:          uuid.New(),
						FirebaseUID: "test-uid",
					}, nil
				}
			},
			token:          "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name: "auth failure breaks chain",
			setupAuth: func() *MockFirebaseAuth {
				mock := NewMockFirebaseAuth()
				mock.SetVerifyIDTokenFunc(func(ctx context.Context, token string) (*auth.Token, error) {
					return nil, errors.New("invalid token")
				})
				return mock
			},
			setupRepo: func(repo *MockUserRepository) {
				// Should not be called
			},
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			repos := setupTest(t)

			if tt.setupRepo != nil {
				tt.setupRepo(repos.Users)
			}

			fa := &FirebaseAuth{client: tt.setupAuth()}
			router.Use(fa.AuthMiddleware())
			router.Use(RequireAuth())
			router.Use(UserIDMiddleware(repos))

			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetterFunctions(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*gin.Context)
		getFunc  func(*gin.Context) string
		expected string
	}{
		{
			name: "get firebase uid - exists",
			setup: func(c *gin.Context) {
				c.Set("firebase_uid", "test-uid")
			},
			getFunc:  GetFirebaseUID,
			expected: "test-uid",
		},
		{
			name:     "get firebase uid - not exists",
			setup:    func(c *gin.Context) {},
			getFunc:  GetFirebaseUID,
			expected: "",
		},
		{
			name: "get user email - exists",
			setup: func(c *gin.Context) {
				c.Set("user_email", "test@example.com")
			},
			getFunc:  GetUserEmail,
			expected: "test@example.com",
		},
		{
			name:     "get user email - not exists",
			setup:    func(c *gin.Context) {},
			getFunc:  GetUserEmail,
			expected: "",
		},
		{
			name: "get user name - exists",
			setup: func(c *gin.Context) {
				c.Set("user_name", "Test User")
			},
			getFunc:  GetUserName,
			expected: "Test User",
		},
		{
			name:     "get user name - not exists",
			setup:    func(c *gin.Context) {},
			getFunc:  GetUserName,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			tt.setup(c)
			result := tt.getFunc(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewFirebaseAuth(t *testing.T) {
	tests := []struct {
		name            string
		credentialsFile string
		expectError     bool
	}{
		{
			name:            "invalid credentials file",
			credentialsFile: "nonexistent.json",
			expectError:     true,
		},
		{
			name:            "empty credentials file",
			credentialsFile: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewFirebaseAuth(tt.credentialsFile)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, auth)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, auth)
			}
		})
	}
}
