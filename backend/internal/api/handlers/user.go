package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/api/response"
	"github.com/jenglund/rlship-tools/internal/middleware"
	"github.com/jenglund/rlship-tools/internal/models"
)

// RepositoryProvider defines the interface for accessing repositories
type RepositoryProvider interface {
	GetUserRepository() models.UserRepository
}

// UserHandler handles user-related requests
type UserHandler struct {
	repos RepositoryProvider
}

// NewUserHandler creates a new user handler
func NewUserHandler(repos RepositoryProvider) *UserHandler {
	return &UserHandler{
		repos: repos,
	}
}

// RegisterRoutes registers the user routes
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		// Public routes
		users.POST("/auth", h.AuthenticateUser)

		// Protected routes
		auth := users.Use(middleware.RequireAuth())
		{
			auth.GET("/me", h.GetCurrentUser)
			auth.PUT("/me", h.UpdateCurrentUser)
			auth.GET("/:id", h.GetUser)
		}
	}
}

// AuthRequest represents the authentication request body
type AuthRequest struct {
	FirebaseUID string `json:"firebase_uid" binding:"required"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Provider    string `json:"provider" binding:"required"`
}

// AuthenticateUser handles user authentication and registration
func (h *UserHandler) AuthenticateUser(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Check if user exists
	user, err := h.repos.GetUserRepository().GetByFirebaseUID(req.FirebaseUID)
	if err != nil {
		// Create new user if not found
		user = &models.User{
			ID:          uuid.New(),
			FirebaseUID: req.FirebaseUID,
			Email:       req.Email,
			Name:        req.Name,
			AvatarURL:   req.AvatarURL,
			Provider:    models.AuthProvider(req.Provider),
		}

		if err := h.repos.GetUserRepository().Create(user); err != nil {
			response.InternalError(c, err)
			return
		}
		response.Created(c, user)
		return
	}

	// Update existing user's information
	user.Email = req.Email
	user.Name = req.Name
	user.AvatarURL = req.AvatarURL
	user.Provider = models.AuthProvider(req.Provider)

	if err := h.repos.GetUserRepository().Update(user); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, user)
}

// GetCurrentUser returns the current authenticated user
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.repos.GetUserRepository().GetByID(userID.(uuid.UUID))
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}

// UpdateUserRequest represents the update user request body
type UpdateUserRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// UpdateCurrentUser updates the current user's profile
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.repos.GetUserRepository().GetByID(userID.(uuid.UUID))
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := h.repos.GetUserRepository().Update(user); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, user)
}

// GetUser returns a specific user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.repos.GetUserRepository().GetByID(id)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}
