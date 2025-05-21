package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jenglund/rlship-tools/internal/models"
)

// DevFirebaseAuth is a development-friendly implementation of Firebase auth
// that accepts all users in development mode but identifies users with
// email addresses ending with @gonetribal.com as dev users with admin access
type DevFirebaseAuth struct {
	repos RepositoryProvider
}

// NewDevFirebaseAuth creates a new development Firebase auth client
func NewDevFirebaseAuth() *DevFirebaseAuth {
	return &DevFirebaseAuth{}
}

// IsDevUser checks if the email belongs to a developer (ends with @gonetribal.com)
func IsDevUser(email string) bool {
	return strings.HasSuffix(email, "@gonetribal.com")
}

// SetRepositoryProvider sets the repository provider for user creation
func (d *DevFirebaseAuth) SetRepositoryProvider(repos RepositoryProvider) {
	d.repos = repos
}

// VerifyIDToken implements the AuthClient interface
// In development mode, it accepts all tokens and generates a fake token
func (d *DevFirebaseAuth) VerifyIDToken(ctx context.Context, token string) (*auth.Token, error) {
	// For tokens with "dev:" prefix, extract the email
	if strings.HasPrefix(token, "dev:") {
		email := strings.TrimPrefix(token, "dev:")
		userName := strings.Split(email, "@")[0]

		// Create a fake token for the user
		return &auth.Token{
			UID: fmt.Sprintf("dev-%s", email),
			Claims: map[string]interface{}{
				"email": email,
				"name":  userName,
			},
		}, nil
	}

	// For any other tokens in dev mode, create a generic token
	// In a real production environment, this would verify with Firebase
	return &auth.Token{
		UID: fmt.Sprintf("dev-user-%s", token),
		Claims: map[string]interface{}{
			"email": "dev-user@example.com",
			"name":  "Development User",
		},
	}, nil
}

// CreateOrGetDevUser ensures the user exists in the database
func (d *DevFirebaseAuth) CreateOrGetDevUser(c *gin.Context, firebaseUID, email, name string) (*models.User, error) {
	if d.repos == nil {
		return nil, fmt.Errorf("repository provider not set")
	}

	// Try to get the user first
	userRepo := d.repos.GetUserRepository()
	user, err := userRepo.GetByFirebaseUID(firebaseUID)

	// If not found, create the user
	if err != nil {
		fmt.Printf("User not found by Firebase UID %s, will create: %v\n", firebaseUID, err)

		// Check if user exists by email as a fallback
		user, emailErr := userRepo.GetByEmail(email)
		if emailErr == nil && user != nil {
			fmt.Printf("Found user by email instead: %s\n", email)
			// Update Firebase UID if it doesn't match
			if user.FirebaseUID != firebaseUID {
				user.FirebaseUID = firebaseUID
				if updateErr := userRepo.Update(user); updateErr != nil {
					fmt.Printf("Error updating user Firebase UID: %v\n", updateErr)
				}
			}
			return user, nil
		}

		// Create a new user
		newUser := &models.User{
			FirebaseUID: firebaseUID,
			Email:       email,
			Name:        name,
			Provider:    models.AuthProviderEmail,
		}

		if err := userRepo.Create(newUser); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		fmt.Printf("Created new user: %s (%s)\n", email, newUser.ID)
		return newUser, nil
	}

	return user, nil
}

// AuthMiddleware returns a Gin middleware function for development authentication
// This implements the AuthMiddleware interface
func (d *DevFirebaseAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		devEmail := c.Request.Header.Get("X-Dev-Email")

		// Case 1: Using X-Dev-Email header
		if devEmail != "" {
			uid := fmt.Sprintf("dev-%s", devEmail)
			name := strings.Split(devEmail, "@")[0]

			c.Set(string(ContextFirebaseUIDKey), uid)
			c.Set(string(ContextUserEmailKey), devEmail)
			c.Set(string(ContextUserNameKey), name)

			// Create or get the user if repositories are available
			if d.repos != nil {
				user, err := d.CreateOrGetDevUser(c, uid, devEmail, name)
				if err != nil {
					fmt.Printf("Warning: Failed to get/create user: %v\n", err)
				} else if user != nil {
					// Set user ID in context
					c.Set(string(ContextUserIDKey), user.ID)
				}
			}

			// All authenticated users are allowed to proceed in development mode
			c.Next()
			return
		}

		// Case 2: Using dev: token
		if token != "" && strings.HasPrefix(token, "dev:") {
			email := strings.TrimPrefix(token, "dev:")
			uid := fmt.Sprintf("dev-%s", email)
			name := strings.Split(email, "@")[0]

			c.Set(string(ContextFirebaseUIDKey), uid)
			c.Set(string(ContextUserEmailKey), email)
			c.Set(string(ContextUserNameKey), name)

			// Create or get the user if repositories are available
			if d.repos != nil {
				user, err := d.CreateOrGetDevUser(c, uid, email, name)
				if err != nil {
					fmt.Printf("Warning: Failed to get/create user: %v\n", err)
				} else if user != nil {
					// Set user ID in context
					c.Set(string(ContextUserIDKey), user.ID)
				}
			}

			// All authenticated users are allowed to proceed in development mode
			c.Next()
			return
		}

		// Case 3: Using any other token in development mode
		if token != "" {
			// For any token in development mode, accept it with generic info
			c.Set(string(ContextFirebaseUIDKey), fmt.Sprintf("dev-user-%s", token))
			email := "dev-user@example.com"
			name := "Development User"
			c.Set(string(ContextUserEmailKey), email)
			c.Set(string(ContextUserNameKey), name)

			// Create or get a generic dev user if repositories are available
			if d.repos != nil {
				uid := fmt.Sprintf("dev-user-%s", token)
				user, err := d.CreateOrGetDevUser(c, uid, email, name)
				if err != nil {
					fmt.Printf("Warning: Failed to get/create generic user: %v\n", err)
				} else if user != nil {
					// Set user ID in context
					c.Set(string(ContextUserIDKey), user.ID)
				}
			}

			// All authenticated users are allowed to proceed in development mode
			c.Next()
			return
		}

		// No credentials provided
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}
