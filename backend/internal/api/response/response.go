package response

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is a generic API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Standard http.ResponseWriter functions

// JSON sends a JSON response using standard http.ResponseWriter
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Error sends a JSON error response using standard http.ResponseWriter
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, struct {
		Success bool `json:"success"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}{
		Success: false,
		Error: struct {
			Message string `json:"message"`
		}{
			Message: message,
		},
	})
}

// NoContent sends a 204 No Content response using standard http.ResponseWriter
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Gin-specific functions

// GinSuccess sends a successful response using Gin
func GinSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// GinCreated sends a 201 Created response using Gin
func GinCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// GinNoContent sends a 204 No Content response using Gin
func GinNoContent(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

// GinError sends an error response using Gin
func GinError(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// GinBadRequest sends a 400 Bad Request response using Gin
func GinBadRequest(c *gin.Context, message string) {
	GinError(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// GinUnauthorized sends a 401 Unauthorized response using Gin
func GinUnauthorized(c *gin.Context, message string) {
	GinError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// GinForbidden sends a 403 Forbidden response using Gin
func GinForbidden(c *gin.Context, message string) {
	GinError(c, http.StatusForbidden, "FORBIDDEN", message)
}

// GinNotFound sends a 404 Not Found response using Gin
func GinNotFound(c *gin.Context, message string) {
	GinError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// GinConflict sends a 409 Conflict response using Gin
func GinConflict(c *gin.Context, message string) {
	GinError(c, http.StatusConflict, "CONFLICT", message)
}

// GinInternalError sends a 500 Internal Server Error response using Gin
func GinInternalError(c *gin.Context, err error) {
	GinError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
}
