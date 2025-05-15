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

func setupTribeTest(t *testing.T) (*gin.Engine, *postgres.Repositories, *testutil.TestUser) {
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

	// Mock the authentication middleware with the test user's Firebase UID
	router.Use(testutil.MockAuthMiddleware(testUser.ID.String(), testUser.FirebaseUID))

	// Set up tribe routes
	handler := NewTribeHandler(repos)
	api := router.Group("/api")
	handler.RegisterRoutes(api)

	return router, repos, &testUser
}

func TestCreateTribe(t *testing.T) {
	router, _, _ := setupTribeTest(t)

	tests := []struct {
		name       string
		request    CreateTribeRequest
		wantStatus int
	}{
		{
			name: "valid tribe",
			request: CreateTribeRequest{
				Name: "Test Tribe",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: CreateTribeRequest{
				Name: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/tribes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var response struct {
					Data *models.Tribe `json:"data"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.request.Name, response.Data.Name)
				assert.NotEmpty(t, response.Data.Members)
			}
		})
	}
}

func TestGetTribe(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe
	tribe := &models.Tribe{
		Name: "Test Tribe",
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		tribeID    string
		wantStatus int
	}{
		{
			name:       "existing tribe",
			tribeID:    tribe.ID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent tribe",
			tribeID:    uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid uuid",
			tribeID:    "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/tribes/"+tt.tribeID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Data *models.Tribe `json:"data"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tribe.ID, response.Data.ID)
			}
		})
	}
}

func TestUpdateTribe(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe
	tribe := &models.Tribe{
		Name: "Original Name",
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		tribeID    string
		request    UpdateTribeRequest
		wantStatus int
	}{
		{
			name:    "valid update",
			tribeID: tribe.ID.String(),
			request: UpdateTribeRequest{
				Name: "Updated Name",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "non-existent tribe",
			tribeID: uuid.New().String(),
			request: UpdateTribeRequest{
				Name: "Updated Name",
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/tribes/"+tt.tribeID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Data *models.Tribe `json:"data"`
				}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.request.Name, response.Data.Name)
			}
		})
	}
}

func TestDeleteTribe(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe
	tribe := &models.Tribe{
		Name: "Test Tribe",
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		tribeID    string
		wantStatus int
	}{
		{
			name:       "existing tribe",
			tribeID:    tribe.ID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "non-existent tribe",
			tribeID:    uuid.New().String(),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid uuid",
			tribeID:    "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/tribes/"+tt.tribeID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAddMember(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe and another user
	tribe := &models.Tribe{
		Name: "Test Tribe",
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	// Create another test user using the same database connection
	newUser := testutil.CreateTestUser(t, repos.DB())

	tests := []struct {
		name       string
		tribeID    string
		request    AddMemberRequest
		wantStatus int
	}{
		{
			name:    "valid member addition",
			tribeID: tribe.ID.String(),
			request: AddMemberRequest{
				UserID: newUser.ID,
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "non-existent tribe",
			tribeID: uuid.New().String(),
			request: AddMemberRequest{
				UserID: newUser.ID,
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "invalid tribe uuid",
			tribeID: "invalid-uuid",
			request: AddMemberRequest{
				UserID: newUser.ID,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/tribes/"+tt.tribeID+"/members", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListMembers(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe and add members
	tribe := &models.Tribe{
		Name: "Test Tribe",
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	// Add another member using the same database connection
	newUser := testutil.CreateTestUser(t, repos.DB())
	err = repos.Tribes.AddMember(tribe.ID, newUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/tribes/"+tribe.ID.String()+"/members", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []*models.TribeMember `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, 2)
}
