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

func TestSuccess(t *testing.T) {
	c, w := setupTest()

	testData := map[string]string{"key": "value"}
	Success(c, testData)

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

func TestCreated(t *testing.T) {
	c, w := setupTest()

	testData := map[string]string{"key": "value"}
	Created(c, testData)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)
}

func TestNoContent(t *testing.T) {
	c, w := setupTest()

	NoContent(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestSendError(t *testing.T) {
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

			SendError(c, tt.status, tt.code, tt.message)

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

func TestBadRequest(t *testing.T) {
	c, w := setupTest()

	message := "Invalid input"
	BadRequest(c, message)

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

func TestUnauthorized(t *testing.T) {
	c, w := setupTest()

	message := "Not authenticated"
	Unauthorized(c, message)

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

func TestForbidden(t *testing.T) {
	c, w := setupTest()

	message := "Access denied"
	Forbidden(c, message)

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

func TestNotFound(t *testing.T) {
	c, w := setupTest()

	message := "Resource not found"
	NotFound(c, message)

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

func TestInternalError(t *testing.T) {
	c, w := setupTest()

	err := errors.New("something went wrong")
	InternalError(c, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, jsonErr)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	assert.Equal(t, "An internal error occurred", response.Error.Message)
}
