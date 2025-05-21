package handlers

import (
	"github.com/jenglund/rlship-tools/internal/middleware"
)

// Re-export the middleware.ContextUserIDKey for backward compatibility
// This allows existing code to continue using handlers.UserIDKey
var UserIDKey = middleware.ContextUserIDKey
