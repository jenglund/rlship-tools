package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/lib/pq"
)

// UserRepository implements models.UserRepository using PostgreSQL
type UserRepository struct {
	db *sql.DB
	tm *TransactionManager
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
		tm: NewTransactionManager(db),
	}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(user *models.User) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			INSERT INTO users (id, firebase_uid, provider, email, name, avatar_url, last_login, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`

		if user.ID == uuid.Nil {
			user.ID = uuid.New()
		}
		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now

		err := tx.QueryRow(
			query,
			user.ID,
			user.FirebaseUID,
			user.Provider,
			user.Email,
			user.Name,
			user.AvatarURL,
			user.LastLogin,
			user.CreatedAt,
			user.UpdatedAt,
		).Scan(&user.ID)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return fmt.Errorf("user already exists: %w", err)
			}
			return fmt.Errorf("error creating user: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var user *models.User

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, firebase_uid, provider, email, name, avatar_url, last_login, created_at, updated_at
			FROM users
			WHERE id = $1`

		user = &models.User{}
		err := tx.QueryRow(query, id).Scan(
			&user.ID,
			&user.FirebaseUID,
			&user.Provider,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		if err != nil {
			return fmt.Errorf("error getting user: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByFirebaseUID retrieves a user by their Firebase UID
func (r *UserRepository) GetByFirebaseUID(firebaseUID string) (*models.User, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var user *models.User

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, firebase_uid, provider, email, name, avatar_url, last_login, created_at, updated_at
			FROM users
			WHERE firebase_uid = $1`

		user = &models.User{}
		err := tx.QueryRow(query, firebaseUID).Scan(
			&user.ID,
			&user.FirebaseUID,
			&user.Provider,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		if err != nil {
			return fmt.Errorf("error getting user: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var user *models.User

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, firebase_uid, provider, email, name, avatar_url, last_login, created_at, updated_at
			FROM users
			WHERE email = $1`

		user = &models.User{}
		err := tx.QueryRow(query, email).Scan(
			&user.ID,
			&user.FirebaseUID,
			&user.Provider,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		if err != nil {
			return fmt.Errorf("error getting user: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates an existing user in the database
func (r *UserRepository) Update(user *models.User) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE users
			SET firebase_uid = $1,
				provider = $2,
				email = $3,
				name = $4,
				avatar_url = $5,
				last_login = $6,
				updated_at = $7
			WHERE id = $8
			RETURNING id`

		user.UpdatedAt = time.Now()

		err := tx.QueryRow(
			query,
			user.FirebaseUID,
			user.Provider,
			user.Email,
			user.Name,
			user.AvatarURL,
			user.LastLogin,
			user.UpdatedAt,
			user.ID,
		).Scan(&user.ID)

		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		if err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}

		return nil
	})
}

// Delete removes a user from the database
func (r *UserRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `DELETE FROM users WHERE id = $1`

		result, err := tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("error deleting user: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error checking rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("user not found")
		}

		return nil
	})
}

// List retrieves a paginated list of users
func (r *UserRepository) List(offset, limit int) ([]*models.User, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var users []*models.User

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, firebase_uid, provider, email, name, avatar_url, last_login, created_at, updated_at
			FROM users
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

		rows, err := tx.Query(query, limit, offset)
		if err != nil {
			return fmt.Errorf("error listing users: %w", err)
		}
		defer rows.Close()

		users = make([]*models.User, 0)
		for rows.Next() {
			user := &models.User{}
			err := rows.Scan(
				&user.ID,
				&user.FirebaseUID,
				&user.Provider,
				&user.Email,
				&user.Name,
				&user.AvatarURL,
				&user.LastLogin,
				&user.CreatedAt,
				&user.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("error scanning user row: %w", err)
			}
			users = append(users, user)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating user rows: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}
