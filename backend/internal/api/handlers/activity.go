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

// ActivityHandler handles activity-related requests
type ActivityHandler struct {
	repos *postgres.Repositories
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(repos *postgres.Repositories) *ActivityHandler {
	return &ActivityHandler{repos: repos}
}

// RegisterRoutes registers the activity routes
func (h *ActivityHandler) RegisterRoutes(r *gin.RouterGroup) {
	activities := r.Group("/activities", middleware.RequireAuth())
	{
		activities.POST("", h.CreateActivity)
		activities.GET("", h.ListActivities)
		activities.GET("/:id", h.GetActivity)
		activities.PUT("/:id", h.UpdateActivity)
		activities.DELETE("/:id", h.DeleteActivity)

		// Activity ownership
		activities.POST("/:id/owners", h.AddOwner)
		activities.DELETE("/:id/owners/:ownerId", h.RemoveOwner)
		activities.GET("/:id/owners", h.ListOwners)

		// Activity sharing
		activities.POST("/:id/share", h.ShareActivity)
		activities.DELETE("/:id/share/:tribeId", h.UnshareActivity)
		activities.GET("/shared", h.ListSharedActivities)
	}
}

// CreateActivityRequest represents the create activity request body
type CreateActivityRequest struct {
	Type        models.ActivityType   `json:"type" binding:"required"`
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	Visibility  models.VisibilityType `json:"visibility"`
	Metadata    interface{}           `json:"metadata,omitempty"`
}

// CreateActivity creates a new activity
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err.Error())
		return
	}

	var metadata models.JSONMap
	if req.Metadata != nil {
		var ok bool
		metadata, ok = req.Metadata.(map[string]interface{})
		if !ok {
			response.GinBadRequest(c, "Metadata must be a JSON object")
			return
		}
	}

	activity := &models.Activity{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repos.Activities.Create(activity); err != nil {
		response.GinInternalError(c, err)
		return
	}

	// Add the current user as an owner
	firebaseUID := middleware.GetFirebaseUID(c)
	if firebaseUID == "" {
		fmt.Printf("Firebase UID not found in context\n")
		response.GinInternalError(c, fmt.Errorf("firebase UID not found in context"))
		return
	}

	user, err := h.repos.Users.GetByFirebaseUID(firebaseUID)
	if err != nil {
		fmt.Printf("Error getting user by Firebase UID: %v\n", err)
		response.GinInternalError(c, err)
		return
	}

	if err := h.repos.Activities.AddOwner(activity.ID, user.ID, "user"); err != nil {
		fmt.Printf("Error adding owner: %v\n", err)
		response.GinInternalError(c, err)
		return
	}

	response.GinCreated(c, activity)
}

// ListActivities returns a paginated list of activities
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	offset := 0
	limit := 50

	activities, err := h.repos.Activities.List(offset, limit)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, activities)
}

// GetActivity returns a single activity by ID
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	activity, err := h.repos.Activities.GetByID(id)
	if err != nil {
		response.GinNotFound(c, "Activity not found")
		return
	}

	response.GinSuccess(c, activity)
}

// UpdateActivityRequest represents the update activity request body
type UpdateActivityRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Visibility  models.VisibilityType `json:"visibility"`
	Metadata    interface{}           `json:"metadata,omitempty"`
}

// UpdateActivity updates an existing activity
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	var req UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err.Error())
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	activity, err := h.repos.Activities.GetByID(activityID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	metadata, ok := req.Metadata.(map[string]interface{})
	if !ok {
		response.GinBadRequest(c, "Metadata must be a JSON object")
		return
	}

	activity.Name = req.Name
	activity.Description = req.Description
	activity.Visibility = req.Visibility
	activity.Metadata = models.JSONMap(metadata)
	activity.UpdatedAt = time.Now()

	if err := h.repos.Activities.Update(activity); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, activity)
}

// DeleteActivity removes an activity
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	if err := h.repos.Activities.Delete(id); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// AddOwnerRequest represents the add owner request body
type AddOwnerRequest struct {
	OwnerID   uuid.UUID `json:"owner_id" binding:"required"`
	OwnerType string    `json:"owner_type" binding:"required,oneof=user tribe"`
}

// AddOwner adds an owner to an activity
func (h *ActivityHandler) AddOwner(c *gin.Context) {
	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	var req AddOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	if err := h.repos.Activities.AddOwner(activityID, req.OwnerID, req.OwnerType); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// RemoveOwner removes an owner from an activity
func (h *ActivityHandler) RemoveOwner(c *gin.Context) {
	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	ownerID, err := uuid.Parse(c.Param("ownerId"))
	if err != nil {
		response.GinBadRequest(c, "Invalid owner ID")
		return
	}

	if err := h.repos.Activities.RemoveOwner(activityID, ownerID); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// ListOwners returns all owners of an activity
func (h *ActivityHandler) ListOwners(c *gin.Context) {
	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	owners, err := h.repos.Activities.GetOwners(activityID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinSuccess(c, owners)
}

// ShareActivityRequest represents the share activity request body
type ShareActivityRequest struct {
	TribeID   uuid.UUID  `json:"tribe_id" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ShareActivity shares an activity with a tribe
func (h *ActivityHandler) ShareActivity(c *gin.Context) {
	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	var req ShareActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	// Get the current user's ID from Firebase UID
	firebaseUID := middleware.GetFirebaseUID(c)
	if firebaseUID == "" {
		response.GinInternalError(c, fmt.Errorf("firebase UID not found in context"))
		return
	}

	user, err := h.repos.Users.GetByFirebaseUID(firebaseUID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	if err := h.repos.Activities.ShareWithTribe(activityID, req.TribeID, user.ID, req.ExpiresAt); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// UnshareActivity removes an activity share from a tribe
func (h *ActivityHandler) UnshareActivity(c *gin.Context) {
	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid activity ID")
		return
	}

	tribeID, err := uuid.Parse(c.Param("tribeId"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	if err := h.repos.Activities.UnshareWithTribe(activityID, tribeID); err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinNoContent(c)
}

// ListSharedActivities returns all activities shared with the user's tribes
func (h *ActivityHandler) ListSharedActivities(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.GinInternalError(c, fmt.Errorf("user ID not found in context"))
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	// Get user's tribes
	tribes, err := h.repos.Tribes.GetUserTribes(uid)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	// Get shared activities for each tribe
	var allSharedActivities []*models.Activity
	for _, tribe := range tribes {
		activities, err := h.repos.Activities.GetSharedActivities(tribe.ID)
		if err != nil {
			response.GinInternalError(c, err)
			return
		}
		allSharedActivities = append(allSharedActivities, activities...)
	}

	response.GinSuccess(c, allSharedActivities)
}
