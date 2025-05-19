package middleware

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Context keys used throughout the application
const (
	ContextFirebaseUIDKey ContextKey = "firebase_uid"
	ContextUserIDKey      ContextKey = "user_id"
	ContextUserEmailKey   ContextKey = "user_email"
	ContextUserNameKey    ContextKey = "user_name"
)
