package service

import (
	"context"
	"time"

	"app-api/internal/mail"
	"app-api/internal/model"
)

type dummyUserRepo struct{}

func (d *dummyUserRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyUserRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyUserRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyUserRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}

type dummyOutboxRepo struct{}

func (d *dummyOutboxRepo) Create(ctx context.Context, m *model.MailOutbox) error { return nil }
func (d *dummyOutboxRepo) FetchPending(ctx context.Context, i int) ([]*model.MailOutbox, error) {
	return nil, nil
}
func (d *dummyOutboxRepo) MarkSent(ctx context.Context, id string, t time.Time) error { return nil }
func (d *dummyOutboxRepo) MarkFailed(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}

type dummyTx struct{}

func (d *dummyTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type dummyClock struct{}

func (d dummyClock) Now() time.Time { return time.Now() }

type dummyUUID struct{}

func (d dummyUUID) NewV7() (string, error) { return "uuid", nil }

type dummyTokenGen struct{}

func (d dummyTokenGen) Generate() (string, error) { return "token", nil }

type dummyTokenHasher struct{}

func (d dummyTokenHasher) Hash(s string) (string, error) { return "hash", nil }

type dummyMailer struct{}

func (d dummyMailer) SendUserRegistrationMail(ctx context.Context, m mail.UserRegistrationMail) error {
	return nil
}
func (d dummyMailer) SendUserAlreadyRegisteredMail(ctx context.Context, e, l string) error {
	return nil
}

type dummyURL struct{}

func (d dummyURL) Build(s string) string { return "url" }

type dummyConfig struct{}

func (d dummyConfig) RegistrationResendIntervalMinutes() int { return 10 }
func (d dummyConfig) RegistrationTokenExpiresMinutes() int   { return 60 }

func newTestUserRegistrationService() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyMailer{},
		dummyURL{},
		dummyConfig{},
	)
}
