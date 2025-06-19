package repo

import (
	"chat-api/internal/model"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type LocationsRepo struct {
	db *sqlx.DB
}

func NewLocationsRepo(db *sqlx.DB) *LocationsRepo {
	return &LocationsRepo{db: db}
}

func (r *LocationsRepo) SetLocation(ctx context.Context, location model.Location) (locationID int64, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO locations (latitude, longitude)
		VALUES ($1, $2)
		RETURNING location_id
	`

	err = tx.QueryRowContext(ctx, query, location.Latitude, location.Longitude).Scan(&locationID)
	if err != nil {
		return 0, err
	}

	return locationID, tx.Commit()
}

func (r *LocationsRepo) GetLocationByLocationID(ctx context.Context, locationID int64) (model.Location, error) {
	var location model.Location

	query := `SELECT location_id, latitude, longitude, created_at FROM locations WHERE location_id = $1`
	err := r.db.QueryRowContext(ctx, query, locationID).Scan(
		&location.LocationID, &location.Latitude, &location.Longitude, &location.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return location, nil
		}
		return location, err
	}

	return location, nil
}

func (r *LocationsRepo) GetLocationsByMessageID(ctx context.Context, messageID int64) ([]model.Location, error) {
	var locations []model.Location

	query := `
		SELECT l.location_id, l.latitude, l.longitude, l.created_at
		FROM locations l
		JOIN messages_locations ml ON l.location_id = ml.location_id
		WHERE ml.message_id = $1
		ORDER BY l.created_at DESC
	`
	err := r.db.SelectContext(ctx, &locations, query, messageID)
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *LocationsRepo) GetLocationsByChatID(ctx context.Context, chatID int64) ([]model.Location, error) {
	var locations []model.Location

	query := `
		SELECT l.location_id, l.latitude, l.longitude, l.created_at
		FROM locations l
		JOIN messages_locations ml ON l.location_id = ml.location_id
		JOIN chat_messages m ON ml.message_id = m.message_id
		WHERE m.chat_id = $1
		ORDER BY l.created_at DESC
	`
	err := r.db.SelectContext(ctx, &locations, query, chatID)
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *LocationsRepo) DeleteLocation(ctx context.Context, locationID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		DELETE FROM locations WHERE location_id = $1
	`, locationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
