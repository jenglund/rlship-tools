package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockListService is a mock implementation of the list service
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

func (m *MockListService) ListLists(offset, limit int) ([]*models.List, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*models.List), args.Error(1)
}

func (m *MockListService) AddListItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListService) UpdateListItem(item *models.ListItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockListService) RemoveListItem(listID, itemID uuid.UUID) error {
	args := m.Called(listID, itemID)
	return args.Error(0)
}

func (m *MockListService) GetListItems(listID uuid.UUID) ([]*models.ListItem, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListService) GenerateMenu(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
	args := m.Called(listIDs, filters)
	return args.Get(0).([]*models.ListItem), args.Error(1)
}

func (m *MockListService) SyncList(listID uuid.UUID) error {
	args := m.Called(listID)
	return args.Error(0)
}

func (m *MockListService) GetListConflicts(listID uuid.UUID) ([]*models.ListConflict, error) {
	args := m.Called(listID)
	return args.Get(0).([]*models.ListConflict), args.Error(1)
}

func (m *MockListService) ResolveListConflict(conflictID uuid.UUID) error {
	args := m.Called(conflictID)
	return args.Error(0)
}

func (m *MockListService) CreateListConflict(conflict *models.ListConflict) error {
	args := m.Called(conflict)
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
		mockService.On("ResolveListConflict", conflictID).Return(nil)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/conflicts/%s/resolve", listID, conflictID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)
	})
}
