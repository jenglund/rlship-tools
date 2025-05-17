package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// ActivityRepository implements models.ActivityRepository using PostgreSQL
type ActivityRepository struct {
	db *sql.DB
	tm *TransactionManager
}

// NewActivityRepository creates a new PostgreSQL activity repository
func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{
		db: db,
		tm: NewTransactionManager(db),
	}
}

// Create inserts a new activity into the database
func (r *ActivityRepository) Create(activity *models.Activity) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			INSERT INTO activities (id, user_id, type, name, description, visibility, metadata, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`

		if activity.ID == uuid.Nil {
			activity.ID = uuid.New()
		}

		// Convert metadata to JSON
		var metadataJSON interface{} = nil
		if activity.Metadata != nil {
			var err error
			jsonBytes, err := json.Marshal(activity.Metadata)
			if err != nil {
				return fmt.Errorf("error encoding metadata: %w", err)
			}
			metadataJSON = jsonBytes
		}

		err := tx.QueryRow(
			query,
			activity.ID,
			activity.UserID,
			activity.Type,
			activity.Name,
			activity.Description,
			activity.Visibility,
			metadataJSON,
			activity.CreatedAt,
			activity.UpdatedAt,
		).Scan(&activity.ID)

		if err != nil {
			return fmt.Errorf("error creating activity: %w", err)
		}

		return nil
	})
}

// convertMetadata is a helper function to convert JSON metadata and handle array type conversions
func convertMetadata(metadata []byte) (map[string]interface{}, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(metadata, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetByID retrieves an activity by its ID
func (r *ActivityRepository) GetByID(id uuid.UUID) (*models.Activity, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var activity *models.Activity

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, type, name, description, visibility, metadata, created_at, updated_at, deleted_at
			FROM activities
			WHERE id = $1 AND deleted_at IS NULL`

		var metadata []byte
		activity = &models.Activity{}
		err := tx.QueryRow(query, id).Scan(
			&activity.ID,
			&activity.Type,
			&activity.Name,
			&activity.Description,
			&activity.Visibility,
			&metadata,
			&activity.CreatedAt,
			&activity.UpdatedAt,
			&activity.DeletedAt,
		)

		if err == sql.ErrNoRows {
			return fmt.Errorf("activity not found")
		}
		if err != nil {
			return fmt.Errorf("error getting activity: %w", err)
		}

		// Convert metadata
		activity.Metadata, err = convertMetadata(metadata)
		if err != nil {
			return fmt.Errorf("error converting metadata: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return activity, nil
}

// Update updates an existing activity
func (r *ActivityRepository) Update(activity *models.Activity) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE activities
			SET type = $1,
				name = $2,
				description = $3,
				visibility = $4,
				metadata = $5,
				updated_at = $6
			WHERE id = $7 AND deleted_at IS NULL
			RETURNING id`

		// Convert metadata to JSON
		var metadataJSON interface{} = nil
		if activity.Metadata != nil {
			var err error
			jsonBytes, err := json.Marshal(activity.Metadata)
			if err != nil {
				return fmt.Errorf("error encoding metadata: %w", err)
			}
			metadataJSON = jsonBytes
		}

		activity.UpdatedAt = time.Now()

		err := tx.QueryRow(
			query,
			activity.Type,
			activity.Name,
			activity.Description,
			activity.Visibility,
			metadataJSON,
			activity.UpdatedAt,
			activity.ID,
		).Scan(&activity.ID)

		if err == sql.ErrNoRows {
			return fmt.Errorf("activity not found")
		}
		if err != nil {
			return fmt.Errorf("error updating activity: %w", err)
		}

		return nil
	})
}

// Delete soft-deletes an activity
func (r *ActivityRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE activities
			SET deleted_at = $1
			WHERE id = $2 AND deleted_at IS NULL`

		result, err := tx.Exec(query, time.Now(), id)
		if err != nil {
			return fmt.Errorf("error deleting activity: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error checking rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("activity not found")
		}

		return nil
	})
}

// List retrieves a paginated list of non-deleted activities
func (r *ActivityRepository) List(offset, limit int) ([]*models.Activity, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var activities []*models.Activity

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, type, name, description, visibility, metadata, created_at, updated_at, deleted_at
			FROM activities
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

		rows, err := tx.Query(query, limit, offset)
		if err != nil {
			return fmt.Errorf("error listing activities: %w", err)
		}
		defer rows.Close()

		activities = make([]*models.Activity, 0)
		for rows.Next() {
			activity := &models.Activity{}
			var metadata []byte
			err := rows.Scan(
				&activity.ID,
				&activity.Type,
				&activity.Name,
				&activity.Description,
				&activity.Visibility,
				&metadata,
				&activity.CreatedAt,
				&activity.UpdatedAt,
				&activity.DeletedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning activity row: %w", err)
			}

			// Convert metadata
			activity.Metadata, err = convertMetadata(metadata)
			if err != nil {
				return fmt.Errorf("error converting metadata: %w", err)
			}

			activities = append(activities, activity)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating activity rows: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return activities, nil
}

// AddOwner adds an owner to an activity
func (r *ActivityRepository) AddOwner(activityID, ownerID uuid.UUID, ownerType string) error {
	query := `
		INSERT INTO activity_owners (activity_id, owner_id, owner_type, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(query, activityID, ownerID, ownerType, time.Now())
	if err != nil {
		return fmt.Errorf("error adding activity owner: %w", err)
	}

	return nil
}

// RemoveOwner removes an owner from an activity
func (r *ActivityRepository) RemoveOwner(activityID, ownerID uuid.UUID) error {
	query := `
		UPDATE activity_owners
		SET deleted_at = $1
		WHERE activity_id = $2 AND owner_id = $3 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), activityID, ownerID)
	if err != nil {
		return fmt.Errorf("error removing activity owner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("activity owner not found")
	}

	return nil
}

// GetOwners retrieves all owners of an activity
func (r *ActivityRepository) GetOwners(activityID uuid.UUID) ([]*models.ActivityOwner, error) {
	query := `
		SELECT activity_id, owner_id, owner_type, created_at, deleted_at
		FROM activity_owners
		WHERE activity_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.Query(query, activityID)
	if err != nil {
		return nil, fmt.Errorf("error getting activity owners: %w", err)
	}
	defer rows.Close()

	var owners []*models.ActivityOwner
	for rows.Next() {
		owner := &models.ActivityOwner{}
		err := rows.Scan(
			&owner.ActivityID,
			&owner.OwnerID,
			&owner.OwnerType,
			&owner.CreatedAt,
			&owner.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning activity owner row: %w", err)
		}
		owners = append(owners, owner)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activity owner rows: %w", err)
	}

	return owners, nil
}

// GetUserActivities retrieves all activities owned by a user
func (r *ActivityRepository) GetUserActivities(userID uuid.UUID) ([]*models.Activity, error) {
	query := `
		SELECT a.id, a.type, a.name, a.description, a.visibility, a.metadata, a.created_at, a.updated_at, a.deleted_at
		FROM activities a
		JOIN activity_owners ao ON ao.activity_id = a.id
		WHERE ao.owner_id = $1 
		AND ao.owner_type = 'user'
		AND ao.deleted_at IS NULL
		AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		var metadataBytes []byte
		err := rows.Scan(
			&activity.ID,
			&activity.Type,
			&activity.Name,
			&activity.Description,
			&activity.Visibility,
			&metadataBytes,
			&activity.CreatedAt,
			&activity.UpdatedAt,
			&activity.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning activity row: %w", err)
		}

		metadata, err := convertMetadata(metadataBytes)
		if err != nil {
			return nil, err
		}
		activity.Metadata = metadata

		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activity rows: %w", err)
	}

	return activities, nil
}

// GetTribeActivities retrieves all activities owned by a tribe
func (r *ActivityRepository) GetTribeActivities(tribeID uuid.UUID) ([]*models.Activity, error) {
	query := `
		SELECT a.id, a.type, a.name, a.description, a.visibility, a.metadata, a.created_at, a.updated_at, a.deleted_at
		FROM activities a
		JOIN activity_owners ao ON ao.activity_id = a.id
		WHERE ao.owner_id = $1 
		AND ao.owner_type = 'tribe'
		AND ao.deleted_at IS NULL
		AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC`

	rows, err := r.db.Query(query, tribeID)
	if err != nil {
		return nil, fmt.Errorf("error getting tribe activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		var metadataBytes []byte
		err := rows.Scan(
			&activity.ID,
			&activity.Type,
			&activity.Name,
			&activity.Description,
			&activity.Visibility,
			&metadataBytes,
			&activity.CreatedAt,
			&activity.UpdatedAt,
			&activity.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning activity row: %w", err)
		}

		metadata, err := convertMetadata(metadataBytes)
		if err != nil {
			return nil, err
		}
		activity.Metadata = metadata

		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activity rows: %w", err)
	}

	return activities, nil
}

// ShareWithTribe shares an activity with a tribe
func (r *ActivityRepository) ShareWithTribe(activityID, tribeID, userID uuid.UUID, expiresAt *time.Time) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First, verify the activity exists and is not deleted
		query := `
			SELECT 1
			FROM activities
			WHERE id = $1 AND deleted_at IS NULL`

		var exists bool
		err := tx.QueryRow(query, activityID).Scan(&exists)
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("error checking activity existence: %w", err)
		}

		// Verify the tribe exists and is not deleted
		query = `
			SELECT 1
			FROM tribes
			WHERE id = $1 AND deleted_at IS NULL`

		err = tx.QueryRow(query, tribeID).Scan(&exists)
		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("error checking tribe existence: %w", err)
		}

		now := time.Now()

		// First, soft delete any existing shares
		deleteQuery := `
			UPDATE activity_shares
			SET deleted_at = $1,
				updated_at = $1
			WHERE activity_id = $2 
			AND tribe_id = $3 
			AND deleted_at IS NULL`

		_, err = tx.Exec(deleteQuery, now, activityID, tribeID)
		if err != nil {
			return fmt.Errorf("error removing old shares: %w", err)
		}

		// Create new share
		insertQuery := `
			INSERT INTO activity_shares (
				activity_id, tribe_id, user_id, expires_at,
				created_at, updated_at, deleted_at
			) 
			VALUES ($1, $2, $3, $4, $5, $5, NULL)`

		result, err := tx.Exec(insertQuery, activityID, tribeID, userID, expiresAt, now)
		if err != nil {
			return fmt.Errorf("error creating activity share: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error checking rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("failed to create share for activity %s and tribe %s", activityID, tribeID)
		}

		// Add tribe as an owner of the activity
		ownerQuery := `
			INSERT INTO activity_owners (activity_id, owner_id, owner_type, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (activity_id, owner_id) DO UPDATE
			SET deleted_at = NULL,
				updated_at = $4
			WHERE activity_owners.activity_id = $1
			AND activity_owners.owner_id = $2`

		_, err = tx.Exec(ownerQuery, activityID, tribeID, models.OwnerTypeTribe, now)
		if err != nil {
			return fmt.Errorf("error adding tribe as activity owner: %w", err)
		}

		return nil
	})
}

// UnshareWithTribe removes an activity share from a tribe
func (r *ActivityRepository) UnshareWithTribe(activityID, tribeID uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		now := time.Now()

		// First, soft delete the share
		shareQuery := `
			UPDATE activity_shares
			SET deleted_at = $1,
				updated_at = $1
			WHERE activity_id = $2 
			AND tribe_id = $3 
			AND deleted_at IS NULL`

		result, err := tx.Exec(shareQuery, now, activityID, tribeID)
		if err != nil {
			return fmt.Errorf("error unsharing activity: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error checking rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("no active share found for activity %s and tribe %s", activityID, tribeID)
		}

		// Remove tribe ownership
		ownerQuery := `
			UPDATE activity_owners
			SET deleted_at = $1,
				updated_at = $1
			WHERE activity_id = $2 
			AND owner_id = $3
			AND owner_type = $4
			AND deleted_at IS NULL`

		_, err = tx.Exec(ownerQuery, now, activityID, tribeID, models.OwnerTypeTribe)
		if err != nil {
			return fmt.Errorf("error removing tribe ownership: %w", err)
		}

		return nil
	})
}

// GetSharedActivities retrieves all activities shared with a tribe
func (r *ActivityRepository) GetSharedActivities(tribeID uuid.UUID) ([]*models.Activity, error) {
	query := `
		WITH latest_shares AS (
			SELECT DISTINCT ON (activity_id) 
				activity_id, deleted_at, expires_at
			FROM activity_shares
			WHERE tribe_id = $1
			ORDER BY activity_id, updated_at DESC
		),
		valid_shares AS (
			SELECT activity_id
			FROM latest_shares
			WHERE deleted_at IS NULL
				AND (expires_at IS NULL OR expires_at > NOW())
		)
		SELECT DISTINCT
			a.id, a.user_id, a.type, a.name, a.description, a.visibility, 
			COALESCE(a.metadata, '{}'::jsonb) as metadata,
			a.created_at, a.updated_at, a.deleted_at
		FROM activities a
		INNER JOIN valid_shares s ON s.activity_id = a.id
		WHERE a.deleted_at IS NULL
		ORDER BY a.created_at DESC`

	rows, err := r.db.Query(query, tribeID)
	if err != nil {
		return nil, fmt.Errorf("error querying shared activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Type,
			&activity.Name,
			&activity.Description,
			&activity.Visibility,
			&activity.Metadata,
			&activity.CreatedAt,
			&activity.UpdatedAt,
			&activity.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning activity row: %w", err)
		}

		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activity rows: %w", err)
	}

	return activities, nil
}
