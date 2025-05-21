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

// SuccessResponse returns a standardized success response structure
func SuccessResponse(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

// ErrorResponse returns a standardized error response structure
func ErrorResponse(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"message": message,
		},
	}
}

// JSON sends a JSON response with the given status code and data
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	// If data is already a map with success/error fields, use it directly
	// Otherwise, wrap it in a success response
	var responseData interface{}

	switch v := data.(type) {
	case map[string]interface{}:
		// Check if it's already in our expected format
		if _, hasSuccess := v["success"]; hasSuccess {
			responseData = v
		} else {
			responseData = SuccessResponse(v)
		}
	default:
		responseData = SuccessResponse(data)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(responseData)
}

// Error sends a standardized error response with the given status code and message
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, ErrorResponse(message))
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Created sends a 201 Created response with the given data
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, SuccessResponse(data))
}

// OK sends a 200 OK response with the given data
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, SuccessResponse(data))
}

// BadRequest sends a 400 Bad Request response with the given message
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized response with the given message
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response with the given message
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message)
}

// NotFound sends a 404 Not Found response with the given message
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

// InternalServerError sends a 500 Internal Server Error response with the given error
func InternalServerError(w http.ResponseWriter, err error) {
	Error(w, http.StatusInternalServerError, err.Error())
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
