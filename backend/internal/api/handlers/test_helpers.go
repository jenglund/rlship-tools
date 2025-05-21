package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/middleware"
)

// setupAuthenticatedRequest adds user authentication context to a request
// It returns a new request with the authentication context added
func setupAuthenticatedRequest(r *http.Request) *http.Request {
	userID := uuid.New()
	return r.WithContext(context.WithValue(r.Context(), middleware.ContextUserIDKey, userID))
}

// setupTestRequestWithParams creates a new test request with route parameters and authentication
// It returns both the response recorder and the authenticated request
func setupTestRequestWithParams(method, path string, params map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)

	// Add route parameters
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}

	// Create a new context with the route context
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)

	// Add user authentication
	userID := uuid.New()
	ctx = context.WithValue(ctx, middleware.ContextUserIDKey, userID)

	// Update the request with the new context
	r = r.WithContext(ctx)

	return w, r
}

// GetTestUserID returns a fixed test user ID that can be used in tests
// where the specific ID needs to be known in advance
func GetTestUserID() uuid.UUID {
	// This is a fixed UUID that can be used for testing
	id, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	return id
}
