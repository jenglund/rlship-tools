package handlers

import (
	"fmt"
	"strconv"
	"strings"
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
		tribes.DELETE("/:id/members/:userID", h.RemoveMember)
		tribes.GET("/:id/members", h.ListMembers)

		// Invitation response
		tribes.POST("/:id/respond", h.RespondToInvitation)
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

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate name
	if req.Name == "" {
		response.GinBadRequest(c, "Tribe name is required")
		return
	}

	if len(req.Name) > 100 {
		response.GinBadRequest(c, "Tribe name cannot be longer than 100 characters")
		return
	}

	// Validate tribe type
	if err := req.Type.Validate(); err != nil {
		response.GinBadRequest(c, fmt.Sprintf("Invalid tribe type: %v", err))
		return
	}

	// Validate visibility
	if err := req.Visibility.Validate(); err != nil {
		response.GinBadRequest(c, fmt.Sprintf("Invalid visibility type: %v", err))
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
	} else {
		// Initialize empty metadata if nil was provided
		metadata = models.JSONMap{}
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		response.GinUnauthorized(c, "Authentication required")
		return
	}

	// Create the tribe object
	now := time.Now()

	// Make sure metadata is never nil - initialize to empty map if needed
	if metadata == nil {
		metadata = models.JSONMap{}
	}

	tribe := &models.Tribe{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Visibility:  req.Visibility,
		Metadata:    metadata,
	}

	// Validate the tribe
	if validationErr := tribe.Validate(); validationErr != nil {
		response.GinBadRequest(c, validationErr.Error())
		return
	}

	// Create tribe and add member in a transaction
	if createErr := h.repos.Tribes.Create(tribe); createErr != nil {
		// Check for duplicate tribe name error
		if strings.Contains(createErr.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(createErr.Error(), "idx_unique_tribe_name") {
			response.GinBadRequest(c, "A tribe with this name already exists")
			return
		}
		response.GinBadRequest(c, fmt.Sprintf("Failed to create tribe: %v", createErr))
		return
	}

	// Add the creator as the first member
	if addMemberErr := h.repos.Tribes.AddMember(tribe.ID, userID, models.MembershipFull, nil, &userID); addMemberErr != nil {
		// Try to clean up the tribe if we couldn't add the member
		response.GinInternalError(c, addMemberErr)
		return
	}

	// Get the updated tribe with members
	tribe, err = h.repos.Tribes.GetByID(tribe.ID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	response.GinCreated(c, tribe)
}

// Helper function to get user ID from context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	// Try to get user ID directly first
	if userIDVal, exists := c.Get("user_id"); exists && userIDVal != nil {
		switch id := userIDVal.(type) {
		case uuid.UUID:
			return id, nil
		case string:
			return uuid.Parse(id)
		}
	}

	// If we couldn't get a valid user ID, return an error
	return uuid.Nil, fmt.Errorf("user ID not found in context")
}

// ListTribes returns a paginated list of all tribes
func (h *TribeHandler) ListTribes(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate pagination parameters
	if page < 1 {
		response.GinBadRequest(c, "Page number must be greater than 0")
		return
	}

	if limit < 1 {
		response.GinBadRequest(c, "Limit must be greater than 0")
		return
	}

	// Check for specific tribe_id parameter
	tribeIDParam := c.Query("tribe_id")
	if tribeIDParam != "" {
		// If tribe_id is specified, fetch just that one tribe
		tribeID, err := uuid.Parse(tribeIDParam)
		if err != nil {
			response.GinNotFound(c, "Tribe not found")
			return
		}

		tribe, err := h.repos.Tribes.GetByID(tribeID)
		if err != nil {
			if err.Error() == "tribe not found" {
				response.GinNotFound(c, "Tribe not found")
			} else {
				response.GinInternalError(c, err)
			}
			return
		}

		// Return just this one tribe
		response.GinSuccess(c, []*models.Tribe{tribe})
		return
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Parse filter parameters
	tribeType := c.Query("type")

	// Validate tribe type if provided
	if tribeType != "" {
		valid := false
		validTypes := []models.TribeType{
			models.TribeTypeCouple,
			models.TribeTypePolyCule,
			models.TribeTypeFriends,
			models.TribeTypeFamily,
			models.TribeTypeRoommates,
			models.TribeTypeCoworkers,
			models.TribeTypeCustom,
		}

		// Convert type to lowercase for case-insensitive matching
		typeToCheck := strings.ToLower(tribeType)

		for _, validType := range validTypes {
			if models.TribeType(typeToCheck) == validType {
				valid = true
				tribeType = string(validType) // Use canonical version
				break
			}
		}

		if !valid {
			response.GinBadRequest(c, "Invalid tribe type")
			return
		}
	}

	var tribes []*models.Tribe
	var err error

	// Fetch tribes based on filter
	if tribeType != "" {
		tribes, err = h.repos.Tribes.GetByType(models.TribeType(tribeType), offset, limit)
		if err != nil {
			response.GinInternalError(c, err)
			return
		}
	} else {
		tribes, err = h.repos.Tribes.List(offset, limit)
		if err != nil {
			response.GinInternalError(c, err)
			return
		}
	}

	response.GinSuccess(c, tribes)
}

// ListMyTribes returns all tribes that the current user is a member of
func (h *TribeHandler) ListMyTribes(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate pagination parameters
	if page < 1 {
		response.GinBadRequest(c, "Page number must be greater than 0")
		return
	}

	if limit < 1 {
		response.GinBadRequest(c, "Limit must be greater than 0")
		return
	}

	// Parse filter parameters
	tribeType := c.Query("type")

	// Validate tribe type if provided
	if tribeType != "" {
		valid := false
		validTypes := []models.TribeType{
			models.TribeTypeCouple,
			models.TribeTypePolyCule,
			models.TribeTypeFriends,
			models.TribeTypeFamily,
			models.TribeTypeRoommates,
			models.TribeTypeCoworkers,
			models.TribeTypeCustom,
		}

		// Convert type to lowercase for case-insensitive matching
		typeToCheck := strings.ToLower(tribeType)

		for _, validType := range validTypes {
			if models.TribeType(typeToCheck) == validType {
				valid = true
				tribeType = string(validType) // Use canonical version
				break
			}
		}

		if !valid {
			response.GinBadRequest(c, "Invalid tribe type")
			return
		}
	}

	firebaseUID := middleware.GetFirebaseUID(c)

	user, err := h.repos.Users.GetByFirebaseUID(firebaseUID)
	if err != nil {
		response.GinNotFound(c, "User not found")
		return
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get user tribes, filtering by type if specified
	tribes, err := h.repos.Tribes.GetUserTribes(user.ID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	// If type filter is specified, filter the results
	if tribeType != "" {
		filteredTribes := make([]*models.Tribe, 0)
		for _, tribe := range tribes {
			if tribe.Type == models.TribeType(tribeType) {
				filteredTribes = append(filteredTribes, tribe)
			}
		}
		tribes = filteredTribes
	}

	// Apply pagination to the results (since GetUserTribes doesn't support pagination directly)
	start := offset
	end := offset + limit

	if start >= len(tribes) {
		// Return empty slice if start is beyond the available results
		tribes = []*models.Tribe{}
	} else if end > len(tribes) {
		// Adjust end if it's beyond the available results
		tribes = tribes[start:]
	} else {
		// Apply pagination
		tribes = tribes[start:end]
	}

	// After pagination, set CurrentUserMembershipType for each tribe (should already be set by repo, but ensure it)
	for _, tribe := range tribes {
		for _, member := range tribe.Members {
			if member.UserID == user.ID {
				tribe.CurrentUserMembershipType = member.MembershipType
				break
			}
		}
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

	userID, err := getUserIDFromContext(c)
	if err != nil {
		response.GinUnauthorized(c, "Authentication required")
		return
	}

	isMember := false
	var currentMembership models.MembershipType
	var currentMember *models.TribeMember
	for _, member := range tribe.Members {
		if member.UserID == userID {
			isMember = true
			currentMembership = member.MembershipType
			currentMember = member
			tribe.CurrentUserMembershipType = member.MembershipType
			break
		}
	}

	if !isMember {
		response.GinNotFound(c, "Tribe not found")
		return
	}

	if currentMembership == models.MembershipPending {
		// Only return minimal info and a pending_invitation flag
		minimal := map[string]interface{}{
			"id":                           tribe.ID,
			"name":                         tribe.Name,
			"type":                         tribe.Type,
			"pending_invitation":           true,
			"current_user_membership_type": tribe.CurrentUserMembershipType,
		}

		// Include invitation information if available
		if currentMember != nil {
			if currentMember.InvitedAt != nil {
				minimal["invited_at"] = currentMember.InvitedAt
			}

			if currentMember.InvitedBy != nil {
				minimal["invited_by"] = currentMember.InvitedBy

				// Try to get the inviter's name
				inviter, err := h.repos.Users.GetByID(*currentMember.InvitedBy)
				if err == nil && inviter != nil {
					minimal["inviter"] = map[string]interface{}{
						"id":   inviter.ID,
						"name": inviter.Name,
					}
				}
			}
		}

		response.GinSuccess(c, minimal)
		return
	}

	response.GinSuccess(c, tribe)
}

// UpdateTribeRequest represents the update tribe request body
type UpdateTribeRequest struct {
	Name        string                `json:"name" binding:"required"`
	Type        models.TribeType      `json:"type"`
	Description string                `json:"description"`
	Visibility  models.VisibilityType `json:"visibility"`
	Metadata    interface{}           `json:"metadata,omitempty"`
	Version     int                   `json:"version" binding:"required"`
}

// UpdateTribe updates a tribe's details
func (h *TribeHandler) UpdateTribe(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	var req UpdateTribeRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.GinBadRequest(c, fmt.Sprintf("Invalid request body: %v", bindErr))
		return
	}

	// Get the existing tribe
	tribe, err := h.repos.Tribes.GetByID(id)
	if err != nil {
		response.GinNotFound(c, "Tribe not found")
		return
	}

	// Get current user ID
	var userID uuid.UUID

	// Try to get user ID directly first
	if userIDVal, exists := c.Get("user_id"); exists {
		switch id := userIDVal.(type) {
		case uuid.UUID:
			userID = id
		case string:
			userID, err = uuid.Parse(id)
			if err != nil {
				response.GinInternalError(c, fmt.Errorf("failed to parse user ID"))
				return
			}
		}
	}

	// If user ID is not set directly, get it from Firebase UID
	if userID == uuid.Nil {
		firebaseUID := middleware.GetFirebaseUID(c)
		user, err := h.repos.Users.GetByFirebaseUID(firebaseUID)
		if err != nil {
			response.GinNotFound(c, "User not found")
			return
		}
		userID = user.ID
	}

	// Check if user is a member of the tribe
	isMember := false
	for _, member := range tribe.Members {
		if member.UserID == userID {
			isMember = true
			// Only full members can update the tribe
			if member.MembershipType != models.MembershipFull {
				response.GinForbidden(c, "Only full members can update tribe details")
				return
			}
			break
		}
	}

	if !isMember {
		response.GinForbidden(c, "You must be a member of the tribe to update it")
		return
	}

	// Check version for optimistic concurrency control
	if tribe.Version != req.Version {
		response.GinConflict(c, "Tribe has been modified since you last retrieved it")
		return
	}

	// Update fields
	tribe.Name = req.Name

	// Only update optional fields if they're provided
	if req.Type != "" {
		tribe.Type = req.Type
	}

	tribe.Description = req.Description

	if req.Visibility != "" {
		tribe.Visibility = req.Visibility
	}

	// Handle metadata if provided
	if req.Metadata != nil {
		var metadata models.JSONMap
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

		tribe.Metadata = metadata
	}

	// Validate the updated tribe
	if err := tribe.Validate(); err != nil {
		response.GinBadRequest(c, err.Error())
		return
	}

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

// AddMemberRequest represents the request to add a member to a tribe
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Force  bool      `json:"force,omitempty"` // Force reinvitation if user was previously a member
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

	// Get the current user as the inviter using the getUserIDFromContext helper
	inviterID, err := getUserIDFromContext(c)
	if err != nil {
		response.GinUnauthorized(c, "Authentication required")
		return
	}

	// Use inviterID directly
	inviter := &inviterID

	// Verify that the user exists
	if _, err := h.repos.Users.GetByID(req.UserID); err != nil {
		response.GinBadRequest(c, "User not found")
		return
	}

	// Check if user is already a member of the tribe
	members, err := h.repos.Tribes.GetMembers(tribeID)
	if err != nil {
		if err.Error() == "tribe not found" {
			response.GinNotFound(c, "Tribe not found")
			return
		}
		response.GinInternalError(c, err)
		return
	}

	for _, member := range members {
		if member.UserID == req.UserID {
			response.GinBadRequest(c, "User is already a member of this tribe")
			return
		}
	}

	// Check if the user was previously a member but has been removed
	wasFormerMember, err := h.repos.Tribes.CheckFormerTribeMember(tribeID, req.UserID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	// Handle former members
	if wasFormerMember {
		if !req.Force {
			// If user was a former member and force flag is not set, return a special error
			response.GinError(c, 409, "former_member", "User was previously a member of this tribe. Set force=true to reinvite them.")
			return
		}

		// If force flag is set, reinvite the former member
		if err := h.repos.Tribes.ReinviteMember(tribeID, req.UserID, models.MembershipPending, nil, inviter); err != nil {
			response.GinInternalError(c, err)
			return
		}
	} else {
		// Add the user as a pending member
		if err := h.repos.Tribes.AddMember(tribeID, req.UserID, models.MembershipPending, nil, inviter); err != nil {
			response.GinInternalError(c, err)
			return
		}
	}

	response.GinNoContent(c)
}

// RemoveMember removes a user from a tribe
func (h *TribeHandler) RemoveMember(c *gin.Context) {
	tribeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fmt.Printf("[DEBUG] RemoveMember: Invalid tribe ID: %v\n", c.Param("id"))
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		fmt.Printf("[DEBUG] RemoveMember: Invalid user ID: %v\n", c.Param("userID"))
		response.GinBadRequest(c, "Invalid user ID")
		return
	}

	fmt.Printf("[DEBUG] RemoveMember: tribeID=%s, userID=%s\n", tribeID, userID)

	// Check if tribe exists before attempting removal
	_, tribeErr := h.repos.Tribes.GetByID(tribeID)
	if tribeErr != nil {
		fmt.Printf("[DEBUG] RemoveMember: Tribe existence error: %v\n", tribeErr)
		if tribeErr.Error() == "tribe not found" {
			response.GinNotFound(c, "Tribe not found")
			return
		}
		response.GinInternalError(c, tribeErr)
		return
	}

	err = h.repos.Tribes.RemoveMember(tribeID, userID)
	if err != nil {
		fmt.Printf("[DEBUG] RemoveMember: RemoveMember error: %v\n", err)
		if err.Error() == "tribe member not found" {
			response.GinBadRequest(c, "Tribe member not found")
			return
		}
		response.GinInternalError(c, err)
		return
	}

	fmt.Printf("[DEBUG] RemoveMember: Successfully removed user %s from tribe %s\n", userID, tribeID)
	response.GinNoContent(c)
}

// ListMembers returns all members of a tribe
func (h *TribeHandler) ListMembers(c *gin.Context) {
	tribeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	// Check if the request includes the include_users query parameter
	includeUsers := c.Query("include_users") == "true"

	// Get the members
	members, err := h.repos.Tribes.GetMembers(tribeID)
	if err != nil {
		if err.Error() == "tribe not found" {
			response.GinNotFound(c, "Tribe not found")
			return
		}
		response.GinInternalError(c, err)
		return
	}

	// If include_users flag is not set, remove sensitive user data
	if !includeUsers {
		// Create a response with limited user info
		for i := range members {
			if members[i].User != nil {
				// Create a copy with minimal information
				members[i].User = &models.User{
					ID:        members[i].User.ID,
					Name:      members[i].User.Name,
					AvatarURL: members[i].User.AvatarURL,
				}
			}
		}
	}

	response.GinSuccess(c, members)
}

// InvitationResponse represents the response to a tribe invitation
type InvitationResponse struct {
	Action string `json:"action" binding:"required"` // "accept" or "reject"
}

// RespondToInvitation handles a user's response to a tribe invitation
func (h *TribeHandler) RespondToInvitation(c *gin.Context) {
	tribeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.GinBadRequest(c, "Invalid tribe ID")
		return
	}

	var req InvitationResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GinBadRequest(c, "Invalid request body")
		return
	}

	// Validate the action
	if req.Action != "accept" && req.Action != "reject" {
		response.GinBadRequest(c, "Action must be either 'accept' or 'reject'")
		return
	}

	// Get current user
	firebaseUID := middleware.GetFirebaseUID(c)
	user, err := h.repos.Users.GetByFirebaseUID(firebaseUID)
	if err != nil {
		response.GinNotFound(c, "User not found")
		return
	}

	// Check if user is a pending member of the tribe
	members, err := h.repos.Tribes.GetMembers(tribeID)
	if err != nil {
		response.GinInternalError(c, err)
		return
	}

	var isMember bool
	var isPending bool

	for _, member := range members {
		if member.UserID == user.ID {
			isMember = true
			isPending = member.MembershipType == models.MembershipPending
			break
		}
	}

	if !isMember {
		response.GinForbidden(c, "You are not a member of this tribe")
		return
	}

	if !isPending {
		response.GinBadRequest(c, "Your membership is not in pending state")
		return
	}

	if req.Action == "accept" {
		// Update membership type to full
		err = h.repos.Tribes.UpdateMember(tribeID, user.ID, models.MembershipFull, nil)
		if err != nil {
			response.GinInternalError(c, err)
			return
		}
		response.GinSuccess(c, gin.H{"message": "Invitation accepted"})
	} else {
		// Remove user from tribe
		err = h.repos.Tribes.RemoveMember(tribeID, user.ID)
		if err != nil {
			response.GinInternalError(c, err)
			return
		}
		response.GinSuccess(c, gin.H{"message": "Invitation rejected"})
	}
}
