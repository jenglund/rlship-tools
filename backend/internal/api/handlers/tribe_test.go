package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

	// Create repositories
	repos := postgres.NewRepositories(db)

	// Create test user
	testUser := testutil.CreateTestUser(t, db)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Set up auth context middleware
	router.Use(func(c *gin.Context) {
		c.Set("firebase_uid", testUser.FirebaseUID)
		c.Set("user_id", testUser.ID)
		c.Next()
	})

	// Register handlers
	tribeHandler := NewTribeHandler(repos)
	api := router.Group("/api")
	tribeHandler.RegisterRoutes(api.Group("/tribes"))

	return router, repos, &testUser
}

func TestCreateTribe(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	tests := []struct {
		name       string
		request    CreateTribeRequest
		wantStatus int
		validate   func(t *testing.T, tribe *models.Tribe)
	}{
		{
			name: "valid tribe with minimal fields",
			request: CreateTribeRequest{
				Name:       "Test Tribe",
				Type:       models.TribeTypeCustom,
				Visibility: models.VisibilityPrivate,
				Metadata:   map[string]interface{}{},
			},
			wantStatus: http.StatusCreated,
			validate: func(t *testing.T, tribe *models.Tribe) {
				assert.NotNil(t, tribe)
				assert.Equal(t, "Test Tribe", tribe.Name)
				assert.Equal(t, models.TribeTypeCustom, tribe.Type)
				assert.Equal(t, models.VisibilityPrivate, tribe.Visibility)
				assert.NotEmpty(t, tribe.ID)
				assert.Equal(t, 1, tribe.Version)

				// Verify the creator was added as a member
				members, err := repos.Tribes.GetMembers(tribe.ID)
				require.NoError(t, err)
				require.Len(t, members, 1, "Creator should be added as a member")
				assert.Equal(t, testUser.ID, members[0].UserID)
				assert.Equal(t, models.MembershipFull, members[0].MembershipType)
			},
		},
		{
			name: "valid tribe with description and metadata",
			request: CreateTribeRequest{
				Name:        "Test Tribe with Description",
				Type:        models.TribeTypeCouple,
				Description: "A description of the tribe",
				Visibility:  models.VisibilityPrivate,
				Metadata:    map[string]interface{}{"key": "value"},
			},
			wantStatus: http.StatusCreated,
			validate: func(t *testing.T, tribe *models.Tribe) {
				assert.NotNil(t, tribe)
				assert.Equal(t, "Test Tribe with Description", tribe.Name)
				assert.Equal(t, models.TribeTypeCouple, tribe.Type)
				assert.Equal(t, "A description of the tribe", tribe.Description)
				assert.Equal(t, models.VisibilityPrivate, tribe.Visibility)
				assert.NotEmpty(t, tribe.Metadata)
				assert.Equal(t, "value", tribe.Metadata["key"])
			},
		},
		{
			name: "missing name",
			request: CreateTribeRequest{
				Name:       "",
				Type:       models.TribeTypeCustom,
				Visibility: models.VisibilityPrivate,
				Metadata:   map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, tribe *models.Tribe) {},
		},
		{
			name: "missing type",
			request: CreateTribeRequest{
				Name:       "Test Tribe",
				Visibility: models.VisibilityPrivate,
				Metadata:   map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, tribe *models.Tribe) {},
		},
		{
			name: "missing visibility",
			request: CreateTribeRequest{
				Name:     "Test Tribe",
				Type:     models.TribeTypeCustom,
				Metadata: map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, tribe *models.Tribe) {},
		},
		{
			name: "invalid type",
			request: CreateTribeRequest{
				Name:       "Test Tribe",
				Type:       "invalid_type",
				Visibility: models.VisibilityPrivate,
				Metadata:   map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, tribe *models.Tribe) {},
		},
		{
			name: "invalid visibility",
			request: CreateTribeRequest{
				Name:       "Test Tribe",
				Type:       models.TribeTypeCustom,
				Visibility: "invalid_visibility",
				Metadata:   map[string]interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, tribe *models.Tribe) {},
		},
		{
			name: "invalid metadata",
			request: CreateTribeRequest{
				Name:       "Test Tribe",
				Type:       models.TribeTypeCustom,
				Visibility: models.VisibilityPrivate,
				Metadata:   "not a map", // This will cause a type assertion failure
			},
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, tribe *models.Tribe) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/tribes/tribes", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusCreated {
				var response struct {
					Data models.Tribe `json:"data"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				tt.validate(t, &response.Data)
			} else {
				tt.validate(t, nil)
			}
		})
	}
}

func TestGetTribe(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		Name:       "Test Tribe",
		Type:       models.TribeTypeCustom,
		Visibility: models.VisibilityPrivate,
		Metadata:   models.JSONMap{},
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
			req := httptest.NewRequest(http.MethodGet, "/api/tribes/tribes/"+tt.tribeID, nil)
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
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		Name:       "Original Name",
		Type:       models.TribeTypeCustom,
		Visibility: models.VisibilityPrivate,
		Metadata:   models.JSONMap{},
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

			req := httptest.NewRequest(http.MethodPut, "/api/tribes/tribes/"+tt.tribeID, bytes.NewReader(body))
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
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		Name:       "Test Tribe",
		Type:       models.TribeTypeCustom,
		Visibility: models.VisibilityPrivate,
		Metadata:   models.JSONMap{},
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
			req := httptest.NewRequest(http.MethodDelete, "/api/tribes/tribes/"+tt.tribeID, nil)
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
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		Name:       "Test Tribe",
		Type:       models.TribeTypeCustom,
		Visibility: models.VisibilityPrivate,
		Metadata:   models.JSONMap{},
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	// Create another test user using the same database connection
	newUser := testutil.CreateTestUser(t, repos.DB())

	// Check initial member count
	initialMembers, err := repos.Tribes.GetMembers(tribe.ID)
	require.NoError(t, err)
	require.Len(t, initialMembers, 1, "Should start with 1 member (the creator)")

	tests := []struct {
		name       string
		tribeID    string
		request    AddMemberRequest
		wantStatus int
		validate   func(t *testing.T, repos *postgres.Repositories, tribeID uuid.UUID)
	}{
		{
			name:    "valid member addition",
			tribeID: tribe.ID.String(),
			request: AddMemberRequest{
				UserID: newUser.ID,
			},
			wantStatus: http.StatusNoContent,
			validate: func(t *testing.T, repos *postgres.Repositories, tribeID uuid.UUID) {
				// Verify the member was added
				members, err := repos.Tribes.GetMembers(tribeID)
				require.NoError(t, err)
				require.Len(t, members, 2, "Should have 2 members after adding one")

				// Check if the new member is present
				var found bool
				for _, member := range members {
					if member.UserID == newUser.ID {
						found = true
						assert.Equal(t, models.MembershipFull, member.MembershipType, "Membership level should match")
						break
					}
				}
				assert.True(t, found, "New member should be in the members list")
			},
		},
		{
			name:    "add already existing member",
			tribeID: tribe.ID.String(),
			request: AddMemberRequest{
				UserID: testUser.ID, // This user is already a member
			},
			wantStatus: http.StatusInternalServerError, // The current implementation returns 500 for this case
			validate: func(t *testing.T, repos *postgres.Repositories, tribeID uuid.UUID) {
				// Member count should remain unchanged
				members, err := repos.Tribes.GetMembers(tribeID)
				require.NoError(t, err)
				assert.Len(t, members, 2, "Member count should not change when adding an existing member")
			},
		},
		{
			name:    "non-existent tribe",
			tribeID: uuid.New().String(),
			request: AddMemberRequest{
				UserID: newUser.ID,
			},
			wantStatus: http.StatusInternalServerError,
			validate:   func(t *testing.T, repos *postgres.Repositories, tribeID uuid.UUID) {},
		},
		{
			name:    "invalid tribe uuid",
			tribeID: "invalid-uuid",
			request: AddMemberRequest{
				UserID: newUser.ID,
			},
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, repos *postgres.Repositories, tribeID uuid.UUID) {},
		},
		{
			name:    "non-existent user",
			tribeID: tribe.ID.String(),
			request: AddMemberRequest{
				UserID: uuid.New(), // Random non-existent user ID
			},
			wantStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, repos *postgres.Repositories, tribeID uuid.UUID) {
				// Member count should remain unchanged
				members, err := repos.Tribes.GetMembers(tribeID)
				require.NoError(t, err)
				assert.Len(t, members, 2, "Member count should not change when adding a non-existent user")
			},
		},
		{
			name:       "malformed request body",
			tribeID:    tribe.ID.String(),
			request:    AddMemberRequest{}, // Empty UserID
			wantStatus: http.StatusBadRequest,
			validate:   func(t *testing.T, repos *postgres.Repositories, tribeID uuid.UUID) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.name == "malformed request body" {
				body = []byte(`{"invalid_field": "value"}`) // Deliberately malformed JSON
			} else {
				body, err = json.Marshal(tt.request)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/tribes/tribes/"+tt.tribeID+"/members", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			// Run additional validations if specified and if the tribe ID is valid
			if tt.validate != nil && tt.tribeID != "invalid-uuid" {
				var tribeID uuid.UUID
				if tt.tribeID == tribe.ID.String() {
					tribeID = tribe.ID
				} else {
					tribeID, _ = uuid.Parse(tt.tribeID)
				}
				tt.validate(t, repos, tribeID)
			}
		})
	}
}

func TestRemoveMember(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe and add test users as members
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		Name:       "Test Tribe",
		Type:       models.TribeTypeCustom,
		Visibility: models.VisibilityPrivate,
		Metadata:   models.JSONMap{},
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)

	// Add the test user
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	// Create and add another test user
	memberToRemove := testutil.CreateTestUser(t, repos.DB())
	err = repos.Tribes.AddMember(tribe.ID, memberToRemove.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	// Create a third user that is not a member
	nonMember := testutil.CreateTestUser(t, repos.DB())

	// Verify initial members
	members, err := repos.Tribes.GetMembers(tribe.ID)
	require.NoError(t, err)
	require.Len(t, members, 2)

	tests := []struct {
		name       string
		tribeID    string
		userID     string
		wantStatus int
	}{
		{
			name:       "remove existing member",
			tribeID:    tribe.ID.String(),
			userID:     memberToRemove.ID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "remove non-existent member",
			tribeID:    tribe.ID.String(),
			userID:     nonMember.ID.String(),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid tribe id",
			tribeID:    "invalid-uuid",
			userID:     memberToRemove.ID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid user id",
			tribeID:    tribe.ID.String(),
			userID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "non-existent tribe",
			tribeID:    uuid.New().String(),
			userID:     memberToRemove.ID.String(),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/tribes/tribes/"+tt.tribeID+"/members/"+tt.userID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			// If the test was for successful removal, verify the member was removed
			if tt.name == "remove existing member" {
				// Verify the member was removed
				updatedMembers, err := repos.Tribes.GetMembers(tribe.ID)
				require.NoError(t, err)
				require.Len(t, updatedMembers, 1, "Should have one less member after removal")

				// The remaining member should be the test user
				assert.Equal(t, testUser.ID, updatedMembers[0].UserID, "The remaining member should be the test user")
			}
		})
	}
}

func TestListMembers(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribe and add members
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		Name:       "Test Tribe",
		Type:       models.TribeTypeCustom,
		Visibility: models.VisibilityPrivate,
		Metadata:   models.JSONMap{},
	}
	err := repos.Tribes.Create(tribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	// Add another member using the same database connection
	newUser := testutil.CreateTestUser(t, repos.DB())
	err = repos.Tribes.AddMember(tribe.ID, newUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/tribes/tribes/"+tribe.ID.String()+"/members", nil)
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

func TestListTribes(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribes
	for i := 0; i < 3; i++ {
		tribe := &models.Tribe{
			BaseModel: models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			Name:       "Test Tribe " + time.Now().String(),
			Type:       models.TribeTypeCustom,
			Visibility: models.VisibilityPrivate,
			Metadata:   models.JSONMap{},
		}
		err := repos.Tribes.Create(tribe)
		require.NoError(t, err)
		err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
		require.NoError(t, err)
	}

	// Test listing all tribes
	req := httptest.NewRequest(http.MethodGet, "/api/tribes/tribes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []*models.Tribe `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response.Data), 3, "Should find at least the 3 tribes we created")
}

func TestListMyTribes(t *testing.T) {
	router, repos, testUser := setupTribeTest(t)

	// Create test tribes where the user is a member
	userTribes := 3
	for i := 0; i < userTribes; i++ {
		tribe := &models.Tribe{
			BaseModel: models.BaseModel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			Name:       "User's Tribe " + time.Now().String(),
			Type:       models.TribeTypeCustom,
			Visibility: models.VisibilityPrivate,
			Metadata:   models.JSONMap{},
		}
		err := repos.Tribes.Create(tribe)
		require.NoError(t, err)
		err = repos.Tribes.AddMember(tribe.ID, testUser.ID, models.MembershipFull, nil, nil)
		require.NoError(t, err)
	}

	// Create a tribe where the user is not a member
	otherUser := testutil.CreateTestUser(t, repos.DB())
	otherTribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Version:   1,
		},
		Name:       "Other User's Tribe",
		Type:       models.TribeTypeCustom,
		Visibility: models.VisibilityPrivate,
		Metadata:   models.JSONMap{},
	}
	err := repos.Tribes.Create(otherTribe)
	require.NoError(t, err)
	err = repos.Tribes.AddMember(otherTribe.ID, otherUser.ID, models.MembershipFull, nil, nil)
	require.NoError(t, err)

	// Test listing the user's tribes
	req := httptest.NewRequest(http.MethodGet, "/api/tribes/tribes/my", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []*models.Tribe `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Data, userTribes, "Should only return tribes where the user is a member")

	// Verify that the other tribe is not included in the results
	for _, tribe := range response.Data {
		assert.NotEqual(t, otherTribe.ID, tribe.ID, "Other user's tribe should not be included")
	}
}
