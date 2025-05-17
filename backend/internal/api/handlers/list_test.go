package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jenglund/rlship-tools/internal/models"
)

type MockListService struct {
	mock.Mock
}

func (m *MockListService) CreateList(list *models.List) error {
	args := m.Called(list)
	return args.Error(0)
}

func (m *MockListService) GetList(id uuid.UUID) (*models.List, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.List), args.Error(1)
}

func (m *MockListService) UpdateList(list *models.List) error {
	args := m.Called(list)
	return args.Error(0)
}

func (m *MockListService) DeleteList(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockListService) AddListItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListService) GetListItems(listID uuid.UUID) ([]*models.ListItem, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListService) UpdateListItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListService) RemoveListItem(listID, itemID uuid.UUID) error {
	args := m.Called(listID, itemID)
	return args.Error(0)
}

func (m *MockListService) GenerateMenu(params *models.MenuParams) ([]*models.List, error) {
	args := m.Called(params)
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) SyncList(listID uuid.UUID) error {
	args := m.Called(listID)
	return args.Error(0)
}

func (m *MockListService) GetListConflicts(listID uuid.UUID) ([]*models.SyncConflict, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.SyncConflict), args.Error(1)
}

func (m *MockListService) ResolveListConflict(listID, conflictID uuid.UUID, resolution string) error {
	args := m.Called(listID, conflictID, resolution)
	return args.Error(0)
}

func (m *MockListService) AddListOwner(listID, ownerID uuid.UUID, ownerType string) error {
	args := m.Called(listID, ownerID, ownerType)
	return args.Error(0)
}

func (m *MockListService) RemoveListOwner(listID, ownerID uuid.UUID) error {
	args := m.Called(listID, ownerID)
	return args.Error(0)
}

func (m *MockListService) GetListOwners(listID uuid.UUID) ([]*models.ListOwner, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListOwner), args.Error(1)
}

func (m *MockListService) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) ShareListWithTribe(listID, tribeID, userID uuid.UUID, expiresAt *time.Time) error {
	args := m.Called(listID, tribeID, userID, expiresAt)
	return args.Error(0)
}

func (m *MockListService) UnshareListWithTribe(listID, tribeID, userID uuid.UUID) error {
	args := m.Called(listID, tribeID, userID)
	return args.Error(0)
}

func (m *MockListService) GetSharedLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) List(offset, limit int) ([]*models.List, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) GetListShares(listID uuid.UUID) ([]*models.ListShare, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListShare), args.Error(1)
}

func timeMatch(expected *time.Time) interface{} {
	return mock.MatchedBy(func(actual *time.Time) bool {
		if expected == nil {
			return actual == nil
		}
		if actual == nil {
			return false
		}
		return expected.Equal(*actual)
	})
}

func TestListHandler(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	t.Run("CreateList", func(t *testing.T) {
		list := &models.List{
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			DefaultWeight: 1.0,
		}

		mockService.On("CreateList", mock.AnythingOfType("*models.List")).Return(nil)

		body, err := json.Marshal(list)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/lists", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetList", func(t *testing.T) {
		listID := uuid.New()
		list := &models.List{
			ID:            listID,
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			DefaultWeight: 1.0,
		}

		mockService.On("GetList", listID).Return(list, nil)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/%s", listID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response models.List
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, list.ID, response.ID)
		assert.Equal(t, list.Name, response.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("ListLists", func(t *testing.T) {
		lists := []*models.List{
			{
				ID:   uuid.New(),
				Name: "List 1",
			},
			{
				ID:   uuid.New(),
				Name: "List 2",
			},
		}

		mockService.On("List", 0, 20).Return(lists, nil)

		req := httptest.NewRequest(http.MethodGet, "/lists", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []*models.List
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)

		mockService.AssertExpectations(t)
	})

	t.Run("GenerateMenu", func(t *testing.T) {
		listIDs := []uuid.UUID{uuid.New(), uuid.New()}
		filters := map[string]interface{}{
			"cooldown_days": float64(7),
			"max_items":     float64(2),
		}

		lists := []*models.List{
			{
				ID:   uuid.New(),
				Name: "List 1",
			},
			{
				ID:   uuid.New(),
				Name: "List 2",
			},
		}

		request := struct {
			ListIDs []uuid.UUID            `json:"list_ids"`
			Filters map[string]interface{} `json:"filters"`
		}{
			ListIDs: listIDs,
			Filters: filters,
		}

		expectedParams := &models.MenuParams{
			ListIDs: listIDs,
			Filters: filters,
		}

		mockService.On("GenerateMenu", expectedParams).Return(lists, nil)

		body, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/lists/menu", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []*models.List
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, lists[0].ID, response[0].ID)
		assert.Equal(t, lists[1].ID, response[1].ID)

		mockService.AssertExpectations(t)
	})

	t.Run("SyncList", func(t *testing.T) {
		listID := uuid.New()
		mockService.On("SyncList", listID).Return(nil)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/sync", listID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetListConflicts", func(t *testing.T) {
		listID := uuid.New()
		now := time.Now()
		conflicts := []*models.SyncConflict{
			{
				ID:         uuid.New(),
				ListID:     listID,
				Type:       "modified",
				LocalData:  "Local Name",
				RemoteData: "Remote Name",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		}

		mockService.On("GetListConflicts", listID).Return(conflicts, nil)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/%s/conflicts", listID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []*models.SyncConflict
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 1)

		mockService.AssertExpectations(t)
	})

	t.Run("ResolveListConflict", func(t *testing.T) {
		listID := uuid.New()
		conflictID := uuid.New()
		resolution := "accept"
		mockService.On("ResolveListConflict", listID, conflictID, resolution).Return(nil)

		reqBody := struct {
			Resolution string `json:"resolution"`
		}{
			Resolution: resolution,
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/conflicts/%s/resolve", listID, conflictID), bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("List Ownership", func(t *testing.T) {
		listID := uuid.New()
		userID := uuid.New()
		tribeID := uuid.New()

		// Test AddListOwner
		ownerReq := struct {
			OwnerID   uuid.UUID `json:"owner_id"`
			OwnerType string    `json:"owner_type"`
		}{
			OwnerID:   userID,
			OwnerType: "user",
		}

		mockService.On("AddListOwner", listID, userID, "user").Return(nil)

		body, err := json.Marshal(ownerReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/owners", listID), bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)

		// Test GetListOwners
		owners := []*models.ListOwner{
			{ListID: listID, OwnerID: userID, OwnerType: models.OwnerTypeUser},
			{ListID: listID, OwnerID: tribeID, OwnerType: models.OwnerTypeTribe},
		}
		mockService.On("GetListOwners", listID).Return(owners, nil)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/%s/owners", listID), nil)
		rec = httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []*models.ListOwner
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		mockService.AssertExpectations(t)

		// Test GetUserLists
		lists := []*models.List{
			{ID: listID, Name: "Test List"},
		}
		mockService.On("GetUserLists", userID).Return(lists, nil)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/user/%s", userID), nil)
		rec = httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var listResponse []*models.List
		err = json.NewDecoder(rec.Body).Decode(&listResponse)
		require.NoError(t, err)
		assert.Len(t, listResponse, 1)
		mockService.AssertExpectations(t)

		// Test GetTribeLists
		mockService.On("GetTribeLists", tribeID).Return(lists, nil)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/tribe/%s", tribeID), nil)
		rec = httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		err = json.NewDecoder(rec.Body).Decode(&listResponse)
		require.NoError(t, err)
		assert.Len(t, listResponse, 1)
		mockService.AssertExpectations(t)
	})

	t.Run("List_Sharing", func(t *testing.T) {
		listID := uuid.New()
		tribeID := uuid.New()
		userID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)

		// Set up the mock expectations
		mockService.On("GetListOwners", listID).Return([]*models.ListOwner{
			{ListID: listID, OwnerID: userID, OwnerType: "user"},
		}, nil)
		mockService.On("ShareListWithTribe", listID, tribeID, userID, mock.MatchedBy(func(t *time.Time) bool {
			return t != nil && t.Unix() == expiresAt.Unix()
		})).Return(nil)

		// Create the request body
		reqBody := struct {
			TribeID   uuid.UUID `json:"tribe_id"`
			ExpiresAt time.Time `json:"expires_at"`
		}{
			TribeID:   tribeID,
			ExpiresAt: expiresAt,
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		// Create the request with user ID in context
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/share", listID), bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
		rec := httptest.NewRecorder()

		// Send the request
		router.ServeHTTP(rec, req)

		// Verify the response
		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("GetListShares", func(t *testing.T) {
		mockService := new(MockListService)
		handler := NewListHandler(mockService)

		validListID := uuid.New()
		validUserID := uuid.New()
		validShares := []*models.ListShare{
			{
				ListID:    validListID,
				TribeID:   uuid.New(),
				UserID:    validUserID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		tests := []struct {
			name           string
			listID         string
			setupAuth      func(*http.Request)
			setupMocks     func()
			expectedStatus int
			expectedError  string
		}{
			{
				name:   "success",
				listID: validListID.String(),
				setupAuth: func(r *http.Request) {
					ctx := context.WithValue(r.Context(), "user_id", validUserID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func() {
					mockService.On("GetListOwners", validListID).Return([]*models.ListOwner{
						{OwnerID: validUserID, OwnerType: "user"},
					}, nil)
					mockService.On("GetListShares", validListID).Return(validShares, nil)
				},
				expectedStatus: http.StatusOK,
			},
			{
				name:   "invalid list ID",
				listID: "invalid-uuid",
				setupAuth: func(r *http.Request) {
					ctx := context.WithValue(r.Context(), "user_id", validUserID)
					*r = *r.WithContext(ctx)
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid list ID",
			},
			{
				name:   "unauthorized",
				listID: validListID.String(),
				setupAuth: func(r *http.Request) {
					ctx := context.WithValue(r.Context(), "user_id", uuid.New())
					*r = *r.WithContext(ctx)
				},
				setupMocks: func() {
					mockService.On("GetListOwners", validListID).Return([]*models.ListOwner{
						{OwnerID: uuid.New(), OwnerType: "user"},
					}, nil)
				},
				expectedStatus: http.StatusForbidden,
				expectedError:  "you do not have permission to view this list's shares",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.listID)

				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

				if tt.setupAuth != nil {
					tt.setupAuth(r)
				}

				if tt.setupMocks != nil {
					tt.setupMocks()
				}

				handler.GetListShares(w, r)

				assert.Equal(t, tt.expectedStatus, w.Code)

				if tt.expectedError != "" {
					var response struct {
						Success bool `json:"success"`
						Error   struct {
							Message string `json:"message"`
						} `json:"error"`
					}
					err := json.NewDecoder(w.Body).Decode(&response)
					require.NoError(t, err)
					assert.False(t, response.Success)
					assert.Equal(t, tt.expectedError, response.Error.Message)
				} else if tt.expectedStatus == http.StatusOK {
					var response []*models.ListShare
					err := json.NewDecoder(w.Body).Decode(&response)
					require.NoError(t, err)

					// Check IDs and other fields, but not timestamps which may differ
					require.Equal(t, len(validShares), len(response))
					for i, share := range validShares {
						assert.Equal(t, share.ListID, response[i].ListID)
						assert.Equal(t, share.TribeID, response[i].TribeID)
						assert.Equal(t, share.UserID, response[i].UserID)
						assert.Equal(t, share.Version, response[i].Version)
					}
				}

				mockService.AssertExpectations(t)
			})
		}
	})
}

func TestListHandler_ShareListWithTribe(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		tribeID        string
		requestBody    []byte
		expectedStatus int
		expectedError  string
		setupAuth      func(*http.Request)
		setupMocks     func(*MockListService)
		setupRequest   func() *http.Request
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			tribeID:        uuid.New().String(),
			requestBody:    []byte(`{"expires_at": null}`),
			expectedStatus: http.StatusNoContent,
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", uuid.New())
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				mockService.On("ShareListWithTribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{"expires_at": null}`)))
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			tribeID:        uuid.New().String(),
			requestBody:    []byte(`{"expires_at": null}`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid list ID",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{"expires_at": null}`)))
			},
		},
		{
			name:           "list not found",
			listID:         uuid.New().String(),
			tribeID:        uuid.New().String(),
			requestBody:    []byte(`{"expires_at": null}`),
			expectedStatus: http.StatusNotFound,
			expectedError:  "not found",
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", uuid.New())
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				mockService.On("ShareListWithTribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.ErrNotFound)
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{"expires_at": null}`)))
			},
		},
		{
			name:           "forbidden",
			listID:         uuid.New().String(),
			tribeID:        uuid.New().String(),
			requestBody:    []byte(`{"expires_at": null}`),
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", uuid.New())
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				mockService.On("ShareListWithTribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(models.ErrForbidden)
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{"expires_at": null}`)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.listID)
			rctx.URLParams.Add("tribeId", tt.tribeID)

			w := httptest.NewRecorder()
			r := tt.setupRequest()
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupAuth != nil {
				tt.setupAuth(r)
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			handler.ShareListWithTribe(w, r)
			//fmt.Printf("DEBUG: Response for %s: code=%d, body=%s\n", tt.name, w.Code, w.Body.String())

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" && w.Code != http.StatusNoContent {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, tt.expectedError, response.Error.Message)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_UnshareListWithTribe(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		tribeID        string
		expectedStatus int
		expectedError  string
		setupAuth      func(*http.Request)
		setupMocks     func(*MockListService)
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusNoContent,
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", uuid.New())
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				mockService.On("UnshareListWithTribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid list ID",
		},
		{
			name:           "list not found",
			listID:         uuid.New().String(),
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  "not found",
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", uuid.New())
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				mockService.On("UnshareListWithTribe", mock.Anything, mock.Anything, mock.Anything).Return(models.ErrNotFound)
			},
		},
		{
			name:           "forbidden",
			listID:         uuid.New().String(),
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", uuid.New())
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				mockService.On("UnshareListWithTribe", mock.Anything, mock.Anything, mock.Anything).Return(models.ErrForbidden)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.listID)
			rctx.URLParams.Add("tribeId", tt.tribeID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("DELETE", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupAuth != nil {
				tt.setupAuth(r)
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			handler.UnshareListWithTribe(w, r)
			//fmt.Printf("DEBUG: Response for %s: code=%d, body=%s\n", tt.name, w.Code, w.Body.String())

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" && w.Code != http.StatusNoContent {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, tt.expectedError, response.Error.Message)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_UpdateList(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		requestBody    []byte
		setupMocks     func(*MockListService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			requestBody:    []byte(`{"name":"Updated List","type":"general","visibility":"private","default_weight":1.5}`),
			expectedStatus: http.StatusOK,
			setupMocks: func(mockService *MockListService) {
				mockService.On("UpdateList", mock.AnythingOfType("*models.List")).Return(nil)
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			requestBody:    []byte(`{"name":"Updated List","type":"general","visibility":"private","default_weight":1.5}`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid list ID",
		},
		{
			name:           "invalid request body",
			listID:         uuid.New().String(),
			requestBody:    []byte(`invalid json`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "list not found",
			listID:         uuid.New().String(),
			requestBody:    []byte(`{"name":"Updated List","type":"general","visibility":"private","default_weight":1.5}`),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService) {
				mockService.On("UpdateList", mock.AnythingOfType("*models.List")).Return(models.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("listID", tt.listID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("PUT", "/", bytes.NewReader(tt.requestBody))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			handler.UpdateList(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_DeleteList(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		setupMocks     func(*MockListService, uuid.UUID)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			expectedStatus: http.StatusNoContent,
			setupMocks: func(mockService *MockListService, listID uuid.UUID) {
				mockService.On("DeleteList", listID).Return(nil)
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid list ID",
		},
		{
			name:           "list not found",
			listID:         uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, listID uuid.UUID) {
				mockService.On("DeleteList", listID).Return(models.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("listID", tt.listID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("DELETE", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				listID, _ := uuid.Parse(tt.listID)
				tt.setupMocks(mockService, listID)
			}

			handler.DeleteList(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_AddListItem(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		requestBody    []byte
		setupMocks     func(*MockListService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			requestBody:    []byte(`{"name":"New Item","description":"Description","weight":1.0,"available":true}`),
			expectedStatus: http.StatusCreated,
			setupMocks: func(mockService *MockListService) {
				mockService.On("AddListItem", mock.AnythingOfType("*models.ListItem")).Return(nil)
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			requestBody:    []byte(`{"name":"New Item","description":"Description","weight":1.0,"available":true}`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid list ID",
		},
		{
			name:           "invalid request body",
			listID:         uuid.New().String(),
			requestBody:    []byte(`invalid json`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "list not found",
			listID:         uuid.New().String(),
			requestBody:    []byte(`{"name":"New Item","description":"Description","weight":1.0,"available":true}`),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService) {
				mockService.On("AddListItem", mock.AnythingOfType("*models.ListItem")).Return(models.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("listID", tt.listID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/", bytes.NewReader(tt.requestBody))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			handler.AddListItem(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_GetListItems(t *testing.T) {
	now := time.Now()
	listID := uuid.New()

	tests := []struct {
		name           string
		listID         string
		setupMocks     func(*MockListService, uuid.UUID)
		expectedStatus int
		expectedError  string
		expectedItems  []*models.ListItem
	}{
		{
			name:           "success",
			listID:         listID.String(),
			expectedStatus: http.StatusOK,
			setupMocks: func(mockService *MockListService, listID uuid.UUID) {
				items := []*models.ListItem{
					{
						ID:          uuid.New(),
						ListID:      listID,
						Name:        "Item 1",
						Description: "Description 1",
						Weight:      1.0,
						Available:   true,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
					{
						ID:          uuid.New(),
						ListID:      listID,
						Name:        "Item 2",
						Description: "Description 2",
						Weight:      2.0,
						Available:   true,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				}
				mockService.On("GetListItems", listID).Return(items, nil)
			},
			expectedItems: []*models.ListItem{
				{
					ListID:      listID,
					Name:        "Item 1",
					Description: "Description 1",
					Weight:      1.0,
					Available:   true,
				},
				{
					ListID:      listID,
					Name:        "Item 2",
					Description: "Description 2",
					Weight:      2.0,
					Available:   true,
				},
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid list ID",
		},
		{
			name:           "list not found",
			listID:         uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, listID uuid.UUID) {
				mockService.On("GetListItems", listID).Return([]*models.ListItem{}, models.ErrNotFound)
			},
		},
		{
			name:           "empty list",
			listID:         uuid.New().String(),
			expectedStatus: http.StatusOK,
			setupMocks: func(mockService *MockListService, listID uuid.UUID) {
				mockService.On("GetListItems", listID).Return([]*models.ListItem{}, nil)
			},
			expectedItems: []*models.ListItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("listID", tt.listID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				listID, _ := uuid.Parse(tt.listID)
				tt.setupMocks(mockService, listID)
			}

			handler.GetListItems(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			} else if tt.expectedItems != nil {
				var items []*models.ListItem
				err := json.NewDecoder(w.Body).Decode(&items)
				require.NoError(t, err)
				assert.Len(t, items, len(tt.expectedItems))

				// We only check the name, description, weight, and available fields
				// as the IDs will be different in the response
				for i, expectedItem := range tt.expectedItems {
					assert.Equal(t, expectedItem.Name, items[i].Name)
					assert.Equal(t, expectedItem.Description, items[i].Description)
					assert.Equal(t, expectedItem.Weight, items[i].Weight)
					assert.Equal(t, expectedItem.Available, items[i].Available)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_UpdateListItem(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		itemID         string
		requestBody    []byte
		setupMocks     func(*MockListService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			itemID:         uuid.New().String(),
			requestBody:    []byte(`{"name":"Updated Item","description":"New Description","weight":2.0,"available":false}`),
			expectedStatus: http.StatusOK,
			setupMocks: func(mockService *MockListService) {
				mockService.On("UpdateListItem", mock.AnythingOfType("*models.ListItem")).Return(nil)
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			itemID:         uuid.New().String(),
			requestBody:    []byte(`{"name":"Updated Item","description":"New Description","weight":2.0,"available":false}`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid list ID",
		},
		{
			name:           "invalid item ID",
			listID:         uuid.New().String(),
			itemID:         "invalid-id",
			requestBody:    []byte(`{"name":"Updated Item","description":"New Description","weight":2.0,"available":false}`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid item ID",
		},
		{
			name:           "invalid request body",
			listID:         uuid.New().String(),
			itemID:         uuid.New().String(),
			requestBody:    []byte(`invalid json`),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "item not found",
			listID:         uuid.New().String(),
			itemID:         uuid.New().String(),
			requestBody:    []byte(`{"name":"Updated Item","description":"New Description","weight":2.0,"available":false}`),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService) {
				mockService.On("UpdateListItem", mock.AnythingOfType("*models.ListItem")).Return(models.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("listID", tt.listID)
			rctx.URLParams.Add("itemID", tt.itemID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("PUT", "/", bytes.NewReader(tt.requestBody))
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			handler.UpdateListItem(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_RemoveListItem(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		itemID         string
		setupMocks     func(*MockListService, uuid.UUID, uuid.UUID)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			itemID:         uuid.New().String(),
			expectedStatus: http.StatusNoContent,
			setupMocks: func(mockService *MockListService, listID, itemID uuid.UUID) {
				mockService.On("RemoveListItem", listID, itemID).Return(nil)
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			itemID:         uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid list ID",
		},
		{
			name:           "invalid item ID",
			listID:         uuid.New().String(),
			itemID:         "invalid-id",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid item ID",
		},
		{
			name:           "item not found",
			listID:         uuid.New().String(),
			itemID:         uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, listID, itemID uuid.UUID) {
				mockService.On("RemoveListItem", listID, itemID).Return(models.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("listID", tt.listID)
			rctx.URLParams.Add("itemID", tt.itemID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("DELETE", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				listID, _ := uuid.Parse(tt.listID)
				itemID, _ := uuid.Parse(tt.itemID)
				tt.setupMocks(mockService, listID, itemID)
			}

			handler.RemoveListItem(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListHandler_GetSharedLists(t *testing.T) {
	tests := []struct {
		name           string
		tribeID        string
		setupMocks     func(*MockListService, uuid.UUID)
		expectedStatus int
		expectedLists  []*models.List
		expectedError  string
	}{
		{
			name:           "success",
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusOK,
			setupMocks: func(mockService *MockListService, tribeID uuid.UUID) {
				lists := []*models.List{
					{
						ID:            uuid.New(),
						Type:          models.ListTypeLocation,
						Name:          "Shared List 1",
						Description:   "Shared list description 1",
						DefaultWeight: 1.0,
						Visibility:    models.VisibilityShared,
					},
					{
						ID:            uuid.New(),
						Type:          models.ListTypeActivity,
						Name:          "Shared List 2",
						Description:   "Shared list description 2",
						DefaultWeight: 1.5,
						Visibility:    models.VisibilityShared,
					},
				}
				mockService.On("GetSharedLists", tribeID).Return(lists, nil)
			},
			expectedLists: []*models.List{
				{
					Name:        "Shared List 1",
					Description: "Shared list description 1",
					Type:        models.ListTypeLocation,
					Visibility:  models.VisibilityShared,
				},
				{
					Name:        "Shared List 2",
					Description: "Shared list description 2",
					Type:        models.ListTypeActivity,
					Visibility:  models.VisibilityShared,
				},
			},
		},
		{
			name:           "empty list",
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusOK,
			setupMocks: func(mockService *MockListService, tribeID uuid.UUID) {
				mockService.On("GetSharedLists", tribeID).Return([]*models.List{}, nil)
			},
			expectedLists: []*models.List{},
		},
		{
			name:           "invalid tribe ID",
			tribeID:        "invalid-id",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid tribe ID",
		},
		{
			name:           "tribe not found",
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, tribeID uuid.UUID) {
				mockService.On("GetSharedLists", tribeID).Return([]*models.List(nil), models.ErrNotFound)
			},
			expectedError: "not found",
		},
		{
			name:           "service error",
			tribeID:        uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, tribeID uuid.UUID) {
				mockService.On("GetSharedLists", tribeID).Return([]*models.List(nil), fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("tribeID", tt.tribeID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				tribeID, _ := uuid.Parse(tt.tribeID)
				tt.setupMocks(mockService, tribeID)
			}

			handler.GetSharedLists(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response struct {
					Success bool `json:"success"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error.Message, tt.expectedError)
			} else if tt.expectedLists != nil {
				var lists []*models.List
				err := json.NewDecoder(w.Body).Decode(&lists)
				require.NoError(t, err)
				assert.Len(t, lists, len(tt.expectedLists))

				// Check only specific fields (we don't care about UUIDs here)
				for i, expectedList := range tt.expectedLists {
					assert.Equal(t, expectedList.Name, lists[i].Name)
					assert.Equal(t, expectedList.Description, lists[i].Description)
					assert.Equal(t, expectedList.Type, lists[i].Type)
					assert.Equal(t, expectedList.Visibility, lists[i].Visibility)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
