package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// ListService defines the interface for list operations
type ListService interface {
	// List management
	CreateList(list *models.List) error
	GetList(id uuid.UUID) (*models.List, error)
	UpdateList(list *models.List) error
	DeleteList(id uuid.UUID) error
	List(offset, limit int) ([]*models.List, error)

	// List items
	AddListItem(item *models.ListItem) error
	GetListItems(listID uuid.UUID) ([]*models.ListItem, error)
	UpdateListItem(item *models.ListItem) error
	RemoveListItem(listID, itemID uuid.UUID) error

	// Menu generation
	GenerateMenu(params *models.MenuParams) ([]*models.List, error)

	// Sync management
	SyncList(listID uuid.UUID) error
	GetListConflicts(listID uuid.UUID) ([]*models.SyncConflict, error)
	ResolveListConflict(listID, conflictID uuid.UUID, resolution string) error

	// Owner management
	AddListOwner(listID, ownerID uuid.UUID, ownerType string) error
	RemoveListOwner(listID, ownerID uuid.UUID) error
	GetListOwners(listID uuid.UUID) ([]*models.ListOwner, error)
	GetUserLists(userID uuid.UUID) ([]*models.List, error)
	GetTribeLists(tribeID uuid.UUID) ([]*models.List, error)

	// Share management
	ShareListWithTribe(listID, tribeID, userID uuid.UUID, expiresAt *time.Time) error
	UnshareListWithTribe(listID, tribeID, userID uuid.UUID) error
	GetSharedLists(tribeID uuid.UUID) ([]*models.List, error)

	// GetListShares retrieves all shares for a list
	GetListShares(listID uuid.UUID) ([]*models.ListShare, error)
}

// listService implements the ListService interface
type listService struct {
	repo models.ListRepository
}

// NewListService creates a new list service
func NewListService(repo models.ListRepository) ListService {
	return &listService{repo: repo}
}

// CreateList creates a new list
func (s *listService) CreateList(list *models.List) error {
	// Validate list
	if err := list.Validate(); err != nil {
		return err
	}

	// Create list
	if err := s.repo.Create(list); err != nil {
		return err
	}

	return nil
}

// GetList retrieves a list by ID
func (s *listService) GetList(id uuid.UUID) (*models.List, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}

	list, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// UpdateList updates an existing list
func (s *listService) UpdateList(list *models.List) error {
	// Validate list
	if err := list.Validate(); err != nil {
		return err
	}

	// Update list
	if err := s.repo.Update(list); err != nil {
		return err
	}

	return nil
}

// DeleteList deletes a list
func (s *listService) DeleteList(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("error deleting list: %w", err)
	}

	return nil
}

// List retrieves a paginated list of lists
func (s *listService) List(offset, limit int) ([]*models.List, error) {
	if offset < 0 {
		return nil, fmt.Errorf("%w: offset cannot be negative", models.ErrInvalidInput)
	}
	if limit <= 0 {
		return nil, fmt.Errorf("%w: limit must be positive", models.ErrInvalidInput)
	}

	lists, err := s.repo.List(offset, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing lists: %w", err)
	}

	return lists, nil
}

// AddListItem adds an item to a list
func (s *listService) AddListItem(item *models.ListItem) error {
	// Validate item
	if item.ListID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}
	if item.Name == "" {
		return fmt.Errorf("%w: item name is required", models.ErrInvalidInput)
	}

	// Verify list exists
	if _, err := s.repo.GetByID(item.ListID); err != nil {
		return fmt.Errorf("error verifying list exists: %w", err)
	}

	// Add item
	if err := s.repo.AddItem(item); err != nil {
		return fmt.Errorf("error adding list item: %w", err)
	}

	return nil
}

// UpdateListItem updates a list item
func (s *listService) UpdateListItem(item *models.ListItem) error {
	// Validate item
	if item.ID == uuid.Nil {
		return fmt.Errorf("%w: item ID is required", models.ErrInvalidInput)
	}
	if item.ListID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}
	if item.Name == "" {
		return fmt.Errorf("%w: item name is required", models.ErrInvalidInput)
	}

	// Update item
	if err := s.repo.UpdateItem(item); err != nil {
		return fmt.Errorf("error updating list item: %w", err)
	}

	return nil
}

// RemoveListItem removes an item from a list
func (s *listService) RemoveListItem(listID, itemID uuid.UUID) error {
	if listID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}
	if itemID == uuid.Nil {
		return fmt.Errorf("%w: item ID is required", models.ErrInvalidInput)
	}

	if err := s.repo.RemoveItem(listID, itemID); err != nil {
		return fmt.Errorf("error removing list item: %w", err)
	}

	return nil
}

// GetListItems retrieves all items in a list
func (s *listService) GetListItems(listID uuid.UUID) ([]*models.ListItem, error) {
	if listID == uuid.Nil {
		return nil, fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}

	items, err := s.repo.GetItems(listID)
	if err != nil {
		return nil, fmt.Errorf("error getting list items: %w", err)
	}

	return items, nil
}

// GenerateMenu generates a menu from multiple lists based on weights and filters
func (s *listService) GenerateMenu(params *models.MenuParams) ([]*models.List, error) {
	// Get lists
	lists := make([]*models.List, 0, len(params.ListIDs))
	for _, listID := range params.ListIDs {
		list, err := s.repo.GetByID(listID)
		if err != nil {
			return nil, fmt.Errorf("error getting list %s: %w", listID, err)
		}
		lists = append(lists, list)
	}

	// Get eligible items for each list
	for _, list := range lists {
		items, err := s.repo.GetEligibleItems([]uuid.UUID{list.ID}, params.Filters)
		if err != nil {
			return nil, fmt.Errorf("error getting eligible items for list %s: %w", list.ID, err)
		}

		// Apply weights and cooldown
		for _, item := range items {
			weight := list.DefaultWeight
			if item.Weight > 0 {
				weight = item.Weight
			}

			// Apply cooldown if specified
			if list.CooldownDays != nil && item.LastUsed != nil {
				cooldownEnds := item.LastUsed.Add(time.Duration(*list.CooldownDays) * 24 * time.Hour)
				if time.Now().Before(cooldownEnds) {
					weight = 0 // Item is in cooldown, set weight to 0
				}
			}

			item.Weight = weight
		}

		// If maxItems is specified in filters, use it, otherwise return all items
		maxItems := len(items)
		if max, ok := params.Filters["max_items"].(int); ok && max < maxItems {
			maxItems = max
		}

		// Sort items by weight and select top maxItems
		sort.Slice(items, func(i, j int) bool {
			return items[i].Weight > items[j].Weight
		})
		if len(items) > maxItems {
			items = items[:maxItems]
		}

		// Update the list's items
		list.Items = items
	}

	return lists, nil
}

// SyncList synchronizes a list with its external source
func (s *listService) SyncList(listID uuid.UUID) error {
	// First try
	list, err := s.repo.GetByID(listID)
	if err != nil {
		return err
	}

	// Check if list is configured for syncing
	if list.SyncSource == "" {
		return models.ErrInvalidInput
	}

	// Update sync status to pending
	err = s.repo.UpdateSyncStatus(listID, models.ListSyncStatusPending)
	if err != nil {
		// Handle concurrent modification by retrying once
		if err == models.ErrConcurrentModification {
			// Retry with fresh data
			list, err = s.repo.GetByID(listID)
			if err != nil {
				return err
			}

			// Try again with the updated list
			err = s.repo.UpdateSyncStatus(listID, models.ListSyncStatusPending)
			if err != nil {
				return err
			}
		} else {
			// For any other error, return immediately
			return err
		}
	}

	// TODO: Implement external source sync (Google Maps API integration)
	// This will be implemented in a separate PR

	// For now, just mark as synced
	return s.repo.UpdateSyncStatus(listID, models.ListSyncStatusSynced)
}

// GetListConflicts retrieves all unresolved conflicts for a list
func (s *listService) GetListConflicts(listID uuid.UUID) ([]*models.SyncConflict, error) {
	return s.repo.GetConflicts(listID)
}

// ResolveListConflict resolves a sync conflict
func (s *listService) ResolveListConflict(listID, conflictID uuid.UUID, resolution string) error {
	return s.repo.ResolveConflict(conflictID)
}

// CreateListConflict creates a new sync conflict
func (s *listService) CreateListConflict(conflict *models.SyncConflict) error {
	return s.repo.CreateConflict(conflict)
}

// AddListOwner adds an owner to a list
func (s *listService) AddListOwner(listID, ownerID uuid.UUID, ownerType string) error {
	owner := &models.ListOwner{
		ListID:    listID,
		OwnerID:   ownerID,
		OwnerType: models.OwnerType(ownerType),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := owner.Validate(); err != nil {
		return fmt.Errorf("invalid owner: %w", err)
	}

	return s.repo.AddOwner(owner)
}

// RemoveListOwner removes an owner from a list
func (s *listService) RemoveListOwner(listID, ownerID uuid.UUID) error {
	return s.repo.RemoveOwner(listID, ownerID)
}

// GetListOwners retrieves all owners of a list
func (s *listService) GetListOwners(listID uuid.UUID) ([]*models.ListOwner, error) {
	return s.repo.GetOwners(listID)
}

// GetUserLists retrieves all lists owned by a user
func (s *listService) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
	return s.repo.GetUserLists(userID)
}

// GetTribeLists retrieves all lists owned by a tribe
func (s *listService) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	return s.repo.GetTribeLists(tribeID)
}

// ShareListWithTribe shares a list with a tribe
func (s *listService) ShareListWithTribe(listID, tribeID, userID uuid.UUID, expiresAt *time.Time) error {
	// Check if list exists
	list, err := s.repo.GetByID(listID)
	if err != nil {
		if err.Error() == "list not found" {
			return models.ErrNotFound
		}
		return err
	}

	// Check if user is an owner of the list
	owners, err := s.repo.GetOwners(listID)
	if err != nil {
		return err
	}

	hasPermission := false
	for _, owner := range owners {
		if owner.OwnerID == userID && owner.OwnerType == "user" {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return models.ErrForbidden
	}

	// Update visibility if needed
	if list.Visibility == models.VisibilityPrivate {
		list.Visibility = models.VisibilityShared
		if err := s.repo.Update(list); err != nil {
			return err
		}
	}

	// Create a ListShare object
	share := &models.ListShare{
		ListID:    listID,
		TribeID:   tribeID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	return s.repo.ShareWithTribe(share)
}

// UnshareListWithTribe removes a list share from a tribe
func (s *listService) UnshareListWithTribe(listID, tribeID, userID uuid.UUID) error {
	// Check if list exists
	_, err := s.repo.GetByID(listID)
	if err != nil {
		if err.Error() == "list not found" {
			return models.ErrNotFound
		}
		return err
	}

	// Check if user is an owner of the list
	owners, err := s.repo.GetOwners(listID)
	if err != nil {
		return err
	}

	hasPermission := false
	for _, owner := range owners {
		if owner.OwnerID == userID && owner.OwnerType == "user" {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return models.ErrForbidden
	}

	return s.repo.UnshareWithTribe(listID, tribeID)
}

// GetSharedLists retrieves all lists shared with a tribe
func (s *listService) GetSharedLists(tribeID uuid.UUID) ([]*models.List, error) {
	return s.repo.GetSharedLists(tribeID)
}

// GetListShares retrieves all shares for a list
func (s *listService) GetListShares(listID uuid.UUID) ([]*models.ListShare, error) {
	// Verify list exists
	_, err := s.repo.GetByID(listID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetListShares(listID)
}
