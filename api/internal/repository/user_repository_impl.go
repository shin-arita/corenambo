package repository

import (
	"context"
	"database/sql"

	"app-api/internal/model"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, entity *model.User) error {
	const query = `
INSERT INTO users (
                id,
                display_name,
                status,
                created_at,
                updated_at
) VALUES (
                $1, $2, $3, $4, $5
)
`

	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		entity.ID,
		entity.DisplayName,
		entity.Status,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
	return err
}
