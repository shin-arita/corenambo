package repository

import (
	"context"
	"database/sql"

	"app-api/internal/model"
)

type userEmailRepository struct {
	db *sql.DB
}

func NewUserEmailRepository(db *sql.DB) UserEmailRepository {
	return &userEmailRepository{db: db}
}

func (r *userEmailRepository) FindByEmail(ctx context.Context, email string) (*model.UserEmail, error) {
	const query = `
SELECT
        id,
        user_id,
        email,
        is_primary,
        verified_at,
        created_at,
        updated_at
FROM user_emails
WHERE LOWER(email) = LOWER($1)
LIMIT 1
`

	entity := &model.UserEmail{}
	err := getExecutor(ctx, r.db).QueryRowContext(ctx, query, email).Scan(
		&entity.ID,
		&entity.UserID,
		&entity.Email,
		&entity.IsPrimary,
		&entity.VerifiedAt,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return entity, nil
}

func (r *userEmailRepository) Create(ctx context.Context, entity *model.UserEmail) error {
	const query = `
INSERT INTO user_emails (
                id,
                user_id,
                email,
                is_primary,
                verified_at,
                created_at,
                updated_at
) VALUES (
                $1, $2, $3, $4, $5, $6, $7
)
`

	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		entity.ID,
		entity.UserID,
		entity.Email,
		entity.IsPrimary,
		entity.VerifiedAt,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
	return err
}
