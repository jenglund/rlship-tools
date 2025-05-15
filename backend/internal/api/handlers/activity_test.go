package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/repository/postgres"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupActivityTest(t *testing.T) (*gin.Engine, *postgres.Repositories, *testutil.TestUser) {
	// Set up test database
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.TeardownTestDB(t, db)
	})

	// Create repositories
	repos := postgres.NewRepositories(db)

	// Create test user
	testUser := testutil.CreateTestUser(t, db)

	// Set up router with authentication middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock the authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("firebase_uid", testUser.FirebaseUID)
		c.Set("user_id", testUser.ID.String())
		c.Next()
	})

	// Set up activity routes
	handler := NewActivityHandler(repos)
	api := router.Group("/api")
	handler.RegisterRoutes(api)

	return router, repos, &testUser
}

func TestCreateActivity(t *testing.T) {
	router, repos, testUser := setupActivityTest(t)

	tests := []struct {
		name       string
		request    CreateActivityRequest
		wantStatus int
	}{
		{
			name: "valid activity",
			request: CreateActivityRequest{
				Type:        models.ActivityTypeLocation,
				Name:        "Test Activity",
				Description: "Test Description",
				Visibility:  models.VisibilityPublic,
				Metadata:    map[string]interface{}{"key": "value"},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			request: CreateActivityRequest{
				Description: "Test Description",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/activities", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var response struct {
					Data *models.Activity `json:"data"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.request.Name, response.Data.Name)
				assert.Equal(t, tt.request.Type, response.Data.Type)

				// Get owners and verify
				owners, err := repos.Activities.GetOwners(response.Data.ID)
				require.NoError(t, err)
				require.Len(t, owners, 1)
				assert.Equal(t, testUser.ID, owners[0].OwnerID)
			}
		})
	}
}

func TestGetActivity(t *testing.T) {
	router, repos, testUser := setupActivityTest(t)

	// Create test activity
	activity := &models.Activity{
		Type:        models.ActivityTypeLocation,
		Name:        "Test Activity",
		Description: "Test Description",
		Visibility:  models.VisibilityPublic,
	}
	err := repos.Activities.Create(activity)
	require.NoError(t, err)
	err = repos.Activities.AddOwner(activity.ID, testUser.ID, "user")
	require.NoError(t, err)

	tests := []struct {
		name       string
		activityID string
		wantStatus int
	}{
		{
			name:       "existing activity",
			activityID: activity.ID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent activity",
			activityID: uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid uuid",
			activityID: "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/activities/"+tt.activityID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Data *models.Activity `json:"data"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, activity.ID, response.Data.ID)

				// Get owners and verify
				owners, err := repos.Activities.GetOwners(response.Data.ID)
				require.NoError(t, err)
				require.Len(t, owners, 1)
				assert.Equal(t, testUser.ID, owners[0].OwnerID)
			}
		})
	}
}

func TestUpdateActivity(t *testing.T) {
	router, repos, testUser := setupActivityTest(t)

	// Create test activity
	activity := &models.Activity{
		Type:        models.ActivityTypeLocation,
		Name:        "Original Name",
		Description: "Original Description",
		Visibility:  models.VisibilityPrivate,
	}
	err := repos.Activities.Create(activity)
	require.NoError(t, err)
	err = repos.Activities.AddOwner(activity.ID, testUser.ID, "user")
	require.NoError(t, err)

	tests := []struct {
		name       string
		activityID string
		request    UpdateActivityRequest
		wantStatus int
	}{
		{
			name:       "valid update",
			activityID: activity.ID.String(),
			request: UpdateActivityRequest{
				Name:        "Updated Name",
				Description: "Updated Description",
				Visibility:  models.VisibilityPublic,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent activity",
			activityID: uuid.New().String(),
			request: UpdateActivityRequest{
				Name: "Updated Name",
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/activities/"+tt.activityID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Data *models.Activity `json:"data"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.request.Name, response.Data.Name)

				// Get owners and verify
				owners, err := repos.Activities.GetOwners(response.Data.ID)
				require.NoError(t, err)
				require.Len(t, owners, 1)
				assert.Equal(t, testUser.ID, owners[0].OwnerID)
			}
		})
	}
}

func TestDeleteActivity(t *testing.T) {
	router, repos, testUser := setupActivityTest(t)

	// Create test activity
	activity := &models.Activity{
		Type:        models.ActivityTypeLocation,
		Name:        "Test Activity",
		Description: "Test Description",
		Visibility:  models.VisibilityPublic,
	}
	err := repos.Activities.Create(activity)
	require.NoError(t, err)
	err = repos.Activities.AddOwner(activity.ID, testUser.ID, "user")
	require.NoError(t, err)

	tests := []struct {
		name       string
		activityID string
		wantStatus int
	}{
		{
			name:       "existing activity",
			activityID: activity.ID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "non-existent activity",
			activityID: uuid.New().String(),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid uuid",
			activityID: "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/activities/"+tt.activityID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestShareActivity(t *testing.T) {
	router, repos, testUser := setupActivityTest(t)

	// Create test activity and tribe
	activity := &models.Activity{
		Type:        models.ActivityTypeLocation,
		Name:        "Test Activity",
		Description: "Test Description",
		Visibility:  models.VisibilityPublic,
	}
	err := repos.Activities.Create(activity)
	require.NoError(t, err)
	err = repos.Activities.AddOwner(activity.ID, testUser.ID, "user")
	require.NoError(t, err)

	// Create test tribe using the same database connection
	tribe := testutil.CreateTestTribe(t, repos.DB(), []testutil.TestUser{*testUser})

	tests := []struct {
		name       string
		activityID string
		request    ShareActivityRequest
		wantStatus int
	}{
		{
			name:       "valid share",
			activityID: activity.ID.String(),
			request: ShareActivityRequest{
				TribeID:   tribe.ID,
				ExpiresAt: nil,
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "non-existent activity",
			activityID: uuid.New().String(),
			request: ShareActivityRequest{
				TribeID: tribe.ID,
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid activity uuid",
			activityID: "invalid-uuid",
			request: ShareActivityRequest{
				TribeID: tribe.ID,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/activities/"+tt.activityID+"/share", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListSharedActivities(t *testing.T) {
	router, repos, testUser := setupActivityTest(t)

	// Create test tribe using the same database connection
	tribe := testutil.CreateTestTribe(t, repos.DB(), []testutil.TestUser{*testUser})

	// Create test activity
	activity := &models.Activity{
		Type:        models.ActivityTypeLocation,
		Name:        "Test Activity",
		Description: "Test Description",
		Visibility:  models.VisibilityPublic,
	}
	err := repos.Activities.Create(activity)
	require.NoError(t, err)
	err = repos.Activities.AddOwner(activity.ID, testUser.ID, "user")
	require.NoError(t, err)

	// Share activity with tribe
	err = repos.Activities.ShareWithTribe(activity.ID, tribe.ID, testUser.ID, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/activities/shared", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []*models.Activity `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Data)
}
