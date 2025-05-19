package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jenglund/rlship-tools/internal/models"
	"google.golang.org/api/option"
)

// AuthMiddleware defines the interface for authentication middleware
type AuthMiddleware interface {
	AuthMiddleware() gin.HandlerFunc
}

// RepositoryProvider defines the interface for accessing repositories
type RepositoryProvider interface {
	DB() *sql.DB
	GetUserRepository() models.UserRepository
}

// AuthClient defines the interface for Firebase auth operations
type AuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// FirebaseAuth middleware for validating Firebase tokens
type FirebaseAuth struct {
	client AuthClient
}

// NewFirebaseAuth creates a new Firebase authentication middleware
func NewFirebaseAuth(credentialsFile string) (*FirebaseAuth, error) {
	// Check if the credentials file exists
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("firebase credentials file does not exist: %s", credentialsFile)
	}

	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting firebase auth client: %w", err)
	}

	return &FirebaseAuth{client: client}, nil
}

// AuthMiddleware returns a Gin middleware function that validates Firebase tokens
func (fa *FirebaseAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no token provided"})
			return
		}

		// Verify the token
		tokenData, err := fa.client.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Add the Firebase UID to the context
		c.Set("firebase_uid", tokenData.UID)
		c.Set("user_email", tokenData.Claims["email"])
		c.Set("user_name", tokenData.Claims["name"])

		c.Next()
	}
}

// RequireAuth returns a middleware that requires Firebase authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get("firebase_uid")
		fmt.Printf("RequireAuth middleware: firebase_uid exists=%v, value=%v\n", exists, uid)

		userId, userIdExists := c.Get("user_id")
		fmt.Printf("RequireAuth middleware: user_id exists=%v, value=%v\n", userIdExists, userId)

		if !exists {
			fmt.Printf("RequireAuth middleware: Authentication failed, aborting\n")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		fmt.Printf("RequireAuth middleware: Authentication successful, continuing\n")
		c.Next()
	}
}

// extractToken gets the token from the Authorization header
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// GetFirebaseUID retrieves the Firebase UID from the Gin context
func GetFirebaseUID(c *gin.Context) string {
	uid, exists := c.Get("firebase_uid")
	if !exists {
		return ""
	}
	return uid.(string)
}

// GetUserEmail retrieves the user's email from the Gin context
func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get("user_email")
	if !exists {
		return ""
	}
	return email.(string)
}

// GetUserName retrieves the user's name from the Gin context
func GetUserName(c *gin.Context) string {
	name, exists := c.Get("user_name")
	if !exists {
		return ""
	}
	return name.(string)
}

// UserIDMiddleware converts firebase_uid to user_id by looking up the user in the database
func UserIDMiddleware(repos RepositoryProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		firebaseUID := GetFirebaseUID(c)
		if firebaseUID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "firebase UID not found in context"})
			return
		}

		userRepo := repos.GetUserRepository()
		if userRepo == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user repository not configured"})
			return
		}

		user, err := userRepo.GetByFirebaseUID(firebaseUID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		c.Set("user_id", user.ID.String())
		c.Next()
	}
}
