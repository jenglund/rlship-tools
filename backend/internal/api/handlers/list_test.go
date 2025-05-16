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

func (m *MockListService) GetListConflicts(listID uuid.UUID) ([]*models.ListConflict, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListConflict), args.Error(1)
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

		mockService.On("ListLists", 0, 20).Return(lists, nil)

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
			"cooldown_days": 7,
			"max_items":     2,
		}

		items := []*models.ListItem{
			{
				ID:     uuid.New(),
				Name:   "Item 1",
				Weight: 1.5,
			},
			{
				ID:     uuid.New(),
				Name:   "Item 2",
				Weight: 1.0,
			},
		}

		request := struct {
			ListIDs []uuid.UUID            `json:"list_ids"`
			Filters map[string]interface{} `json:"filters"`
		}{
			ListIDs: listIDs,
			Filters: filters,
		}

		mockService.On("GenerateMenu", listIDs, filters).Return(items, nil)

		body, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/lists/menu", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []*models.ListItem
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)

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
		conflicts := []*models.ListConflict{
			{
				ID:           uuid.New(),
				ListID:       listID,
				ConflictType: "modified",
			},
		}

		mockService.On("GetListConflicts", listID).Return(conflicts, nil)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/%s/conflicts", listID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []*models.ListConflict
		err := json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 1)

		mockService.AssertExpectations(t)
	})

	t.Run("ResolveListConflict", func(t *testing.T) {
		listID := uuid.New()
		conflictID := uuid.New()
		mockService.On("ResolveListConflict", listID, conflictID, "").Return(nil)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/conflicts/%s/resolve", listID, conflictID), nil)
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
			{ListID: listID, OwnerID: userID, OwnerType: "user"},
			{ListID: listID, OwnerID: tribeID, OwnerType: "tribe"},
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

	t.Run("List Sharing", func(t *testing.T) {
		listID := uuid.New()
		userID := uuid.New()
		tribeID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)

		// Test ShareList
		shareReq := struct {
			TribeID   uuid.UUID  `json:"tribe_id"`
			UserID    uuid.UUID  `json:"user_id"`
			ExpiresAt *time.Time `json:"expires_at,omitempty"`
		}{
			TribeID:   tribeID,
			UserID:    userID,
			ExpiresAt: &expiresAt,
		}

		mockService.On("ShareListWithTribe", listID, tribeID, userID, &expiresAt).Return(nil)

		body, err := json.Marshal(shareReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/share", listID), bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)

		// Test UnshareList
		// Create a context with user ID
		ctx := context.WithValue(context.Background(), "user_id", userID)
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/lists/%s/share/%s", listID, tribeID), nil)
		req = req.WithContext(ctx)
		rec = httptest.NewRecorder()

		mockService.On("UnshareListWithTribe", listID, tribeID, userID).Return(nil)

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)

		// Test GetSharedLists
		lists := []*models.List{
			{ID: listID, Name: "Test List"},
		}
		mockService.On("GetSharedLists", tribeID).Return(lists, nil)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/shared/%s", tribeID), nil)
		rec = httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response []*models.List
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 1)
		mockService.AssertExpectations(t)
	})
}
