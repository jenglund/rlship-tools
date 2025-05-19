package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/jenglund/rlship-tools/internal/models"
)

// DevFirebaseAuth is a development-friendly implementation of Firebase auth
// that automatically accepts users with email addresses matching dev_user[0-9]+@gonetribal.com
type DevFirebaseAuth struct {
	devEmailPattern *regexp.Regexp
	repos           RepositoryProvider
}

// NewDevFirebaseAuth creates a new development Firebase auth client
// that accepts dev users without actual authentication
func NewDevFirebaseAuth() *DevFirebaseAuth {
	return &DevFirebaseAuth{
		devEmailPattern: regexp.MustCompile(`^dev_user[0-9]+@gonetribal\.com$`),
	}
}

// SetRepositoryProvider sets the repository provider for user creation
func (d *DevFirebaseAuth) SetRepositoryProvider(repos RepositoryProvider) {
	d.repos = repos
}

// VerifyIDToken implements the AuthClient interface
// For dev users, it generates a fake token with the user's email
func (d *DevFirebaseAuth) VerifyIDToken(ctx context.Context, token string) (*auth.Token, error) {
	// Check if this is a dev user token
	if strings.HasPrefix(token, "dev:") {
		email := strings.TrimPrefix(token, "dev:")

		// Validate that it matches our dev email pattern
		if !d.devEmailPattern.MatchString(email) {
			return nil, fmt.Errorf("invalid dev user email format")
		}

		// Create a fake token for the dev user
		return &auth.Token{
			UID: fmt.Sprintf("dev-%s", email),
			Claims: map[string]interface{}{
				"email": email,
				"name":  fmt.Sprintf("Dev User %s", strings.Split(email, "@")[0][8:]),
			},
		}, nil
	}

	// For non-dev tokens, reject them in dev mode
	return nil, fmt.Errorf("only dev users are supported in development mode")
}

// CreateOrGetDevUser ensures the dev user exists in the database
func (d *DevFirebaseAuth) CreateOrGetDevUser(c *gin.Context, firebaseUID, email, name string) (*models.User, error) {
	if d.repos == nil {
		return nil, fmt.Errorf("repository provider not set")
	}

	// Try to get the user first
	userRepo := d.repos.GetUserRepository()
	user, err := userRepo.GetByFirebaseUID(firebaseUID)

	// If not found, create the user
	if err != nil {
		return nil, fmt.Errorf("failed to create dev user: %w", err)
	}

	return user, nil
}

// AuthMiddleware returns a Gin middleware function for development authentication
// This implements the AuthMiddleware interface
func (d *DevFirebaseAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		if token == "" {
			// For development, allow the email to be passed in a header
			devEmail := c.Request.Header.Get("X-Dev-Email")

			if devEmail != "" && d.devEmailPattern.MatchString(devEmail) {
				// Set a fake Firebase UID and user info
				uid := fmt.Sprintf("dev-%s", devEmail)
				name := fmt.Sprintf("Dev User %s", strings.Split(devEmail, "@")[0][8:])

				c.Set("firebase_uid", uid)
				c.Set("user_email", devEmail)
				c.Set("user_name", name)

				// Create or get the dev user if repositories are available
				if d.repos != nil {
					user, err := d.CreateOrGetDevUser(c, uid, devEmail, name)
					if err == nil {
						c.Set("user_id", user.ID)
					}
				}

				c.Next()
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no token or dev email provided"})
			return
		}

		// For tokens, check if it's a dev token
		if strings.HasPrefix(token, "dev:") {
			email := strings.TrimPrefix(token, "dev:")

			// Validate that it matches our dev email pattern
			if !d.devEmailPattern.MatchString(email) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid dev user email format"})
				return
			}

			// Set a fake Firebase UID and user info
			uid := fmt.Sprintf("dev-%s", email)
			name := fmt.Sprintf("Dev User %s", strings.Split(email, "@")[0][8:])

			c.Set("firebase_uid", uid)
			c.Set("user_email", email)
			c.Set("user_name", name)

			// Create or get the dev user if repositories are available
			if d.repos != nil {
				user, err := d.CreateOrGetDevUser(c, uid, email, name)
				if err == nil {
					c.Set("user_id", user.ID)
				}
			}

			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "only dev users are supported in development mode"})
	}
}
