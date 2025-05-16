package service

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

type ListService struct {
	repo models.ListRepository
}

func NewListService(repo models.ListRepository) *ListService {
	return &ListService{repo: repo}
}

// CreateList creates a new list
func (s *ListService) CreateList(list *models.List) error {
	return s.repo.Create(list)
}

// GetList retrieves a list by ID
func (s *ListService) GetList(id uuid.UUID) (*models.List, error) {
	return s.repo.GetByID(id)
}

// UpdateList updates an existing list
func (s *ListService) UpdateList(list *models.List) error {
	return s.repo.Update(list)
}

// DeleteList deletes a list
func (s *ListService) DeleteList(id uuid.UUID) error {
	return s.repo.Delete(id)
}

// ListLists retrieves a paginated list of lists
func (s *ListService) ListLists(offset, limit int) ([]*models.List, error) {
	return s.repo.List(offset, limit)
}

// AddListItem adds an item to a list
func (s *ListService) AddListItem(item *models.ListItem) error {
	return s.repo.AddItem(item)
}

// UpdateListItem updates a list item
func (s *ListService) UpdateListItem(item *models.ListItem) error {
	return s.repo.UpdateItem(item)
}

// RemoveListItem removes an item from a list
func (s *ListService) RemoveListItem(listID, itemID uuid.UUID) error {
	return s.repo.RemoveItem(listID, itemID)
}

// GetListItems retrieves all items in a list
func (s *ListService) GetListItems(listID uuid.UUID) ([]*models.ListItem, error) {
	return s.repo.GetItems(listID)
}

// GenerateMenu generates a menu from multiple lists based on weights and filters
func (s *ListService) GenerateMenu(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
	// Get eligible items
	items, err := s.repo.GetEligibleItems(listIDs, filters)
	if err != nil {
		return nil, err
	}

	// If maxItems is specified in filters, use it, otherwise return all items
	maxItems := len(items)
	if max, ok := filters["max_items"].(int); ok && max < maxItems {
		maxItems = max
	}

	// Calculate total weight
	var totalWeight float64
	for _, item := range items {
		totalWeight += item.Weight
	}

	// Randomly select items based on weights
	selected := make([]*models.ListItem, 0, maxItems)
	remainingItems := make([]*models.ListItem, len(items))
	copy(remainingItems, items)

	for len(selected) < maxItems && len(remainingItems) > 0 {
		// Calculate current total weight
		currentTotal := 0.0
		for _, item := range remainingItems {
			currentTotal += item.Weight
		}

		if currentTotal <= 0 {
			break
		}

		// Generate random number between 0 and total weight
		r := rand.Float64() * currentTotal

		// Find the selected item
		var selectedIndex int
		var runningTotal float64

		for i, item := range remainingItems {
			runningTotal += item.Weight
			if runningTotal > r {
				selectedIndex = i
				break
			}
		}

		// Add selected item to result and remove from remaining items
		selected = append(selected, remainingItems[selectedIndex])
		remainingItems = append(remainingItems[:selectedIndex], remainingItems[selectedIndex+1:]...)
	}

	// Update selection stats for chosen items
	for _, item := range selected {
		if err := s.repo.UpdateItemStats(item.ID, true); err != nil {
			return nil, fmt.Errorf("failed to update item stats: %v", err)
		}
	}

	return selected, nil
}

// SyncList synchronizes a list with its external source
func (s *ListService) SyncList(listID uuid.UUID) error {
	list, err := s.repo.GetByID(listID)
	if err != nil {
		return err
	}

	// Check if list is configured for syncing
	if list.SyncSource == "" {
		return fmt.Errorf("list %v is not configured for syncing", listID)
	}

	// Update sync status to pending
	if err := s.repo.UpdateSyncStatus(listID, models.ListSyncStatusPending); err != nil {
		return err
	}

	// TODO: Implement external source sync (Google Maps API integration)
	// This will be implemented in a separate PR

	// For now, just mark as synced
	return s.repo.UpdateSyncStatus(listID, models.ListSyncStatusSynced)
}

// GetListConflicts retrieves all unresolved conflicts for a list
func (s *ListService) GetListConflicts(listID uuid.UUID) ([]*models.ListConflict, error) {
	return s.repo.GetConflicts(listID)
}

// ResolveListConflict resolves a sync conflict
func (s *ListService) ResolveListConflict(conflictID uuid.UUID) error {
	return s.repo.ResolveConflict(conflictID)
}

// CreateListConflict creates a new sync conflict
func (s *ListService) CreateListConflict(conflict *models.ListConflict) error {
	return s.repo.CreateConflict(conflict)
}
