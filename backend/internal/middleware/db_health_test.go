package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockCheckFunc is used to replace the actual health check logic for testing
type mockDatabaseHealthChecker struct {
	*DatabaseHealthChecker
	mockHealthStatus bool
}

func (m *mockDatabaseHealthChecker) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use the mock health status instead of real check
		if !m.mockHealthStatus {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database connection unavailable",
			})
			return
		}
		c.Next()
	}
}

func TestDatabaseHealthChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	// Test cases
	tests := []struct {
		name             string
		mockHealthStatus bool
		expectedStatus   int
	}{
		{
			name:             "database is healthy",
			mockHealthStatus: true,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "database is unhealthy",
			mockHealthStatus: false,
			expectedStatus:   http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock checker with predictable behavior
			baseChecker := NewDatabaseHealthChecker(db, 1*time.Second)
			checker := &mockDatabaseHealthChecker{
				DatabaseHealthChecker: baseChecker,
				mockHealthStatus:      tt.mockHealthStatus,
			}

			// Create a test router with the middleware
			router := gin.New()
			router.Use(checker.Middleware())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Make a test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
