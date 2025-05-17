package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestSyncListHandler tests the SyncList endpoint with a variety of scenarios
func TestSyncListHandler(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	testCases := []struct {
		name           string
		listID         string // Use string to allow testing invalid UUIDs
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Happy Path",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("SyncList", mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "Invalid List ID",
			listID:         "not-a-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid list ID format",
		},
		{
			name:   "List Not Found",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("SyncList", mock.AnythingOfType("uuid.UUID")).Return(models.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "List not found",
		},
		{
			name:   "Sync Disabled",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("SyncList", mock.AnythingOfType("uuid.UUID")).Return(models.ErrSyncDisabled).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Sync is not enabled for this list",
		},
		{
			name:   "External Source Unavailable",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("SyncList", mock.AnythingOfType("uuid.UUID")).Return(models.ErrExternalSourceUnavailable).Once()
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "External sync source is unavailable",
		},
		{
			name:   "External Source Timeout",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("SyncList", mock.AnythingOfType("uuid.UUID")).Return(models.ErrExternalSourceTimeout).Once()
			},
			expectedStatus: http.StatusGatewayTimeout,
			expectedBody:   "External sync source timed out",
		},
		{
			name:   "External Source Error",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("SyncList", mock.AnythingOfType("uuid.UUID")).Return(models.ErrExternalSourceError).Once()
			},
			expectedStatus: http.StatusBadGateway,
			expectedBody:   "External sync source error",
		},
		{
			name:   "Internal Server Error",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("SyncList", mock.AnythingOfType("uuid.UUID")).Return(errors.New("internal error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to sync list",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock expectations
			tc.setupMock()

			// Create and execute request
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/sync", tc.listID), nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// Verify response
			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetListConflictsHandler tests the GetListConflicts endpoint with a variety of scenarios
func TestGetListConflictsHandler(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	now := time.Now()
	itemID := uuid.New()

	// Create test conflicts
	validListID := uuid.New()
	testConflicts := []*models.SyncConflict{
		{
			ID:         uuid.New(),
			ListID:     validListID,
			ItemID:     &itemID,
			Type:       "name_conflict",
			LocalData:  "Local Name",
			RemoteData: "Remote Name",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			ListID:     validListID,
			ItemID:     nil,
			Type:       "list_deletion_conflict",
			LocalData:  "Deleted locally",
			RemoteData: "Still exists remotely",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	testCases := []struct {
		name           string
		listID         string // Use string to allow testing invalid UUIDs
		setupMock      func()
		expectedStatus int
		checkResponse  func(*httptest.ResponseRecorder)
	}{
		{
			name:   "Happy Path - Multiple Conflicts",
			listID: validListID.String(),
			setupMock: func() {
				mockService.On("GetListConflicts", validListID).Return(testConflicts, nil).Once()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				var conflicts []*models.SyncConflict
				err := json.NewDecoder(rec.Body).Decode(&conflicts)
				require.NoError(t, err)
				assert.Len(t, conflicts, 2)
				assert.Equal(t, "name_conflict", conflicts[0].Type)
				assert.Equal(t, "list_deletion_conflict", conflicts[1].Type)
			},
		},
		{
			name:   "Happy Path - No Conflicts",
			listID: validListID.String(),
			setupMock: func() {
				mockService.On("GetListConflicts", validListID).Return([]*models.SyncConflict{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				var conflicts []*models.SyncConflict
				err := json.NewDecoder(rec.Body).Decode(&conflicts)
				require.NoError(t, err)
				assert.Empty(t, conflicts)
			},
		},
		{
			name:           "Invalid List ID",
			listID:         "not-a-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Invalid list ID format")
			},
		},
		{
			name:   "List Not Found",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("GetListConflicts", mock.AnythingOfType("uuid.UUID")).Return(nil, models.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "List not found")
			},
		},
		{
			name:   "Sync Disabled",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("GetListConflicts", mock.AnythingOfType("uuid.UUID")).Return(nil, models.ErrSyncDisabled).Once()
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Sync is not enabled for this list")
			},
		},
		{
			name:   "Internal Server Error",
			listID: uuid.New().String(),
			setupMock: func() {
				mockService.On("GetListConflicts", mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Failed to get list conflicts")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock expectations
			tc.setupMock()

			// Create and execute request
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/lists/%s/conflicts", tc.listID), nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// Verify response
			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.checkResponse(rec)
			mockService.AssertExpectations(t)
		})
	}
}

// TestResolveListConflictHandler tests the ResolveListConflict endpoint with a variety of scenarios
func TestResolveListConflictHandler(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	validListID := uuid.New()
	validConflictID := uuid.New()

	createRequestBody := func(resolution string) []byte {
		body, _ := json.Marshal(map[string]string{
			"resolution": resolution,
		})
		return body
	}

	testCases := []struct {
		name           string
		listID         string
		conflictID     string
		requestBody    []byte
		setupMock      func()
		expectedStatus int
		checkResponse  func(*httptest.ResponseRecorder)
	}{
		{
			name:        "Happy Path - Accept Local",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("accept_local"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "accept_local").Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
			checkResponse:  func(rec *httptest.ResponseRecorder) {},
		},
		{
			name:        "Happy Path - Accept Remote",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("accept_remote"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "accept_remote").Return(nil).Once()
			},
			expectedStatus: http.StatusNoContent,
			checkResponse:  func(rec *httptest.ResponseRecorder) {},
		},
		{
			name:           "Invalid List ID",
			listID:         "not-a-uuid",
			conflictID:     validConflictID.String(),
			requestBody:    createRequestBody("accept_local"),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Invalid list ID format")
			},
		},
		{
			name:           "Invalid Conflict ID",
			listID:         validListID.String(),
			conflictID:     "not-a-uuid",
			requestBody:    createRequestBody("accept_local"),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Invalid conflict ID format")
			},
		},
		{
			name:           "Invalid Request Body",
			listID:         validListID.String(),
			conflictID:     validConflictID.String(),
			requestBody:    []byte("not valid json"),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Invalid request body")
			},
		},
		{
			name:        "Empty Request Body",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: []byte("{}"),
			setupMock: func() {
				// Need to mock empty resolution string call since the json will decode to an empty string
				mockService.On("ResolveListConflict", validListID, validConflictID, "").Return(models.ErrInvalidResolution).Once()
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Invalid resolution")
			},
		},
		{
			name:        "List Not Found",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("accept_local"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "accept_local").Return(models.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "List not found")
			},
		},
		{
			name:        "Conflict Not Found",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("accept_local"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "accept_local").Return(models.ErrConflictNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Conflict not found")
			},
		},
		{
			name:        "Conflict Already Resolved",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("accept_local"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "accept_local").Return(models.ErrConflictAlreadyResolved).Once()
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Conflict already resolved")
			},
		},
		{
			name:        "Invalid Resolution",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("invalid_resolution"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "invalid_resolution").Return(models.ErrInvalidResolution).Once()
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Invalid resolution")
			},
		},
		{
			name:        "Sync Disabled",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("accept_local"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "accept_local").Return(models.ErrSyncDisabled).Once()
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Sync is not enabled for this list")
			},
		},
		{
			name:        "Internal Server Error",
			listID:      validListID.String(),
			conflictID:  validConflictID.String(),
			requestBody: createRequestBody("accept_local"),
			setupMock: func() {
				mockService.On("ResolveListConflict", validListID, validConflictID, "accept_local").Return(errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "Failed to resolve conflict")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock expectations
			tc.setupMock()

			// Create and execute request
			endpoint := fmt.Sprintf("/lists/%s/conflicts/%s/resolve", tc.listID, tc.conflictID)
			req := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// Verify response
			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.checkResponse(rec)
			mockService.AssertExpectations(t)
		})
	}
}

// TestEdgeCasesForSyncHandlers tests edge cases for the sync-related handlers
func TestEdgeCasesForSyncHandlers(t *testing.T) {
	mockService := new(MockListService)
	handler := NewListHandler(mockService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	t.Run("SyncList with missing body", func(t *testing.T) {
		// This should still work as sync doesn't require a body
		listID := uuid.New()
		mockService.On("SyncList", listID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/sync", listID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ResolveConflict with invalid JSON", func(t *testing.T) {
		listID := uuid.New()
		conflictID := uuid.New()

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/conflicts/%s/resolve", listID, conflictID),
			bytes.NewReader([]byte(`{invalid json`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid request body")
	})

	t.Run("ResolveConflict with empty body", func(t *testing.T) {
		listID := uuid.New()
		conflictID := uuid.New()

		// Mock the service call with an empty resolution string
		mockService.On("ResolveListConflict", listID, conflictID, "").Return(models.ErrInvalidResolution).Once()

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/lists/%s/conflicts/%s/resolve", listID, conflictID),
			bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid resolution")
	})

	t.Run("GetListConflicts with bad request", func(t *testing.T) {
		// Using a simulated broken client that sends the wrong HTTP method
		listID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/lists/%s/conflicts", listID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// Chi router should return Method Not Allowed
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}

// MockReadCloser is a helper struct for testing with request bodies
type MockReadCloser struct {
	*bytes.Reader
}

func (m *MockReadCloser) Close() error {
	return nil
}
