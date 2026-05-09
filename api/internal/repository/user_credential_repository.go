package repository

import (
	"context"

	"app-api/internal/model"
)

type UserCredentialRepository interface {
	Create(ctx context.Context, entity *model.UserCredential) error
}
