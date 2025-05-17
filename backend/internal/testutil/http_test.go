package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExecuteRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Test endpoint that echoes back the request body and headers
	router.POST("/test", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// Echo back the request body and custom header
		response := map[string]interface{}{
			"body":    body,
			"headers": make(map[string]string),
		}

		// Get all headers
		headers := response["headers"].(map[string]string)
		for key := range c.Request.Header {
			headers[key] = c.GetHeader(key)
		}

		// Set response headers
		for key, value := range c.Request.Header {
			if len(value) > 0 {
				c.Header(key, value[0])
			}
		}

		c.JSON(http.StatusOK, response)
	})

	tests := []struct {
		name           string
		request        TestRequest
		expectedCode   int
		expectedBody   map[string]interface{}
		expectedHeader string
	}{
		{
			name: "valid request with body and headers",
			request: TestRequest{
				Method: "POST",
				Path:   "/test",
				Body: map[string]interface{}{
					"key": "value",
				},
				Header: map[string]string{
					"X-Custom-Header": "test-value",
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{
				"body": map[string]interface{}{
					"key": "value",
				},
				"headers": map[string]string{
					"Content-Type":    "application/json",
					"X-Custom-Header": "test-value",
				},
			},
			expectedHeader: "test-value",
		},
		{
			name: "request without body",
			request: TestRequest{
				Method: "POST",
				Path:   "/test",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "EOF",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ExecuteRequest(t, router, tt.request)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedBody != nil {
				var actualBody map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &actualBody)
				assert.NoError(t, err)

				// Convert headers to string map for comparison
				if headers, ok := actualBody["headers"].(map[string]interface{}); ok {
					stringHeaders := make(map[string]string)
					for k, v := range headers {
						stringHeaders[k] = v.(string)
					}
					actualBody["headers"] = stringHeaders
				}

				assert.Equal(t, tt.expectedBody, actualBody)
			}

			if tt.request.Header != nil {
				for key, value := range tt.request.Header {
					assert.Equal(t, value, w.Header().Get(key))
				}
			}
		})
	}
}

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name          string
		recorder      *httptest.ResponseRecorder
		expected      TestResponse
		shouldSucceed bool
	}{
		{
			name: "matching response",
			recorder: func() *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				w.Code = http.StatusOK
				w.Body.WriteString(`{"message":"success"}`)
				return w
			}(),
			expected: TestResponse{
				Code: http.StatusOK,
				Body: map[string]interface{}{
					"message": "success",
				},
			},
			shouldSucceed: true,
		},
		{
			name: "mismatched status code",
			recorder: func() *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				w.Code = http.StatusBadRequest
				w.Body.WriteString(`{"message":"success"}`)
				return w
			}(),
			expected: TestResponse{
				Code: http.StatusOK,
				Body: map[string]interface{}{
					"message": "success",
				},
			},
			shouldSucceed: false,
		},
		{
			name: "mismatched body",
			recorder: func() *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				w.Code = http.StatusOK
				w.Body.WriteString(`{"message":"error"}`)
				return w
			}(),
			expected: TestResponse{
				Code: http.StatusOK,
				Body: map[string]interface{}{
					"message": "success",
				},
			},
			shouldSucceed: false,
		},
		{
			name: "nil expected body",
			recorder: func() *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				w.Code = http.StatusOK
				return w
			}(),
			expected: TestResponse{
				Code: http.StatusOK,
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSucceed {
				assert.NotPanics(t, func() {
					CheckResponse(t, tt.recorder, tt.expected)
				})
			} else {
				mockT := &testing.T{}
				CheckResponse(mockT, tt.recorder, tt.expected)
				assert.True(t, mockT.Failed())
			}
		})
	}
}

func TestMockAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	const (
		testUserID      = "test-user-id"
		testFirebaseUID = "test-firebase-uid"
	)

	// Add test endpoint with mock auth
	router.GET("/protected",
		MockAuthMiddleware(testUserID, testFirebaseUID),
		func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			assert.True(t, exists)
			assert.Equal(t, testUserID, userID)

			firebaseUID, exists := c.Get("firebase_uid")
			assert.True(t, exists)
			assert.Equal(t, testFirebaseUID, firebaseUID)

			c.JSON(http.StatusOK, gin.H{"status": "success"})
		},
	)

	req := TestRequest{
		Method: "GET",
		Path:   "/protected",
	}

	w := ExecuteRequest(t, router, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}
