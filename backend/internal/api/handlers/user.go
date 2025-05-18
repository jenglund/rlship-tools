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
		users.POST("/check-email", h.CheckEmailExists)
		users.POST("/register", h.RegisterUser)
		users.POST("/login", h.LoginUser)

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
		response.GinBadRequest(c, "Invalid request body")
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
			response.GinInternalError(c, err)
			return
		}
		response.GinCreated(c, user)
		return
	}

	// Update existing user's information
	user.Email = req.Email
	user.Name = req.Name
	user.AvatarURL = req.AvatarURL
	user.Provider = models.AuthProvider(req.Provider)

	if err := h.repos.GetUserRepository().Update(user); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, user)
}

// CheckEmailRequest represents the email check request
type CheckEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// CheckEmailResponse represents the email check response
type CheckEmailResponse struct {
	Exists bool `json:"exists"`
}

// CheckEmailExists checks if an email already exists in the system
func (h *UserHandler) CheckEmailExists(c *gin.Context) {
	var req CheckEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	// Check if user exists
	user, err := h.repos.GetUserRepository().GetByEmail(req.Email)
	exists := err == nil && user != nil

	response.GinSuccess(c, CheckEmailResponse{Exists: exists})
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Name      string `json:"name" binding:"required"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider" binding:"required"`
}

// RegisterUser handles new user registration
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	// Check if user already exists
	existingUser, _ := h.repos.GetUserRepository().GetByEmail(req.Email)
	if existingUser != nil {
		response.GinBadRequest(c, "User with this email already exists")
		return
	}

	// Create new user
	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
		Provider:  models.AuthProvider(req.Provider),
	}

	if err := h.repos.GetUserRepository().Create(user); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinCreated(c, user)
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// LoginUser handles user login
func (h *UserHandler) LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	// Find user by email
	user, err := h.repos.GetUserRepository().GetByEmail(req.Email)
	if err != nil || user == nil {
		response.GinNotFound(c, "User not found")
		return
	}

	// In a real application, we would check credentials here
	// For this demo, we'll simply return the user

	response.GinSuccess(c, user)
}

// GetCurrentUser returns the current authenticated user
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GinUnauthorized(c, "User not authenticated")
		return
	}

	user, err := h.repos.GetUserRepository().GetByID(userID.(uuid.UUID))
	if err != nil {
		response.GinNotFound(c, "User not found")
		return
	}

	response.GinSuccess(c, user)
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
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.GinUnauthorized(c, "User not authenticated")
		return
	}

	user, err := h.repos.GetUserRepository().GetByID(userID.(uuid.UUID))
	if err != nil {
		response.GinNotFound(c, "User not found")
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
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, user)
}

// GetUser returns a specific user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.repos.GetUserRepository().GetByID(id)
	if err != nil {
		response.GinNotFound(c, "User not found")
		return
	}

	response.GinSuccess(c, user)
}
