package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/middleware"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockListService provides a mock for the ListService interface
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) SyncList(listID uuid.UUID) error {
	args := m.Called(listID)
	return args.Error(0)
}

func (m *MockListService) GetListConflicts(listID uuid.UUID) ([]*models.SyncConflict, error) {
	args := m.Called(listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListOwner), args.Error(1)
}

func (m *MockListService) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	args := m.Called(tribeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ListShare), args.Error(1)
}

func (m *MockListService) CreateListConflict(conflict *models.SyncConflict) error {
	args := m.Called(conflict)
	return args.Error(0)
}

// CleanupExpiredShares removes expired shares
func (m *MockListService) CleanupExpiredShares() error {
	args := m.Called()
	return args.Error(0)
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

		var response struct {
			Success bool        `json:"success"`
			Data    models.List `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, list.ID, response.Data.ID)
		assert.Equal(t, list.Name, response.Data.Name)

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

		var response struct {
			Success bool           `json:"success"`
			Data    []*models.List `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Len(t, response.Data, 2)

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

		var response struct {
			Success bool           `json:"success"`
			Data    []*models.List `json:"data"`
		}
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, lists[0].ID, response.Data[0].ID)
		assert.Equal(t, lists[1].ID, response.Data[1].ID)

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

		var response struct {
			Success bool                   `json:"success"`
			Data    []*models.SyncConflict `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Len(t, response.Data, 1)

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
		localMockService := new(MockListService)
		localHandler := NewListHandler(localMockService)

		localRouter := chi.NewRouter()
		localRouter.Post("/lists/{listID}/share", localHandler.ShareList)

		listID := uuid.New()
		tribeID := uuid.New()
		userID := uuid.New()

		// Setup service expectations
		localMockService.On("GetListOwners", listID).Return([]*models.ListOwner{
			{OwnerID: userID, OwnerType: "user"},
		}, nil)
		localMockService.On("ShareListWithTribe", listID, tribeID, userID, mock.Anything).Return(nil)

		// Create the request body
		body, err := json.Marshal(map[string]interface{}{
			"tribe_id":   tribeID,
			"expires_at": nil,
		})
		require.NoError(t, err)

		// Create the request with user ID in context
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/share", listID), bytes.NewReader(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("listID", listID.String())
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		ctx = context.WithValue(ctx, middleware.ContextUserIDKey, userID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		// Send the request
		localRouter.ServeHTTP(rec, req)

		// Verify the response
		assert.Equal(t, http.StatusNoContent, rec.Code)
		localMockService.AssertExpectations(t)
	})

	t.Run("GetListShares", func(t *testing.T) {
		localMockService := new(MockListService)
		localHandler := NewListHandler(localMockService)

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
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, validUserID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func() {
					localMockService.On("GetListOwners", validListID).Return([]*models.ListOwner{
						{OwnerID: validUserID, OwnerType: "user"},
					}, nil)
					localMockService.On("GetListShares", validListID).Return(validShares, nil)
				},
				expectedStatus: http.StatusOK,
			},
			{
				name:   "invalid list ID",
				listID: "invalid-uuid",
				setupAuth: func(r *http.Request) {
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, validUserID)
					*r = *r.WithContext(ctx)
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid list ID",
			},
			{
				name:   "unauthorized",
				listID: validListID.String(),
				setupAuth: func(r *http.Request) {
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, uuid.New())
					*r = *r.WithContext(ctx)
				},
				setupMocks: func() {
					localMockService.On("GetListOwners", validListID).Return([]*models.ListOwner{
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

				localHandler.GetListShares(w, r)

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

				localMockService.AssertExpectations(t)
			})
		}
	})

	t.Run("UnshareListWithTribe", func(t *testing.T) {
		localMockService := new(MockListService)
		localHandler := NewListHandler(localMockService)
		localRouter := chi.NewRouter()
		localHandler.RegisterRoutes(localRouter)

		tests := []struct {
			name           string
			listID         string
			tribeID        string
			setupAuth      func(*http.Request)
			setupMocks     func(*MockListService, uuid.UUID, uuid.UUID, uuid.UUID)
			expectedStatus int
			expectedError  string
		}{
			{
				name:           "success",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				expectedStatus: http.StatusNoContent,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID) {
					mockService.On("UnshareListWithTribe", listID, tribeID, userID).Return(nil)
				},
			},
			{
				name:           "invalid list ID",
				listID:         "invalid-id",
				tribeID:        uuid.New().String(),
				expectedStatus: http.StatusBadRequest,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				expectedError: "invalid list ID",
			},
			{
				name:           "invalid tribe ID",
				listID:         uuid.New().String(),
				tribeID:        "invalid-id",
				expectedStatus: http.StatusBadRequest,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				expectedError: "invalid tribe ID",
			},
			{
				name:           "missing user ID in context",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				expectedStatus: http.StatusUnauthorized,
				setupAuth:      func(r *http.Request) {}, // No user ID in context
				expectedError:  "user not authenticated",
			},
			{
				name:           "invalid user ID type in context",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				expectedStatus: http.StatusUnauthorized,
				setupAuth: func(r *http.Request) {
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, "not-a-uuid")
					*r = *r.WithContext(ctx)
				},
				expectedError: "invalid user authentication",
			},
			{
				name:           "service returns not found error",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				expectedStatus: http.StatusNotFound,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID) {
					mockService.On("UnshareListWithTribe", listID, tribeID, userID).Return(models.ErrNotFound)
				},
				expectedError: "not found",
			},
			{
				name:           "service returns forbidden error",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				expectedStatus: http.StatusForbidden,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID) {
					mockService.On("UnshareListWithTribe", listID, tribeID, userID).Return(models.ErrForbidden)
				},
				expectedError: "forbidden",
			},
			{
				name:           "service returns internal server error",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				expectedStatus: http.StatusInternalServerError,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID) {
					mockService.On("UnshareListWithTribe", listID, tribeID, userID).Return(fmt.Errorf("database error"))
				},
				expectedError: "database error",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Reset mock between tests
				localMockService.ExpectedCalls = nil
				localMockService.Calls = nil

				// Setup direct request
				req := httptest.NewRequest(http.MethodDelete, "/", nil)
				rec := httptest.NewRecorder()

				// Setup Chi route context with URL params
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tc.listID)
				rctx.URLParams.Add("tribeID", tc.tribeID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				// Setup authentication if provided
				if tc.setupAuth != nil {
					tc.setupAuth(req)
				}

				// Setup mocks if provided
				if tc.setupMocks != nil {
					listID, _ := uuid.Parse(tc.listID)
					tribeID, _ := uuid.Parse(tc.tribeID)
					userID := uuid.UUID{}
					if userIDVal := req.Context().Value(middleware.ContextUserIDKey); userIDVal != nil {
						if id, ok := userIDVal.(uuid.UUID); ok {
							userID = id
						}
					}

					tc.setupMocks(localMockService, listID, tribeID, userID)
				}

				// Call handler directly instead of using router
				localHandler.UnshareListWithTribe(rec, req)

				// Check results
				assert.Equal(t, tc.expectedStatus, rec.Code)

				// Check error message in the response if expected
				if tc.expectedError != "" && rec.Code != http.StatusNoContent {
					var errorResponse struct {
						Success bool `json:"success"`
						Error   struct {
							Message string `json:"message"`
						} `json:"error"`
					}
					err := json.NewDecoder(rec.Body).Decode(&errorResponse)
					require.NoError(t, err)
					assert.Contains(t, errorResponse.Error.Message, tc.expectedError)
				}

				localMockService.AssertExpectations(t)
			})
		}
	})

	t.Run("UpdateList", func(t *testing.T) {
		listID := uuid.New()
		list := &models.List{
			ID:            listID,
			Type:          models.ListTypeLocation,
			Name:          "Test List",
			Description:   "Test Description",
			DefaultWeight: 1.0,
		}

		mockService.On("UpdateList", mock.AnythingOfType("*models.List")).Return(nil)

		body, err := json.Marshal(list)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/lists/%s", listID), bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ShareListWithTribe", func(t *testing.T) {
		mockService := new(MockListService)
		handler := NewListHandler(mockService)

		tests := []struct {
			name           string
			listID         string
			tribeID        string
			requestBody    string
			setupAuth      func(*http.Request)
			setupMocks     func(*MockListService, uuid.UUID, uuid.UUID, uuid.UUID, *time.Time)
			expectedStatus int
			expectedError  string
		}{
			{
				name:        "success with no expiration",
				listID:      uuid.New().String(),
				tribeID:     uuid.New().String(),
				requestBody: `{}`,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID, expiresAt *time.Time) {
					mockService.On("ShareListWithTribe", listID, tribeID, userID, expiresAt).Return(nil)
				},
				expectedStatus: http.StatusNoContent,
			},
			{
				name:        "success with expiration",
				listID:      uuid.New().String(),
				tribeID:     uuid.New().String(),
				requestBody: `{"expires_at": "2030-01-01T00:00:00Z"}`,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID, expiresAt *time.Time) {
					mockService.On("ShareListWithTribe", listID, tribeID, userID, expiresAt).Return(nil)
				},
				expectedStatus: http.StatusNoContent,
			},
			{
				name:           "invalid list ID",
				listID:         "invalid-id",
				tribeID:        uuid.New().String(),
				requestBody:    `{}`,
				expectedStatus: http.StatusBadRequest,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				expectedError: "invalid list ID",
			},
			{
				name:           "invalid tribe ID",
				listID:         uuid.New().String(),
				tribeID:        "invalid-id",
				requestBody:    `{}`,
				expectedStatus: http.StatusBadRequest,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				expectedError: "invalid tribe ID",
			},
			{
				name:           "invalid request body",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				requestBody:    `{invalid json`,
				expectedStatus: http.StatusBadRequest,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				expectedError: "invalid request body",
			},
			{
				name:           "missing user ID in context",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				requestBody:    `{}`,
				expectedStatus: http.StatusUnauthorized,
				setupAuth:      func(r *http.Request) {}, // No user ID in context
				expectedError:  "user not authenticated",
			},
			{
				name:           "invalid user ID type in context",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				requestBody:    `{}`,
				expectedStatus: http.StatusUnauthorized,
				setupAuth: func(r *http.Request) {
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, "not-a-uuid")
					*r = *r.WithContext(ctx)
				},
				expectedError: "invalid user authentication",
			},
			{
				name:           "service returns not found error",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				requestBody:    `{}`,
				expectedStatus: http.StatusNotFound,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID, expiresAt *time.Time) {
					mockService.On("ShareListWithTribe", listID, tribeID, userID, expiresAt).Return(models.ErrNotFound)
				},
				expectedError: "not found",
			},
			{
				name:           "service returns forbidden error",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				requestBody:    `{}`,
				expectedStatus: http.StatusForbidden,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID, expiresAt *time.Time) {
					mockService.On("ShareListWithTribe", listID, tribeID, userID, expiresAt).Return(models.ErrForbidden)
				},
				expectedError: "forbidden",
			},
			{
				name:           "service returns internal server error",
				listID:         uuid.New().String(),
				tribeID:        uuid.New().String(),
				requestBody:    `{}`,
				expectedStatus: http.StatusInternalServerError,
				setupAuth: func(r *http.Request) {
					userID := uuid.New()
					ctx := context.WithValue(r.Context(), middleware.ContextUserIDKey, userID)
					*r = *r.WithContext(ctx)
				},
				setupMocks: func(mockService *MockListService, listID, tribeID, userID uuid.UUID, expiresAt *time.Time) {
					mockService.On("ShareListWithTribe", listID, tribeID, userID, expiresAt).Return(fmt.Errorf("database error"))
				},
				expectedError: "database error",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				// Reset mock between tests
				mockService.ExpectedCalls = nil
				mockService.Calls = nil

				// Import strings package at the top of the file
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tc.requestBody)))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()

				// Setup Chi route context with URL params
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tc.listID)
				rctx.URLParams.Add("tribeID", tc.tribeID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				// Setup authentication if provided
				if tc.setupAuth != nil {
					tc.setupAuth(req)
				}

				// Setup mocks if provided
				if tc.setupMocks != nil {
					listID, _ := uuid.Parse(tc.listID)
					tribeID, _ := uuid.Parse(tc.tribeID)
					userID := uuid.UUID{}
					var expiresAt *time.Time

					if userIDVal := req.Context().Value(middleware.ContextUserIDKey); userIDVal != nil {
						if id, ok := userIDVal.(uuid.UUID); ok {
							userID = id
						}
					}

					// Parse request body to extract expiresAt if valid JSON
					if tc.requestBody != "" && tc.requestBody != "{invalid json" {
						var body struct {
							ExpiresAt *time.Time `json:"expires_at"`
						}
						if err := json.Unmarshal([]byte(tc.requestBody), &body); err == nil {
							expiresAt = body.ExpiresAt
						}
					}

					tc.setupMocks(mockService, listID, tribeID, userID, expiresAt)
				}

				// Call handler directly
				handler.ShareListWithTribe(rec, req)

				// Check results
				assert.Equal(t, tc.expectedStatus, rec.Code)

				// Check error message in the response if expected
				if tc.expectedError != "" {
					var errorResponse struct {
						Success bool `json:"success"`
						Error   struct {
							Message string `json:"message"`
						} `json:"error"`
					}
					err := json.NewDecoder(rec.Body).Decode(&errorResponse)
					require.NoError(t, err)
					assert.Contains(t, errorResponse.Error.Message, tc.expectedError)
				}

				mockService.AssertExpectations(t)
			})
		}
	})
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
				// Debug log the response
				respBody, _ := io.ReadAll(w.Body)
				t.Logf("Response body: %s", string(respBody))
				// Reset the response body for decoding
				w.Body = httptest.NewRecorder().Body
				w.Body.Write(respBody)

				var response struct {
					Success bool `json:"success"`
					Data    struct {
						Success bool               `json:"success"`
						Data    []*models.ListItem `json:"data"`
					} `json:"data"`
				}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.True(t, response.Data.Success)
				assert.Len(t, response.Data.Data, len(tt.expectedItems))

				// We only check the name, description, weight, and available fields
				// as the IDs will be different in the response
				for i, expectedItem := range tt.expectedItems {
					assert.Equal(t, expectedItem.Name, response.Data.Data[i].Name)
					assert.Equal(t, expectedItem.Description, response.Data.Data[i].Description)
					assert.Equal(t, expectedItem.Weight, response.Data.Data[i].Weight)
					assert.Equal(t, expectedItem.Available, response.Data.Data[i].Available)
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

func TestListHandler_RemoveListOwner(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		ownerID        string
		setupMocks     func(*MockListService, uuid.UUID, uuid.UUID)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "success",
			listID:         uuid.New().String(),
			ownerID:        uuid.New().String(),
			expectedStatus: http.StatusNoContent,
			setupMocks: func(mockService *MockListService, listID, ownerID uuid.UUID) {
				mockService.On("RemoveListOwner", listID, ownerID).Return(nil)
			},
		},
		{
			name:           "invalid list ID",
			listID:         "invalid-id",
			ownerID:        uuid.New().String(),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid list ID",
		},
		{
			name:           "invalid owner ID",
			listID:         uuid.New().String(),
			ownerID:        "invalid-id",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid owner ID",
		},
		{
			name:           "list not found",
			listID:         uuid.New().String(),
			ownerID:        uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, listID, ownerID uuid.UUID) {
				mockService.On("RemoveListOwner", listID, ownerID).Return(models.ErrNotFound)
			},
		},
		{
			name:           "permission denied",
			listID:         uuid.New().String(),
			ownerID:        uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, listID, ownerID uuid.UUID) {
				mockService.On("RemoveListOwner", listID, ownerID).Return(models.ErrForbidden)
			},
		},
		{
			name:           "last owner error",
			listID:         uuid.New().String(),
			ownerID:        uuid.New().String(),
			expectedStatus: http.StatusInternalServerError,
			setupMocks: func(mockService *MockListService, listID, ownerID uuid.UUID) {
				mockService.On("RemoveListOwner", listID, ownerID).Return(fmt.Errorf("cannot remove last owner"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockListService)
			handler := NewListHandler(mockService)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("listID", tt.listID)
			rctx.URLParams.Add("ownerID", tt.ownerID)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("DELETE", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			if tt.setupMocks != nil {
				listID, _ := uuid.Parse(tt.listID)
				ownerID, _ := uuid.Parse(tt.ownerID)
				tt.setupMocks(mockService, listID, ownerID)
			}

			handler.RemoveListOwner(w, r)

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
			expectedError: "resource not found",
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

			// Use the helper function to create a request with parameters and authentication
			w, r := setupTestRequestWithParams("GET", "/lists/shared/"+tt.tribeID, map[string]string{
				"tribeID": tt.tribeID,
			})

			if tt.setupMocks != nil && tt.tribeID != "invalid-id" {
				tt.setupMocks(mockService, uuid.MustParse(tt.tribeID))
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
				assert.Equal(t, tt.expectedError, response.Error.Message)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// TestCleanupExpiredShares tests the admin endpoint for cleaning up expired shares
func TestCleanupExpiredShares(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	// Test successful cleanup
	t.Run("Successful Cleanup", func(t *testing.T) {
		mockService.On("CleanupExpiredShares").Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/lists/admin/cleanup-expired-shares", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		mockService.AssertExpectations(t)

		// Debug log the response
		respBody, _ := io.ReadAll(rec.Body)
		t.Logf("Response body: %s", string(respBody))
		// Reset the response body for decoding
		rec.Body = httptest.NewRecorder().Body
		rec.Body.Write(respBody)

		var response struct {
			Success bool `json:"success"`
			Data    struct {
				Message string `json:"message"`
			} `json:"data"`
		}
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Data.Message, "successfully")
	})

	// Test error during cleanup
	t.Run("Cleanup Error", func(t *testing.T) {
		mockService.On("CleanupExpiredShares").Return(errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodPost, "/lists/admin/cleanup-expired-shares", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		mockService.AssertExpectations(t)
		assert.Contains(t, rec.Body.String(), "Failed to clean up expired shares")
	})
}

// TestCreateListValidation tests the validation rules for the CreateList handler
func TestCreateListValidation(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func(*MockListService)
		setupAuth      func(*http.Request)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "empty_name",
			requestBody: map[string]interface{}{
				"name":        "",
				"description": "Test description",
				"type":        "general",
			},
			setupAuth: func(r *http.Request) {
				// Set up auth context
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "List name cannot be empty",
		},
		{
			name: "whitespace_only_name",
			requestBody: map[string]interface{}{
				"name":        "   ",
				"description": "Test description",
				"type":        "general",
			},
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "List name cannot be empty",
		},
		{
			name: "name_too_long",
			requestBody: map[string]interface{}{
				"name":        strings.Repeat("a", 101), // 101 characters
				"description": "Test description",
				"type":        "general",
			},
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "List name is too long",
		},
		{
			name: "description_too_long",
			requestBody: map[string]interface{}{
				"name":        "Test List",
				"description": strings.Repeat("a", 1001), // 1001 characters
				"type":        "general",
			},
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "List description is too long",
		},
		{
			name: "create_list_for_another_user",
			requestBody: map[string]interface{}{
				"name":        "Test List",
				"description": "Test description",
				"type":        "general",
				"owner_id":    uuid.New().String(),
				"owner_type":  "user",
			},
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "You can only create lists for yourself or tribes you belong to",
		},
		{
			name: "duplicate_list",
			requestBody: map[string]interface{}{
				"name":        "Duplicate List",
				"description": "Test description",
				"type":        "general",
			},
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				// Create the list object that would be passed to the service
				list := &models.List{
					Name:        "Duplicate List",
					Description: "Test description",
					Type:        "general",
				}
				// Mock the service to return a duplicate error
				mockService.On("CreateList", mock.MatchedBy(func(l *models.List) bool {
					return l.Name == list.Name
				})).Return(models.ErrDuplicate).Once()

				// Mock the GetUserLists call that happens after a duplicate error
				existingLists := []*models.List{
					{
						ID:          uuid.New(),
						Name:        "Duplicate List",
						Description: "Existing list with the same name",
						Type:        "general",
					},
				}
				mockService.On("GetUserLists", userID).Return(existingLists, nil).Once()
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "A list with the name 'Duplicate List' already exists",
		},
		{
			name: "trim_whitespace_valid_creation",
			requestBody: map[string]interface{}{
				"name":        "  Test List  ",
				"description": "  Test description  ",
				"type":        "general",
			},
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				// The handler should trim whitespace before passing to service
				mockService.On("CreateList", mock.MatchedBy(func(l *models.List) bool {
					return l.Name == "Test List" && l.Description == "Test description"
				})).Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "success_with_defaults",
			requestBody: map[string]interface{}{
				"name": "Valid List",
			},
			setupAuth: func(r *http.Request) {
				ctx := context.WithValue(r.Context(), "user_id", userID)
				*r = *r.WithContext(ctx)
			},
			setupMocks: func(mockService *MockListService) {
				mockService.On("CreateList", mock.MatchedBy(func(l *models.List) bool {
					return l.Name == "Valid List" &&
						l.Type == models.ListTypeGeneral &&
						l.Visibility == models.VisibilityPrivate &&
						l.DefaultWeight == 1.0
				})).Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Setup mocks if provided
			if tc.setupMocks != nil {
				tc.setupMocks(mockService)
			}

			// Create request
			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/lists", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Setup auth if provided
			if tc.setupAuth != nil {
				tc.setupAuth(req)
			}

			rec := httptest.NewRecorder()

			// Call the handler directly
			handler.CreateList(rec, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Check error message if expected
			if tc.expectedError != "" {
				// For duplicate errors, the response format is different
				if tc.name == "duplicate_list" {
					var response map[string]interface{}
					err := json.NewDecoder(rec.Body).Decode(&response)
					require.NoError(t, err)

					errorMsg, ok := response["error"].(string)
					if ok {
						assert.Contains(t, errorMsg, tc.expectedError)
					} else {
						t.Logf("Response body: %v", response)
						assert.Fail(t, "Expected error message not found in duplicate list response")
					}
				} else {
					// Regular error format
					var response struct {
						Success bool `json:"success"`
						Error   struct {
							Message string `json:"message"`
						} `json:"error"`
					}
					err := json.NewDecoder(rec.Body).Decode(&response)
					require.NoError(t, err)
					assert.Contains(t, response.Error.Message, tc.expectedError)
				}
			}

			// Verify all mocks were called
			mockService.AssertExpectations(t)
		})
	}
}
