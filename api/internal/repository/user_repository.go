package repository

import (
	"context"

	"app-api/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, entity *model.User) error
}
