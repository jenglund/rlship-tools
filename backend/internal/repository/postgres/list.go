package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
	"github.com/lib/pq"
)

type ListRepository struct {
	BaseRepository
	tm *TransactionManager
}

// NewListRepository creates a new PostgreSQL-backed list repository
func NewListRepository(db interface{}) models.ListRepository {
	baseRepo := NewBaseRepository(db)

	// Log schema info
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("Creating ListRepository with test schema: %s\n", testSchema)
	}

	return &ListRepository{
		BaseRepository: baseRepo,
		tm:             NewTransactionManager(baseRepo.GetQueryDB()),
	}
}

// Create creates a new list and adds the primary owner to the list_owners table
func (r *ListRepository) Create(list *models.List) error {
	// Validate required fields
	if list.Name == "" {
		return fmt.Errorf("%w: name is required", models.ErrInvalidInput)
	}

	// Validate list type
	if err := list.Type.Validate(); err != nil {
		return fmt.Errorf("invalid list type: %w", err)
	}

	// Validate and set owners
	if err := r.validateAndSetOwners(list); err != nil {
		return fmt.Errorf("error validating owners: %w", err)
	}

	log.Printf("Creating list '%s' with primary owner (ID: %s, Type: %s)", list.Name, *list.OwnerID, *list.OwnerType)

	// Set creation timestamp if not set
	now := time.Now()
	if list.CreatedAt.IsZero() {
		list.CreatedAt = now
	}
	if list.UpdatedAt.IsZero() {
		list.UpdatedAt = now
	}

	// Generate ID if not provided
	if list.ID == uuid.Nil {
		list.ID = uuid.New()
	}

	ctx := context.Background()
	opts := DefaultTransactionOptions()
	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Create the list
		_, err := tx.Exec(`
			INSERT INTO lists (
				id, type, name, description, visibility,
				default_weight, max_items, cooldown_days,
				sync_status, sync_source, sync_id, last_sync_at,
				owner_id, owner_type,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11, $12,
				$13, $14,
				$15, $16
			)`,
			list.ID, list.Type, list.Name, list.Description, list.Visibility,
			list.DefaultWeight, list.MaxItems, list.CooldownDays,
			list.SyncStatus, list.SyncSource, list.SyncID, list.LastSyncAt,
			list.OwnerID, list.OwnerType,
			list.CreatedAt, list.UpdatedAt)

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
		log.Printf("Adding primary owner (ID: %s, Type: %s) to list %s", *list.OwnerID, *list.OwnerType, list.ID)
		_, err = tx.Exec(`
			INSERT INTO list_owners (
				list_id, owner_id, owner_type,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5)`,
			list.ID, *list.OwnerID, *list.OwnerType, list.CreatedAt, list.UpdatedAt)
		if err != nil {
			return fmt.Errorf("error adding primary owner: %w", err)
		}

		log.Printf("Successfully added primary owner (ID: %s, Type: %s) to list %s", *list.OwnerID, *list.OwnerType, list.ID)

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
					log.Printf("Skipping duplicate primary owner (ID: %s, Type: %s) for list %s", owner.OwnerID, owner.OwnerType, list.ID)
					continue
				}

				owner.ListID = list.ID
				owner.CreatedAt = list.CreatedAt
				owner.UpdatedAt = list.UpdatedAt

				log.Printf("Adding additional owner (ID: %s, Type: %s) to list %s", owner.OwnerID, owner.OwnerType, list.ID)
				_, err = stmt.Exec(
					owner.ListID,
					owner.OwnerID,
					owner.OwnerType,
					owner.CreatedAt,
					owner.UpdatedAt,
				)
				if err != nil {
					if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
						log.Printf("Skipping duplicate owner (ID: %s, Type: %s) for list %s", owner.OwnerID, owner.OwnerType, list.ID)
						continue // Skip duplicate owners
					}
					return fmt.Errorf("error adding list owner: %w", err)
				}
				log.Printf("Successfully added additional owner (ID: %s, Type: %s) to list %s", owner.OwnerID, owner.OwnerType, list.ID)
			}
		}

		log.Printf("Successfully created list '%s' (ID: %s) with primary owner (ID: %s, Type: %s)", list.Name, list.ID, *list.OwnerID, *list.OwnerType)
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
	// Get schema for debugging
	var schemaName string
	schemaErr := tx.QueryRow("SHOW search_path").Scan(&schemaName)
	if schemaErr == nil {
		fmt.Printf("loadListData: Using schema %s (verification handled by TransactionManager)\n", schemaName)
	}

	// Initialize list's related fields to empty slices even if loading fails
	list.Items = make([]*models.ListItem, 0)
	list.Owners = make([]*models.ListOwner, 0)
	list.Shares = make([]*models.ListShare, 0)

	// Load items - use only the fields we know are in the database schema
	// This is a more defensive approach to prevent errors with missing columns
	itemsQuery := `
		SELECT id, list_id, name, description, 
			metadata, external_id, 
			latitude, longitude, address, 
			weight, last_chosen, chosen_count,
			created_at, updated_at, deleted_at
		FROM list_items
		WHERE list_id = $1 AND deleted_at IS NULL
		ORDER BY name`

	// Use transaction for query instead of r.db to maintain schema context
	var items []*models.ListItem
	itemsRows, err := tx.Query(itemsQuery, list.ID)
	if err != nil {
		log.Printf("Error loading list items for list %s: %v", list.ID, err)
		// Don't return error, just continue with empty items
		// This allows the application to continue functioning with partial data
	} else {
		defer safeClose(itemsRows)

		for itemsRows.Next() {
			item := &models.ListItem{}
			var metadata []byte
			if err := itemsRows.Scan(
				&item.ID, &item.ListID, &item.Name, &item.Description,
				&metadata, &item.ExternalID,
				&item.Latitude, &item.Longitude, &item.Address,
				&item.Weight, &item.LastChosen, &item.ChosenCount,
				&item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
			); err != nil {
				log.Printf("Error scanning list item: %v", err)
				continue // Skip this item but continue with others
			}

			if len(metadata) > 0 {
				if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
					log.Printf("Error decoding item metadata: %v", err)
					item.Metadata = make(models.JSONMap) // Use empty metadata rather than failing
				}
			} else {
				item.Metadata = make(models.JSONMap)
			}

			items = append(items, item)
		}

		// Check for iterator errors but don't fail the entire operation
		if err := itemsRows.Err(); err != nil {
			log.Printf("Error iterating list items: %v", err)
		}

		list.Items = items
	}

	// Load list owners
	ownersQuery := `
		SELECT list_id, owner_id, owner_type, created_at, updated_at, deleted_at
		FROM list_owners
		WHERE list_id = $1 AND deleted_at IS NULL`

	var owners []*models.ListOwner
	ownersRows, err := tx.Query(ownersQuery, list.ID)
	if err != nil {
		log.Printf("Error loading list owners for list %s: %v", list.ID, err)
		// Continue with empty owners
	} else {
		defer safeClose(ownersRows)

		for ownersRows.Next() {
			owner := &models.ListOwner{}
			if err := ownersRows.Scan(
				&owner.ListID, &owner.OwnerID, &owner.OwnerType,
				&owner.CreatedAt, &owner.UpdatedAt, &owner.DeletedAt,
			); err != nil {
				log.Printf("Error scanning list owner: %v", err)
				continue // Skip this owner but continue with others
			}
			owners = append(owners, owner)
		}

		// Check for iterator errors but don't fail the entire operation
		if err := ownersRows.Err(); err != nil {
			log.Printf("Error iterating list owners: %v", err)
		}

		list.Owners = owners
	}

	// Load list sharing
	sharesQuery := `
		SELECT list_id, tribe_id, user_id, expires_at, created_at, updated_at, deleted_at, version
		FROM list_sharing
		WHERE list_id = $1 AND deleted_at IS NULL`

	var shares []*models.ListShare
	sharesRows, err := tx.Query(sharesQuery, list.ID)
	if err != nil {
		log.Printf("Error loading list shares for list %s: %v", list.ID, err)
		// Continue with empty shares
	} else {
		defer safeClose(sharesRows)

		for sharesRows.Next() {
			share := &models.ListShare{}
			if err := sharesRows.Scan(
				&share.ListID, &share.TribeID, &share.UserID, &share.ExpiresAt,
				&share.CreatedAt, &share.UpdatedAt, &share.DeletedAt, &share.Version,
			); err != nil {
				log.Printf("Error scanning list share: %v", err)
				continue // Skip this share but continue with others
			}
			shares = append(shares, share)
		}

		// Check for iterator errors but don't fail the entire operation
		if err := sharesRows.Err(); err != nil {
			log.Printf("Error iterating list shares: %v", err)
		}

		list.Shares = shares
	}

	return nil
}

// GetByID retrieves a list by its ID
func (r *ListRepository) GetByID(id uuid.UUID) (*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	var list models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Query the list
		query := `
			SELECT id, type, name, description, visibility,
				sync_status, sync_source, sync_id, last_sync_at,
				default_weight, max_items, cooldown_days,
				owner_id, owner_type, 
				created_at, updated_at, deleted_at
			FROM lists
			WHERE id = $1 AND deleted_at IS NULL`

		var maxItems, cooldownDays sql.NullInt32
		var syncID sql.NullString
		var lastSyncAt sql.NullTime
		var ownerType models.OwnerType
		var ownerID uuid.UUID
		var deletedAt sql.NullTime

		err := tx.QueryRow(query, id).Scan(
			&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
			&list.SyncStatus, &list.SyncSource, &syncID, &lastSyncAt,
			&list.DefaultWeight, &maxItems, &cooldownDays,
			&ownerID, &ownerType,
			&list.CreatedAt, &list.UpdatedAt, &deletedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return models.ErrNotFound
			}
			return fmt.Errorf("error getting list: %w", err)
		}

		// Set nullable fields
		if maxItems.Valid {
			maxItemsVal := int(maxItems.Int32)
			list.MaxItems = &maxItemsVal
		}
		if cooldownDays.Valid {
			cooldownDaysVal := int(cooldownDays.Int32)
			list.CooldownDays = &cooldownDaysVal
		}
		if syncID.Valid {
			list.SyncID = syncID.String
		}
		if lastSyncAt.Valid {
			syncTime := lastSyncAt.Time
			list.LastSyncAt = &syncTime
		}
		if deletedAt.Valid {
			list.DeletedAt = &deletedAt.Time
		}

		// Set owner fields
		list.OwnerID = &ownerID
		list.OwnerType = &ownerType

		// Load related data
		if err := r.loadListData(tx, &list); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &list, nil
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
		if validateErr := list.Validate(); validateErr != nil {
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
				WHERE table_schema = current_schema() 
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

	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
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

		rows, err := tx.Query(query, limit, offset)
		if err != nil {
			return fmt.Errorf("error listing lists: %w", err)
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

			lists = append(lists, list)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating lists: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
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
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE list_items SET
				deleted_at = NOW()
			WHERE id = $1 AND list_id = $2 AND deleted_at IS NULL`

		result, err := tx.Exec(query, itemID, listID)
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
	})
}

// GetItems retrieves all items from a list
func (r *ListRepository) GetItems(listID uuid.UUID) ([]*models.ListItem, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var items []*models.ListItem

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Use a simpler query with standard table reference
		// The transaction manager already sets the correct search path
		query := `
			SELECT id, list_id, name, description,
				metadata, external_id,
				weight, last_chosen, chosen_count,
				latitude, longitude, address,
				created_at, updated_at, deleted_at
			FROM list_items
			WHERE list_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC`

		rows, err := tx.Query(query, listID)
		if err != nil {
			return fmt.Errorf("error querying list items: %w", err)
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

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetEligibleItems retrieves items eligible for menu generation
func (r *ListRepository) GetEligibleItems(listIDs []uuid.UUID, filters map[string]interface{}) ([]*models.ListItem, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var items []*models.ListItem

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
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

		rows, err := tx.Query(query, pq.Array(listIDs))
		if err != nil {
			return err
		}
		defer safeClose(rows)

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
				return err
			}

			if len(metadata) > 0 {
				if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
					return err
				}
			} else {
				item.Metadata = make(models.JSONMap)
			}

			// Apply list's default weight if item weight is not set
			if item.Weight == 0 {
				item.Weight = defaultWeight
			}

			items = append(items, item)
		}

		return nil
	})

	if err != nil {
		return nil, err
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
	// Get test schema context if available
	var ctx context.Context
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("GetListsBySource: Using schema %s (verification handled by TransactionManager)\n", testSchema)
		// Use context with schema for better isolation
		schemaCtx := testutil.GetSchemaContext(testSchema)
		if schemaCtx != nil {
			ctx = schemaCtx.WithContext(context.Background())
		} else {
			ctx = context.Background()
		}
	} else {
		ctx = context.Background()
	}

	// Create a context with a longer timeout for this operation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	opts := DefaultTransactionOptions()
	// Increase statement timeout to match context timeout
	opts.StatementTimeout = 25 * time.Second
	// Use a more permissive isolation level for read-only operations
	opts.IsolationLevel = sql.LevelReadCommitted

	var lists []*models.List

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Verify schema is set correctly
		var schemaName string
		err := tx.QueryRow("SELECT current_schema()").Scan(&schemaName)
		if err != nil {
			return fmt.Errorf("error verifying schema: %w", err)
		}
		fmt.Printf("GetListsBySource: Using schema: %s\n", schemaName)

		// First get the lists basic info with schema-qualified table names
		query := fmt.Sprintf(`
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.owner_id, l.owner_type,
				l.created_at, l.updated_at, l.deleted_at
			FROM %s.lists l
			WHERE l.sync_source = $1
				AND l.deleted_at IS NULL
			ORDER BY l.created_at DESC`, pq.QuoteIdentifier(schemaName))

		rows, err := tx.Query(query, source)
		if err != nil {
			return fmt.Errorf("error getting lists by source: %w", err)
		}
		defer safeClose(rows)

		var listIDs []uuid.UUID // Track list IDs for batch loading

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
			listIDs = append(listIDs, list.ID)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating lists: %w", err)
		}

		// If no lists were found, return early
		if len(lists) == 0 {
			return nil
		}

		fmt.Printf("GetListsBySource: Found %d lists for source '%s'\n", len(lists), source)

		// Load all the related data for each list
		for _, list := range lists {
			if err := r.loadListData(tx, list); err != nil {
				// Log the error but continue to load other lists
				fmt.Printf("GetListsBySource: Error loading data for list %s: %v\n", list.ID, err)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("GetListsBySource: Error retrieving lists: %v\n", err)
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
		// Use simple table reference - the transaction manager handles schema
		query := `
			SELECT list_id, owner_id, owner_type, created_at, updated_at
			FROM list_owners
			WHERE list_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC`

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

	// Only log the schema, don't try to verify it (TransactionManager handles this)
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("GetUserLists: Using schema %s (verification handled by TransactionManager)\n", testSchema)
	}

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

		// Initialize lists slice and collect list IDs for batch loading
		var listIDs []uuid.UUID

		for rows.Next() {
			list := &models.List{
				// Initialize empty slices to prevent nil errors
				Items:  []*models.ListItem{},
				Owners: []*models.ListOwner{},
				Shares: []*models.ListShare{},
			}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
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

		// Batch load items for all lists at once
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
			err := itemRows.Scan(
				&item.ID,
				&item.ListID,
				&item.Name,
				&item.Description,
				&item.Metadata,
				&item.ExternalID,
				&item.Weight,
				&item.LastChosen,
				&item.ChosenCount,
				&item.Latitude,
				&item.Longitude,
				&item.Address,
				&item.CreatedAt,
				&item.UpdatedAt,
				&item.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list item: %w", err)
			}
			itemsByListID[item.ListID] = append(itemsByListID[item.ListID], item)
		}

		// Batch load owners for all lists at once
		ownersQuery := `
			SELECT list_id, owner_id, owner_type, created_at, updated_at, deleted_at
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
			err := ownerRows.Scan(
				&owner.ListID,
				&owner.OwnerID,
				&owner.OwnerType,
				&owner.CreatedAt,
				&owner.UpdatedAt,
				&owner.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list owner: %w", err)
			}
			ownersByListID[owner.ListID] = append(ownersByListID[owner.ListID], owner)
		}

		// Batch load shares for all lists at once
		sharesQuery := `
			SELECT list_id, tribe_id, user_id, expires_at, created_at, updated_at, deleted_at, version
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
				&share.UserID,
				&share.ExpiresAt,
				&share.CreatedAt,
				&share.UpdatedAt,
				&share.DeletedAt,
				&share.Version,
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

// GetTribeLists retrieves all lists owned by a tribe
func (r *ListRepository) GetTribeLists(tribeID uuid.UUID) ([]*models.List, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var lists []*models.List

	// Only log the schema, don't try to verify it (TransactionManager handles this)
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("GetTribeLists: Using schema %s (verification handled by TransactionManager)\n", testSchema)
	}

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Fix the parentheses in the query to properly group conditions
		query := `
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.created_at, l.updated_at, l.deleted_at
			FROM lists l
			LEFT JOIN list_owners lo ON l.id = lo.list_id
			LEFT JOIN list_sharing ls ON l.id = ls.list_id
			WHERE ((lo.owner_id = $1 AND lo.owner_type = 'tribe' AND lo.deleted_at IS NULL)
				OR (ls.tribe_id = $1 AND ls.deleted_at IS NULL))
				AND l.deleted_at IS NULL
			ORDER BY l.created_at DESC`

		rows, err := tx.Query(query, tribeID)
		if err != nil {
			return fmt.Errorf("error getting tribe lists: %w", err)
		}
		defer safeClose(rows)

		// Initialize lists slice and collect list IDs for batch loading
		var listIDs []uuid.UUID

		for rows.Next() {
			list := &models.List{
				// Initialize empty slices to prevent nil errors
				Items:  []*models.ListItem{},
				Owners: []*models.ListOwner{},
				Shares: []*models.ListShare{},
			}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
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

		// Batch load items for all lists at once
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
			err := itemRows.Scan(
				&item.ID,
				&item.ListID,
				&item.Name,
				&item.Description,
				&item.Metadata,
				&item.ExternalID,
				&item.Weight,
				&item.LastChosen,
				&item.ChosenCount,
				&item.Latitude,
				&item.Longitude,
				&item.Address,
				&item.CreatedAt,
				&item.UpdatedAt,
				&item.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list item: %w", err)
			}
			itemsByListID[item.ListID] = append(itemsByListID[item.ListID], item)
		}

		// Batch load owners for all lists at once
		ownersQuery := `
			SELECT list_id, owner_id, owner_type, created_at, updated_at, deleted_at
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
			err := ownerRows.Scan(
				&owner.ListID,
				&owner.OwnerID,
				&owner.OwnerType,
				&owner.CreatedAt,
				&owner.UpdatedAt,
				&owner.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list owner: %w", err)
			}
			ownersByListID[owner.ListID] = append(ownersByListID[owner.ListID], owner)
		}

		// Batch load shares for all lists at once
		sharesQuery := `
			SELECT list_id, tribe_id, user_id, expires_at, created_at, updated_at, deleted_at, version
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
				&share.UserID,
				&share.ExpiresAt,
				&share.CreatedAt,
				&share.UpdatedAt,
				&share.DeletedAt,
				&share.Version,
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

// ShareWithTribe shares a list with a tribe
func (r *ListRepository) ShareWithTribe(share *models.ListShare) error {
	// Get test schema context if available
	var ctx context.Context
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("ShareWithTribe: Using schema %s (verification handled by TransactionManager)\n", testSchema)
		// Use context with schema for better isolation
		schemaCtx := testutil.GetSchemaContext(testSchema)
		if schemaCtx != nil {
			ctx = schemaCtx.WithContext(context.Background())
		} else {
			ctx = context.Background()
		}
	} else {
		ctx = context.Background()
	}

	// Validate the share input
	if share == nil {
		return fmt.Errorf("%w: share cannot be nil", models.ErrInvalidInput)
	}

	if share.ListID == uuid.Nil {
		return fmt.Errorf("%w: list ID is required", models.ErrInvalidInput)
	}

	if share.TribeID == uuid.Nil {
		return fmt.Errorf("%w: tribe ID is required", models.ErrInvalidInput)
	}

	if share.UserID == uuid.Nil {
		return fmt.Errorf("%w: user ID is required", models.ErrInvalidInput)
	}

	// Ensure expiry date is in the future if it's set
	if share.ExpiresAt != nil && share.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("%w: expiration date must be in the future", models.ErrInvalidInput)
	}

	// Set timestamp fields if not already set
	now := time.Now()
	if share.CreatedAt.IsZero() {
		share.CreatedAt = now
	}
	if share.UpdatedAt.IsZero() {
		share.UpdatedAt = now
	}

	// Default the version to 1 for new shares
	if share.Version <= 0 {
		share.Version = 1
	}

	opts := DefaultTransactionOptions()
	// Use a higher isolation level for safety
	opts.IsolationLevel = sql.LevelReadCommitted
	// Set a reasonable statement timeout, but not too short
	opts.StatementTimeout = 30 * time.Second

	startTime := time.Now()

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First, check that the list exists
		var listExists bool
		err := tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM lists 
				WHERE id = $1 AND deleted_at IS NULL
			)`,
			share.ListID,
		).Scan(&listExists)
		if err != nil {
			return fmt.Errorf("error checking if list exists: %w", err)
		}
		if !listExists {
			fmt.Printf("ShareWithTribe: List not found: %s\n", share.ListID)
			return fmt.Errorf("%w: list not found", models.ErrNotFound)
		}

		// Check that the tribe exists
		var tribeExists bool
		err = tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM tribes
				WHERE id = $1 AND deleted_at IS NULL
			)`,
			share.TribeID,
		).Scan(&tribeExists)
		if err != nil {
			return fmt.Errorf("error checking if tribe exists: %w", err)
		}
		if !tribeExists {
			fmt.Printf("ShareWithTribe: Tribe not found: %s\n", share.TribeID)
			return fmt.Errorf("%w: tribe not found", models.ErrNotFound)
		}

		// Check that the user exists
		var userExists bool
		err = tx.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM users
				WHERE id = $1 AND deleted_at IS NULL
			)`,
			share.UserID,
		).Scan(&userExists)
		if err != nil {
			return fmt.Errorf("error checking if user exists: %w", err)
		}
		if !userExists {
			fmt.Printf("ShareWithTribe: User not found: %s\n", share.UserID)
			return fmt.Errorf("%w: user not found", models.ErrNotFound)
		}

		// Check if share already exists and is active
		var currentVersion int
		err = tx.QueryRow(`
			SELECT version FROM list_sharing
			WHERE list_id = $1 AND tribe_id = $2 AND deleted_at IS NULL
		`, share.ListID, share.TribeID).Scan(&currentVersion)

		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error checking existing share: %w", err)
		}

		if err == sql.ErrNoRows {
			// No active share exists, create a new one
			// Use CTE to ensure we get the created row back
			query := `
				WITH new_share AS (
					INSERT INTO list_sharing (
						list_id, tribe_id, user_id, 
						created_at, updated_at, expires_at,
						version
					) VALUES (
						$1, $2, $3, 
						$4, $5, $6, 
						$7
					)
					RETURNING version
				)
				SELECT version FROM new_share;
			`
			var version int
			err = tx.QueryRow(query,
				share.ListID, share.TribeID, share.UserID,
				share.CreatedAt, share.UpdatedAt, share.ExpiresAt,
				share.Version,
			).Scan(&version)

			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok {
					if pqErr.Code == "23505" { // Unique violation
						return fmt.Errorf("%w: share already exists", models.ErrDuplicate)
					}
				}
				return fmt.Errorf("error creating share: %w", err)
			}

			share.Version = version
			fmt.Printf("ShareWithTribe: Created new share for list %s with tribe %s (version %d)\n",
				share.ListID, share.TribeID, share.Version)

			// Now add the tribe as an owner of the list (if not already)
			var isOwner bool
			err = tx.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM list_owners
					WHERE list_id = $1 AND owner_id = $2 AND owner_type = $3 AND deleted_at IS NULL
				)`,
				share.ListID, share.TribeID, models.OwnerTypeTribe,
			).Scan(&isOwner)
			if err != nil {
				return fmt.Errorf("error checking if tribe is owner: %w", err)
			}

			if !isOwner {
				// Add tribe as owner
				_, err = tx.Exec(`
					INSERT INTO list_owners (
						list_id, owner_id, owner_type,
						created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5)`,
					share.ListID, share.TribeID, models.OwnerTypeTribe,
					now, now,
				)
				if err != nil {
					return fmt.Errorf("error adding tribe as list owner: %w", err)
				}
			}
		} else {
			// Share already exists, update it
			// Use CTE to get the updated row back
			query := `
				WITH updated_share AS (
					UPDATE list_sharing
					SET 
						user_id = $1,
						expires_at = $2,
						updated_at = $3,
						version = version + 1
					WHERE list_id = $4 AND tribe_id = $5 AND deleted_at IS NULL
					RETURNING version
				)
				SELECT version FROM updated_share;
			`
			var newVersion int
			err = tx.QueryRow(query,
				share.UserID, share.ExpiresAt, now,
				share.ListID, share.TribeID,
			).Scan(&newVersion)

			if err != nil {
				return fmt.Errorf("error updating share: %w", err)
			}

			fmt.Printf("ShareWithTribe: Updated existing share for list %s to tribe %s (version %d -> %d)\n",
				share.ListID, share.TribeID, currentVersion, newVersion)
			share.Version = newVersion
		}

		return nil
	})

	elapsedTime := time.Since(startTime).Milliseconds()
	if err != nil {
		fmt.Printf("ShareWithTribe: Failed to share list: %v (took %dms)\n", err, elapsedTime)

		// Check for validation errors
		if errors.Is(err, models.ErrInvalidInput) {
			fmt.Printf("ShareWithTribe: Validation failed: %v\n", err)
			return err
		}

		// Check for not found errors
		if errors.Is(err, models.ErrNotFound) {
			return err
		}

		return fmt.Errorf("error sharing list: %w", err)
	}

	fmt.Printf("ShareWithTribe: Successfully shared list %s with tribe %s (took %dms)\n",
		share.ListID, share.TribeID, elapsedTime)
	return nil
}

// UnshareWithTribe removes a tribe's access to a list
func (r *ListRepository) UnshareWithTribe(listID, tribeID uuid.UUID) error {
	// Get test schema context if available
	var ctx context.Context
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("UnshareWithTribe: Using schema %s (verification handled by TransactionManager)\n", testSchema)
		// Use context with schema for better isolation
		schemaCtx := testutil.GetSchemaContext(testSchema)
		if schemaCtx != nil {
			ctx = schemaCtx.WithContext(context.Background())
		} else {
			ctx = context.Background()
		}
	} else {
		ctx = context.Background()
	}

	opts := DefaultTransactionOptions()
	// Set a higher isolation level to prevent race conditions
	opts.IsolationLevel = sql.LevelSerializable

	fmt.Printf("[UnshareWithTribe] START: ListID=%s, TribeID=%s\n", listID, tribeID)
	startTime := time.Now()

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First check if the list exists
		listQuery := `
			SELECT EXISTS (
				SELECT 1 FROM lists 
				WHERE id = $1 
				AND deleted_at IS NULL
			)`

		var listExists bool
		err := tx.QueryRow(listQuery, listID).Scan(&listExists)
		if err != nil {
			fmt.Printf("[UnshareWithTribe] ERROR: Error checking list existence: %v\n", err)
			return fmt.Errorf("error checking list existence: %w", err)
		}

		if !listExists {
			fmt.Printf("[UnshareWithTribe] ERROR: List with ID %s not found\n", listID)
			return fmt.Errorf("%w: list with ID %s not found", models.ErrNotFound, listID)
		}

		// Check if the share actually exists before trying to update it
		shareQuery := `
			SELECT EXISTS (
				SELECT 1 FROM list_sharing 
				WHERE list_id = $1 
				AND tribe_id = $2
				AND deleted_at IS NULL
			)`

		var shareExists bool
		err = tx.QueryRow(shareQuery, listID, tribeID).Scan(&shareExists)
		if err != nil {
			fmt.Printf("[UnshareWithTribe] ERROR: Error checking share existence: %v\n", err)
			return fmt.Errorf("error checking share existence: %w", err)
		}

		if !shareExists {
			fmt.Printf("[UnshareWithTribe] WARNING: No active share found for list %s and tribe %s\n", listID, tribeID)
			// This is not necessarily an error - if the share doesn't exist, then we've already achieved the desired state
			return nil
		}

		// Soft delete the share by setting deleted_at, updated_at, and incrementing version
		query := `
			UPDATE list_sharing
			SET deleted_at = NOW(),
				updated_at = NOW(),
				version = version + 1
			WHERE list_id = $1 
			AND tribe_id = $2 
			AND deleted_at IS NULL
			RETURNING version`

		var newVersion int
		err = tx.QueryRow(query, listID, tribeID).Scan(&newVersion)
		if err == sql.ErrNoRows {
			// This could happen if another transaction deleted the share between our check and update
			fmt.Printf("[UnshareWithTribe] WARNING: Share was deleted by another process between check and update\n")
			return nil
		}
		if err != nil {
			fmt.Printf("[UnshareWithTribe] ERROR: Error unsharing list: %v\n", err)
			return fmt.Errorf("error unsharing list: %w", err)
		}

		fmt.Printf("[UnshareWithTribe] INFO: Successfully unshared list %s from tribe %s (new version: %d)\n",
			listID, tribeID, newVersion)

		// Remove the tribe as an owner if they are one
		ownerQuery := `
			UPDATE list_owners
			SET deleted_at = NOW(), 
			    updated_at = NOW()
			WHERE list_id = $1 
			AND owner_id = $2 
			AND owner_type = 'tribe' 
			AND deleted_at IS NULL`

		ownerResult, err := tx.Exec(ownerQuery, listID, tribeID)
		if err != nil {
			fmt.Printf("[UnshareWithTribe] ERROR: Error removing list owner: %v\n", err)
			return fmt.Errorf("error removing list owner: %w", err)
		}

		ownerRows, _ := ownerResult.RowsAffected()
		if ownerRows > 0 {
			fmt.Printf("[UnshareWithTribe] INFO: Removed tribe %s as owner of list %s\n", tribeID, listID)
		}

		return nil
	})

	elapsedTime := time.Since(startTime).Milliseconds()
	if err != nil {
		fmt.Printf("[UnshareWithTribe] FAILED: Error unsharing list (took %dms): %v\n", elapsedTime, err)
		return err
	}

	fmt.Printf("[UnshareWithTribe] SUCCESS: List unshared successfully (took %dms)\n", elapsedTime)
	return nil
}

// GetListShares retrieves all shares for a list
func (r *ListRepository) GetListShares(listID uuid.UUID) ([]*models.ListShare, error) {
	// Get test schema context if available
	var ctx context.Context
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("GetListShares: Using schema %s (verification handled by TransactionManager)\n", testSchema)
		// Use context with schema for better isolation
		schemaCtx := testutil.GetSchemaContext(testSchema)
		if schemaCtx != nil {
			ctx = schemaCtx.WithContext(context.Background())
		} else {
			ctx = context.Background()
		}
	} else {
		ctx = context.Background()
	}

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

	// Only log the schema, don't try to verify it (TransactionManager handles this)
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("GetSharedLists: Using schema %s (verification handled by TransactionManager)\n", testSchema)
		// Use context with schema for better isolation
		schemaCtx := testutil.GetSchemaContext(testSchema)
		if schemaCtx != nil {
			ctx = schemaCtx.WithContext(context.Background())
		}
	}

	// Determine if we should run cleanup based on context
	runCleanup := true

	// Skip cleanup in test mode unless explicitly enabled
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		// Default to skipping cleanup in tests unless explicitly requested
		runCleanup = false

		// Check for explicit flag
		if forceCleanup, ok := ctx.Value("force_cleanup").(bool); ok && forceCleanup {
			runCleanup = true
		}
	}

	// First, clean up any expired shares to ensure we don't return lists with expired shares
	if runCleanup {
		_, err := r.CleanupExpiredShares(ctx)
		if err != nil {
			// Only log warning in non-test environment
			if testutil.GetCurrentTestSchema() == "" {
				fmt.Printf("Warning: Failed to clean up expired shares: %v\n", err)
			}
			// Continue despite error - we'll still filter out expired shares in the query
		}
	}

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

		// Initialize lists slice and collect list IDs for batch loading
		var listIDs []uuid.UUID

		for rows.Next() {
			list := &models.List{
				// Initialize empty slices to prevent nil errors
				Items:  []*models.ListItem{},
				Owners: []*models.ListOwner{},
				Shares: []*models.ListShare{},
			}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
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

		// Batch load items for all lists at once
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
			err := itemRows.Scan(
				&item.ID,
				&item.ListID,
				&item.Name,
				&item.Description,
				&item.Metadata,
				&item.ExternalID,
				&item.Weight,
				&item.LastChosen,
				&item.ChosenCount,
				&item.Latitude,
				&item.Longitude,
				&item.Address,
				&item.CreatedAt,
				&item.UpdatedAt,
				&item.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list item: %w", err)
			}
			itemsByListID[item.ListID] = append(itemsByListID[item.ListID], item)
		}

		// Batch load owners for all lists at once
		ownersQuery := `
			SELECT list_id, owner_id, owner_type, created_at, updated_at, deleted_at
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
			err := ownerRows.Scan(
				&owner.ListID,
				&owner.OwnerID,
				&owner.OwnerType,
				&owner.CreatedAt,
				&owner.UpdatedAt,
				&owner.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning list owner: %w", err)
			}
			ownersByListID[owner.ListID] = append(ownersByListID[owner.ListID], owner)
		}

		// Batch load shares for all lists at once
		sharesQuery := `
			SELECT list_id, tribe_id, user_id, expires_at, created_at, updated_at, deleted_at, version
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
				&share.UserID,
				&share.ExpiresAt,
				&share.CreatedAt,
				&share.UpdatedAt,
				&share.DeletedAt,
				&share.Version,
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

	log.Printf("Successfully retrieved %d lists shared with tribe %s", len(lists), tribeID)
	return lists, nil
}

// CleanupExpiredShares removes sharing records that have passed their expiration date
// Returns the number of shares that were cleaned up, or an error if the operation failed
func (r *ListRepository) CleanupExpiredShares(ctx context.Context) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Add a reasonable timeout to the context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check if we're in a test environment and should use a specific schema
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		schemaCtx := testutil.GetSchemaContext(testSchema)
		if schemaCtx != nil {
			ctx = schemaCtx.WithContext(ctx)
			fmt.Printf("[CleanupExpiredShares] Using schema context for %s\n", testSchema)
		}
	}

	// Skip cleanup if the context has a skip_cleanup flag set to true
	// This allows tests to disable cleanup when needed
	if skipCleanup, ok := ctx.Value("skip_cleanup").(bool); ok && skipCleanup {
		return 0, nil
	}

	opts := DefaultTransactionOptions()
	// Use a higher isolation level to ensure consistency
	opts.IsolationLevel = sql.LevelReadCommitted

	startTime := time.Now()
	fmt.Printf("[CleanupExpiredShares] Starting cleanup of expired list shares\n")

	// First, let's check if there are any expired shares to clean up
	var expiredCount int
	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		err := tx.QueryRow(`
			SELECT COUNT(*) FROM list_sharing
			WHERE expires_at < NOW() 
			AND deleted_at IS NULL
		`).Scan(&expiredCount)

		if err != nil {
			return fmt.Errorf("error counting expired shares: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("error checking for expired shares: %w", err)
	}

	if expiredCount == 0 {
		fmt.Printf("[CleanupExpiredShares] No expired shares found, skipping cleanup\n")
		return 0, nil
	}

	fmt.Printf("[CleanupExpiredShares] Found %d expired shares to clean up\n", expiredCount)

	// Now perform the cleanup in a separate transaction
	var cleanedCount int
	err = r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Use a CTE to get all the details of what's being updated for better logging
		query := `
			WITH expired_shares AS (
				SELECT list_id, tribe_id, expires_at
				FROM list_sharing
				WHERE expires_at < NOW() 
				AND deleted_at IS NULL
			),
			updated AS (
				UPDATE list_sharing ls
				SET deleted_at = NOW(),
					updated_at = NOW(),
					version = version + 1
				FROM expired_shares es
				WHERE ls.list_id = es.list_id 
				AND ls.tribe_id = es.tribe_id
				AND ls.deleted_at IS NULL
				RETURNING ls.list_id, ls.tribe_id, ls.version, ls.expires_at
			)
			SELECT list_id, tribe_id, version, expires_at FROM updated`

		rows, err := tx.Query(query)
		if err != nil {
			return fmt.Errorf("error cleaning up expired shares: %w", err)
		}
		defer safeClose(rows)

		count := 0
		for rows.Next() {
			var listID, tribeID uuid.UUID
			var version int
			var expiresAt time.Time
			if err := rows.Scan(&listID, &tribeID, &version, &expiresAt); err != nil {
				return fmt.Errorf("error scanning expired share row: %w", err)
			}
			count++
			fmt.Printf("[CleanupExpiredShares] Cleaned up expired share: list_id=%s, tribe_id=%s, version=%d, expired_at=%s\n",
				listID, tribeID, version, expiresAt.Format(time.RFC3339))
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating expired shares: %w", err)
		}

		// Double-check if we cleaned up everything we expected
		if count != expiredCount {
			fmt.Printf("[CleanupExpiredShares] Warning: Expected to clean up %d shares but only cleaned up %d\n",
				expiredCount, count)
		}

		cleanedCount = count
		elapsedTime := time.Since(startTime).Milliseconds()
		fmt.Printf("[CleanupExpiredShares] Completed cleanup of %d expired list shares (took %dms)\n",
			count, elapsedTime)
		return nil
	})

	if err != nil {
		elapsedTime := time.Since(startTime).Milliseconds()
		fmt.Printf("[CleanupExpiredShares] Failed to clean up expired shares (took %dms): %v\n",
			elapsedTime, err)
		return 0, err
	}

	return cleanedCount, nil
}

// MarkItemChosen marks an item as chosen and updates its stats
func (r *ListRepository) MarkItemChosen(itemID uuid.UUID) error {
	return r.UpdateItemStats(itemID, true)
}

// GetTribeListsWithContext gets all lists owned by or shared with a tribe, with explicit context for schema information
func (r *ListRepository) GetTribeListsWithContext(ctx context.Context, tribeID uuid.UUID) ([]*models.List, error) {
	// Create a timeout context if one wasn't provided
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Use stricter transaction options for better stability
	opts := DefaultTransactionOptions()
	opts.IsolationLevel = sql.LevelReadCommitted // Less strict isolation level for better performance

	var lists []*models.List

	// Get schema from context if available
	if schemaName := testutil.GetSchemaFromContext(ctx); schemaName != "" {
		fmt.Printf("GetTribeListsWithContext: Using schema from context: %s\n", schemaName)
	}

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Verify the search path is set correctly within this transaction
		var searchPath string
		err := tx.QueryRow("SHOW search_path").Scan(&searchPath)
		if err != nil {
			return fmt.Errorf("error checking search_path: %w", err)
		}

		// Fix the parentheses in the query to properly group conditions
		query := `
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.created_at, l.updated_at, l.deleted_at
			FROM lists l
			LEFT JOIN list_owners lo ON l.id = lo.list_id
			LEFT JOIN list_sharing ls ON l.id = ls.list_id
			WHERE ((lo.owner_id = $1 AND lo.owner_type = 'tribe' AND lo.deleted_at IS NULL)
				OR (ls.tribe_id = $1 AND ls.deleted_at IS NULL))
				AND l.deleted_at IS NULL
			ORDER BY l.created_at DESC`

		rows, err := tx.Query(query, tribeID)
		if err != nil {
			return fmt.Errorf("error getting tribe lists: %w", err)
		}

		// IMPORTANT: Get all the list IDs and basic info first before attempting to load related data
		var loadLists []*models.List
		for rows.Next() {
			list := &models.List{}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
				safeClose(rows)
				return fmt.Errorf("error scanning list: %w", err)
			}
			loadLists = append(loadLists, list)
		}

		safeClose(rows)

		// Now load related data for each list in a separate transaction
		for _, list := range loadLists {
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

// GetUserListsWithContext gets all lists owned by or shared with a user, with explicit context for schema information
func (r *ListRepository) GetUserListsWithContext(ctx context.Context, userID uuid.UUID) ([]*models.List, error) {
	// Create a timeout context if one wasn't provided
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Use stricter transaction options for better stability
	opts := DefaultTransactionOptions()
	opts.IsolationLevel = sql.LevelReadCommitted // Less strict isolation level for better performance

	var lists []*models.List

	// Get schema from context if available
	if schemaName := testutil.GetSchemaFromContext(ctx); schemaName != "" {
		fmt.Printf("GetUserListsWithContext: Using schema from context: %s\n", schemaName)
	}

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// Verify the search path is set correctly within this transaction
		var searchPath string
		err := tx.QueryRow("SHOW search_path").Scan(&searchPath)
		if err != nil {
			return fmt.Errorf("error checking search_path: %w", err)
		}

		// Get lists the user owns directly or via tribe membership
		query := `
			WITH user_tribes AS (
				SELECT tribe_id 
				FROM tribe_members 
				WHERE user_id = $1 AND deleted_at IS NULL
			)
			SELECT DISTINCT l.id, l.type, l.name, l.description, l.visibility,
				l.sync_status, l.sync_source, l.sync_id, l.last_sync_at,
				l.default_weight, l.max_items, l.cooldown_days,
				l.created_at, l.updated_at, l.deleted_at
			FROM lists l
			LEFT JOIN list_owners lo ON l.id = lo.list_id
			LEFT JOIN list_sharing ls ON l.id = ls.list_id
			WHERE (
				-- User owns the list directly
				(lo.owner_id = $1 AND lo.owner_type = 'user' AND lo.deleted_at IS NULL)
				-- List is owned by a tribe the user is a member of
				OR (lo.owner_type = 'tribe' AND lo.owner_id IN (SELECT tribe_id FROM user_tribes) AND lo.deleted_at IS NULL)
				-- List is shared with a tribe the user is a member of
				OR (ls.tribe_id IN (SELECT tribe_id FROM user_tribes) AND ls.deleted_at IS NULL)
			)
			AND l.deleted_at IS NULL
			ORDER BY l.created_at DESC`

		rows, err := tx.Query(query, userID)
		if err != nil {
			return fmt.Errorf("error getting user lists: %w", err)
		}

		// IMPORTANT: Get all the list IDs and basic info first before attempting to load related data
		var loadLists []*models.List
		for rows.Next() {
			list := &models.List{}
			err := rows.Scan(
				&list.ID, &list.Type, &list.Name, &list.Description, &list.Visibility,
				&list.SyncStatus, &list.SyncSource, &list.SyncID, &list.LastSyncAt,
				&list.DefaultWeight, &list.MaxItems, &list.CooldownDays,
				&list.CreatedAt, &list.UpdatedAt, &list.DeletedAt,
			)
			if err != nil {
				safeClose(rows)
				return fmt.Errorf("error scanning list: %w", err)
			}
			loadLists = append(loadLists, list)
		}

		safeClose(rows)

		// Now load related data for each list in a separate transaction
		for _, list := range loadLists {
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
			SELECT list_id, owner_id, owner_type, created_at, updated_at, deleted_at
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
				&owner.UpdatedAt,
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
	// Get test schema context if available
	var ctx context.Context
	if testSchema := testutil.GetCurrentTestSchema(); testSchema != "" {
		fmt.Printf("GetSharedTribes: Using schema %s (verification handled by TransactionManager)\n", testSchema)
		// Use context with schema for better isolation
		schemaCtx := testutil.GetSchemaContext(testSchema)
		if schemaCtx != nil {
			ctx = schemaCtx.WithContext(context.Background())
		} else {
			ctx = context.Background()
		}
	} else {
		ctx = context.Background()
	}

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
