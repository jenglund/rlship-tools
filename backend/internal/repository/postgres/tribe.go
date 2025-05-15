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
		INSERT INTO tribes (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
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
		SELECT id, name, created_at, updated_at, deleted_at
		FROM tribes
		WHERE id = $1 AND deleted_at IS NULL`

	tribe := &models.Tribe{}
	err := r.db.QueryRow(query, id).Scan(
		&tribe.ID,
		&tribe.Name,
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
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING id`

	tribe.UpdatedAt = time.Now()

	err := r.db.QueryRow(
		query,
		tribe.Name,
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
		INSERT INTO tribe_members (tribe_id, user_id, type, joined_at, expires_at, invited_by)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query, tribeID, userID, memberType, time.Now(), expiresAt, invitedBy)
	if err != nil {
		return fmt.Errorf("error adding tribe member: %w", err)
	}

	return nil
}

// UpdateMember updates a member's type and expiration
func (r *TribeRepository) UpdateMember(tribeID, userID uuid.UUID, memberType models.MembershipType, expiresAt *time.Time) error {
	query := `
		UPDATE tribe_members
		SET type = $1,
			expires_at = $2
		WHERE tribe_id = $3 AND user_id = $4`

	result, err := r.db.Exec(query, memberType, expiresAt, tribeID, userID)
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
	query := `DELETE FROM tribe_members WHERE tribe_id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, tribeID, userID)
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
		SELECT tm.tribe_id, tm.user_id, tm.type, tm.joined_at, tm.expires_at, tm.invited_by,
			   u.firebase_uid, u.provider, u.email, u.name, u.avatar_url, u.last_login, u.created_at, u.updated_at
		FROM tribe_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.tribe_id = $1
		ORDER BY tm.joined_at ASC`

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
			&member.TribeID,
			&member.UserID,
			&member.Type,
			&member.JoinedAt,
			&member.ExpiresAt,
			&member.InvitedBy,
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

// GetUserTribes retrieves all tribes that a user is a member of
func (r *TribeRepository) GetUserTribes(userID uuid.UUID) ([]*models.Tribe, error) {
	query := `
		SELECT t.id, t.name, t.created_at, t.updated_at, t.deleted_at
		FROM tribes t
		JOIN tribe_members tm ON tm.tribe_id = t.id
		WHERE tm.user_id = $1 AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC`

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
			&tribe.CreatedAt,
			&tribe.UpdatedAt,
			&tribe.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user tribe row: %w", err)
		}

		tribes = append(tribes, tribe)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user tribe rows: %w", err)
	}

	return tribes, nil
}

// GetExpiredGuestMemberships retrieves all expired guest memberships
func (r *TribeRepository) GetExpiredGuestMemberships() ([]*models.TribeMember, error) {
	query := `
		SELECT tm.tribe_id, tm.user_id, tm.type, tm.joined_at, tm.expires_at, tm.invited_by,
			   u.firebase_uid, u.provider, u.email, u.name, u.avatar_url, u.last_login, u.created_at, u.updated_at
		FROM tribe_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.type = $1 AND tm.expires_at < NOW()`

	rows, err := r.db.Query(query, models.MembershipGuest)
	if err != nil {
		return nil, fmt.Errorf("error getting expired guest memberships: %w", err)
	}
	defer rows.Close()

	var members []*models.TribeMember
	for rows.Next() {
		member := &models.TribeMember{
			User: &models.User{},
		}
		err := rows.Scan(
			&member.TribeID,
			&member.UserID,
			&member.Type,
			&member.JoinedAt,
			&member.ExpiresAt,
			&member.InvitedBy,
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
			return nil, fmt.Errorf("error scanning expired guest membership row: %w", err)
		}
		member.User.ID = member.UserID
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired guest membership rows: %w", err)
	}

	return members, nil
}
