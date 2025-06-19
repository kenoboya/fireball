package repo

import (
	"chat-api/internal/model"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type MediaRepo struct {
	db *sqlx.DB
}

func NewMediaRepo(db *sqlx.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

func (r *MediaRepo) SetMedia(ctx context.Context, media model.Media) (mediaID int64, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO media (url, type, size)
		VALUES ($1, $2, $3)
		RETURNING media_id
	`

	err = tx.QueryRowContext(ctx, query, media.URL, media.Type, media.Size).Scan(&mediaID)
	if err != nil {
		return 0, err
	}

	return mediaID, tx.Commit()
}

func (r *MediaRepo) GetMediaByMediaID(ctx context.Context, mediaID int64) (model.Media, error) {
	var media model.Media

	query := `
		SELECT media_id, url, type, size, uploaded_at
		FROM media
		WHERE media_id = $1
	`

	err := r.db.QueryRowContext(ctx, query, mediaID).Scan(&media.MediaID, &media.URL, &media.Type, &media.Size, &media.UploadedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return media, nil
		}
		return media, err
	}

	return media, nil
}

func (r *MediaRepo) GetMediaFileByMessageID(ctx context.Context, messageID int64) ([]model.Media, error) {
	var mediaFiles []model.Media

	query := `
		SELECT m.media_id, m.url, m.type, m.size, m.uploaded_at
		FROM media m
		JOIN messages_media mm ON m.media_id = mm.media_id
		WHERE mm.message_id = $1
		ORDER BY m.uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &mediaFiles, query, messageID)
	if err != nil {
		return nil, err
	}

	return mediaFiles, nil
}

func (r *MediaRepo) GetMediaByChatID(ctx context.Context, chatID int64) ([]model.Media, error) {
	var mediaFiles []model.Media

	query := `
		SELECT m.media_id, m.url, m.type, m.size, m.uploaded_at
		FROM media m
		JOIN messages_media mm ON m.media_id = mm.media_id
		JOIN chat_messages chm ON mm.message_id = chm.message_id
		WHERE chm.chat_id = $1
		ORDER BY m.uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &mediaFiles, query, chatID)
	if err != nil {
		return nil, err
	}

	return mediaFiles, nil
}

func (r *MediaRepo) DeleteMedia(ctx context.Context, mediaID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		DELETE FROM media
		WHERE media_id = $1
	`, mediaID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
