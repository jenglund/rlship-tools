package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateUser(t *testing.T) {
	setup := testutil.SetupTest(t)
	handler := NewUserHandler(setup.Repositories)

	tests := []struct {
		name          string
		firebaseUID   string
		provider      string
		email         string
		userName      string
		avatarURL     string
		setupMocks    func()
		expectedCode  int
		expectedError bool
	}{
		{
			name:         "new user",
			firebaseUID:  "new-firebase-uid",
			provider:     "google",
			email:        "new@example.com",
			userName:     "New User",
			avatarURL:    "https://example.com/avatar.jpg",
			expectedCode: http.StatusCreated,
		},
		{
			name:         "existing user",
			firebaseUID:  setup.TestUser.FirebaseUID,
			provider:     string(setup.TestUser.Provider),
			email:        setup.TestUser.Email,
			userName:     setup.TestUser.Name,
			avatarURL:    setup.TestUser.AvatarURL,
			expectedCode: http.StatusOK,
		},
		{
			name:          "missing required fields",
			firebaseUID:   "",
			provider:      "",
			email:         "",
			userName:      "",
			avatarURL:     "",
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
		{
			name:        "create error",
			firebaseUID: "new-firebase-uid",
			provider:    "google",
			email:       "new@example.com",
			userName:    "New User",
			avatarURL:   "https://example.com/avatar.jpg",
			setupMocks: func() {
				setup.Repositories.Users.CreateFunc = func(user *models.User) error {
					return errors.New("create error")
				}
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: true,
		},
		{
			name:        "update error",
			firebaseUID: setup.TestUser.FirebaseUID,
			provider:    string(setup.TestUser.Provider),
			email:       setup.TestUser.Email,
			userName:    setup.TestUser.Name,
			avatarURL:   setup.TestUser.AvatarURL,
			setupMocks: func() {
				setup.Repositories.Users.UpdateFunc = func(user *models.User) error {
					return errors.New("update error")
				}
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request body
			body := map[string]interface{}{
				"firebase_uid": tt.firebaseUID,
				"provider":     tt.provider,
				"email":        tt.email,
				"name":         tt.userName,
				"avatar_url":   tt.avatarURL,
			}
			jsonBody, err := json.Marshal(body)
			require.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/api/v1/users/authenticate", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.AuthenticateUser(c)

			assert.Equal(t, tt.expectedCode, w.Code)

			if !tt.expectedError {
				var response struct {
					Success bool        `json:"success"`
					Data    models.User `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, tt.firebaseUID, response.Data.FirebaseUID)
				assert.Equal(t, models.AuthProvider(tt.provider), response.Data.Provider)
				assert.Equal(t, tt.email, response.Data.Email)
				assert.Equal(t, tt.userName, response.Data.Name)
				assert.Equal(t, tt.avatarURL, response.Data.AvatarURL)
			}
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	setup := testutil.SetupTest(t)
	handler := NewUserHandler(setup.Repositories)

	tests := []struct {
		name          string
		setupContext  func(*gin.Context)
		expectedCode  int
		expectedError bool
	}{
		{
			name: "existing user",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", setup.TestUser.ID)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "user not found",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			expectedCode:  http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "missing user_id",
			setupContext:  func(c *gin.Context) {},
			expectedCode:  http.StatusUnauthorized,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/v1/users/me", nil)
			tt.setupContext(c)

			handler.GetCurrentUser(c)

			assert.Equal(t, tt.expectedCode, w.Code)

			if !tt.expectedError {
				var response struct {
					Success bool        `json:"success"`
					Data    models.User `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, setup.TestUser.ID, response.Data.ID)
				assert.Equal(t, setup.TestUser.FirebaseUID, response.Data.FirebaseUID)
				assert.Equal(t, setup.TestUser.Email, response.Data.Email)
			}
		})
	}
}

func TestUpdateCurrentUser(t *testing.T) {
	setup := testutil.SetupTest(t)
	handler := NewUserHandler(setup.Repositories)

	tests := []struct {
		name          string
		setupContext  func(*gin.Context)
		updateData    map[string]interface{}
		setupMocks    func()
		expectedCode  int
		expectedError bool
	}{
		{
			name: "valid update",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", setup.TestUser.ID)
			},
			updateData: map[string]interface{}{
				"name":       "Updated Name",
				"avatar_url": "https://example.com/new-avatar.jpg",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "user not found",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", uuid.New())
			},
			updateData: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedCode:  http.StatusNotFound,
			expectedError: true,
		},
		{
			name:         "missing user_id",
			setupContext: func(c *gin.Context) {},
			updateData: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: true,
		},
		{
			name: "invalid request body",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", setup.TestUser.ID)
			},
			updateData: map[string]interface{}{
				"invalid_field": make(chan int), // This will cause JSON marshaling to fail
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
		{
			name: "update error",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", setup.TestUser.ID)
			},
			updateData: map[string]interface{}{
				"name": "Updated Name",
			},
			setupMocks: func() {
				setup.Repositories.Users.UpdateFunc = func(user *models.User) error {
					return errors.New("update error")
				}
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setupContext(c)

			var jsonBody []byte
			var err error
			if tt.updateData != nil {
				jsonBody, err = json.Marshal(tt.updateData)
				if err == nil {
					c.Request = httptest.NewRequest("PUT", "/api/v1/users/me", bytes.NewBuffer(jsonBody))
					c.Request.Header.Set("Content-Type", "application/json")
				}
			}

			handler.UpdateCurrentUser(c)

			assert.Equal(t, tt.expectedCode, w.Code)

			if !tt.expectedError {
				var response struct {
					Success bool        `json:"success"`
					Data    models.User `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				if tt.updateData["name"] != nil {
					assert.Equal(t, tt.updateData["name"], response.Data.Name)
				}
				if tt.updateData["avatar_url"] != nil {
					assert.Equal(t, tt.updateData["avatar_url"], response.Data.AvatarURL)
				}
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	setup := testutil.SetupTest(t)
	handler := NewUserHandler(setup.Repositories)

	tests := []struct {
		name          string
		userID        string
		expectedCode  int
		expectedError bool
	}{
		{
			name:         "existing user",
			userID:       setup.TestUser.ID.String(),
			expectedCode: http.StatusOK,
		},
		{
			name:          "non-existent user",
			userID:        uuid.New().String(),
			expectedCode:  http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "invalid uuid",
			userID:        "invalid-uuid",
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/v1/users/"+tt.userID, nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.userID}}

			handler.GetUser(c)

			assert.Equal(t, tt.expectedCode, w.Code)

			if !tt.expectedError {
				var response struct {
					Success bool        `json:"success"`
					Data    models.User `json:"data"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response.Success)
				assert.Equal(t, setup.TestUser.ID, response.Data.ID)
				assert.Equal(t, setup.TestUser.FirebaseUID, response.Data.FirebaseUID)
				assert.Equal(t, setup.TestUser.Email, response.Data.Email)
			}
		})
	}
}
