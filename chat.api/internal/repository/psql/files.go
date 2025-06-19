package repo

import (
	"chat-api/internal/model"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type FilesRepo struct {
	db *sqlx.DB
}

func NewFilesRepo(db *sqlx.DB) *FilesRepo {
	return &FilesRepo{db: db}
}

func (r *FilesRepo) SetFile(ctx context.Context, file model.File) (fileID int64, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO files (url, type, size)
		VALUES ($1, $2, $3)
		RETURNING file_id
	`

	err = tx.QueryRowContext(ctx, query, file.URL, file.Type, file.Size).Scan(&fileID)
	if err != nil {
		return 0, err
	}

	return fileID, tx.Commit()
}

func (r *FilesRepo) GetFileByFileID(ctx context.Context, fileID int64) (model.File, error) {
	var file model.File

	query := `
		SELECT f.file_id, f.url, f.type, f.size, f.uploaded_at
		FROM files f
		WHERE f.file_id = $1
	`

	err := r.db.QueryRowContext(ctx, query, fileID).Scan(&file.FileID, &file.URL, &file.Type, &file.Size, &file.UploadedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return file, nil
		}
		return file, err
	}

	return file, nil
}

func (r *FilesRepo) GetFilesByMessageID(ctx context.Context, messageID int64) ([]model.File, error) {
	var files []model.File

	query := `
		SELECT f.file_id, f.url, f.type, f.size, f.uploaded_at
		FROM files f
		JOIN messages_files mf ON f.file_id = mf.file_id
		WHERE mf.message_id = $1
		ORDER BY f.uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &files, query, messageID)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (r *FilesRepo) GetFilesByChatID(ctx context.Context, chatID int64) ([]model.File, error) {
	var files []model.File

	query := `
		SELECT f.file_id, f.url, f.type, f.size, f.uploaded_at
		FROM files f
		JOIN messages_files mf ON f.file_id = mf.file_id
		JOIN chat_messages cm ON mf.message_id = cm.message_id
		WHERE cm.chat_id = $1
		ORDER BY f.uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &files, query, chatID)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (r *FilesRepo) DeleteFile(ctx context.Context, fileID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		DELETE FROM files WHERE file_id = $1
	`, fileID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
