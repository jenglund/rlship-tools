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

// GetContextKey returns the string value of a ContextKey
// This is useful when using context.WithValue which requires the same exact key type
func GetContextKey(key ContextKey) string {
	return string(key)
}
