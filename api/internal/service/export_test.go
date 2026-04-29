package service

import (
	"context"
	"errors"
	"time"

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

type dummyErrorUserRepo struct{}

func (d *dummyErrorUserRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, errors.New("db error")
}
func (d *dummyErrorUserRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyErrorUserRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyErrorUserRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}

type dummyErrorCreateUserRepo struct{}

func (d *dummyErrorCreateUserRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyErrorCreateUserRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return errors.New("create error")
}
func (d *dummyErrorCreateUserRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyErrorCreateUserRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}

type dummyErrorUpdateTokenUserRepo struct{}

func (d *dummyErrorUpdateTokenUserRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	now := time.Now()
	lastSentAt := now.Add(-20 * time.Minute)
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      email,
		TokenHash:  "hash",
		ExpiresAt:  now.Add(60 * time.Minute),
		VerifiedAt: nil,
		LastSentAt: &lastSentAt,
		CreatedAt:  now,
	}, nil
}
func (d *dummyErrorUpdateTokenUserRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyErrorUpdateTokenUserRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return errors.New("update token error")
}
func (d *dummyErrorUpdateTokenUserRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}

type dummyVerifiedUserRepo struct{}

func (d *dummyVerifiedUserRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	now := time.Now()
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      email,
		TokenHash:  "hash",
		ExpiresAt:  now.Add(60 * time.Minute),
		VerifiedAt: &now,
		CreatedAt:  now,
	}, nil
}
func (d *dummyVerifiedUserRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyVerifiedUserRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyVerifiedUserRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}

type dummyRecentlySentUserRepo struct{}

func (d *dummyRecentlySentUserRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	now := time.Now()
	lastSentAt := now.Add(-1 * time.Minute)
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      email,
		TokenHash:  "hash",
		ExpiresAt:  now.Add(60 * time.Minute),
		VerifiedAt: nil,
		LastSentAt: &lastSentAt,
		CreatedAt:  now,
	}, nil
}
func (d *dummyRecentlySentUserRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyRecentlySentUserRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyRecentlySentUserRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}

type dummyExpiredUserRepo struct{}

func (d *dummyExpiredUserRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	now := time.Now()
	lastSentAt := now.Add(-20 * time.Minute)
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      email,
		TokenHash:  "hash",
		ExpiresAt:  now.Add(60 * time.Minute),
		VerifiedAt: nil,
		LastSentAt: &lastSentAt,
		CreatedAt:  now,
	}, nil
}
func (d *dummyExpiredUserRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyExpiredUserRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyExpiredUserRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}

type dummyOutboxRepo struct{}

func (d *dummyOutboxRepo) Create(ctx context.Context, m *model.MailOutbox) error { return nil }
func (d *dummyOutboxRepo) FetchPending(ctx context.Context, i int) ([]*model.MailOutbox, error) {
	return nil, nil
}
func (d *dummyOutboxRepo) MarkProcessing(ctx context.Context, id string) error        { return nil }
func (d *dummyOutboxRepo) MarkSent(ctx context.Context, id string, t time.Time) error { return nil }
func (d *dummyOutboxRepo) MarkFailed(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}
func (d *dummyOutboxRepo) ResetStuckProcessing(ctx context.Context, stuckBefore time.Time) error {
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

type dummyErrorFirstUUID struct {
	count int
}

func (d *dummyErrorFirstUUID) NewV7() (string, error) {
	d.count++
	if d.count == 1 {
		return "", errors.New("uuid error")
	}
	return "uuid", nil
}

type dummyErrorSecondUUID struct {
	count int
}

func (d *dummyErrorSecondUUID) NewV7() (string, error) {
	d.count++
	if d.count == 2 {
		return "", errors.New("uuid error")
	}
	return "uuid", nil
}

type dummyTokenGen struct{}

func (d dummyTokenGen) Generate() (string, error) { return "token", nil }

type dummyErrorTokenGen struct{}

func (d dummyErrorTokenGen) Generate() (string, error) { return "", errors.New("token error") }

type dummyTokenHasher struct{}

func (d dummyTokenHasher) Hash(s string) (string, error) { return "hash", nil }

type dummyErrorTokenHasher struct{}

func (d dummyErrorTokenHasher) Hash(s string) (string, error) { return "", errors.New("hash error") }

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
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithVerifiedUser() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyVerifiedUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithRecentlySentUser() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyRecentlySentUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithExpiredUser() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyExpiredUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithDBError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyErrorUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithTokenGenError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyErrorTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithTokenHashError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyErrorTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithFirstUUIDError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		&dummyErrorFirstUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithSecondUUIDError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		&dummyErrorSecondUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithCreateUserError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyErrorCreateUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithUpdateTokenError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyErrorUpdateTokenUserRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}
