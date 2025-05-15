package postgres

import (
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
}

// NewActivityRepository creates a new PostgreSQL activity repository
func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create inserts a new activity into the database
func (r *ActivityRepository) Create(activity *models.Activity) error {
	query := `
		INSERT INTO activities (id, type, name, description, visibility, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

	err := r.db.QueryRow(
		query,
		activity.ID,
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
}

// GetByID retrieves an activity by its ID
func (r *ActivityRepository) GetByID(id uuid.UUID) (*models.Activity, error) {
	query := `
		SELECT id, type, name, description, visibility, metadata, created_at, updated_at, deleted_at
		FROM activities
		WHERE id = $1 AND deleted_at IS NULL`

	activity := &models.Activity{}
	err := r.db.QueryRow(query, id).Scan(
		&activity.ID,
		&activity.Type,
		&activity.Name,
		&activity.Description,
		&activity.Visibility,
		&activity.Metadata,
		&activity.CreatedAt,
		&activity.UpdatedAt,
		&activity.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("activity not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting activity: %w", err)
	}

	return activity, nil
}

// Update updates an existing activity in the database
func (r *ActivityRepository) Update(activity *models.Activity) error {
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

	activity.UpdatedAt = time.Now()

	err := r.db.QueryRow(
		query,
		activity.Type,
		activity.Name,
		activity.Description,
		activity.Visibility,
		activity.Metadata,
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
}

// Delete soft-deletes an activity
func (r *ActivityRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE activities
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), id)
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
}

// List retrieves a paginated list of non-deleted activities
func (r *ActivityRepository) List(offset, limit int) ([]*models.Activity, error) {
	query := `
		SELECT id, type, name, description, visibility, metadata, created_at, updated_at, deleted_at
		FROM activities
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		err := rows.Scan(
			&activity.ID,
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
		err := rows.Scan(
			&activity.ID,
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
		err := rows.Scan(
			&activity.ID,
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

// ShareWithTribe shares an activity with a tribe
func (r *ActivityRepository) ShareWithTribe(activityID, tribeID, userID uuid.UUID, expiresAt *time.Time) error {
	query := `
		INSERT INTO activity_shares (activity_id, tribe_id, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(query, activityID, tribeID, userID, time.Now(), expiresAt)
	if err != nil {
		return fmt.Errorf("error sharing activity: %w", err)
	}

	return nil
}

// UnshareWithTribe removes an activity share from a tribe
func (r *ActivityRepository) UnshareWithTribe(activityID, tribeID uuid.UUID) error {
	query := `
		UPDATE activity_shares
		SET deleted_at = $1
		WHERE activity_id = $2 AND tribe_id = $3 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), activityID, tribeID)
	if err != nil {
		return fmt.Errorf("error unsharing activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("activity share not found")
	}

	return nil
}

// GetSharedActivities retrieves all activities shared with a tribe
func (r *ActivityRepository) GetSharedActivities(tribeID uuid.UUID) ([]*models.Activity, error) {
	query := `
		SELECT a.id, a.type, a.name, a.description, a.visibility, a.metadata, a.created_at, a.updated_at, a.deleted_at
		FROM activities a
		JOIN activity_shares ash ON ash.activity_id = a.id
		WHERE ash.tribe_id = $1 
		AND ash.deleted_at IS NULL
		AND (ash.expires_at IS NULL OR ash.expires_at > NOW())
		AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC`

	rows, err := r.db.Query(query, tribeID)
	if err != nil {
		return nil, fmt.Errorf("error getting shared activities: %w", err)
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		activity := &models.Activity{}
		err := rows.Scan(
			&activity.ID,
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

// MarkForDeletion marks an activity for deletion
func (r *ActivityRepository) MarkForDeletion(activityID uuid.UUID) error {
	return r.Delete(activityID)
}

// CleanupOrphanedActivities removes activities that have no owners
func (r *ActivityRepository) CleanupOrphanedActivities() error {
	query := `
		UPDATE activities a
		SET deleted_at = NOW()
		WHERE NOT EXISTS (
			SELECT 1 FROM activity_owners ao
			WHERE ao.activity_id = a.id
			AND ao.deleted_at IS NULL
		)
		AND a.deleted_at IS NULL`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error cleaning up orphaned activities: %w", err)
	}

	return nil
}
