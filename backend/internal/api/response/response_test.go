package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestGinSuccess(t *testing.T) {
	c, w := setupTest()

	testData := map[string]string{"key": "value"}
	GinSuccess(c, testData)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)

	// Check the actual data
	dataBytes, err := json.Marshal(response.Data)
	require.NoError(t, err)

	var returnedData map[string]string
	err = json.Unmarshal(dataBytes, &returnedData)
	require.NoError(t, err)
	assert.Equal(t, testData, returnedData)
}

func TestGinCreated(t *testing.T) {
	c, w := setupTest()

	testData := map[string]string{"key": "value"}
	GinCreated(c, testData)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)
}

func TestGinNoContent(t *testing.T) {
	c, w := setupTest()

	GinNoContent(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestGinError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		code       string
		message    string
		wantStatus int
	}{
		{
			name:       "custom error",
			status:     http.StatusTeapot,
			code:       "TEAPOT",
			message:    "I'm a teapot",
			wantStatus: http.StatusTeapot,
		},
		{
			name:       "server error",
			status:     http.StatusInternalServerError,
			code:       "SERVER_ERROR",
			message:    "Server error occurred",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTest()

			GinError(c, tt.status, tt.code, tt.message)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.Nil(t, response.Data)
			assert.NotNil(t, response.Error)
			assert.Equal(t, tt.code, response.Error.Code)
			assert.Equal(t, tt.message, response.Error.Message)
		})
	}
}

func TestGinBadRequest(t *testing.T) {
	c, w := setupTest()

	message := "Invalid input"
	GinBadRequest(c, message)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "BAD_REQUEST", response.Error.Code)
	assert.Equal(t, message, response.Error.Message)
}

func TestGinUnauthorized(t *testing.T) {
	c, w := setupTest()

	message := "Not authenticated"
	GinUnauthorized(c, message)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
	assert.Equal(t, message, response.Error.Message)
}

func TestGinForbidden(t *testing.T) {
	c, w := setupTest()

	message := "Access denied"
	GinForbidden(c, message)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "FORBIDDEN", response.Error.Code)
	assert.Equal(t, message, response.Error.Message)
}

func TestGinNotFound(t *testing.T) {
	c, w := setupTest()

	message := "Resource not found"
	GinNotFound(c, message)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, message, response.Error.Message)
}

func TestGinInternalError(t *testing.T) {
	c, w := setupTest()

	err := errors.New("something went wrong")
	GinInternalError(c, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	assert.Equal(t, "An internal error occurred", response.Error.Message)
}

// Test standard http.ResponseWriter functions
func TestStandardHTTPFunctions(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	// Test JSON function
	JSON(w, http.StatusOK, data)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	// Reset recorder
	w = httptest.NewRecorder()

	// Test Error function
	Error(w, http.StatusBadRequest, "test error")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	// Reset recorder
	w = httptest.NewRecorder()

	// Test NoContent function
	NoContent(w)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}
