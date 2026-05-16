package repository

import (
	"context"
	"errors"

	"app-api/internal/model"
)

// ErrDuplicateEmail はメールアドレスの一意制約違反を表すセンチネルエラー。
// concurrent INSERTで発生した場合は成功として扱う。
var ErrDuplicateEmail = errors.New("duplicate email")

type UserRegistrationRequestRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error)
	FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.UserRegistrationRequest, error)
	FindByTokenHashForUpdate(ctx context.Context, tokenHash string) (*model.UserRegistrationRequest, error)
	Create(ctx context.Context, entity *model.UserRegistrationRequest) error
	UpdateToken(ctx context.Context, entity *model.UserRegistrationRequest) error
	UpdateVerifiedAt(ctx context.Context, entity *model.UserRegistrationRequest) error
}
