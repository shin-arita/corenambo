package service

import (
	"context"
	"errors"
	"time"

	"app-api/internal/model"
	"app-api/internal/repository"
	"app-api/internal/token"
)

// ──── UserRegistrationRequestRepository dummies ────

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
func (d *dummyUserRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyUserRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyUserRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
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
func (d *dummyErrorUserRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, errors.New("db error")
}
func (d *dummyErrorUserRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, errors.New("db error")
}
func (d *dummyErrorUserRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
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
func (d *dummyErrorCreateUserRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyErrorCreateUserRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyErrorCreateUserRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
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
func (d *dummyErrorUpdateTokenUserRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return d.FindByEmail(ctx, email)
}
func (d *dummyErrorUpdateTokenUserRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyErrorUpdateTokenUserRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
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
func (d *dummyVerifiedUserRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	now := time.Now()
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      "test@example.com",
		TokenHash:  hash,
		ExpiresAt:  now.Add(60 * time.Minute),
		VerifiedAt: &now,
		CreatedAt:  now,
	}, nil
}
func (d *dummyVerifiedUserRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return d.FindByEmail(ctx, email)
}
func (d *dummyVerifiedUserRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
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
func (d *dummyRecentlySentUserRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyRecentlySentUserRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return d.FindByEmail(ctx, email)
}
func (d *dummyRecentlySentUserRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
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
func (d *dummyExpiredUserRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyExpiredUserRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return d.FindByEmail(ctx, email)
}
func (d *dummyExpiredUserRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

// ──── MailOutboxRepository dummies ────

type dummyOutboxRepo struct{}

func (d *dummyOutboxRepo) Create(ctx context.Context, m *model.MailOutbox) error { return nil }
func (d *dummyOutboxRepo) FetchPending(ctx context.Context, i int) ([]*model.MailOutbox, error) {
	return nil, nil
}
func (d *dummyOutboxRepo) MarkProcessing(ctx context.Context, id string) error        { return nil }
func (d *dummyOutboxRepo) MarkSent(ctx context.Context, id string, t time.Time) error { return nil }
func (d *dummyOutboxRepo) MarkRetry(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}
func (d *dummyOutboxRepo) MarkFailed(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}
func (d *dummyOutboxRepo) ResetStuckProcessing(ctx context.Context, stuckBefore time.Time) error {
	return nil
}

// ──── TxManager dummy ────

type dummyTx struct{}

func (d *dummyTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// ──── Clock dummies ────

type dummyClock struct{}

func (d dummyClock) Now() time.Time { return time.Now() }

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time { return c.t }

// ──── UUID dummies ────

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

// ──── Token dummies ────

type dummyTokenGen struct{}

func (d dummyTokenGen) Generate() (string, error) { return "token", nil }

type dummyErrorTokenGen struct{}

func (d dummyErrorTokenGen) Generate() (string, error) { return "", errors.New("token error") }

type dummyEmptyTokenGen struct{}

func (d dummyEmptyTokenGen) Generate() (string, error) { return "", nil }

type dummyTokenHasher struct{}

func (d dummyTokenHasher) Hash(s string) (string, error) { return "hash", nil }

type dummyErrorTokenHasher struct{}

func (d dummyErrorTokenHasher) Hash(s string) (string, error) { return "", errors.New("hash error") }

// ──── URL builder dummies ────

type dummyURL struct{}

func (d dummyURL) Build(s string) string { return "url" }

type captureURLBuilder struct {
	capturedToken string
}

func (d *captureURLBuilder) Build(token string) string {
	d.capturedToken = token
	return "http://localhost:5173/registration/verify?token=" + token
}

// ──── Config dummy ────

type dummyConfig struct{}

func (d dummyConfig) RegistrationResendIntervalMinutes() int { return 10 }
func (d dummyConfig) RegistrationTokenExpiresMinutes() int   { return 60 }

// ──── UserRepository (model) dummies ────

type dummyUserModelRepo struct{}

func (d *dummyUserModelRepo) Create(ctx context.Context, u *model.User) error { return nil }

type dummyErrorUserModelRepo struct{}

func (d *dummyErrorUserModelRepo) Create(ctx context.Context, u *model.User) error {
	return errors.New("user create error")
}

// ──── UserEmailRepository dummies ────

type dummyUserEmailRepo struct{}

func (d *dummyUserEmailRepo) FindByEmail(ctx context.Context, email string) (*model.UserEmail, error) {
	return nil, nil
}
func (d *dummyUserEmailRepo) Create(ctx context.Context, e *model.UserEmail) error { return nil }

type dummyErrorUserEmailRepo struct{}

func (d *dummyErrorUserEmailRepo) FindByEmail(ctx context.Context, email string) (*model.UserEmail, error) {
	return nil, nil
}
func (d *dummyErrorUserEmailRepo) Create(ctx context.Context, e *model.UserEmail) error {
	return errors.New("user email create error")
}

type dummyExistingEmailRepo struct{}

func (d *dummyExistingEmailRepo) FindByEmail(ctx context.Context, email string) (*model.UserEmail, error) {
	now := time.Now()
	return &model.UserEmail{
		ID:         "existing-id",
		UserID:     "existing-user",
		Email:      email,
		IsPrimary:  true,
		VerifiedAt: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
func (d *dummyExistingEmailRepo) Create(ctx context.Context, e *model.UserEmail) error { return nil }

type dummyErrorFindEmailRepo struct{}

func (d *dummyErrorFindEmailRepo) FindByEmail(ctx context.Context, email string) (*model.UserEmail, error) {
	return nil, errors.New("find email error")
}
func (d *dummyErrorFindEmailRepo) Create(ctx context.Context, e *model.UserEmail) error { return nil }

// ──── UserCredentialRepository dummies ────

type dummyUserCredentialRepo struct{}

func (d *dummyUserCredentialRepo) Create(ctx context.Context, c *model.UserCredential) error {
	return nil
}

type dummyErrorUserCredentialRepo struct{}

func (d *dummyErrorUserCredentialRepo) Create(ctx context.Context, c *model.UserCredential) error {
	return errors.New("credential create error")
}

// ──── Capture repos ────

type captureCreateRepo struct {
	capturedEmail string
}

func (d *captureCreateRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	d.capturedEmail = req.Email
	return nil
}
func (d *captureCreateRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *captureCreateRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

type captureOutboxRepo struct {
	capturedPayload string
}

func (d *captureOutboxRepo) Create(ctx context.Context, m *model.MailOutbox) error {
	d.capturedPayload = m.Payload
	return nil
}
func (d *captureOutboxRepo) FetchPending(ctx context.Context, i int) ([]*model.MailOutbox, error) {
	return nil, nil
}
func (d *captureOutboxRepo) MarkProcessing(ctx context.Context, id string) error        { return nil }
func (d *captureOutboxRepo) MarkSent(ctx context.Context, id string, t time.Time) error { return nil }
func (d *captureOutboxRepo) MarkRetry(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}
func (d *captureOutboxRepo) MarkFailed(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}
func (d *captureOutboxRepo) ResetStuckProcessing(ctx context.Context, stuckBefore time.Time) error {
	return nil
}

type captureCreateUserRegistrationRepo struct {
	capturedTokenHash string
}

func (d *captureCreateUserRegistrationRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateUserRegistrationRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	d.capturedTokenHash = req.TokenHash
	return nil
}
func (d *captureCreateUserRegistrationRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *captureCreateUserRegistrationRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateUserRegistrationRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateUserRegistrationRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *captureCreateUserRegistrationRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

type captureFullOutboxRepo struct {
	capturedOutbox *model.MailOutbox
}

func (d *captureFullOutboxRepo) Create(ctx context.Context, m *model.MailOutbox) error {
	d.capturedOutbox = m
	return nil
}
func (d *captureFullOutboxRepo) FetchPending(ctx context.Context, i int) ([]*model.MailOutbox, error) {
	return nil, nil
}
func (d *captureFullOutboxRepo) MarkProcessing(ctx context.Context, id string) error { return nil }
func (d *captureFullOutboxRepo) MarkSent(ctx context.Context, id string, t time.Time) error {
	return nil
}
func (d *captureFullOutboxRepo) MarkRetry(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}
func (d *captureFullOutboxRepo) MarkFailed(ctx context.Context, id string, s string, t time.Time) error {
	return nil
}
func (d *captureFullOutboxRepo) ResetStuckProcessing(ctx context.Context, stuckBefore time.Time) error {
	return nil
}

// ──── Verify-specific registration request repos ────

type tokenFoundUserRegRepo struct {
	expiresAt time.Time
}

func (d *tokenFoundUserRegRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenFoundUserRegRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *tokenFoundUserRegRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *tokenFoundUserRegRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenFoundUserRegRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      "test@example.com",
		TokenHash:  hash,
		ExpiresAt:  d.expiresAt,
		VerifiedAt: nil,
		CreatedAt:  time.Now(),
	}, nil
}
func (d *tokenFoundUserRegRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenFoundUserRegRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

type tokenNotFoundUserRegRepo struct{}

func (d *tokenNotFoundUserRegRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenNotFoundUserRegRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *tokenNotFoundUserRegRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *tokenNotFoundUserRegRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenNotFoundUserRegRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenNotFoundUserRegRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenNotFoundUserRegRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

type tokenExpiredUserRegRepo struct{}

func (d *tokenExpiredUserRegRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenExpiredUserRegRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *tokenExpiredUserRegRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *tokenExpiredUserRegRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenExpiredUserRegRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	past := time.Now().Add(-2 * time.Hour)
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      "test@example.com",
		TokenHash:  hash,
		ExpiresAt:  past,
		VerifiedAt: nil,
		CreatedAt:  past,
	}, nil
}
func (d *tokenExpiredUserRegRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *tokenExpiredUserRegRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

type errorUpdateVerifiedAtRepo struct {
	expiresAt time.Time
}

func (d *errorUpdateVerifiedAtRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *errorUpdateVerifiedAtRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *errorUpdateVerifiedAtRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *errorUpdateVerifiedAtRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *errorUpdateVerifiedAtRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      "test@example.com",
		TokenHash:  hash,
		ExpiresAt:  d.expiresAt,
		VerifiedAt: nil,
		CreatedAt:  time.Now(),
	}, nil
}
func (d *errorUpdateVerifiedAtRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *errorUpdateVerifiedAtRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return errors.New("update verified at error")
}

// ──── Concurrent duplicate email dummy ────

type dummyDuplicateEmailCreateRepo struct{}

func (d *dummyDuplicateEmailCreateRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyDuplicateEmailCreateRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyDuplicateEmailCreateRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return repository.ErrDuplicateEmail
}
func (d *dummyDuplicateEmailCreateRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *dummyDuplicateEmailCreateRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyDuplicateEmailCreateRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *dummyDuplicateEmailCreateRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

// ──── Service factory helpers ────

func newTestUserRegistrationService() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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

func newTestUserRegistrationServiceWithCaptureRepo(repo *captureCreateRepo) UserRegistrationService {
	return NewUserRegistrationService(
		repo,
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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

func newTestUserRegistrationServiceWithEmptyTokenGen() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyEmptyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithCaptureURLBuilder(builder *captureURLBuilder) UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		builder,
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithCaptureOutbox(outbox *captureOutboxRepo) UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
		outbox,
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		dummyClock{},
		&captureURLBuilder{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithCaptureUserRegistrationRepo(repo *captureCreateUserRegistrationRepo) UserRegistrationService {
	return NewUserRegistrationService(
		repo,
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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

func newTestUserRegistrationServiceWithCaptureFullOutbox(outbox *captureFullOutboxRepo, clk fixedClock) UserRegistrationService {
	return NewUserRegistrationService(
		&dummyUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
		outbox,
		&dummyTx{},
		dummyTokenGen{},
		dummyTokenHasher{},
		dummyUUID{},
		clk,
		dummyURL{},
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithVerifiedUserAndTokenGenError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyVerifiedUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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

func newTestUserRegistrationServiceWithRecentlySentUserAndTokenGenError() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyRecentlySentUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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

func newTestUserRegistrationServiceWithRealHasherAndCapture(
	repo *captureCreateUserRegistrationRepo,
	builder *captureURLBuilder,
) UserRegistrationService {
	return NewUserRegistrationService(
		repo,
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
		&dummyOutboxRepo{},
		&dummyTx{},
		token.DefaultGenerator{},
		token.SHA256Hasher{},
		dummyUUID{},
		dummyClock{},
		builder,
		dummyConfig{},
	)
}

func newTestUserRegistrationServiceWithDuplicateEmailRepo() UserRegistrationService {
	return NewUserRegistrationService(
		&dummyDuplicateEmailCreateRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
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

// ──── CheckToken-specific registration request repos ────

type checkTokenFoundRepo struct {
	expiresAt  time.Time
	verifiedAt *time.Time
}

func (d *checkTokenFoundRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenFoundRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *checkTokenFoundRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *checkTokenFoundRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return &model.UserRegistrationRequest{
		ID:         "uuid",
		Email:      "test@example.com",
		TokenHash:  hash,
		ExpiresAt:  d.expiresAt,
		VerifiedAt: d.verifiedAt,
		CreatedAt:  time.Now(),
	}, nil
}
func (d *checkTokenFoundRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenFoundRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenFoundRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

type checkTokenNotFoundRepo struct{}

func (d *checkTokenNotFoundRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenNotFoundRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *checkTokenNotFoundRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *checkTokenNotFoundRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenNotFoundRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenNotFoundRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenNotFoundRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

type checkTokenFindErrorRepo struct{}

func (d *checkTokenFindErrorRepo) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenFindErrorRepo) Create(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *checkTokenFindErrorRepo) UpdateToken(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}
func (d *checkTokenFindErrorRepo) FindByTokenHash(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, errors.New("db error")
}
func (d *checkTokenFindErrorRepo) FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenFindErrorRepo) FindByEmailForUpdate(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	return nil, nil
}
func (d *checkTokenFindErrorRepo) UpdateVerifiedAt(ctx context.Context, req *model.UserRegistrationRequest) error {
	return nil
}

func newCheckTokenService(
	regRepo interface {
		FindByTokenHash(context.Context, string) (*model.UserRegistrationRequest, error)
		FindByEmailForUpdate(context.Context, string) (*model.UserRegistrationRequest, error)
		FindByTokenHashForUpdate(context.Context, string) (*model.UserRegistrationRequest, error)
		FindByEmail(context.Context, string) (*model.UserRegistrationRequest, error)
		Create(context.Context, *model.UserRegistrationRequest) error
		UpdateToken(context.Context, *model.UserRegistrationRequest) error
		UpdateVerifiedAt(context.Context, *model.UserRegistrationRequest) error
	},
	userEmailRepo interface {
		FindByEmail(context.Context, string) (*model.UserEmail, error)
		Create(context.Context, *model.UserEmail) error
	},
) UserRegistrationService {
	return NewUserRegistrationService(
		regRepo,
		&dummyUserModelRepo{},
		userEmailRepo,
		&dummyUserCredentialRepo{},
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

func newVerifyService(
	regRepo interface {
		FindByEmailForUpdate(context.Context, string) (*model.UserRegistrationRequest, error)
		FindByTokenHashForUpdate(context.Context, string) (*model.UserRegistrationRequest, error)
		FindByEmail(context.Context, string) (*model.UserRegistrationRequest, error)
		Create(context.Context, *model.UserRegistrationRequest) error
		UpdateToken(context.Context, *model.UserRegistrationRequest) error
		FindByTokenHash(context.Context, string) (*model.UserRegistrationRequest, error)
		UpdateVerifiedAt(context.Context, *model.UserRegistrationRequest) error
	},
	userModelRepo interface {
		Create(context.Context, *model.User) error
	},
	userEmailRepo interface {
		FindByEmail(context.Context, string) (*model.UserEmail, error)
		Create(context.Context, *model.UserEmail) error
	},
	userCredentialRepo interface {
		Create(context.Context, *model.UserCredential) error
	},
) UserRegistrationService {
	return NewUserRegistrationService(
		regRepo,
		userModelRepo,
		userEmailRepo,
		userCredentialRepo,
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
