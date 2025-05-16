package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/lib/pq"
)

type listRepository struct {
	db *sql.DB
	tm *TransactionManager
}

// NewListRepository creates a new PostgreSQL-backed list repository
func NewListRepository(db *sql.DB) models.ListRepository {
	return &listRepository{
		db: db,
		tm: NewTransactionManager(db),
	}
}

// Create inserts a new list into the database
func (r *listRepository) Create(list *models.List) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Validate list
		if err := list.Validate(); err != nil {
			return err
		}

		// Generate ID if not provided
		if list.ID == uuid.Nil {
			list.ID = uuid.New()
		}

		// Insert list
		query := `
			INSERT INTO lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, $11, $12,
				NOW(), NOW()
			)`

		_, err := tx.Exec(query,
			list.ID, list.Type, list.Name, list.Description, list.Visibility,
			list.SyncStatus, list.SyncSource, list.SyncID, list.LastSyncAt,
			list.DefaultWeight, list.MaxItems, list.CooldownDays,
		)
		if err != nil {
			return fmt.Errorf("error creating list: %w", err)
		}

		// Add owner if provided
		if list.OwnerID != nil && list.OwnerType != nil {
			owner := &models.ListOwner{
				ListID:    list.ID,
				OwnerID:   *list.OwnerID,
				OwnerType: *list.OwnerType,
			}
			if err := owner.Validate(); err != nil {
				return err
			}

			query = `
				INSERT INTO list_owners (
					list_id, owner_id, owner_type,
					created_at, updated_at
				) VALUES ($1, $2, $3, NOW(), NOW())`

			_, err = tx.Exec(query,
				owner.ListID,
				owner.OwnerID,
				owner.OwnerType,
			)
			if err != nil {
				return fmt.Errorf("error adding list owner: %w", err)
			}
		}

		return nil
	})
}

// loadListData loads all related data for a list
func (r *listRepository) loadListData(tx *sql.Tx, list *models.List) error {
	// Load items
	query := `
		SELECT id, list_id, name, description,
			metadata, external_id,
			weight, last_chosen, chosen_count,
			latitude, longitude, address,
			created_at, updated_at, deleted_at
		FROM list_items
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err := tx.Query(query, list.ID)
	if err != nil {
		return fmt.Errorf("error loading list items: %w", err)
	}
	defer rows.Close()

	var items []*models.ListItem
	for rows.Next() {
		item := &models.ListItem{}
		var metadata []byte
		err := rows.Scan(
			&item.ID, &item.ListID, &item.Name, &item.Description,
			&metadata, &item.ExternalID,
			&item.Weight, &item.LastChosen, &item.ChosenCount,
			&item.Latitude, &item.Longitude, &item.Address,
			&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("error scanning list item: %w", err)
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
				return fmt.Errorf("error decoding item metadata: %w", err)
			}
		}

		items = append(items, item)
	}
	list.Items = items

	// Load owners
	query = `
		SELECT list_id, owner_id, owner_type, created_at, deleted_at
		FROM list_owners
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err = tx.Query(query, list.ID)
	if err != nil {
		return fmt.Errorf("error loading list owners: %w", err)
	}
	defer rows.Close()

	var owners []*models.ListOwner
	for rows.Next() {
		owner := &models.ListOwner{}
		err := rows.Scan(
			&owner.ListID,
			&owner.OwnerID,
			&owner.OwnerType,
			&owner.CreatedAt,
			&owner.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("error scanning list owner: %w", err)
		}
		owners = append(owners, owner)
	}
	list.Owners = owners

	// Load shares
	query = `
		SELECT list_id, user_id, tribe_id, created_at, updated_at, expires_at, deleted_at
		FROM list_sharing
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err = tx.Query(query, list.ID)
	if err != nil {
		return fmt.Errorf("error loading list shares: %w", err)
	}
	defer rows.Close()

	var shares []*models.ListShare
	for rows.Next() {
		share := &models.ListShare{}
		err := rows.Scan(
			&share.ListID,
			&share.UserID,
			&share.TribeID,
			&share.CreatedAt,
			&share.UpdatedAt,
			&share.ExpiresAt,
			&share.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("error scanning list share: %w", err)
		}
		shares = append(shares, share)
	}
	list.Shares = shares

	return nil
}

// GetByID retrieves a list by its ID
func (r *listRepository) GetByID(id uuid.UUID) (*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var list *models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT 
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				created_at, updated_at, deleted_at
			FROM lists
			WHERE id = $1 AND deleted_at IS NULL`

		list = &models.List{}
		err := tx.QueryRow(query, id).Scan(
			&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
			&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
			&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
			&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
		)
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("error getting list: %w", err)
		}

		// Load all related data
		if err := r.loadListData(tx, list); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}

// Update updates an existing list in the database
func (r *listRepository) Update(list *models.List) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Validate list
		if err := list.Validate(); err != nil {
			return err
		}

		// Update list
		query := `
			UPDATE lists SET
				type = $1,
				name = $2,
				description = $3,
				visibility = $4,
				sync_status = $5,
				sync_source = $6,
				sync_id = $7,
				last_sync_at = $8,
				default_weight = $9,
				max_items = $10,
				cooldown_days = $11,
				updated_at = NOW()
			WHERE id = $12 AND deleted_at IS NULL`

		result, err := tx.Exec(query,
			list.Type,
			list.Name,
			list.Description,
			list.Visibility,
			list.SyncStatus,
			list.SyncSource,
			list.SyncID,
			list.LastSyncAt,
			list.DefaultWeight,
			list.MaxItems,
			list.CooldownDays,
			list.ID,
		)
		if err != nil {
			return fmt.Errorf("error updating list: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if rows == 0 {
			return models.ErrNotFound
		}

		// Update owner if provided
		if list.OwnerID != nil && list.OwnerType != nil {
			owner := &models.ListOwner{
				ListID:    list.ID,
				OwnerID:   *list.OwnerID,
				OwnerType: *list.OwnerType,
			}
			if err := owner.Validate(); err != nil {
				return err
			}

			// First, soft delete any existing owners
			query = `
				UPDATE list_owners
				SET deleted_at = NOW()
				WHERE list_id = $1 AND deleted_at IS NULL`

			_, err = tx.Exec(query, list.ID)
			if err != nil {
				return fmt.Errorf("error removing old owners: %w", err)
			}

			// Then add the new owner
			query = `
				INSERT INTO list_owners (
					list_id, owner_id, owner_type,
					created_at, updated_at
				) VALUES ($1, $2, $3, NOW(), NOW())`

			_, err = tx.Exec(query,
				owner.ListID,
				owner.OwnerID,
				owner.OwnerType,
			)
			if err != nil {
				return fmt.Errorf("error adding new owner: %w", err)
			}
		}

		return nil
	})
}

// Delete soft-deletes a list and all its related data
func (r *listRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Soft delete list
		query := `
			UPDATE lists
			SET deleted_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL`

		result, err := tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("error deleting list: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if rows == 0 {
			return models.ErrNotFound
		}

		// Soft delete list items
		query = `
			UPDATE list_items
			SET deleted_at = NOW()
			WHERE list_id = $1 AND deleted_at IS NULL`

		_, err = tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("error deleting list items: %w", err)
		}

		// Soft delete list owners
		query = `
			UPDATE list_owners
			SET deleted_at = NOW()
			WHERE list_id = $1 AND deleted_at IS NULL`

		_, err = tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("error deleting list owners: %w", err)
		}

		// Soft delete list shares
		query = `
			UPDATE list_sharing
			SET deleted_at = NOW()
			WHERE list_id = $1 AND deleted_at IS NULL`

		_, err = tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("error deleting list shares: %w", err)
		}

		return nil
	})
}

// List retrieves a paginated list of lists
func (r *listRepository) List(offset, limit int) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT 
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				created_at, updated_at, deleted_at
			FROM lists
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

		rows, err := tx.Query(query, limit, offset)
		if err != nil {
			return fmt.Errorf("error listing lists: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			list := &models.List{}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list: %w", err)
			}

			// Load all related data
			if err := r.loadListData(tx, list); err != nil {
				return err
			}

			lists = append(lists, list)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}

// AddItem adds a new item to a list
func (r *listRepository) AddItem(item *models.ListItem) error {
	query := `
		INSERT INTO list_items (
			id, list_id, name, description,
			metadata, external_id,
			weight, last_chosen, chosen_count,
			latitude, longitude, address,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8, $9,
			$10, $11, $12,
			NOW(), NOW()
		)`

	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}

	metadata, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(query,
		item.ID, item.ListID, item.Name, item.Description,
		metadata, item.ExternalID,
		item.Weight, item.LastChosen, item.ChosenCount,
		item.Latitude, item.Longitude, item.Address,
	)
	return err
}

// UpdateItem updates an existing list item
func (r *listRepository) UpdateItem(item *models.ListItem) error {
	query := `
		UPDATE list_items SET
			name = $1,
			description = $2,
			metadata = $3,
			external_id = $4,
			weight = $5,
			last_chosen = $6,
			chosen_count = $7,
			latitude = $8,
			longitude = $9,
			address = $10,
			updated_at = NOW()
		WHERE id = $11 AND list_id = $12 AND deleted_at IS NULL`

	metadata, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(query,
		item.Name, item.Description,
		metadata, item.ExternalID,
		item.Weight, item.LastChosen, item.ChosenCount,
		item.Latitude, item.Longitude, item.Address,
		item.ID, item.ListID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found: %v", item.ID)
	}

	return nil
}

// RemoveItem soft-deletes an item from a list
func (r *listRepository) RemoveItem(listID, itemID uuid.UUID) error {
	query := `
		UPDATE list_items SET
			deleted_at = NOW()
		WHERE id = $1 AND list_id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, itemID, listID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found: %v", itemID)
	}

	return nil
}

// GetItems retrieves all items in a list
func (r *listRepository) GetItems(listID uuid.UUID) ([]*models.ListItem, error) {
	query := `
		SELECT 
			id, list_id, name, description,
			metadata, external_id,
			weight, last_chosen, chosen_count,
			latitude, longitude, address,
			created_at, updated_at, deleted_at
		FROM list_items
		WHERE list_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.ListItem
	for rows.Next() {
		item := &models.ListItem{}
		var metadata []byte
		err := rows.Scan(
			&item.ID, &item.ListID, &item.Name, &item.Description,
			&metadata, &item.ExternalID,
			&item.Weight, &item.LastChosen, &item.ChosenCount,
			&item.Latitude, &item.Longitude, &item.Address,
			&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
				return nil, err
			}
		}

		items = append(items, item)
	}

	return items, nil
}

// GetEligibleItems retrieves items eligible for menu generation
func (r *listRepository) GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
	query := `
		SELECT 
			i.id, i.list_id, i.name, i.description,
			i.metadata, i.external_id,
			i.weight, i.last_chosen, i.chosen_count,
			i.latitude, i.longitude, i.address,
			i.created_at, i.updated_at, i.deleted_at,
			l.default_weight, l.cooldown_days
		FROM list_items i
		JOIN lists l ON i.list_id = l.id
		WHERE i.list_id = ANY($1)
		AND i.deleted_at IS NULL
		AND l.deleted_at IS NULL`

	// Apply filters
	if cooldown, ok := filters["cooldown_days"].(int); ok {
		query += fmt.Sprintf(" AND (i.last_chosen IS NULL OR i.last_chosen < NOW() - INTERVAL '%d days')", cooldown)
	}
	if maxItems, ok := filters["max_items"].(int); ok {
		query += fmt.Sprintf(" LIMIT %d", maxItems)
	}

	rows, err := r.db.Query(query, pq.Array(listIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.ListItem
	for rows.Next() {
		item := &models.ListItem{}
		var metadata []byte
		var defaultWeight float64
		var cooldownDays *int
		err := rows.Scan(
			&item.ID, &item.ListID, &item.Name, &item.Description,
			&metadata, &item.ExternalID,
			&item.Weight, &item.LastChosen, &item.ChosenCount,
			&item.Latitude, &item.Longitude, &item.Address,
			&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
			&defaultWeight, &cooldownDays,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
				return nil, err
			}
		}

		// Apply list's default weight if item weight is not set
		if item.Weight == 0 {
			item.Weight = defaultWeight
		}

		items = append(items, item)
	}

	return items, nil
}

// UpdateItemStats updates an item's selection statistics
func (r *listRepository) UpdateItemStats(itemID uuid.UUID, chosen bool) error {
	query := `
		UPDATE list_items SET
			last_chosen = CASE WHEN $2 THEN NOW() ELSE last_chosen END,
			chosen_count = chosen_count + CASE WHEN $2 THEN 1 ELSE 0 END,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, itemID, chosen)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found: %v", itemID)
	}

	return nil
}

// UpdateSyncStatus updates the sync status of a list
func (r *listRepository) UpdateSyncStatus(listID uuid.UUID, status models.ListSyncStatus) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Validate status
		if err := status.Validate(); err != nil {
			return err
		}

		query := `
			UPDATE lists
			SET sync_status = $1, updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL`

		result, err := tx.Exec(query, status, listID)
		if err != nil {
			return fmt.Errorf("error updating sync status: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if rows == 0 {
			return models.ErrNotFound
		}

		return nil
	})
}

// GetListsBySource retrieves all lists from a specific sync source
func (r *listRepository) GetListsBySource(source string) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.created_at, l.updated_at, l.deleted_at
			FROM lists l
			WHERE l.sync_source = $1
				AND l.deleted_at IS NULL
			ORDER BY l.created_at DESC`

		rows, err := tx.Query(query, source)
		if err != nil {
			return fmt.Errorf("error getting lists by source: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			list := &models.List{}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list: %w", err)
			}

			// Load all related data
			if err := r.loadListData(tx, list); err != nil {
				return err
			}

			lists = append(lists, list)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}

// CreateConflict creates a new list conflict
func (r *listRepository) CreateConflict(conflict *models.ListConflict) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Validate conflict
		if err := conflict.Validate(); err != nil {
			return err
		}

		// Generate ID if not provided
		if conflict.ID == uuid.Nil {
			conflict.ID = uuid.New()
		}

		// Marshal JSON data
		localData, err := json.Marshal(conflict.LocalData)
		if err != nil {
			return fmt.Errorf("error marshaling local data: %w", err)
		}
		externalData, err := json.Marshal(conflict.ExternalData)
		if err != nil {
			return fmt.Errorf("error marshaling external data: %w", err)
		}

		query := `
			INSERT INTO list_conflicts (
				id, list_id, item_id, conflict_type,
				local_data, external_data,
				created_at, updated_at, version
			) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), 1)`

		_, err = tx.Exec(query,
			conflict.ID,
			conflict.ListID,
			conflict.ItemID,
			conflict.ConflictType,
			localData,
			externalData,
		)
		if err != nil {
			return fmt.Errorf("error creating conflict: %w", err)
		}

		return nil
	})
}

// GetConflicts retrieves all unresolved conflicts for a list
func (r *listRepository) GetConflicts(listID uuid.UUID) ([]*models.ListConflict, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var conflicts []*models.ListConflict

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, list_id, item_id, conflict_type,
				local_data, external_data,
				created_at, updated_at, resolved_at, version
			FROM list_conflicts
			WHERE list_id = $1
				AND resolved_at IS NULL
				AND deleted_at IS NULL
			ORDER BY created_at DESC`

		rows, err := tx.Query(query, listID)
		if err != nil {
			return fmt.Errorf("error getting conflicts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			conflict := &models.ListConflict{}
			var localData, externalData []byte

			err := rows.Scan(
				&conflict.ID,
				&conflict.ListID,
				&conflict.ItemID,
				&conflict.ConflictType,
				&localData,
				&externalData,
				&conflict.CreatedAt,
				&conflict.UpdatedAt,
				&conflict.ResolvedAt,
				&conflict.Version,
			)
			if err != nil {
				return fmt.Errorf("error scanning conflict: %w", err)
			}

			// Parse JSON data
			if err := json.Unmarshal(localData, &conflict.LocalData); err != nil {
				return fmt.Errorf("error parsing local data: %w", err)
			}
			if err := json.Unmarshal(externalData, &conflict.ExternalData); err != nil {
				return fmt.Errorf("error parsing external data: %w", err)
			}

			conflicts = append(conflicts, conflict)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return conflicts, nil
}

// ResolveConflict marks a conflict as resolved
func (r *listRepository) ResolveConflict(conflictID uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE list_conflicts
			SET resolved_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND resolved_at IS NULL`

		result, err := tx.Exec(query, conflictID)
		if err != nil {
			return fmt.Errorf("error resolving conflict: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if rows == 0 {
			return models.ErrNotFound
		}

		return nil
	})
}

// AddOwner adds an owner to a list
func (r *listRepository) AddOwner(owner *models.ListOwner) error {
	query := `
		INSERT INTO list_owners (
			list_id, owner_id, owner_type,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5)`

	now := time.Now()
	_, err := r.db.Exec(query,
		owner.ListID,
		owner.OwnerID,
		owner.OwnerType,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("error adding list owner: %w", err)
	}

	return nil
}

// RemoveOwner removes an owner from a list
func (r *listRepository) RemoveOwner(listID, ownerID uuid.UUID) error {
	query := `
		UPDATE list_owners
		SET deleted_at = NOW()
		WHERE list_id = $1 AND owner_id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, listID, ownerID)
	if err != nil {
		return fmt.Errorf("error removing list owner: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrNotFound
	}

	return nil
}

// GetOwners retrieves all owners of a list
func (r *listRepository) GetOwners(listID uuid.UUID) ([]*models.ListOwner, error) {
	query := `
		SELECT list_id, owner_id, owner_type, created_at, deleted_at
		FROM list_owners
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.Query(query, listID)
	if err != nil {
		return nil, fmt.Errorf("error getting list owners: %w", err)
	}
	defer rows.Close()

	var owners []*models.ListOwner
	for rows.Next() {
		owner := &models.ListOwner{}
		err := rows.Scan(
			&owner.ListID,
			&owner.OwnerID,
			&owner.OwnerType,
			&owner.CreatedAt,
			&owner.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning list owner row: %w", err)
		}
		owners = append(owners, owner)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating list owner rows: %w", err)
	}

	return owners, nil
}

// GetUserLists retrieves all lists owned by a user
func (r *listRepository) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.created_at, l.updated_at, l.deleted_at
			FROM lists l
			JOIN list_owners lo ON l.id = lo.list_id
			WHERE lo.owner_id = $1
				AND lo.owner_type = 'user'
				AND l.deleted_at IS NULL
				AND lo.deleted_at IS NULL
			ORDER BY l.created_at DESC`

		rows, err := tx.Query(query, userID)
		if err != nil {
			return fmt.Errorf("error getting user lists: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			list := &models.List{}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list: %w", err)
			}

			// Load all related data
			if err := r.loadListData(tx, list); err != nil {
				return err
			}

			lists = append(lists, list)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}

// GetTribeLists retrieves all lists owned by a tribe
func (r *listRepository) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.created_at, l.updated_at, l.deleted_at
			FROM lists l
			LEFT JOIN list_owners lo ON l.id = lo.list_id
			LEFT JOIN list_sharing ls ON l.id = ls.list_id
			WHERE (lo.owner_id = $1 AND lo.owner_type = 'tribe' AND lo.deleted_at IS NULL)
				OR (ls.tribe_id = $1 AND ls.deleted_at IS NULL)
				AND l.deleted_at IS NULL
			ORDER BY l.created_at DESC`

		rows, err := tx.Query(query, tribeID)
		if err != nil {
			return fmt.Errorf("error getting tribe lists: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			list := &models.List{}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list: %w", err)
			}

			// Load all related data
			if err := r.loadListData(tx, list); err != nil {
				return err
			}

			lists = append(lists, list)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}

// ShareWithTribe shares a list with a tribe
func (r *listRepository) ShareWithTribe(share *models.ListShare) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Validate share
		if err := share.Validate(); err != nil {
			return err
		}

		// Check if list exists and is not deleted
		query := `
			SELECT 1
			FROM lists
			WHERE id = $1 AND deleted_at IS NULL`

		var exists bool
		err := tx.QueryRow(query, share.ListID).Scan(&exists)
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("error checking list existence: %w", err)
		}

		// Check if tribe exists and is not deleted
		query = `
			SELECT 1
			FROM tribes
			WHERE id = $1 AND deleted_at IS NULL`

		err = tx.QueryRow(query, share.TribeID).Scan(&exists)
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("error checking tribe existence: %w", err)
		}

		// Check if user exists and is not deleted
		query = `
			SELECT 1
			FROM users
			WHERE id = $1`

		err = tx.QueryRow(query, share.UserID).Scan(&exists)
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("error checking user existence: %w", err)
		}

		// First, soft delete any existing shares for this list-tribe combination
		query = `
			UPDATE list_sharing
			SET deleted_at = NOW()
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL`

		_, err = tx.Exec(query, share.ListID, share.TribeID)
		if err != nil {
			return fmt.Errorf("error removing old shares: %w", err)
		}

		// Then add the new share
		query = `
			INSERT INTO list_sharing (
				list_id, tribe_id, user_id,
				created_at, updated_at, expires_at
			) VALUES ($1, $2, $3, NOW(), NOW(), $4)`

		_, err = tx.Exec(query,
			share.ListID,
			share.TribeID,
			share.UserID,
			share.ExpiresAt,
		)
		if err != nil {
			return fmt.Errorf("error sharing list: %w", err)
		}

		return nil
	})
}

// UnshareWithTribe removes a tribe's access to a list
func (r *listRepository) UnshareWithTribe(listID, tribeID uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE list_sharing
			SET deleted_at = NOW()
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL`

		result, err := tx.Exec(query, listID, tribeID)
		if err != nil {
			return fmt.Errorf("error unsharing list: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if rows == 0 {
			return models.ErrNotFound
		}

		return nil
	})
}

// getListShares retrieves all shares for a list (internal helper)
func (r *listRepository) getListShares(listID uuid.UUID) ([]*models.ListShare, error) {
	query := `
		SELECT list_id, tribe_id, user_id, created_at, expires_at, deleted_at
		FROM list_shares
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.Query(query, listID)
	if err != nil {
		return nil, fmt.Errorf("error getting list shares: %w", err)
	}
	defer rows.Close()

	var shares []*models.ListShare
	for rows.Next() {
		share := &models.ListShare{}
		err := rows.Scan(
			&share.ListID,
			&share.TribeID,
			&share.UserID,
			&share.CreatedAt,
			&share.ExpiresAt,
			&share.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning list share row: %w", err)
		}
		shares = append(shares, share)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating list share rows: %w", err)
	}

	return shares, nil
}

// GetSharedLists retrieves all lists shared with a tribe
func (r *listRepository) GetSharedLists(tribeID uuid.UUID) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.created_at, l.updated_at, l.deleted_at
			FROM lists l
			JOIN list_sharing ls ON l.id = ls.list_id
			WHERE ls.tribe_id = $1
				AND ls.deleted_at IS NULL
				AND l.deleted_at IS NULL
				AND (ls.expires_at IS NULL OR ls.expires_at > NOW())
			ORDER BY l.created_at DESC`

		rows, err := tx.Query(query, tribeID)
		if err != nil {
			return fmt.Errorf("error getting shared lists: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			list := &models.List{}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list: %w", err)
			}

			// Load all related data
			if err := r.loadListData(tx, list); err != nil {
				return err
			}

			lists = append(lists, list)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}
