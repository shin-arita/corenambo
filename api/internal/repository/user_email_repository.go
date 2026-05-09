package repository

import (
	"context"

	"app-api/internal/model"
)

type UserEmailRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.UserEmail, error)
	Create(ctx context.Context, entity *model.UserEmail) error
}
