package repository

import (
	"context"

	"app-api/internal/model"
)

type UserRegistrationRequestRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error)
	Create(ctx context.Context, entity *model.UserRegistrationRequest) error
	UpdateToken(ctx context.Context, entity *model.UserRegistrationRequest) error
}
