package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/api/response"
	"github.com/jenglund/rlship-tools/internal/middleware"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/repository/postgres"
)

// TribeHandler handles tribe-related requests
type TribeHandler struct {
	repos *postgres.Repositories
}

// NewTribeHandler creates a new tribe handler
func NewTribeHandler(repos *postgres.Repositories) *TribeHandler {
	return &TribeHandler{repos: repos}
}

// RegisterRoutes registers the tribe routes
func (h *TribeHandler) RegisterRoutes(r *gin.RouterGroup) {
	tribes := r.Group("/tribes", middleware.RequireAuth())
	{
		tribes.POST("", h.CreateTribe)
		tribes.GET("", h.ListTribes)
		tribes.GET("/my", h.ListMyTribes)
		tribes.GET("/:id", h.GetTribe)
		tribes.PUT("/:id", h.UpdateTribe)
		tribes.DELETE("/:id", h.DeleteTribe)

		// Member management
		tribes.POST("/:id/members", h.AddMember)
		tribes.DELETE("/:id/members/:userId", h.RemoveMember)
		tribes.GET("/:id/members", h.ListMembers)
	}
}

// CreateTribeRequest represents the create tribe request body
type CreateTribeRequest struct {
	Name        string                `json:"name" binding:"required"`
	Type        models.TribeType      `json:"type" binding:"required"`
	Description string                `json:"description"`
	Visibility  models.VisibilityType `json:"visibility" binding:"required"`
	Metadata    interface{}           `json:"metadata,omitempty"`
}

// CreateTribe creates a new tribe
func (h *TribeHandler) CreateTribe(c *gin.Context) {
	var req CreateTribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("DEBUG: Failed to bind JSON: %v\n", err)
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	// Convert metadata to JSONMap if provided
	var metadata models.JSONMap
	if req.Metadata != nil {
		switch m := req.Metadata.(type) {
		case map[string]interface{}:
			metadata = models.JSONMap(m)
		case models.JSONMap:
			metadata = m
		default:
			response.GinBadRequest(c, "Metadata must be a valid JSON object")
			return
		}

		// Validate metadata
		if err := metadata.Validate(); err != nil {
			response.GinBadRequest(c, err.Error())
			return
		}
	}

	now := time.Now()
	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Visibility:  req.Visibility,
		Metadata:    metadata,
	}

	// Validate the tribe
	if err := tribe.Validate(); err != nil {
		response.GinBadRequest(c, err.Error())
		return
	}

	if err := h.repos.Tribes.Create(tribe); err != nil {
		fmt.Printf("DEBUG: Failed to create tribe: %v\n", err)
		response.GinInternalError(c, err)
		return
	}

	// Add the creator as the first member
	creatorID := middleware.GetFirebaseUID(c)
	fmt.Printf("DEBUG: Creator Firebase UID: %v\n", creatorID)

	user, err := h.repos.Users.GetByFirebaseUID(creatorID)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get user by Firebase UID: %v\n", err)
		response.GinInternalError(c, err)
		return
	}
	fmt.Printf("DEBUG: Found user: %+v\n", user)

	if err := h.repos.Tribes.AddMember(tribe.ID, user.ID, models.MembershipFull, nil, nil); err != nil {
		fmt.Printf("DEBUG: Failed to add member to tribe: %v\n", err)
		response.GinInternalError(c, err)
		return
	}

	// Get the updated tribe with members
	tribe, err = h.repos.Tribes.GetByID(tribe.ID)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get updated tribe: %v\n", err)
		response.GinInternalError(c, err)
		return
	}

	response.GinCreated(c, tribe)
}

// ListTribes returns a paginated list of all tribes
func (h *TribeHandler) ListTribes(c *gin.Context) {
	offset := 0
	limit := 20

	tribes, err := h.repos.Tribes.List(offset, limit)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, tribes)
}

// ListMyTribes returns all tribes that the current user is a member of
func (h *TribeHandler) ListMyTribes(c *gin.Context) {
	firebaseUID := middleware.GetFirebaseUID(c)
	user, err := h.repos.Users.GetByFirebaseUID(firebaseUID)
	if err != nil {
		response.GinNotFound(c, "User not found")
		return
	}

	tribes, err := h.repos.Tribes.GetUserTribes(user.ID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, tribes)
}

// GetTribe returns a specific tribe by ID
func (h *TribeHandler) GetTribe(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	tribe, err := h.repos.Tribes.GetByID(id)
	if err != nil {
		response.GinNotFound(c, "Tribe not found")
		return
	}

	response.GinSuccess(c, tribe)
}

// UpdateTribeRequest represents the update tribe request body
type UpdateTribeRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateTribe updates a tribe's details
func (h *TribeHandler) UpdateTribe(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	var req UpdateTribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	tribe, err := h.repos.Tribes.GetByID(id)
	if err != nil {
		response.GinNotFound(c, "Tribe not found")
		return
	}

	tribe.Name = req.Name
	if err := h.repos.Tribes.Update(tribe); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, tribe)
}

// DeleteTribe removes a tribe and all its associations
func (h *TribeHandler) DeleteTribe(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	if err := h.repos.Tribes.Delete(id); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// AddMemberRequest represents the add member request body
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// AddMember adds a user to a tribe
func (h *TribeHandler) AddMember(c *gin.Context) {
	tribeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	// Get the current user as the inviter
	inviterID := c.GetString("user_id")
	var inviter *uuid.UUID
	if inviterID != "" {
		uid, err := uuid.Parse(inviterID)
		if err == nil {
			inviter = &uid
		}
	}

	if err := h.repos.Tribes.AddMember(tribeID, req.UserID, models.MembershipFull, nil, inviter); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// RemoveMember removes a user from a tribe
func (h *TribeHandler) RemoveMember(c *gin.Context) {
	tribeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.GinBadRequest(c, "Invalid user ID")
		return
	}

	if err := h.repos.Tribes.RemoveMember(tribeID, userID); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// ListMembers returns all members of a tribe
func (h *TribeHandler) ListMembers(c *gin.Context) {
	tribeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	members, err := h.repos.Tribes.GetMembers(tribeID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, members)
}
