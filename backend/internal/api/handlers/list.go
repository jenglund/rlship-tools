package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/api/response"
	"github.com/jenglund/rlship-tools/internal/api/service"
	"github.com/jenglund/rlship-tools/internal/models"
)

type ListHandler struct {
	service *service.ListService
}

func NewListHandler(service *service.ListService) *ListHandler {
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

	lists, err := h.service.ListLists(offset, limit)
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
	var req struct {
		ListIDs []uuid.UUID            `json:"list_ids"`
		Filters map[string]interface{} `json:"filters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.ListIDs) == 0 {
		response.Error(w, http.StatusBadRequest, "At least one list ID is required")
		return
	}

	items, err := h.service.GenerateMenu(req.ListIDs, req.Filters)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// SyncList handles synchronizing a list with its external source
func (h *ListHandler) SyncList(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	if err := h.service.SyncList(listID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.NoContent(w)
}

// GetListConflicts handles retrieving all unresolved conflicts for a list
func (h *ListHandler) GetListConflicts(w http.ResponseWriter, r *http.Request) {
	listID, err := uuid.Parse(chi.URLParam(r, "listID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	conflicts, err := h.service.GetListConflicts(listID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, conflicts)
}

// ResolveListConflict handles resolving a sync conflict
func (h *ListHandler) ResolveListConflict(w http.ResponseWriter, r *http.Request) {
	conflictID, err := uuid.Parse(chi.URLParam(r, "conflictID"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid conflict ID")
		return
	}

	if err := h.service.ResolveListConflict(conflictID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.NoContent(w)
}
