package repository

import (
	"context"
	"database/sql"

	"app-api/internal/model"
)

type userRegistrationRequestRepository struct {
	db *sql.DB
}

func NewUserRegistrationRequestRepository(db *sql.DB) UserRegistrationRequestRepository {
	return &userRegistrationRequestRepository{db: db}
}

func (r *userRegistrationRequestRepository) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	const query = `
SELECT
	id,
	email,
	token_hash,
	expires_at,
	verified_at,
	created_at
FROM user_registration_requests
WHERE email = $1
LIMIT 1
`

	entity := &model.UserRegistrationRequest{}
	err := getExecutor(ctx, r.db).QueryRowContext(ctx, query, email).Scan(
		&entity.ID,
		&entity.Email,
		&entity.TokenHash,
		&entity.ExpiresAt,
		&entity.VerifiedAt,
		&entity.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return entity, nil
}

func (r *userRegistrationRequestRepository) Create(ctx context.Context, entity *model.UserRegistrationRequest) error {
	const query = `
INSERT INTO user_registration_requests (
		id,
		email,
		token_hash,
		expires_at,
		verified_at,
		created_at
) VALUES (
		$1, $2, $3, $4, $5, $6
)
`

	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		entity.ID,
		entity.Email,
		entity.TokenHash,
		entity.ExpiresAt,
		entity.VerifiedAt,
		entity.CreatedAt,
	)
	return err
}

func (r *userRegistrationRequestRepository) UpdateToken(ctx context.Context, entity *model.UserRegistrationRequest) error {
	const query = `
UPDATE user_registration_requests
SET
		token_hash = $2,
		expires_at = $3
WHERE id = $1
`

	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		entity.ID,
		entity.TokenHash,
		entity.ExpiresAt,
	)
	return err
}
