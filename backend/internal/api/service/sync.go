package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// SyncService handles list synchronization operations
type SyncService struct {
	listRepo models.ListRepository
}

// NewSyncService creates a new sync service
func NewSyncService(listRepo models.ListRepository) *SyncService {
	return &SyncService{
		listRepo: listRepo,
	}
}

// ConfigureSync sets up sync configuration for a list
func (s *SyncService) ConfigureSync(ctx context.Context, listID uuid.UUID, source models.SyncSource, syncID string) error {
	list, err := s.listRepo.GetByID(listID)
	if err != nil {
		if err == models.ErrNotFound {
			return fmt.Errorf("list not found: %w", err)
		}
		return fmt.Errorf("failed to get list: %w", err)
	}

	// Check if sync is already enabled
	if list.SyncStatus != models.ListSyncStatusNone {
		return fmt.Errorf("cannot configure sync: %w", models.ErrSyncAlreadyEnabled)
	}

	// Create new sync config
	config := &models.SyncConfig{
		Source: source,
		ID:     syncID,
		Status: models.ListSyncStatusPending,
	}

	// Validate the config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("%w: %v", models.ErrInvalidSyncConfig, err)
	}

	// Validate the state transition
	if err := models.ValidateTransition(list.SyncStatus, config.Status, "configure_sync"); err != nil {
		return fmt.Errorf("%w: %v", models.ErrInvalidSyncTransition, err)
	}

	// Update list with new sync config
	list.SyncConfig = config
	list.SyncStatus = config.Status
	list.SyncSource = config.Source
	list.SyncID = config.ID
	list.LastSyncAt = nil

	if err := list.Validate(); err != nil {
		return fmt.Errorf("%w: %v", models.ErrInvalidSyncConfig, err)
	}

	if err := s.listRepo.Update(list); err != nil {
		return fmt.Errorf("failed to update list: %w", err)
	}

	return nil
}

// DisableSync removes sync configuration from a list
func (s *SyncService) DisableSync(ctx context.Context, listID uuid.UUID) error {
	list, err := s.listRepo.GetByID(listID)
	if err != nil {
		if err == models.ErrNotFound {
			return fmt.Errorf("list not found: %w", err)
		}
		return fmt.Errorf("failed to get list: %w", err)
	}

	// Check if sync is already disabled
	if list.SyncStatus == models.ListSyncStatusNone {
		return fmt.Errorf("sync already disabled: %w", models.ErrSyncDisabled)
	}

	// Validate the state transition
	if err := models.ValidateTransition(list.SyncStatus, models.ListSyncStatusNone, "disable_sync"); err != nil {
		return fmt.Errorf("%w: %v", models.ErrInvalidSyncTransition, err)
	}

	// Clear sync configuration
	list.SyncConfig = nil
	list.SyncStatus = models.ListSyncStatusNone
	list.SyncSource = models.SyncSourceNone
	list.SyncID = ""
	list.LastSyncAt = nil

	if err := list.Validate(); err != nil {
		return fmt.Errorf("%w: %v", models.ErrInvalidSyncConfig, err)
	}

	if err := s.listRepo.Update(list); err != nil {
		return fmt.Errorf("failed to update list: %w", err)
	}

	return nil
}

// UpdateSyncStatus transitions a list to a new sync status
func (s *SyncService) UpdateSyncStatus(ctx context.Context, listID uuid.UUID, newStatus models.ListSyncStatus, action string) error {
	list, err := s.listRepo.GetByID(listID)
	if err != nil {
		if err == models.ErrNotFound {
			return fmt.Errorf("list not found: %w", err)
		}
		return fmt.Errorf("failed to get list: %w", err)
	}

	// Check if sync is enabled
	if list.SyncSource == "" {
		return fmt.Errorf("cannot update sync status: %w", models.ErrSyncDisabled)
	}

	// Validate the state transition
	if err := models.ValidateTransition(list.SyncStatus, newStatus, action); err != nil {
		return fmt.Errorf("%w: %s", models.ErrInvalidSyncTransition, err.Error())
	}

	// Update sync status
	if list.SyncConfig == nil {
		list.SyncConfig = &models.SyncConfig{
			Source: list.SyncSource,
			ID:     list.SyncID,
			Status: newStatus,
		}
	}
	list.SyncConfig.Status = newStatus
	list.SyncStatus = newStatus

	// Update last sync time if transitioning to synced
	if newStatus == models.ListSyncStatusSynced {
		now := time.Now()
		list.SyncConfig.LastSyncAt = &now
		list.LastSyncAt = &now
	}

	if err := list.Validate(); err != nil {
		return fmt.Errorf("%w: %v", models.ErrInvalidSyncConfig, err)
	}

	err = s.listRepo.Update(list)
	if err != nil {
		if err == models.ErrConcurrentModification {
			// Get the latest state and retry the operation
			list, err = s.listRepo.GetByID(listID)
			if err != nil {
				return fmt.Errorf("failed to get latest list state: %w", err)
			}
			// Validate the transition with the latest state
			if err := models.ValidateTransition(list.SyncStatus, newStatus, action); err != nil {
				return fmt.Errorf("%w: %s", models.ErrInvalidSyncTransition, err.Error())
			}
			// Update the list with the new status
			list.SyncStatus = newStatus
			if list.SyncConfig == nil {
				list.SyncConfig = &models.SyncConfig{
					Source: list.SyncSource,
					ID:     list.SyncID,
					Status: newStatus,
				}
			}
			list.SyncConfig.Status = newStatus
			if newStatus == models.ListSyncStatusSynced {
				now := time.Now()
				list.SyncConfig.LastSyncAt = &now
				list.LastSyncAt = &now
			}
			if err := list.Validate(); err != nil {
				return fmt.Errorf("%w: %v", models.ErrInvalidSyncConfig, err)
			}
			err = s.listRepo.Update(list)
			if err != nil {
				return fmt.Errorf("failed to update list after retry: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to update list: %w", err)
	}

	return nil
}

// CreateConflict creates a new sync conflict
func (s *SyncService) CreateConflict(ctx context.Context, conflict *models.SyncConflict) error {
	if err := conflict.Validate(); err != nil {
		return fmt.Errorf("%w: %v", models.ErrInvalidSyncConfig, err)
	}

	// Transition list to conflict state
	if err := s.UpdateSyncStatus(ctx, conflict.ListID, models.ListSyncStatusConflict, "conflict_detected"); err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	if err := s.listRepo.CreateConflict(conflict); err != nil {
		return fmt.Errorf("failed to create conflict: %w", err)
	}

	return nil
}

// ResolveConflict resolves a sync conflict
func (s *SyncService) ResolveConflict(ctx context.Context, conflictID uuid.UUID, resolution string) error {
	// Get the conflict
	conflicts, err := s.listRepo.GetConflicts(conflictID)
	if err != nil {
		return fmt.Errorf("failed to get conflict: %w", err)
	}
	if len(conflicts) == 0 {
		return fmt.Errorf("conflict not found: %w", models.ErrNotFound)
	}
	conflict := conflicts[0]

	// Check if the conflict is already resolved
	if conflict.ResolvedAt != nil {
		return fmt.Errorf("conflict already resolved: %w", models.ErrConflictAlreadyResolved)
	}

	// Try to resolve the conflict
	err = s.listRepo.ResolveConflict(conflictID)
	if err != nil {
		if err == models.ErrConcurrentModification {
			// Check if the conflict was resolved by another operation
			conflicts, err = s.listRepo.GetConflicts(conflictID)
			if err != nil {
				return fmt.Errorf("failed to get latest conflict state: %w", err)
			}
			if len(conflicts) == 0 {
				return fmt.Errorf("conflict not found: %w", models.ErrNotFound)
			}
			conflict = conflicts[0]
			if conflict.ResolvedAt != nil {
				return fmt.Errorf("conflict already resolved: %w", models.ErrConflictAlreadyResolved)
			}
			// Try to resolve again
			err = s.listRepo.ResolveConflict(conflictID)
			if err != nil {
				return fmt.Errorf("failed to resolve conflict after retry: %w", err)
			}
		} else {
			return fmt.Errorf("failed to resolve conflict: %w", err)
		}
	}

	// Update list status to pending for re-sync
	return s.UpdateSyncStatus(ctx, conflict.ListID, models.ListSyncStatusPending, "resolve_conflict")
}

// GetConflicts returns all unresolved conflicts for a list
func (s *SyncService) GetConflicts(ctx context.Context, listID uuid.UUID) ([]*models.SyncConflict, error) {
	conflicts, err := s.listRepo.GetConflicts(listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicts: %w", err)
	}
	return conflicts, nil
}
