package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// TribeRepository implements models.TribeRepository using PostgreSQL
type TribeRepository struct {
	db *sql.DB
	tm *TransactionManager
}

// NewTribeRepository creates a new PostgreSQL tribe repository
func NewTribeRepository(db *sql.DB) *TribeRepository {
	return &TribeRepository{
		db: db,
		tm: NewTransactionManager(db),
	}
}

// Create inserts a new tribe into the database
func (r *TribeRepository) Create(tribe *models.Tribe) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			INSERT INTO tribes (
				id, name, type, description, visibility,
				metadata, created_at, updated_at, version
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING version`

		if tribe.ID == uuid.Nil {
			tribe.ID = uuid.New()
		}
		now := time.Now()
		tribe.CreatedAt = now
		tribe.UpdatedAt = now
		tribe.Version = 1

		err := tx.QueryRow(
			query,
			tribe.ID,
			tribe.Name,
			tribe.Type,
			tribe.Description,
			tribe.Visibility,
			tribe.Metadata,
			tribe.CreatedAt,
			tribe.UpdatedAt,
			tribe.Version,
		).Scan(&tribe.Version)

		if err != nil {
			return fmt.Errorf("error creating tribe: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a tribe by its ID
func (r *TribeRepository) GetByID(id uuid.UUID) (*models.Tribe, error) {
	// Start transaction for consistent read
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	query := `
		SELECT 
			id, name, type, description, visibility,
			metadata, created_at, updated_at, deleted_at, version
		FROM tribes
		WHERE id = $1 AND deleted_at IS NULL`

	tribe := &models.Tribe{
		Members: []*models.TribeMember{},
	}
	err = tx.QueryRow(query, id).Scan(
		&tribe.ID,
		&tribe.Name,
		&tribe.Type,
		&tribe.Description,
		&tribe.Visibility,
		&tribe.Metadata,
		&tribe.CreatedAt,
		&tribe.UpdatedAt,
		&tribe.DeletedAt,
		&tribe.Version,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tribe not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tribe: %w", err)
	}

	// Get members directly instead of using getMembersWithTx
	membersQuery := `
		SELECT tm.id, tm.tribe_id, tm.user_id, tm.membership_type, COALESCE(tm.display_name, 'Member') as display_name, 
			   tm.expires_at, tm.metadata, tm.created_at, tm.updated_at, tm.deleted_at,
			   u.firebase_uid, u.provider, u.email, u.name, u.avatar_url, u.last_login, u.created_at, u.updated_at
		FROM tribe_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.tribe_id = $1 AND tm.deleted_at IS NULL
		ORDER BY tm.created_at ASC`

	memberRows, err := tx.Query(membersQuery, tribe.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting tribe members: %w", err)
	}
	defer safeClose(memberRows)

	var members []*models.TribeMember
	for memberRows.Next() {
		member := &models.TribeMember{
			User: &models.User{},
		}
		err := memberRows.Scan(
			&member.ID,
			&member.TribeID,
			&member.UserID,
			&member.MembershipType,
			&member.DisplayName,
			&member.ExpiresAt,
			&member.Metadata,
			&member.CreatedAt,
			&member.UpdatedAt,
			&member.DeletedAt,
			&member.User.FirebaseUID,
			&member.User.Provider,
			&member.User.Email,
			&member.User.Name,
			&member.User.AvatarURL,
			&member.User.LastLogin,
			&member.User.CreatedAt,
			&member.User.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tribe member row: %w", err)
		}
		member.User.ID = member.UserID
		members = append(members, member)
	}

	if err = memberRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tribe member rows: %w", err)
	}

	tribe.Members = members

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return tribe, nil
}

// Update updates an existing tribe in the database
func (r *TribeRepository) Update(tribe *models.Tribe) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE tribes
			SET name = $1,
				type = $2,
				description = $3,
				visibility = $4,
				metadata = $5,
				updated_at = $6
			WHERE id = $7 
			AND version = $8
			AND deleted_at IS NULL
			RETURNING version`

		tribe.UpdatedAt = time.Now()
		var newVersion int

		err := tx.QueryRow(
			query,
			tribe.Name,
			tribe.Type,
			tribe.Description,
			tribe.Visibility,
			tribe.Metadata,
			tribe.UpdatedAt,
			tribe.ID,
			tribe.Version,
		).Scan(&newVersion)

		if err == sql.ErrNoRows {
			return fmt.Errorf("tribe not found or was modified by another user")
		}
		if err != nil {
			return fmt.Errorf("error updating tribe: %w", err)
		}

		tribe.Version = newVersion
		return nil
	})
}

// Delete soft-deletes a tribe and its members
func (r *TribeRepository) Delete(id uuid.UUID) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	now := time.Now()

	// Soft delete tribe
	query := `
		UPDATE tribes
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := tx.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("error deleting tribe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tribe not found")
	}

	// Soft delete tribe members
	query = `
		UPDATE tribe_members
		SET deleted_at = $1
		WHERE tribe_id = $2 AND deleted_at IS NULL`

	_, err = tx.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("error deleting tribe members: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// List retrieves all tribes
func (r *TribeRepository) List(offset, limit int) ([]*models.Tribe, error) {
	// Start transaction for consistent read
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	// First, get basic tribe information with a simple query
	query := `
		SELECT id, name, type, description, visibility,
		       metadata, created_at, updated_at, deleted_at, version
		FROM tribes
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := tx.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing tribes: %w", err)
	}
	defer safeClose(rows)

	var tribes []*models.Tribe

	for rows.Next() {
		tribe := &models.Tribe{
			Members: []*models.TribeMember{},
		}

		err := rows.Scan(
			&tribe.ID,
			&tribe.Name,
			&tribe.Type,
			&tribe.Description,
			&tribe.Visibility,
			&tribe.Metadata,
			&tribe.CreatedAt,
			&tribe.UpdatedAt,
			&tribe.DeletedAt,
			&tribe.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tribe: %w", err)
		}

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tribes: %w", err)
	}

	// If no tribes were found, return empty slice
	if len(tribes) == 0 {
		return tribes, nil
	}

	// For each tribe, get its members in a separate query to avoid the parse error
	for i, tribe := range tribes {
		membersQuery := `
			SELECT tm.id, tm.user_id, tm.tribe_id, COALESCE(tm.display_name, 'Member') as display_name, 
				   tm.membership_type, tm.metadata, tm.expires_at,
				   tm.created_at, tm.updated_at, tm.deleted_at, tm.version
			FROM tribe_members tm
			WHERE tm.tribe_id = $1
			AND tm.deleted_at IS NULL`

		memberRows, err := tx.Query(membersQuery, tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tribe members: %w", err)
		}
		defer safeClose(memberRows)

		for memberRows.Next() {
			member := &models.TribeMember{}

			err := memberRows.Scan(
				&member.ID,
				&member.UserID,
				&member.TribeID,
				&member.DisplayName,
				&member.MembershipType,
				&member.Metadata,
				&member.ExpiresAt,
				&member.CreatedAt,
				&member.UpdatedAt,
				&member.DeletedAt,
				&member.Version,
			)
			if err != nil {
				safeClose(memberRows)
				return nil, fmt.Errorf("error scanning tribe member: %w", err)
			}

			tribes[i].Members = append(tribes[i].Members, member)
		}

		if err = memberRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating tribe members: %w", err)
		}
		safeClose(memberRows)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return tribes, nil
}

// AddMember adds a user to a tribe
func (r *TribeRepository) AddMember(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time, invitedBy *uuid.UUID) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	// Get user's name for display_name
	var displayName sql.NullString
	err = tx.QueryRow("SELECT name FROM users WHERE id = $1", userID).Scan(&displayName)
	if err != nil {
		return fmt.Errorf("error getting user name: %w", err)
	}

	// Use displayName if valid, otherwise use a default name
	finalDisplayName := "Member"
	if displayName.Valid && displayName.String != "" {
		finalDisplayName = displayName.String
	}

	query := `
		INSERT INTO tribe_members (
			id, tribe_id, user_id, membership_type, display_name,
			expires_at, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	now := time.Now()
	id := uuid.New()

	_, err = tx.Exec(
		query,
		id,
		tribeID,
		userID,
		memberType,
		finalDisplayName,
		expiresAt,
		models.JSONMap{},
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("error adding tribe member: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// UpdateMember updates a member's type and expiration
func (r *TribeRepository) UpdateMember(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		// First get the current version
		var currentVersion int
		err := tx.QueryRow("SELECT version FROM tribe_members WHERE tribe_id = $1 AND user_id = $2 AND deleted_at IS NULL", tribeID, userID).Scan(&currentVersion)
		if err == sql.ErrNoRows {
			return fmt.Errorf("tribe member not found")
		}
		if err != nil {
			return fmt.Errorf("error getting tribe member version: %w", err)
		}

		query := `
			UPDATE tribe_members
			SET membership_type = $1,
				expires_at = $2,
				updated_at = $3
			WHERE tribe_id = $4 
			AND user_id = $5 
			AND version = $6
			AND deleted_at IS NULL
			RETURNING version`

		var newVersion int
		err = tx.QueryRow(query, memberType, expiresAt, time.Now(), tribeID, userID, currentVersion).Scan(&newVersion)
		if err == sql.ErrNoRows {
			return fmt.Errorf("tribe member was modified by another user")
		}
		if err != nil {
			return fmt.Errorf("error updating tribe member: %w", err)
		}

		return nil
	})
}

// RemoveMember removes a user from a tribe
func (r *TribeRepository) RemoveMember(tribeID, userID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	query := `
		UPDATE tribe_members
		SET deleted_at = $1
		WHERE tribe_id = $2 AND user_id = $3 AND deleted_at IS NULL`

	result, err := tx.Exec(query, time.Now(), tribeID, userID)
	if err != nil {
		return fmt.Errorf("error removing tribe member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tribe member not found")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetMembers retrieves all members of a tribe
func (r *TribeRepository) GetMembers(tribeID uuid.UUID) ([]*models.TribeMember, error) {
	// Start transaction for consistent read
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	// First check if the tribe exists
	exists := false
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM tribes WHERE id = $1 AND deleted_at IS NULL)", tribeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking if tribe exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("tribe not found")
	}

	// Get members directly with a query similar to other methods
	membersQuery := `
		SELECT tm.id, tm.tribe_id, tm.user_id, tm.membership_type, COALESCE(tm.display_name, 'Member') as display_name, 
			   tm.expires_at, tm.metadata, tm.created_at, tm.updated_at, tm.deleted_at,
			   u.firebase_uid, u.provider, u.email, u.name, u.avatar_url, u.last_login, u.created_at, u.updated_at
		FROM tribe_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.tribe_id = $1 AND tm.deleted_at IS NULL
		ORDER BY tm.created_at ASC`

	memberRows, err := tx.Query(membersQuery, tribeID)
	if err != nil {
		return nil, fmt.Errorf("error getting tribe members: %w", err)
	}
	defer safeClose(memberRows)

	var members []*models.TribeMember
	for memberRows.Next() {
		member := &models.TribeMember{
			User: &models.User{},
		}
		err := memberRows.Scan(
			&member.ID,
			&member.TribeID,
			&member.UserID,
			&member.MembershipType,
			&member.DisplayName,
			&member.ExpiresAt,
			&member.Metadata,
			&member.CreatedAt,
			&member.UpdatedAt,
			&member.DeletedAt,
			&member.User.FirebaseUID,
			&member.User.Provider,
			&member.User.Email,
			&member.User.Name,
			&member.User.AvatarURL,
			&member.User.LastLogin,
			&member.User.CreatedAt,
			&member.User.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tribe member row: %w", err)
		}
		member.User.ID = member.UserID
		members = append(members, member)
	}

	if err = memberRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tribe member rows: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return members, nil
}

// GetExpiredGuestMemberships retrieves all expired guest memberships
func (r *TribeRepository) GetExpiredGuestMemberships() ([]*models.TribeMember, error) {
	// Start transaction for consistent read
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	query := `
		SELECT 
			id, tribe_id, user_id, membership_type,
			display_name, expires_at, metadata,
			created_at, updated_at, deleted_at
		FROM tribe_members
		WHERE membership_type = 'guest'
		AND expires_at < NOW()
		AND deleted_at IS NULL`

	rows, err := tx.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting expired guest memberships: %w", err)
	}
	defer safeClose(rows)

	var members []*models.TribeMember
	for rows.Next() {
		member := &models.TribeMember{}
		err := rows.Scan(
			&member.ID,
			&member.TribeID,
			&member.UserID,
			&member.MembershipType,
			&member.DisplayName,
			&member.ExpiresAt,
			&member.Metadata,
			&member.CreatedAt,
			&member.UpdatedAt,
			&member.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning expired guest membership: %w", err)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired guest memberships: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return members, nil
}

// GetUserTribes retrieves all tribes a user is a member of
func (r *TribeRepository) GetUserTribes(userID uuid.UUID) ([]*models.Tribe, error) {
	// Start transaction for consistent read
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer safeClose(tx)

	query := `
		SELECT DISTINCT t.id, t.name, t.type, t.description, t.visibility,
			t.metadata, t.created_at, t.updated_at, t.deleted_at, t.version
		FROM tribes t
		INNER JOIN tribe_members tm ON t.id = tm.tribe_id
		WHERE tm.user_id = $1
		AND t.deleted_at IS NULL
		AND tm.deleted_at IS NULL`

	rows, err := tx.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user tribes: %w", err)
	}
	defer safeClose(rows)

	var tribes []*models.Tribe
	for rows.Next() {
		tribe := &models.Tribe{
			Members: []*models.TribeMember{},
		}
		err := rows.Scan(
			&tribe.ID,
			&tribe.Name,
			&tribe.Type,
			&tribe.Description,
			&tribe.Visibility,
			&tribe.Metadata,
			&tribe.CreatedAt,
			&tribe.UpdatedAt,
			&tribe.DeletedAt,
			&tribe.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user tribe: %w", err)
		}

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user tribes: %w", err)
	}

	// If no tribes were found, return empty slice
	if len(tribes) == 0 {
		return tribes, nil
	}

	// For each tribe, get its members in a separate query to avoid the parse error
	for i, tribe := range tribes {
		membersQuery := `
			SELECT tm.id, tm.user_id, tm.tribe_id, COALESCE(tm.display_name, 'Member') as display_name, 
				   tm.membership_type, tm.metadata, tm.expires_at,
				   tm.created_at, tm.updated_at, tm.deleted_at, tm.version
			FROM tribe_members tm
			WHERE tm.tribe_id = $1
			AND tm.deleted_at IS NULL`

		memberRows, err := tx.Query(membersQuery, tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tribe members: %w", err)
		}
		defer safeClose(memberRows)

		for memberRows.Next() {
			member := &models.TribeMember{}

			err := memberRows.Scan(
				&member.ID,
				&member.UserID,
				&member.TribeID,
				&member.DisplayName,
				&member.MembershipType,
				&member.Metadata,
				&member.ExpiresAt,
				&member.CreatedAt,
				&member.UpdatedAt,
				&member.DeletedAt,
				&member.Version,
			)
			if err != nil {
				safeClose(memberRows)
				return nil, fmt.Errorf("error scanning tribe member: %w", err)
			}

			tribes[i].Members = append(tribes[i].Members, member)
		}

		if err = memberRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating tribe members: %w", err)
		}
		safeClose(memberRows)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return tribes, nil
}

// GetByType retrieves tribes by type
func (r *TribeRepository) GetByType(tribeType models.TribeType, offset, limit int) ([]*models.Tribe, error) {
	query := `
		SELECT id, name, type, description, visibility,
			metadata, created_at, updated_at, deleted_at, version
		FROM tribes
		WHERE type = $1
		AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, tribeType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting tribes by type: %w", err)
	}
	defer safeClose(rows)

	var tribes []*models.Tribe
	for rows.Next() {
		tribe := &models.Tribe{
			Members: []*models.TribeMember{},
		}
		err := rows.Scan(
			&tribe.ID,
			&tribe.Name,
			&tribe.Type,
			&tribe.Description,
			&tribe.Visibility,
			&tribe.Metadata,
			&tribe.CreatedAt,
			&tribe.UpdatedAt,
			&tribe.DeletedAt,
			&tribe.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tribe by type: %w", err)
		}

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tribes by type: %w", err)
	}

	// Load members for each tribe separately without using a transaction
	for _, tribe := range tribes {
		members, err := r.GetMembers(tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error loading members for tribe %s: %w", tribe.ID, err)
		}
		tribe.Members = members
	}

	return tribes, nil
}

// Search searches for tribes by name or description
func (r *TribeRepository) Search(query string, offset, limit int) ([]*models.Tribe, error) {
	sqlQuery := `
		SELECT id, name, type, description, visibility,
			metadata, created_at, updated_at, deleted_at, version
		FROM tribes
		WHERE (name ILIKE $1 OR description ILIKE $1)
		AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(sqlQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error searching tribes: %w", err)
	}
	defer safeClose(rows)

	var tribes []*models.Tribe
	for rows.Next() {
		tribe := &models.Tribe{
			Members: []*models.TribeMember{},
		}
		err := rows.Scan(
			&tribe.ID,
			&tribe.Name,
			&tribe.Type,
			&tribe.Description,
			&tribe.Visibility,
			&tribe.Metadata,
			&tribe.CreatedAt,
			&tribe.UpdatedAt,
			&tribe.DeletedAt,
			&tribe.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning searched tribe: %w", err)
		}

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating searched tribes: %w", err)
	}

	// Load members for each tribe separately without using a transaction
	for _, tribe := range tribes {
		members, err := r.GetMembers(tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error loading members for tribe %s: %w", tribe.ID, err)
		}
		tribe.Members = members
	}

	return tribes, nil
}
