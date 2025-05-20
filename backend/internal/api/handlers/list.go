package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/api/response"
	"github.com/jenglund/rlship-tools/internal/api/service"
	"github.com/jenglund/rlship-tools/internal/models"
)

// Define a custom key type to avoid string key collisions
type contextKey string

const (
	// userIDKey is the context key for the user ID
	userIDKey contextKey = "user_id"
)

type ListHandler struct {
	service service.ListService
}

func NewListHandler(service service.ListService) *ListHandler {
	return &ListHandler{service: service}
}

// RegisterRoutes registers the list management routes
func (h *ListHandler) RegisterRoutes(r chi.Router) {
	r.Route("/lists", func(r chi.Router) {
		r.Post("/", h.CreateList)
		r.Get("/", h.ListLists)
		r.Get("/{listID}", h.GetList)
		r.Put("/{listID}", h.UpdateList)
		r.Delete("/{listID}", h.DeleteList)

		// List items
		r.Post("/{listID}/items", h.AddListItem)
		r.Get("/{listID}/items", h.GetListItems)
		r.Put("/{listID}/items/{itemID}", h.UpdateListItem)
		r.Delete("/{listID}/items/{itemID}", h.RemoveListItem)

		// Menu generation
		r.Post("/menu", h.GenerateMenu)

		// Sync management
		r.Post("/{listID}/sync", h.SyncList)
		r.Get("/{listID}/conflicts", h.GetListConflicts)
		r.Post("/{listID}/conflicts/{conflictID}/resolve", h.ResolveListConflict)

		// Owner management
		r.Post("/{listID}/owners", h.AddListOwner)
		r.Delete("/{listID}/owners/{ownerID}", h.RemoveListOwner)
		r.Get("/{listID}/owners", h.GetListOwners)
		r.Get("/user/{userID}", h.GetUserLists)
		r.Get("/tribe/{tribeID}", h.GetTribeLists)

		// Sharing management
		r.Post("/{listID}/share", h.ShareList)
		r.Delete("/{listID}/share/{tribeID}", h.UnshareList)
		r.Get("/shared/{tribeID}", h.GetSharedLists)

		// New handlers
		r.Get("/{id}/shares", h.GetListShares)
		r.Post("/{id}/share/{tribeID}", h.ShareListWithTribe)
		r.Delete("/{id}/share/{tribeID}", h.UnshareListWithTribe)
	})
}

// CreateList handles the creation of a new list
func (h *ListHandler) CreateList(w http.ResponseWriter, r *http.Request) {
	var list models.List
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.CreateList(&list); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, list)
}

// ListLists handles retrieving a paginated list of lists
func (h *ListHandler) ListLists(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	lists, err := h.service.List(offset, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, lists)
}

// GetList handles retrieving a single list by ID
func (h *ListHandler) GetList(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	list, err := h.service.GetList(listID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, list)
}

// UpdateList handles updating an existing list
func (h *ListHandler) UpdateList(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	var list models.List
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	list.ID = listID
	if err := h.service.UpdateList(&list); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, list)
}

// DeleteList handles deleting a list
func (h *ListHandler) DeleteList(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	if err := h.service.DeleteList(listID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.NoContent(w)
}

// AddListItem handles adding a new item to a list
func (h *ListHandler) AddListItem(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	var item models.ListItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item.ListID = listID
	if err := h.service.AddListItem(&item); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, item)
}

// GetListItems handles retrieving all items in a list
func (h *ListHandler) GetListItems(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	items, err := h.service.GetListItems(listID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// UpdateListItem handles updating an existing list item
func (h *ListHandler) UpdateListItem(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var item models.ListItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item.ID = itemID
	item.ListID = listID
	if err := h.service.UpdateListItem(&item); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, item)
}

// RemoveListItem handles removing an item from a list
func (h *ListHandler) RemoveListItem(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	if err := h.service.RemoveListItem(listID, itemID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.NoContent(w)
}

// GenerateMenu handles generating a menu from multiple lists
func (h *ListHandler) GenerateMenu(w http.ResponseWriter, r *http.Request) {
	var params models.MenuParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	lists, err := h.service.GenerateMenu(&params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, lists)
}

// SyncList handles syncing a list with its external source
func (h *ListHandler) SyncList(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID format")
		return
	}

	if err := h.service.SyncList(listID); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			response.Error(w, http.StatusNotFound, "List not found")
		case errors.Is(err, models.ErrSyncDisabled):
			response.Error(w, http.StatusBadRequest, "Sync is not enabled for this list")
		case errors.Is(err, models.ErrExternalSourceUnavailable):
			response.Error(w, http.StatusServiceUnavailable, "External sync source is unavailable")
		case errors.Is(err, models.ErrExternalSourceTimeout):
			response.Error(w, http.StatusGatewayTimeout, "External sync source timed out")
		case errors.Is(err, models.ErrExternalSourceError):
			response.Error(w, http.StatusBadGateway, "External sync source error")
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to sync list")
		}
		return
	}

	response.NoContent(w)
}

// GetListConflicts handles retrieving all unresolved conflicts for a list
func (h *ListHandler) GetListConflicts(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID format")
		return
	}

	conflicts, err := h.service.GetListConflicts(listID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			response.Error(w, http.StatusNotFound, "List not found")
		case errors.Is(err, models.ErrSyncDisabled):
			response.Error(w, http.StatusBadRequest, "Sync is not enabled for this list")
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to get list conflicts")
		}
		return
	}

	response.JSON(w, http.StatusOK, conflicts)
}

// ResolveListConflict handles resolving a list sync conflict
func (h *ListHandler) ResolveListConflict(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID format")
		return
	}

	conflictID, err := uuid.Parse(chi.URLParam(r, "conflictID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid conflict ID format")
		return
	}

	var resolution struct {
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&resolution); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.ResolveListConflict(listID, conflictID, resolution.Resolution); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			response.Error(w, http.StatusNotFound, "List not found")
		case errors.Is(err, models.ErrConflictNotFound):
			response.Error(w, http.StatusNotFound, "Conflict not found")
		case errors.Is(err, models.ErrConflictAlreadyResolved):
			response.Error(w, http.StatusConflict, "Conflict already resolved")
		case errors.Is(err, models.ErrInvalidResolution):
			response.Error(w, http.StatusBadRequest, "Invalid resolution")
		case errors.Is(err, models.ErrSyncDisabled):
			response.Error(w, http.StatusBadRequest, "Sync is not enabled for this list")
		default:
			response.Error(w, http.StatusInternalServerError, "Failed to resolve conflict")
		}
		return
	}

	response.NoContent(w)
}

// AddListOwner handles adding a new owner to a list
func (h *ListHandler) AddListOwner(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	var req struct {
		OwnerID   uuid.UUID `json:"owner_id"`
		OwnerType string    `json:"owner_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.AddListOwner(listID, req.OwnerID, req.OwnerType); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.NoContent(w)
}

// RemoveListOwner handles removing an owner from a list
func (h *ListHandler) RemoveListOwner(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	ownerID, err := uuid.Parse(chi.URLParam(r, "ownerID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid owner ID")
		return
	}

	if err := h.service.RemoveListOwner(listID, ownerID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.NoContent(w)
}

// GetListOwners handles retrieving all owners of a list
func (h *ListHandler) GetListOwners(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	owners, err := h.service.GetListOwners(listID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, owners)
}

// GetUserLists handles retrieving all lists owned by a user
func (h *ListHandler) GetUserLists(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	lists, err := h.service.GetUserLists(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, lists)
}

// GetTribeLists handles retrieving all lists owned by a tribe
func (h *ListHandler) GetTribeLists(w http.ResponseWriter, r *http.Request) {
	tribeID, err := uuid.Parse(chi.URLParam(r, "tribeID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid tribe ID")
		return
	}

	lists, err := h.service.GetTribeLists(tribeID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, lists)
}

// ShareList handles sharing a list with a tribe
func (h *ListHandler) ShareList(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	var req struct {
		TribeID   uuid.UUID  `json:"tribe_id"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
	}

	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the user ID from the authenticated context
	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "invalid user authentication")
		return
	}

	// Check if the user is an owner of the list
	owners, err := h.service.GetListOwners(listID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	isOwner := false
	for _, owner := range owners {
		if owner.OwnerID == userID && owner.OwnerType == "user" {
			isOwner = true
			break
		}
	}

	if !isOwner {
		response.Error(w, http.StatusForbidden, "you do not have permission to share this list")
		return
	}

	err = h.service.ShareListWithTribe(listID, req.TribeID, userID, req.ExpiresAt)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// UnshareList handles removing a list share from a tribe
func (h *ListHandler) UnshareList(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	tribeID, err := uuid.Parse(chi.URLParam(r, "tribeID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid tribe ID")
		return
	}

	// Get the user ID from the authenticated context
	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "invalid user authentication")
		return
	}

	if err := h.service.UnshareListWithTribe(listID, tribeID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// GetSharedLists handles retrieving all lists shared with a tribe
func (h *ListHandler) GetSharedLists(w http.ResponseWriter, r *http.Request) {
	tribeID, err := uuid.Parse(chi.URLParam(r, "tribeID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid tribe ID")
		return
	}

	lists, err := h.service.GetSharedLists(tribeID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, lists)
}

// GetListShares handles retrieving all shares for a list
func (h *ListHandler) GetListShares(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	// Get the user ID from the authenticated context
	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "invalid user authentication")
		return
	}

	// Check if the user has permission to view the list shares
	owners, err := h.service.GetListOwners(listID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	hasPermission := false
	for _, owner := range owners {
		if owner.OwnerID == userID && owner.OwnerType == "user" {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		response.Error(w, http.StatusForbidden, "you do not have permission to view this list's shares")
		return
	}

	shares, err := h.service.GetListShares(listID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, shares)
}

// handleError converts service errors to appropriate HTTP responses
func (h *ListHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrNotFound):
		response.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, models.ErrInvalidInput):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrUnauthorized):
		response.Error(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, models.ErrForbidden):
		response.Error(w, http.StatusForbidden, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, err.Error())
	}
}

// ShareListWithTribe handles the request to share a list with a tribe
func (h *ListHandler) ShareListWithTribe(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	tribeID, err := uuid.Parse(chi.URLParam(r, "tribeID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid tribe ID")
		return
	}

	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "invalid user authentication")
		return
	}

	var req struct {
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil && decodeErr != io.EOF {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.ShareListWithTribe(listID, tribeID, userID, req.ExpiresAt)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnshareListWithTribe handles the request to unshare a list from a tribe
func (h *ListHandler) UnshareListWithTribe(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid list ID")
		return
	}

	tribeID, err := uuid.Parse(chi.URLParam(r, "tribeID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid tribe ID")
		return
	}

	userIDValue := r.Context().Value(userIDKey)
	if userIDValue == nil {
		response.Error(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "invalid user authentication")
		return
	}

	err = h.service.UnshareListWithTribe(listID, tribeID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
