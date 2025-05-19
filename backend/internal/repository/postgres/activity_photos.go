package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
	"github.com/jenglund/rlship-tools/internal/testutil"
)

type ActivityPhotosRepository struct {
	db *sql.DB
	tm *TransactionManager
}

func NewActivityPhotosRepository(db interface{}) *ActivityPhotosRepository {
	var sqlDB *sql.DB

	switch d := db.(type) {
	case *sql.DB:
		sqlDB = d
	case *testutil.SchemaDB:
		sqlDB = d.UnwrapDB()
	default:
		if db == nil {
			sqlDB = nil
		} else {
			panic(fmt.Sprintf("Unsupported DB type: %T", db))
		}
	}

	return &ActivityPhotosRepository{
		db: sqlDB,
		tm: NewTransactionManager(sqlDB),
	}
}

func (r *ActivityPhotosRepository) Create(photo *models.ActivityPhoto) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		if photo.ID == uuid.Nil {
			photo.ID = uuid.New()
		}

		now := time.Now()
		if photo.CreatedAt.IsZero() {
			photo.CreatedAt = now
		}
		if photo.UpdatedAt.IsZero() {
			photo.UpdatedAt = now
		}

		query := `
			INSERT INTO activity_photos (
				id, activity_id, url, caption, metadata, created_at, updated_at, version
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING version
		`

		err := tx.QueryRow(query,
			photo.ID,
			photo.ActivityID,
			photo.URL,
			photo.Caption,
			photo.Metadata,
			photo.CreatedAt,
			photo.UpdatedAt,
			1, // Initial version
		).Scan(&photo.Version)

		if err != nil {
			return fmt.Errorf("error creating activity photo: %w", err)
		}

		return nil
	})
}

func (r *ActivityPhotosRepository) GetByID(id uuid.UUID) (*models.ActivityPhoto, error) {
	ctx := context.Background()
	opts := DefaultTransactionOptions()
	var photo *models.ActivityPhoto

	err := r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			SELECT id, activity_id, url, caption, metadata, created_at, updated_at, version
			FROM activity_photos
			WHERE id = $1 AND deleted_at IS NULL
		`
		photo = &models.ActivityPhoto{}
		err := tx.QueryRow(query, id).Scan(
			&photo.ID,
			&photo.ActivityID,
			&photo.URL,
			&photo.Caption,
			&photo.Metadata,
			&photo.CreatedAt,
			&photo.UpdatedAt,
			&photo.Version,
		)

		if err == sql.ErrNoRows {
			return models.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("error getting activity photo: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return photo, nil
}

func (r *ActivityPhotosRepository) Update(photo *models.ActivityPhoto) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		photo.UpdatedAt = time.Now()

		query := `
			UPDATE activity_photos
			SET url = $1, caption = $2, metadata = $3, updated_at = $4
			WHERE id = $5 AND deleted_at IS NULL
			RETURNING version
		`
		result, err := tx.Exec(query,
			photo.URL,
			photo.Caption,
			photo.Metadata,
			photo.UpdatedAt,
			photo.ID,
		)
		if err != nil {
			return fmt.Errorf("error updating activity photo: %w", err)
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

func (r *ActivityPhotosRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	opts := DefaultTransactionOptions()

	return r.tm.WithTransaction(ctx, opts, func(tx *sql.Tx) error {
		query := `
			UPDATE activity_photos
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL
		`
		result, err := tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("error deleting activity photo: %w", err)
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
