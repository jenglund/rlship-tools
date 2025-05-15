package models

import (
	"time"

	"github.com/google/uuid"
)

// AuthProvider represents the authentication provider type
type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderEmail  AuthProvider = "email"
	AuthProviderPhone  AuthProvider = "phone"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	FirebaseUID string       `json:"firebase_uid" db:"firebase_uid"`
	Provider    AuthProvider `json:"provider" db:"provider"`
	Email       string       `json:"email" db:"email"`
	Name        string       `json:"name" db:"name"`
	AvatarURL   string       `json:"avatar_url,omitempty" db:"avatar_url"`
	LastLogin   *time.Time   `json:"last_login,omitempty" db:"last_login"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByFirebaseUID(firebaseUID string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	List(offset, limit int) ([]*User, error)
}
