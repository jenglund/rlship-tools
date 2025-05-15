package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())

	// Add a test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			checkHeaders:   true,
		},
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				headers := w.Header()
				assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", headers.Get("Access-Control-Allow-Credentials"))
				assert.Contains(t, headers.Get("Access-Control-Allow-Headers"), "Authorization")
				assert.Contains(t, headers.Get("Access-Control-Allow-Methods"), "POST")
				assert.Contains(t, headers.Get("Access-Control-Allow-Methods"), "GET")
				assert.Contains(t, headers.Get("Access-Control-Allow-Methods"), "PUT")
				assert.Contains(t, headers.Get("Access-Control-Allow-Methods"), "DELETE")
			}
		})
	}
}
