package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
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
	CreateListConflict(conflict *models.SyncConflict) error

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

	// CleanupExpiredShares removes expired shares
	CleanupExpiredShares() error
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
	// Trim whitespace from name and description
	list.Name = strings.TrimSpace(list.Name)
	list.Description = strings.TrimSpace(list.Description)

	// Validate the list
	if err := list.Validate(); err != nil {
		return err
	}

	// Additional validation
	if list.Name == "" {
		return fmt.Errorf("%w: list name cannot be empty", models.ErrInvalidInput)
	}

	// Set creation timestamp
	now := time.Now()
	list.CreatedAt = now
	list.UpdatedAt = now

	// If owner information is provided, add the owner after creating the list
	hasOwner := list.OwnerID != nil && list.OwnerType != nil

	// Check if a list with the same name already exists for this owner
	if hasOwner {
		// Search for lists with the same name for this owner
		existingLists, err := s.repo.GetListsByOwner(*list.OwnerID, *list.OwnerType)
		if err != nil {
			return fmt.Errorf("error checking for existing lists: %w", err)
		}

		// Normalize the list name for comparison
		normalizedName := strings.ToLower(strings.TrimSpace(list.Name))
		for _, existing := range existingLists {
			existingName := strings.ToLower(strings.TrimSpace(existing.Name))
			if existingName == normalizedName {
				return models.ErrDuplicate
			}
		}
	}

	// Create the list - note that the repository's Create method already adds the primary owner
	// to the list_owners table, so we don't need to add it separately
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
			_, err = s.repo.GetByID(listID)
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
	// The resolution parameter might need to be used to determine how to resolve the conflict
	// For now, we're just passing conflictID to the repository method
	return s.repo.ResolveConflict(conflictID)
}

// CreateListConflict creates a new sync conflict
func (s *listService) CreateListConflict(conflict *models.SyncConflict) error {
	// Add basic validation to handle nil conflicts
	if conflict == nil {
		return fmt.Errorf("%w: conflict cannot be nil", models.ErrInvalidInput)
	}

	// Delegate to the repository for detailed validation and creation
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
	// Validate input parameters
	if listID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}
	if tribeID == uuid.Nil {
		return fmt.Errorf("%w: tribe ID is required", models.ErrInvalidInput)
	}
	if userID == uuid.Nil {
		return fmt.Errorf("%w: user ID is required", models.ErrInvalidInput)
	}
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return fmt.Errorf("%w: expiration date must be in the future", models.ErrInvalidInput)
	}

	// Check if list exists
	list, err := s.repo.GetByID(listID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return fmt.Errorf("%w: list with ID %s not found", models.ErrNotFound, listID)
		}
		return fmt.Errorf("failed to get list: %w", err)
	}

	// Check if user can share this list
	canShare, err := s.canUserShareList(userID, list)
	if err != nil {
		return fmt.Errorf("failed to check if user can share list: %w", err)
	}
	if !canShare {
		return fmt.Errorf("%w: user %s is not authorized to share list %s", models.ErrUnauthorized, userID, listID)
	}

	// Since we don't have a tribeRepo field, we'll just verify the tribe exists by checking if sharing fails
	// Create share record
	share := &models.ListShare{
		ListID:    listID,
		TribeID:   tribeID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Share the list with the tribe
	if err := s.repo.ShareWithTribe(share); err != nil {
		return fmt.Errorf("failed to share list with tribe: %w", err)
	}

	return nil
}

// Helper method to check if a user can share a list
func (s *listService) canUserShareList(userID uuid.UUID, list *models.List) (bool, error) {
	// If user is the owner, they can share
	if list.OwnerType != nil && *list.OwnerType == models.OwnerTypeUser &&
		list.OwnerID != nil && *list.OwnerID == userID {
		return true, nil
	}

	// If list is owned by a tribe, check if user is a member with appropriate permissions
	if list.OwnerType != nil && *list.OwnerType == models.OwnerTypeTribe && list.OwnerID != nil {
		// Since we don't have direct access to tribe membership info,
		// we'll check if the user is in the owners list instead
		for _, owner := range list.Owners {
			if owner.OwnerType == models.OwnerTypeUser && owner.OwnerID == userID {
				return true, nil
			}
		}
	}

	// For other owner types, or if not the owner, check if user is an additional owner
	for _, owner := range list.Owners {
		if owner.OwnerType == models.OwnerTypeUser && owner.OwnerID == userID {
			return true, nil
		}
	}

	return false, nil
}

// UnshareListWithTribe removes a list share from a tribe
func (s *listService) UnshareListWithTribe(listID, tribeID, userID uuid.UUID) error {
	// Validate input parameters
	if listID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}
	if tribeID == uuid.Nil {
		return fmt.Errorf("%w: tribe ID is required", models.ErrInvalidInput)
	}
	if userID == uuid.Nil {
		return fmt.Errorf("%w: user ID is required", models.ErrInvalidInput)
	}

	// Check if list exists
	_, err := s.repo.GetByID(listID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("%w: list with ID %s not found", models.ErrNotFound, listID)
		}
		return fmt.Errorf("failed to get list: %w", err)
	}

	// Check if user is an owner of the list
	owners, err := s.repo.GetOwners(listID)
	if err != nil {
		return fmt.Errorf("failed to get list owners: %w", err)
	}

	hasPermission := false
	for _, owner := range owners {
		if owner.OwnerID == userID && owner.OwnerType == "user" {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return fmt.Errorf("%w: you do not have permission to unshare this list", models.ErrForbidden)
	}

	// Check if share exists by getting list shares
	shares, err := s.repo.GetListShares(listID)
	if err != nil {
		return fmt.Errorf("failed to check if list is shared with tribe: %w", err)
	}

	shareExists := false
	for _, share := range shares {
		if share.TribeID == tribeID {
			shareExists = true
			break
		}
	}

	if !shareExists {
		return fmt.Errorf("%w: list is not shared with the specified tribe", models.ErrNotFound)
	}

	// Remove the share
	if err := s.repo.UnshareWithTribe(listID, tribeID); err != nil {
		return fmt.Errorf("failed to unshare list from tribe: %w", err)
	}

	return nil
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

// CleanupExpiredShares removes expired shares
func (s *listService) CleanupExpiredShares() error {
	ctx := context.Background()
	_, err := s.repo.CleanupExpiredShares(ctx)
	if err != nil {
		return fmt.Errorf("error cleaning up expired shares: %w", err)
	}
	return nil
}
