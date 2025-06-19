package repo

import (
	"context"
	"fmt"
	"notification-api/internal/model"

	"github.com/jmoiron/sqlx"
)

type VerificationRepository struct {
	db *sqlx.DB
}

func NewVerificationRepository(db *sqlx.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

func (r *VerificationRepository) SetRecordVerificationLog(ctx context.Context, vc model.VerifyCodeInput, method string) error {
	query := `
        INSERT INTO verification_log (recipient, code, method)
        VALUES (?, ?, ?)
    `

	_, err := r.db.ExecContext(ctx, query, vc.Recipient, vc.Code, method)
	if err != nil {
		return fmt.Errorf("failed to insert verification log: %w", err)
	}

	return nil
}
