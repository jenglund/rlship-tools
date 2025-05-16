package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

// TribeRepository implements models.TribeRepository using PostgreSQL
type TribeRepository struct {
	db *sql.DB
}

// NewTribeRepository creates a new PostgreSQL tribe repository
func NewTribeRepository(db *sql.DB) *TribeRepository {
	return &TribeRepository{db: db}
}

// Create inserts a new tribe into the database
func (r *TribeRepository) Create(tribe *models.Tribe) error {
	query := `
		INSERT INTO tribes (
			id, name, type, description, visibility,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	if tribe.ID == uuid.Nil {
		tribe.ID = uuid.New()
	}
	now := time.Now()
	tribe.CreatedAt = now
	tribe.UpdatedAt = now

	err := r.db.QueryRow(
		query,
		tribe.ID,
		tribe.Name,
		tribe.Type,
		tribe.Description,
		tribe.Visibility,
		tribe.Metadata,
		tribe.CreatedAt,
		tribe.UpdatedAt,
	).Scan(&tribe.ID)

	if err != nil {
		return fmt.Errorf("error creating tribe: %w", err)
	}

	return nil
}

// GetByID retrieves a tribe by its ID
func (r *TribeRepository) GetByID(id uuid.UUID) (*models.Tribe, error) {
	query := `
		SELECT 
			id, name, type, description, visibility,
			metadata, created_at, updated_at, deleted_at
		FROM tribes
		WHERE id = $1 AND deleted_at IS NULL`

	tribe := &models.Tribe{}
	err := r.db.QueryRow(query, id).Scan(
		&tribe.ID,
		&tribe.Name,
		&tribe.Type,
		&tribe.Description,
		&tribe.Visibility,
		&tribe.Metadata,
		&tribe.CreatedAt,
		&tribe.UpdatedAt,
		&tribe.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tribe not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tribe: %w", err)
	}

	// Get tribe members
	members, err := r.GetMembers(tribe.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting tribe members: %w", err)
	}
	tribe.Members = members

	return tribe, nil
}

// Update updates an existing tribe in the database
func (r *TribeRepository) Update(tribe *models.Tribe) error {
	query := `
		UPDATE tribes
		SET name = $1,
			type = $2,
			description = $3,
			visibility = $4,
			metadata = $5,
			updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
		RETURNING id`

	tribe.UpdatedAt = time.Now()

	err := r.db.QueryRow(
		query,
		tribe.Name,
		tribe.Type,
		tribe.Description,
		tribe.Visibility,
		tribe.Metadata,
		tribe.UpdatedAt,
		tribe.ID,
	).Scan(&tribe.ID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("tribe not found")
	}
	if err != nil {
		return fmt.Errorf("error updating tribe: %w", err)
	}

	return nil
}

// Delete soft-deletes a tribe
func (r *TribeRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE tribes
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), id)
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

	return nil
}

// List retrieves a paginated list of tribes
func (r *TribeRepository) List(offset, limit int) ([]*models.Tribe, error) {
	query := `
		SELECT id, name, created_at, updated_at, deleted_at
		FROM tribes
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing tribes: %w", err)
	}
	defer rows.Close()

	var tribes []*models.Tribe
	for rows.Next() {
		tribe := &models.Tribe{}
		err := rows.Scan(
			&tribe.ID,
			&tribe.Name,
			&tribe.CreatedAt,
			&tribe.UpdatedAt,
			&tribe.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tribe row: %w", err)
		}

		// Get tribe members
		members, err := r.GetMembers(tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tribe members: %w", err)
		}
		tribe.Members = members

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tribe rows: %w", err)
	}

	return tribes, nil
}

// AddMember adds a user to a tribe
func (r *TribeRepository) AddMember(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time, invitedBy *uuid.UUID) error {
	query := `
		INSERT INTO tribe_members (
			id, tribe_id, user_id, membership_type, display_name,
			expires_at, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Get user's name for display_name
	var displayName string
	err := r.db.QueryRow("SELECT name FROM users WHERE id = $1", userID).Scan(&displayName)
	if err != nil {
		return fmt.Errorf("error getting user name: %w", err)
	}

	now := time.Now()
	id := uuid.New()

	_, err = r.db.Exec(
		query,
		id,
		tribeID,
		userID,
		memberType,
		displayName,
		expiresAt,
		models.JSONMap{},
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("error adding tribe member: %w", err)
	}

	return nil
}

// UpdateMember updates a member's type and expiration
func (r *TribeRepository) UpdateMember(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time) error {
	query := `
		UPDATE tribe_members
		SET membership_type = $1,
			expires_at = $2,
			updated_at = $3
		WHERE tribe_id = $4 AND user_id = $5 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, memberType, expiresAt, time.Now(), tribeID, userID)
	if err != nil {
		return fmt.Errorf("error updating tribe member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tribe member not found")
	}

	return nil
}

// RemoveMember removes a user from a tribe
func (r *TribeRepository) RemoveMember(tribeID, userID uuid.UUID) error {
	query := `
		UPDATE tribe_members
		SET deleted_at = $1
		WHERE tribe_id = $2 AND user_id = $3 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), tribeID, userID)
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

	return nil
}

// GetMembers retrieves all members of a tribe
func (r *TribeRepository) GetMembers(tribeID uuid.UUID) ([]*models.TribeMember, error) {
	// First check if the tribe exists
	exists := false
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tribes WHERE id = $1 AND deleted_at IS NULL)", tribeID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking tribe existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("tribe not found")
	}

	query := `
		SELECT 
			tm.id, tm.tribe_id, tm.user_id, tm.membership_type, tm.display_name, 
			tm.expires_at, tm.metadata, tm.created_at, tm.updated_at, tm.deleted_at,
			u.firebase_uid, u.provider, u.email, u.name, u.avatar_url, u.last_login, u.created_at, u.updated_at
		FROM tribe_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.tribe_id = $1 AND tm.deleted_at IS NULL
		ORDER BY tm.created_at ASC`

	rows, err := r.db.Query(query, tribeID)
	if err != nil {
		return nil, fmt.Errorf("error getting tribe members: %w", err)
	}
	defer rows.Close()

	var members []*models.TribeMember
	for rows.Next() {
		member := &models.TribeMember{
			User: &models.User{},
		}
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tribe member rows: %w", err)
	}

	return members, nil
}

// GetExpiredGuestMemberships retrieves all expired guest memberships
func (r *TribeRepository) GetExpiredGuestMemberships() ([]*models.TribeMember, error) {
	query := `
		SELECT 
			id, tribe_id, user_id, membership_type,
			display_name, expires_at, metadata,
			created_at, updated_at, deleted_at
		FROM tribe_members
		WHERE membership_type = 'guest'
		AND expires_at < NOW()
		AND deleted_at IS NULL`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting expired guest memberships: %w", err)
	}
	defer rows.Close()

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

	return members, nil
}

// GetUserTribes retrieves all tribes a user is a member of
func (r *TribeRepository) GetUserTribes(userID uuid.UUID) ([]*models.Tribe, error) {
	query := `
		SELECT DISTINCT t.id, t.name, t.type, t.description, t.visibility,
			t.metadata, t.created_at, t.updated_at, t.deleted_at
		FROM tribes t
		INNER JOIN tribe_members tm ON t.id = tm.tribe_id
		WHERE tm.user_id = $1
		AND t.deleted_at IS NULL
		AND tm.deleted_at IS NULL`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user tribes: %w", err)
	}
	defer rows.Close()

	var tribes []*models.Tribe
	for rows.Next() {
		tribe := &models.Tribe{}
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
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user tribe: %w", err)
		}

		// Get tribe members
		members, err := r.GetMembers(tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tribe members: %w", err)
		}
		tribe.Members = members

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user tribes: %w", err)
	}

	return tribes, nil
}

// GetByType retrieves tribes by type
func (r *TribeRepository) GetByType(tribeType models.TribeType, offset, limit int) ([]*models.Tribe, error) {
	query := `
		SELECT id, name, type, description, visibility,
			metadata, created_at, updated_at, deleted_at
		FROM tribes
		WHERE type = $1
		AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, tribeType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting tribes by type: %w", err)
	}
	defer rows.Close()

	var tribes []*models.Tribe
	for rows.Next() {
		tribe := &models.Tribe{}
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
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tribe by type: %w", err)
		}

		// Get tribe members
		members, err := r.GetMembers(tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tribe members: %w", err)
		}
		tribe.Members = members

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tribes by type: %w", err)
	}

	return tribes, nil
}

// Search searches for tribes by name or description
func (r *TribeRepository) Search(query string, offset, limit int) ([]*models.Tribe, error) {
	sqlQuery := `
		SELECT id, name, type, description, visibility,
			metadata, created_at, updated_at, deleted_at
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
	defer rows.Close()

	var tribes []*models.Tribe
	for rows.Next() {
		tribe := &models.Tribe{}
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
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning searched tribe: %w", err)
		}

		// Get tribe members
		members, err := r.GetMembers(tribe.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting tribe members: %w", err)
		}
		tribe.Members = members

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating searched tribes: %w", err)
	}

	return tribes, nil
}
