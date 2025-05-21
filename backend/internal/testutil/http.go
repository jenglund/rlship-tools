package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRequest represents a test HTTP request
type TestRequest struct {
	Method string
	Path   string
	Body   interface{}
	Header map[string]string
}

// TestResponse represents a test HTTP response
type TestResponse struct {
	Code int
	Body interface{}
}

// ExecuteRequest executes a test HTTP request and returns the response
func ExecuteRequest(t *testing.T, router *gin.Engine, req TestRequest) *httptest.ResponseRecorder {
	t.Helper()

	var body io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			t.Fatalf("Error marshaling request body: %v", err)
		}
		body = bytes.NewBuffer(jsonBody)
	}

	r := httptest.NewRequest(req.Method, req.Path, body)
	r.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range req.Header {
		r.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	return w
}

// CheckResponse checks if the response matches the expected response
func CheckResponse(t *testing.T, w *httptest.ResponseRecorder, expected TestResponse) {
	t.Helper()

	if w.Code != expected.Code {
		t.Errorf("Expected status code %d, got %d", expected.Code, w.Code)
	}

	if expected.Body != nil {
		var actualBody interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &actualBody); err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		expectedJSON, err := json.Marshal(expected.Body)
		if err != nil {
			t.Fatalf("Error marshaling expected body: %v", err)
		}

		actualJSON, err := json.Marshal(actualBody)
		if err != nil {
			t.Fatalf("Error marshaling actual body: %v", err)
		}

		if string(expectedJSON) != string(actualJSON) {
			t.Errorf("Expected body %s, got %s", expectedJSON, actualJSON)
		}
	}
}

// MockAuthMiddleware creates a mock authentication middleware for testing
func MockAuthMiddleware(userID string, firebaseUID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("firebase_uid", firebaseUID)
		c.Next()
	}
}
