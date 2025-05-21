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
	activities := r.Group("/activities")
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
		activities.DELETE("/:id/share/:tribeID", h.UnshareActivity)
		activities.GET("/shared", h.ListSharedActivities)
	}
}

// CreateActivityRequest represents the create activity request body
type CreateActivityRequest struct {
	Type        models.ActivityType   `json:"type" binding:"required"`
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	Visibility  models.VisibilityType `json:"visibility" binding:"required"`
	Metadata    interface{}           `json:"metadata,omitempty"`
}

// CreateActivity creates a new activity
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, err.Error())
		return
	}

	// Get the current user
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
	activity := &models.Activity{
		ID:          uuid.New(),
		UserID:      user.ID,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Visibility:  req.Visibility,
		Metadata:    metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Validate the activity
	if err := activity.Validate(); err != nil {
		response.GinBadRequest(c, err.Error())
		return
	}

	if err := h.repos.Activities.Create(activity); err != nil {
		response.GinInternalError(c, err)
		return
	}

	// Add the current user as an owner
	if err := h.repos.Activities.AddOwner(activity.ID, user.ID, "user"); err != nil {
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
		if err.Error() == "activity not found" {
			response.GinNotFound(c, "Activity not found")
			return
		}
		response.GinInternalError(c, err)
		return
	}

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

	activity.Name = req.Name
	activity.Description = req.Description
	activity.Visibility = req.Visibility
	activity.Metadata = metadata
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
	OwnerID   uuid.UUID        `json:"owner_id" binding:"required"`
	OwnerType models.OwnerType `json:"owner_type" binding:"required,oneof=user tribe"`
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
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
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

	tribeID, err := uuid.Parse(c.Param("tribeID"))
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
		response.GinSuccess(c, []*models.Activity{})
		return
	}

	// Use a map to deduplicate activities by ID
	activityMap := make(map[uuid.UUID]*models.Activity)

	// Get shared activities for each tribe
	for _, tribe := range tribes {
		activities, sharedActivitiesErr := h.repos.Activities.GetSharedActivities(tribe.ID)
		if sharedActivitiesErr != nil {
			continue
		}

		// Add activities to map to deduplicate
		for _, activity := range activities {
			activityMap[activity.ID] = activity
		}
	}

	// Convert map to slice
	var deduplicatedActivities []*models.Activity
	for _, activity := range activityMap {
		deduplicatedActivities = append(deduplicatedActivities, activity)
	}

	response.GinSuccess(c, deduplicatedActivities)
}
