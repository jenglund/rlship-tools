package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jenglund/rlship-tools/internal/models"
)

type ActivityPhotosRepository struct {
	db *sql.DB
}

func NewActivityPhotosRepository(db *sql.DB) *ActivityPhotosRepository {
	return &ActivityPhotosRepository{db: db}
}

func (r *ActivityPhotosRepository) Create(photo *models.ActivityPhoto) error {
	query := `
		INSERT INTO activity_photos (
			id, activity_id, url, caption, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query,
		photo.ID,
		photo.ActivityID,
		photo.URL,
		photo.Caption,
		photo.Metadata,
		photo.CreatedAt,
		photo.UpdatedAt,
	)
	return err
}

func (r *ActivityPhotosRepository) GetByID(id uuid.UUID) (*models.ActivityPhoto, error) {
	query := `
		SELECT id, activity_id, url, caption, metadata, created_at, updated_at
		FROM activity_photos
		WHERE id = $1 AND deleted_at IS NULL
	`
	photo := &models.ActivityPhoto{}
	err := r.db.QueryRow(query, id).Scan(
		&photo.ID,
		&photo.ActivityID,
		&photo.URL,
		&photo.Caption,
		&photo.Metadata,
		&photo.CreatedAt,
		&photo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return photo, nil
}

func (r *ActivityPhotosRepository) Update(photo *models.ActivityPhoto) error {
	query := `
		UPDATE activity_photos
		SET url = $1, caption = $2, metadata = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query,
		photo.URL,
		photo.Caption,
		photo.Metadata,
		photo.UpdatedAt,
		photo.ID,
	)
	return err
}

func (r *ActivityPhotosRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE activity_photos
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, id)
	return err
}
