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

type ListRepository struct {
	db *sql.DB
	tm *TransactionManager
}

// NewListRepository creates a new PostgreSQL-backed list repository
func NewListRepository(db *sql.DB) models.ListRepository {
	return &ListRepository{
		db: db,
		tm: NewTransactionManager(db),
	}
}

// Create inserts a new list into the database
func (r *ListRepository) Create(list *models.List) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	opts.IsolationLevel = sql.LevelSerializable // Ensure consistency for list creation

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Generate ID if not provided
		if list.ID == uuid.Nil {
			list.ID = uuid.New()
		}

		// Validate list
		if err := list.Validate(); err != nil {
			return fmt.Errorf("list validation failed: %w", err)
		}

		// Validate and set owner fields
		if err := r.validateAndSetOwners(list); err != nil {
			return fmt.Errorf("owner validation failed: %w", err)
		}

		// Check for duplicate list name for the same owner
		var count int
		err := tx.QueryRow(`
			SELECT COUNT(*)
			FROM lists
			WHERE name = $1
				AND deleted_at IS NULL
				AND owner_id = $2
				AND owner_type = $3`,
			list.Name, *list.OwnerID, *list.OwnerType).Scan(&count)
		if err != nil {
			return fmt.Errorf("error checking for duplicate list: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("%w: list with name %q already exists for this owner", models.ErrDuplicate, list.Name)
		}

		// Insert list
		query := `
			INSERT INTO lists (
				id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				owner_id, owner_type,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, $11, $12,
				$13, $14,
				NOW(), NOW()
			) RETURNING created_at, updated_at`

		err = tx.QueryRow(query,
			list.ID, list.Type, list.Name, list.Description, list.Visibility,
			list.SyncStatus, list.SyncSource, list.SyncID, list.LastSyncAt,
			list.DefaultWeight, list.MaxItems, list.CooldownDays,
			*list.OwnerID, *list.OwnerType,
		).Scan(&list.CreatedAt, &list.UpdatedAt)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code.Name() {
				case "unique_violation":
					return fmt.Errorf("%w: list with this ID already exists", models.ErrDuplicate)
				case "foreign_key_violation":
					return fmt.Errorf("%w: referenced owner does not exist", models.ErrNotFound)
				}
			}
			return fmt.Errorf("error creating list: %w", err)
		}

		// Now add the primary owner to the list_owners table
		_, err = tx.Exec(`
			INSERT INTO list_owners (
				list_id, owner_id, owner_type,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5)`,
			list.ID, *list.OwnerID, *list.OwnerType, list.CreatedAt, list.UpdatedAt)
		if err != nil {
			return fmt.Errorf("error adding primary owner: %w", err)
		}

		// Add additional owners if provided
		if len(list.Owners) > 1 {
			// Prepare statement for better performance with multiple inserts
			stmt, err := tx.Prepare(`
				INSERT INTO list_owners (
					list_id, owner_id, owner_type,
					created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5)`)
			if err != nil {
				return fmt.Errorf("error preparing owner insert statement: %w", err)
			}
			defer safeClose(stmt)

			for _, owner := range list.Owners {
				// Skip the primary owner as it's already handled above
				if owner.OwnerID == *list.OwnerID && owner.OwnerType == *list.OwnerType {
					continue
				}

				owner.ListID = list.ID
				owner.CreatedAt = list.CreatedAt
				owner.UpdatedAt = list.UpdatedAt

				_, err = stmt.Exec(
					owner.ListID,
					owner.OwnerID,
					owner.OwnerType,
					owner.CreatedAt,
					owner.UpdatedAt,
				)
				if err != nil {
					if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
						continue // Skip duplicate owners
					}
					return fmt.Errorf("error adding list owner: %w", err)
				}
			}
		}

		return nil
	})
}

// validateAndSetOwners validates and sets the owner fields for a list
func (r *ListRepository) validateAndSetOwners(list *models.List) error {
	if list.OwnerID == nil || list.OwnerType == nil {
		if len(list.Owners) == 0 {
			return fmt.Errorf("at least one owner is required")
		}

		// Use the first owner as the primary owner
		list.OwnerID = &list.Owners[0].OwnerID
		list.OwnerType = &list.Owners[0].OwnerType
	}

	// Validate owner type
	if err := (*list.OwnerType).Validate(); err != nil {
		return fmt.Errorf("invalid owner type: %w", err)
	}

	// Ensure owner exists in owners list
	ownerExists := false
	for _, owner := range list.Owners {
		if owner.OwnerID == *list.OwnerID && owner.OwnerType == *list.OwnerType {
			ownerExists = true
			break
		}
	}

	if !ownerExists {
		// Add primary owner to owners list if not present
		list.Owners = append(list.Owners, &models.ListOwner{
			OwnerID:   *list.OwnerID,
			OwnerType: *list.OwnerType,
		})
	}

	return nil
}

// loadListData loads all related data for a list
func (r *ListRepository) loadListData(tx *sql.Tx, list *models.List) error {
	// Load items
	itemsQuery := `
		SELECT id, list_id, name, description,
			metadata, external_id,
			weight, last_chosen, chosen_count,
			latitude, longitude, address,
			created_at, updated_at, deleted_at
		FROM list_items
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err := tx.Query(itemsQuery, list.ID)
	if err != nil {
		return fmt.Errorf("error loading list items: %w", err)
	}
	defer safeClose(rows)

	var items []*models.ListItem
	for rows.Next() {
		item := &models.ListItem{}
		var metadata []byte
		if err := rows.Scan(
			&item.ID, &item.ListID, &item.Name, &item.Description,
			&metadata, &item.ExternalID,
			&item.Weight, &item.LastChosen, &item.ChosenCount,
			&item.Latitude, &item.Longitude, &item.Address,
			&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
		); err != nil {
			return fmt.Errorf("error scanning list item: %w", err)
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
				return fmt.Errorf("error decoding item metadata: %w", err)
			}
		} else {
			item.Metadata = make(models.JSONMap)
		}

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating list items: %w", err)
	}
	list.Items = items

	// Load owners
	ownersQuery := `
		SELECT list_id, owner_id, owner_type, created_at, deleted_at
		FROM list_owners
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err = tx.Query(ownersQuery, list.ID)
	if err != nil {
		return fmt.Errorf("error loading list owners: %w", err)
	}
	defer safeClose(rows)

	var owners []*models.ListOwner
	for rows.Next() {
		owner := &models.ListOwner{}
		if err := rows.Scan(
			&owner.ListID,
			&owner.OwnerID,
			&owner.OwnerType,
			&owner.CreatedAt,
			&owner.DeletedAt,
		); err != nil {
			return fmt.Errorf("error scanning list owner: %w", err)
		}
		owners = append(owners, owner)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating list owners: %w", err)
	}
	list.Owners = owners

	// Load shares
	sharesQuery := `
		SELECT list_id, tribe_id, expires_at, created_at, updated_at, deleted_at
		FROM list_sharing
		WHERE list_id = $1 AND deleted_at IS NULL`

	rows, err = tx.Query(sharesQuery, list.ID)
	if err != nil {
		return fmt.Errorf("error loading list shares: %w", err)
	}
	defer safeClose(rows)

	var shares []*models.ListShare
	for rows.Next() {
		share := &models.ListShare{}
		err := rows.Scan(
			&share.ListID,
			&share.TribeID,
			&share.ExpiresAt,
			&share.CreatedAt,
			&share.UpdatedAt,
			&share.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("error scanning list share: %w", err)
		}
		shares = append(shares, share)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating list shares: %w", err)
	}
	list.Shares = shares

	return nil
}

// GetByID retrieves a list by its ID
func (r *ListRepository) GetByID(id uuid.UUID) (*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var list *models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				owner_id, owner_type,
				created_at, updated_at
			FROM lists
			WHERE id = $1 AND deleted_at IS NULL`

		list = &models.List{}
		err := tx.QueryRow(query, id).Scan(
			&list.ID,
			&list.Type,
			&list.Name,
			&list.Description,
			&list.Visibility,
			&list.SyncStatus,
			&list.SyncSource,
			&list.SyncID,
			&list.LastSyncAt,
			&list.DefaultWeight,
			&list.MaxItems,
			&list.CooldownDays,
			&list.OwnerID,
			&list.OwnerType,
			&list.CreatedAt,
			&list.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrNotFound
			}
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
func (r *ListRepository) Update(list *models.List) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First check if the list exists to avoid non-existent list updates
		var exists bool
		err := tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM lists 
				WHERE id = $1 AND deleted_at IS NULL
			)`, list.ID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking if list exists: %w", err)
		}
		if !exists {
			return models.ErrNotFound
		}

		// Get existing list to preserve owner fields and other data if not provided
		query := `
			SELECT id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				owner_id, owner_type,
				created_at, updated_at
			FROM lists
			WHERE id = $1 AND deleted_at IS NULL`

		existingList := &models.List{}
		err = tx.QueryRow(query, list.ID).Scan(
			&existingList.ID,
			&existingList.Type,
			&existingList.Name,
			&existingList.Description,
			&existingList.Visibility,
			&existingList.SyncStatus,
			&existingList.SyncSource,
			&existingList.SyncID,
			&existingList.LastSyncAt,
			&existingList.DefaultWeight,
			&existingList.MaxItems,
			&existingList.CooldownDays,
			&existingList.OwnerID,
			&existingList.OwnerType,
			&existingList.CreatedAt,
			&existingList.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrNotFound
			}
			return fmt.Errorf("error getting existing list: %w", err)
		}

		// Preserve fields that weren't provided in the update
		if list.Type == "" {
			list.Type = existingList.Type
		}
		if list.Name == "" {
			list.Name = existingList.Name
		}
		if list.Description == "" {
			list.Description = existingList.Description
		}
		if list.Visibility == "" {
			list.Visibility = existingList.Visibility
		}
		if list.SyncStatus == "" {
			list.SyncStatus = existingList.SyncStatus
		}
		if list.SyncSource == "" {
			list.SyncSource = existingList.SyncSource
		}
		if list.SyncID == "" {
			list.SyncID = existingList.SyncID
		}
		if list.LastSyncAt == nil {
			list.LastSyncAt = existingList.LastSyncAt
		}
		if list.DefaultWeight == 0 {
			list.DefaultWeight = existingList.DefaultWeight
		}
		if list.MaxItems == nil {
			list.MaxItems = existingList.MaxItems
		}
		if list.CooldownDays == nil {
			list.CooldownDays = existingList.CooldownDays
		}

		// Handle owner fields
		if list.OwnerID == nil || list.OwnerType == nil {
			list.OwnerID = existingList.OwnerID
			list.OwnerType = existingList.OwnerType
		}

		// Ensure owners list contains at least one owner (the primary owner)
		if len(list.Owners) == 0 {
			list.Owners = append(list.Owners, &models.ListOwner{
				ListID:    list.ID,
				OwnerID:   *list.OwnerID,
				OwnerType: *list.OwnerType,
				CreatedAt: existingList.CreatedAt,
				UpdatedAt: time.Now(),
			})
		}

		// Validate list
		if validateErr := list.Validate(); err != nil {
			return fmt.Errorf("list validation failed: %w", validateErr)
		}

		// Update list
		updateQuery := `
			UPDATE lists
			SET name = $1,
				description = $2,
				type = $3,
				visibility = $4,
				sync_status = $5,
				sync_source = $6,
				sync_id = $7,
				last_sync_at = $8,
				default_weight = $9,
				max_items = $10,
				cooldown_days = $11,
				updated_at = NOW(),
				owner_id = $12,
				owner_type = $13
			WHERE id = $14 AND deleted_at IS NULL
			RETURNING updated_at`

		err = tx.QueryRow(updateQuery,
			list.Name,
			list.Description,
			list.Type,
			list.Visibility,
			list.SyncStatus,
			list.SyncSource,
			list.SyncID,
			list.LastSyncAt,
			list.DefaultWeight,
			list.MaxItems,
			list.CooldownDays,
			*list.OwnerID,
			*list.OwnerType,
			list.ID,
		).Scan(&list.UpdatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrNotFound
			}
			return fmt.Errorf("error updating list: %w", err)
		}

		return nil
	})
}

// Delete soft-deletes a list and all its related data
func (r *ListRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	opts.IsolationLevel = sql.LevelSerializable // Ensure consistency for deletion

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First check if the list exists
		var exists bool
		err := tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM lists 
				WHERE id = $1 AND deleted_at IS NULL
			)`, id).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking if list exists: %w", err)
		}
		if !exists {
			return models.ErrNotFound
		}

		now := time.Now()

		// Soft delete list
		_, err = tx.Exec(`
			UPDATE lists
			SET deleted_at = $1
			WHERE id = $2 AND deleted_at IS NULL`,
			now, id)
		if err != nil {
			return fmt.Errorf("error deleting list: %w", err)
		}

		// Soft delete list items
		_, err = tx.Exec(`
			UPDATE list_items
			SET deleted_at = $1
			WHERE list_id = $2 AND deleted_at IS NULL`,
			now, id)
		if err != nil {
			return fmt.Errorf("error deleting list items: %w", err)
		}

		// Soft delete list owners
		_, err = tx.Exec(`
			UPDATE list_owners
			SET deleted_at = $1
			WHERE list_id = $2 AND deleted_at IS NULL`,
			now, id)
		if err != nil {
			return fmt.Errorf("error deleting list owners: %w", err)
		}

		// Soft delete list shares
		_, err = tx.Exec(`
			UPDATE list_sharing
			SET deleted_at = $1
			WHERE list_id = $2 AND deleted_at IS NULL`,
			now, id)
		if err != nil {
			return fmt.Errorf("error deleting list shares: %w", err)
		}

		// Check if list_conflicts table exists before attempting to delete from it
		var tableExists bool
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'list_conflicts'
			)`).Scan(&tableExists)
		if err != nil {
			// Log the error but continue, as this isn't critical
			fmt.Printf("Error checking if list_conflicts table exists: %v\n", err)
		} else if tableExists {
			// Only try to delete from list_conflicts if the table exists
			_, err = tx.Exec(`
				UPDATE list_conflicts
				SET deleted_at = $1
				WHERE list_id = $2 AND deleted_at IS NULL`,
				now, id)
			if err != nil {
				// Log the error but don't fail the transaction
				fmt.Printf("Error deleting list conflicts: %v\n", err)
			}
		}

		return nil
	})
}

// List retrieves a paginated list of lists
func (r *ListRepository) List(offset, limit int) ([]*models.List, error) {
	// Check if limit is 0, which means no results should be returned
	if limit == 0 {
		return []*models.List{}, nil
	}

	// First, fetch the basic list details
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

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing lists: %w", err)
	}
	defer safeClose(rows)

	var lists []*models.List
	for rows.Next() {
		list := &models.List{}
		err := rows.Scan(
			&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
			&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
			&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
			&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning list: %w", err)
		}

		lists = append(lists, list)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lists: %w", err)
	}

	// If there are no lists, return early
	if len(lists) == 0 {
		return lists, nil
	}

	// For each list, load items, owners, and shares
	for _, list := range lists {
		// Load items
		items, err := r.GetItems(list.ID)
		if err != nil {
			return nil, fmt.Errorf("error loading items for list %s: %w", list.ID, err)
		}
		list.Items = items

		// Load owners
		owners, err := r.GetOwners(list.ID)
		if err != nil {
			return nil, fmt.Errorf("error loading owners for list %s: %w", list.ID, err)
		}
		list.Owners = owners

		// Load shares
		shares, err := r.GetListShares(list.ID)
		if err != nil {
			return nil, fmt.Errorf("error loading shares for list %s: %w", list.ID, err)
		}
		list.Shares = shares
	}

	return lists, nil
}

// AddItem adds a new item to a list
func (r *ListRepository) AddItem(item *models.ListItem) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Generate ID if not provided
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}

		// Initialize metadata if nil
		if item.Metadata == nil {
			item.Metadata = make(map[string]interface{})
		}

		metadata, err := json.Marshal(item.Metadata)
		if err != nil {
			return fmt.Errorf("error marshaling metadata: %w", err)
		}

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
			) RETURNING created_at, updated_at`

		err = tx.QueryRow(query,
			item.ID, item.ListID, item.Name, item.Description,
			metadata, item.ExternalID,
			item.Weight, item.LastChosen, item.ChosenCount,
			item.Latitude, item.Longitude, item.Address,
		).Scan(&item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return fmt.Errorf("error adding list item: %w", err)
		}

		return nil
	})
}

// UpdateItem updates an existing list item
func (r *ListRepository) UpdateItem(item *models.ListItem) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		metadata, err := json.Marshal(item.Metadata)
		if err != nil {
			return fmt.Errorf("error marshaling metadata: %w", err)
		}

		item.UpdatedAt = time.Now()

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
				updated_at = $11
			WHERE id = $12 AND list_id = $13 AND deleted_at IS NULL`

		result, err := tx.Exec(query,
			item.Name, item.Description,
			metadata, item.ExternalID,
			item.Weight, item.LastChosen, item.ChosenCount,
			item.Latitude, item.Longitude, item.Address,
			item.UpdatedAt,
			item.ID, item.ListID,
		)
		if err != nil {
			return fmt.Errorf("error updating list item: %w", err)
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

// RemoveItem soft-deletes an item from a list
func (r *ListRepository) RemoveItem(listID, itemID uuid.UUID) error {
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

// GetItems retrieves all items from a list
func (r *ListRepository) GetItems(listID uuid.UUID) ([]*models.ListItem, error) {
	// Create a transaction to ensure schema search path is correctly set
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var items []*models.ListItem

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First confirm the table exists in this schema
		var currentSchema string
		err := tx.QueryRow(`SELECT current_schema()`).Scan(&currentSchema)
		if err != nil {
			return fmt.Errorf("error getting current schema: %w", err)
		}

		var tableExists bool
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = $1
				AND table_name = 'list_items'
			)`, currentSchema).Scan(&tableExists)

		if err != nil {
			return fmt.Errorf("error checking if list_items table exists: %w", err)
		}

		if !tableExists {
			return fmt.Errorf("list_items table does not exist in schema %s", currentSchema)
		}

		// Now get the items with fully qualified table name
		query := fmt.Sprintf(`
			SELECT 
				id, list_id, name, description,
				metadata, external_id,
				weight, last_chosen, chosen_count,
				latitude, longitude, address,
				created_at, updated_at, deleted_at
			FROM %s.list_items
			WHERE list_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC`, currentSchema)

		rows, err := tx.Query(query, listID)
		if err != nil {
			return err
		}
		defer safeClose(rows)

		for rows.Next() {
			item := &models.ListItem{}
			var metadata []byte
			if err := rows.Scan(
				&item.ID, &item.ListID, &item.Name, &item.Description,
				&metadata, &item.ExternalID,
				&item.Weight, &item.LastChosen, &item.ChosenCount,
				&item.Latitude, &item.Longitude, &item.Address,
				&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
			); err != nil {
				return err
			}

			if len(metadata) > 0 {
				if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
					return err
				}
			}

			items = append(items, item)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetEligibleItems retrieves items eligible for menu generation
func (r *ListRepository) GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
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
	defer safeClose(rows)

	var items []*models.ListItem
	for rows.Next() {
		item := &models.ListItem{}
		var metadata []byte
		var defaultWeight float64
		var cooldownDays *int
		if err := rows.Scan(
			&item.ID, &item.ListID, &item.Name, &item.Description,
			&metadata, &item.ExternalID,
			&item.Weight, &item.LastChosen, &item.ChosenCount,
			&item.Latitude, &item.Longitude, &item.Address,
			&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
			&defaultWeight, &cooldownDays,
		); err != nil {
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

// UpdateItemStats updates the statistics for a list item
func (r *ListRepository) UpdateItemStats(itemID uuid.UUID, chosen bool) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE list_items SET
				last_chosen = CASE WHEN $2 THEN NOW() ELSE last_chosen END,
				chosen_count = chosen_count + CASE WHEN $2 THEN 1 ELSE 0 END,
				updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL`

		result, err := tx.Exec(query, itemID, chosen)
		if err != nil {
			return fmt.Errorf("error updating item stats: %w", err)
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

// UpdateSyncStatus updates the sync status of a list
func (r *ListRepository) UpdateSyncStatus(listID uuid.UUID, status models.ListSyncStatus) error {
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
func (r *ListRepository) GetListsBySource(source string) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Verify current schema and tables
		var currentSchema string
		err := tx.QueryRow(`SELECT current_schema()`).Scan(&currentSchema)
		if err != nil {
			return fmt.Errorf("error getting current schema: %w", err)
		}

		// Check if the lists table exists in the current schema
		var tableExists bool
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = $1
				AND table_name = 'lists'
			)`, currentSchema).Scan(&tableExists)

		if err != nil {
			return fmt.Errorf("error checking if lists table exists: %w", err)
		}

		if !tableExists {
			return fmt.Errorf("lists table does not exist in schema %s", currentSchema)
		}

		// First get the lists basic info with fully qualified table name
		query := fmt.Sprintf(`
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.owner_id, l.owner_type,
				l.created_at, l.updated_at, l.deleted_at
			FROM %s.lists l
			WHERE l.sync_source = $1
				AND l.deleted_at IS NULL
			ORDER BY l.created_at DESC`, currentSchema)

		rows, err := tx.Query(query, source)
		if err != nil {
			return fmt.Errorf("error getting lists by source: %w", err)
		}
		defer safeClose(rows)

		for rows.Next() {
			list := &models.List{
				Items:  []*models.ListItem{},
				Owners: []*models.ListOwner{},
			}
			if err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.OwnerID, &list.OwnerType,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			); err != nil {
				return fmt.Errorf("error scanning list: %w", err)
			}

			lists = append(lists, list)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating lists: %w", err)
		}

		// Load items and owners separately for each list to avoid parse errors
		for _, list := range lists {
			// Get items
			items, err := r.GetItems(list.ID)
			if err != nil {
				return fmt.Errorf("error loading items for list %s: %w", list.ID, err)
			}
			list.Items = items

			// Get owners
			owners, err := r.GetOwners(list.ID)
			if err != nil {
				return fmt.Errorf("error loading owners for list %s: %w", list.ID, err)
			}
			list.Owners = owners
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}

// CreateConflict creates a new list conflict
func (r *ListRepository) CreateConflict(conflict *models.SyncConflict) error {
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

		remoteData, err := json.Marshal(conflict.RemoteData)
		if err != nil {
			return fmt.Errorf("error marshaling remote data: %w", err)
		}

		query := `
			INSERT INTO list_conflicts (
				id, list_id, type,
				local_data, remote_data,
				created_at, updated_at
			) VALUES (
				$1, $2, $3,
				$4, $5,
				NOW(), NOW()
			)`

		_, err = tx.Exec(query,
			conflict.ID,
			conflict.ListID,
			conflict.Type,
			localData,
			remoteData,
		)
		if err != nil {
			return fmt.Errorf("error creating conflict: %w", err)
		}

		return nil
	})
}

// GetConflicts retrieves all unresolved conflicts for a list
func (r *ListRepository) GetConflicts(listID uuid.UUID) ([]*models.SyncConflict, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var conflicts []*models.SyncConflict

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, list_id, type,
				local_data, remote_data,
				created_at, updated_at, resolved_at
			FROM list_conflicts
			WHERE list_id = $1
				AND resolved_at IS NULL
				AND deleted_at IS NULL
			ORDER BY created_at DESC`

		rows, err := tx.Query(query, listID)
		if err != nil {
			return fmt.Errorf("error getting conflicts: %w", err)
		}
		defer safeClose(rows)

		for rows.Next() {
			conflict := &models.SyncConflict{}
			var localData, remoteData []byte

			err := rows.Scan(
				&conflict.ID,
				&conflict.ListID,
				&conflict.Type,
				&localData,
				&remoteData,
				&conflict.CreatedAt,
				&conflict.UpdatedAt,
				&conflict.ResolvedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning conflict: %w", err)
			}

			// Parse JSON data
			if err := json.Unmarshal(localData, &conflict.LocalData); err != nil {
				return fmt.Errorf("error parsing local data: %w", err)
			}
			if err := json.Unmarshal(remoteData, &conflict.RemoteData); err != nil {
				return fmt.Errorf("error parsing remote data: %w", err)
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
func (r *ListRepository) ResolveConflict(conflictID uuid.UUID) error {
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
func (r *ListRepository) AddOwner(owner *models.ListOwner) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		now := time.Now()
		if owner.CreatedAt.IsZero() {
			owner.CreatedAt = now
		}
		if owner.UpdatedAt.IsZero() {
			owner.UpdatedAt = now
		}

		query := `
			INSERT INTO list_owners (
				list_id, owner_id, owner_type,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5)`

		_, err := tx.Exec(query,
			owner.ListID,
			owner.OwnerID,
			owner.OwnerType,
			owner.CreatedAt,
			owner.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error adding list owner: %w", err)
		}

		return nil
	})
}

// RemoveOwner removes an owner from a list
func (r *ListRepository) RemoveOwner(listID, ownerID uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE list_owners
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE list_id = $1 AND owner_id = $2 AND deleted_at IS NULL`

		result, err := tx.Exec(query, listID, ownerID)
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
	})
}

// GetOwners retrieves all owners of a list
func (r *ListRepository) GetOwners(listID uuid.UUID) ([]*models.ListOwner, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var owners []*models.ListOwner

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Get current schema
		var currentSchema string
		err := tx.QueryRow(`SELECT current_schema()`).Scan(&currentSchema)
		if err != nil {
			return fmt.Errorf("error getting current schema: %w", err)
		}

		// Check if the table exists
		var tableExists bool
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = $1
				AND table_name = 'list_owners'
			)`, currentSchema).Scan(&tableExists)

		if err != nil {
			return fmt.Errorf("error checking if list_owners table exists: %w", err)
		}

		if !tableExists {
			return fmt.Errorf("list_owners table does not exist in schema %s", currentSchema)
		}

		// Use fully qualified table name
		query := fmt.Sprintf(`
			SELECT list_id, owner_id, owner_type, created_at, updated_at
			FROM %s.list_owners
			WHERE list_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC`, currentSchema)

		rows, err := tx.Query(query, listID)
		if err != nil {
			return fmt.Errorf("error getting list owners: %w", err)
		}
		defer safeClose(rows)

		for rows.Next() {
			owner := &models.ListOwner{}
			if err := rows.Scan(
				&owner.ListID,
				&owner.OwnerID,
				&owner.OwnerType,
				&owner.CreatedAt,
				&owner.UpdatedAt,
			); err != nil {
				return fmt.Errorf("error scanning list owner: %w", err)
			}
			owners = append(owners, owner)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating list owners: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return owners, nil
}

// GetUserLists retrieves all lists owned by a user
func (r *ListRepository) GetUserLists(userID uuid.UUID) ([]*models.List, error) {
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
		defer safeClose(rows)

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
func (r *ListRepository) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
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
		defer safeClose(rows)

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
func (r *ListRepository) ShareWithTribe(share *models.ListShare) error {
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

		// Check if there's an existing share
		var existingShare bool
		err = tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM list_sharing 
				WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
			)`,
			share.ListID, share.TribeID,
		).Scan(&existingShare)
		if err != nil {
			return fmt.Errorf("error checking existing share: %w", err)
		}

		// If exists, soft delete it first
		if existingShare {
			_, err = tx.Exec(`
				UPDATE list_sharing
				SET deleted_at = NOW()
				WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL`,
				share.ListID, share.TribeID,
			)
			if err != nil {
				return fmt.Errorf("error removing old shares: %w", err)
			}
		}

		// Insert the new share
		_, err = tx.Exec(`
			INSERT INTO list_sharing (
				list_id, tribe_id, user_id,
				created_at, updated_at, expires_at
			) VALUES ($1, $2, $3, NOW(), NOW(), $4)`,
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
func (r *ListRepository) UnshareWithTribe(listID, tribeID uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Mark the share record as deleted
		shareQuery := `
			UPDATE list_sharing
			SET deleted_at = NOW()
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL`

		shareResult, err := tx.Exec(shareQuery, listID, tribeID)
		if err != nil {
			return fmt.Errorf("error unsharing list: %w", err)
		}

		shareRows, err := shareResult.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if shareRows == 0 {
			return models.ErrNotFound
		}

		// Remove the tribe as an owner
		ownerQuery := `
			UPDATE list_owners
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE list_id = $1 AND owner_id = $2 AND owner_type = 'tribe' AND deleted_at IS NULL`

		_, err = tx.Exec(ownerQuery, listID, tribeID)
		if err != nil {
			return fmt.Errorf("error removing list owner: %w", err)
		}

		return nil
	})
}

// GetListShares retrieves all shares for a list
func (r *ListRepository) GetListShares(listID uuid.UUID) ([]*models.ListShare, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var shares []*models.ListShare

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT list_id, tribe_id, user_id, created_at, updated_at, expires_at, deleted_at, version
			FROM list_sharing
			WHERE list_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC`

		rows, err := tx.Query(query, listID)
		if err != nil {
			return fmt.Errorf("error getting list shares: %w", err)
		}
		defer safeClose(rows)

		for rows.Next() {
			share := &models.ListShare{}
			err := rows.Scan(
				&share.ListID,
				&share.TribeID,
				&share.UserID,
				&share.CreatedAt,
				&share.UpdatedAt,
				&share.ExpiresAt,
				&share.DeletedAt,
				&share.Version,
			)
			if err != nil {
				return fmt.Errorf("error scanning list share: %w", err)
			}
			shares = append(shares, share)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return shares, nil
}

// GetSharedLists retrieves all lists shared with a tribe
func (r *ListRepository) GetSharedLists(tribeID uuid.UUID) ([]*models.List, error) {
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
		defer safeClose(rows)

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

// GetListsByOwner retrieves all lists owned by a specific owner
func (r *ListRepository) GetListsByOwner(ownerID uuid.UUID, ownerType models.OwnerType) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	lists := make([]*models.List, 0)

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First retrieve basic list information - filter on the lists table to ensure we only get primary owners
		query := `
			SELECT id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				owner_id, owner_type,
				created_at, updated_at
			FROM lists
			WHERE owner_id = $1 
			  AND owner_type = $2
			  AND deleted_at IS NULL
			ORDER BY created_at DESC`

		rows, err := tx.Query(query, ownerID, ownerType)
		if err != nil {
			return fmt.Errorf("error getting lists by owner: %w", err)
		}
		defer safeClose(rows)

		var listIDs []uuid.UUID
		for rows.Next() {
			list := &models.List{}
			if err := rows.Scan(
				&list.ID,
				&list.Type,
				&list.Name,
				&list.Description,
				&list.Visibility,
				&list.SyncStatus,
				&list.SyncSource,
				&list.SyncID,
				&list.LastSyncAt,
				&list.DefaultWeight,
				&list.MaxItems,
				&list.CooldownDays,
				&list.OwnerID,
				&list.OwnerType,
				&list.CreatedAt,
				&list.UpdatedAt,
			); err != nil {
				return fmt.Errorf("error scanning list: %w", err)
			}

			lists = append(lists, list)
			listIDs = append(listIDs, list.ID)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating lists: %w", err)
		}

		// If no lists were found, return empty array
		if len(listIDs) == 0 {
			return nil
		}

		// Now load items for all lists at once
		itemsQuery := `
			SELECT id, list_id, name, description,
				metadata, external_id,
				weight, last_chosen, chosen_count,
				latitude, longitude, address,
				created_at, updated_at, deleted_at
			FROM list_items
			WHERE list_id = ANY($1) AND deleted_at IS NULL`

		itemRows, err := tx.Query(itemsQuery, pq.Array(listIDs))
		if err != nil {
			return fmt.Errorf("error loading list items: %w", err)
		}
		defer safeClose(itemRows)

		// Map to store items by list ID
		itemsByListID := make(map[uuid.UUID][]*models.ListItem)

		for itemRows.Next() {
			item := &models.ListItem{}
			var metadata []byte
			if err := itemRows.Scan(
				&item.ID, &item.ListID, &item.Name, &item.Description,
				&metadata, &item.ExternalID,
				&item.Weight, &item.LastChosen, &item.ChosenCount,
				&item.Latitude, &item.Longitude, &item.Address,
				&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
			); err != nil {
				return fmt.Errorf("error scanning list item: %w", err)
			}

			if len(metadata) > 0 {
				if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
					return fmt.Errorf("error decoding item metadata: %w", err)
				}
			} else {
				item.Metadata = make(models.JSONMap)
			}

			itemsByListID[item.ListID] = append(itemsByListID[item.ListID], item)
		}

		// Load owners
		ownersQuery := `
			SELECT list_id, owner_id, owner_type, created_at, deleted_at
			FROM list_owners
			WHERE list_id = ANY($1) AND deleted_at IS NULL`

		ownerRows, err := tx.Query(ownersQuery, pq.Array(listIDs))
		if err != nil {
			return fmt.Errorf("error loading list owners: %w", err)
		}
		defer safeClose(ownerRows)

		// Map to store owners by list ID
		ownersByListID := make(map[uuid.UUID][]*models.ListOwner)

		for ownerRows.Next() {
			owner := &models.ListOwner{}
			if err := ownerRows.Scan(
				&owner.ListID,
				&owner.OwnerID,
				&owner.OwnerType,
				&owner.CreatedAt,
				&owner.DeletedAt,
			); err != nil {
				return fmt.Errorf("error scanning list owner: %w", err)
			}
			ownersByListID[owner.ListID] = append(ownersByListID[owner.ListID], owner)
		}

		// Load shares
		sharesQuery := `
			SELECT list_id, tribe_id, expires_at, created_at, updated_at, deleted_at
			FROM list_sharing
			WHERE list_id = ANY($1) AND deleted_at IS NULL`

		shareRows, err := tx.Query(sharesQuery, pq.Array(listIDs))
		if err != nil {
			return fmt.Errorf("error loading list shares: %w", err)
		}
		defer safeClose(shareRows)

		// Map to store shares by list ID
		sharesByListID := make(map[uuid.UUID][]*models.ListShare)

		for shareRows.Next() {
			share := &models.ListShare{}
			err := shareRows.Scan(
				&share.ListID,
				&share.TribeID,
				&share.ExpiresAt,
				&share.CreatedAt,
				&share.UpdatedAt,
				&share.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list share: %w", err)
			}
			sharesByListID[share.ListID] = append(sharesByListID[share.ListID], share)
		}

		// Attach all related data to each list
		for _, list := range lists {
			list.Items = itemsByListID[list.ID]
			list.Owners = ownersByListID[list.ID]
			list.Shares = sharesByListID[list.ID]
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}

// GetSharedTribes retrieves all tribes that a list is shared with
func (r *ListRepository) GetSharedTribes(listID uuid.UUID) ([]*models.Tribe, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var tribes []*models.Tribe

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT DISTINCT t.id, t.name, t.type, t.description, t.visibility,
				t.metadata, t.created_at, t.updated_at, t.deleted_at
			FROM tribes t
			JOIN list_sharing ls ON t.id = ls.tribe_id
			WHERE ls.list_id = $1
				AND ls.deleted_at IS NULL
				AND t.deleted_at IS NULL
				AND (ls.expires_at IS NULL OR ls.expires_at > NOW())
			ORDER BY t.created_at DESC`

		rows, err := tx.Query(query, listID)
		if err != nil {
			return fmt.Errorf("error getting shared tribes: %w", err)
		}
		defer safeClose(rows)

		for rows.Next() {
			tribe := &models.Tribe{}
			var metadata []byte
			err := rows.Scan(
				&tribe.ID,
				&tribe.Name,
				&tribe.Type,
				&tribe.Description,
				&tribe.Visibility,
				&metadata,
				&tribe.CreatedAt,
				&tribe.UpdatedAt,
				&tribe.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning tribe: %w", err)
			}

			if len(metadata) > 0 {
				if err := json.Unmarshal(metadata, &tribe.Metadata); err != nil {
					return fmt.Errorf("error parsing tribe metadata: %w", err)
				}
			}

			tribes = append(tribes, tribe)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tribes, nil
}

// MarkItemChosen marks an item as chosen and updates its stats
func (r *ListRepository) MarkItemChosen(itemID uuid.UUID) error {
	return r.UpdateItemStats(itemID, true)
}
