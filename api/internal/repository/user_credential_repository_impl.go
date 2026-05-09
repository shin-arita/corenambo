package repository

import (
	"context"
	"database/sql"

	"app-api/internal/model"
)

type userCredentialRepository struct {
	db *sql.DB
}

func NewUserCredentialRepository(db *sql.DB) UserCredentialRepository {
	return &userCredentialRepository{db: db}
}

func (r *userCredentialRepository) Create(ctx context.Context, entity *model.UserCredential) error {
	const query = `
INSERT INTO user_credentials (
                user_id,
                password_hash,
                password_changed_at,
                created_at
) VALUES (
                $1, $2, $3, $4
)
`

	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		entity.UserID,
		entity.PasswordHash,
		entity.PasswordChangedAt,
		entity.CreatedAt,
	)
	return err
}
