package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"app-api/internal/app_error"
	"app-api/internal/i18n"
	"app-api/internal/mail"
	"app-api/internal/model"
)

type mockUserRegistrationRequestRepository struct {
	findResult *model.UserRegistrationRequest
	findErr    error
	createErr  error
	updateErr  error

	findCalled   bool
	createCalled bool
	updateCalled bool

	findEmail  string
	createData *model.UserRegistrationRequest
	updateData *model.UserRegistrationRequest
}

func (m *mockUserRegistrationRequestRepository) FindByEmail(ctx context.Context, email string) (*model.UserRegistrationRequest, error) {
	m.findCalled = true
	m.findEmail = email

	return m.findResult, m.findErr
}

func (m *mockUserRegistrationRequestRepository) Create(ctx context.Context, entity *model.UserRegistrationRequest) error {
	m.createCalled = true
	m.createData = entity

	return m.createErr
}

func (m *mockUserRegistrationRequestRepository) UpdateToken(ctx context.Context, entity *model.UserRegistrationRequest) error {
	m.updateCalled = true
	m.updateData = entity

	return m.updateErr
}

type mockTxManager struct {
	err    error
	called bool
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	m.called = true

	if m.err != nil {
		return m.err
	}

	return fn(ctx)
}

type mockTokenGenerator struct {
	err    error
	called bool
}

func (m *mockTokenGenerator) Generate() (string, error) {
	m.called = true

	if m.err != nil {
		return "", m.err
	}

	return "plain-token", nil
}

type mockTokenHasher struct {
	err    error
	called bool
	value  string
}

func (m *mockTokenHasher) Hash(value string) (string, error) {
	m.called = true
	m.value = value

	if m.err != nil {
		return "", m.err
	}

	return "hashed-token", nil
}

type mockUUIDGenerator struct {
	err    error
	called bool
}

func (m *mockUUIDGenerator) NewV7() (string, error) {
	m.called = true

	if m.err != nil {
		return "", m.err
	}

	return "uuid-v7", nil
}

type mockClock struct{}

func (m *mockClock) Now() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

type mockMailer struct {
	err    error
	called bool
	data   mail.UserRegistrationMail
}

func (m *mockMailer) SendUserRegistrationMail(ctx context.Context, mailData mail.UserRegistrationMail) error {
	m.called = true
	m.data = mailData

	return m.err
}

type mockRegistrationURLBuilder struct {
	called bool
	token  string
}

func (m *mockRegistrationURLBuilder) Build(token string) string {
	m.called = true
	m.token = token

	return "http://example.com/user-registration/verify?token=" + token
}

type mockConfig struct{}

func (m *mockConfig) RegistrationTokenExpiresMinutes() int {
	return 60
}

func newTestUserRegistrationService(
	repo *mockUserRegistrationRequestRepository,
	txManager *mockTxManager,
	tokenGenerator *mockTokenGenerator,
	tokenHasher *mockTokenHasher,
	uuidGenerator *mockUUIDGenerator,
	mailer *mockMailer,
	urlBuilder *mockRegistrationURLBuilder,
) UserRegistrationService {
	return NewUserRegistrationService(
		repo,
		txManager,
		tokenGenerator,
		tokenHasher,
		uuidGenerator,
		&mockClock{},
		mailer,
		urlBuilder,
		&mockConfig{},
	)
}

func TestUserRegistrationServiceCreate(t *testing.T) {
	repo := &mockUserRegistrationRequestRepository{}
	txManager := &mockTxManager{}
	tokenGenerator := &mockTokenGenerator{}
	tokenHasher := &mockTokenHasher{}
	uuidGenerator := &mockUUIDGenerator{}
	mailer := &mockMailer{}
	urlBuilder := &mockRegistrationURLBuilder{}

	svc := newTestUserRegistrationService(
		repo,
		txManager,
		tokenGenerator,
		tokenHasher,
		uuidGenerator,
		mailer,
		urlBuilder,
	)

	out, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if out.Code != i18n.CodeUserRegistrationRequestCreated {
		t.Fatalf("unexpected code: %s", out.Code)
	}

	if !repo.findCalled {
		t.Fatal("FindByEmail not called")
	}

	if repo.findEmail != "test@example.com" {
		t.Fatalf("unexpected find email: %s", repo.findEmail)
	}

	if !tokenGenerator.called {
		t.Fatal("token generator not called")
	}

	if !tokenHasher.called {
		t.Fatal("token hasher not called")
	}

	if tokenHasher.value != "plain-token" {
		t.Fatalf("unexpected hash value: %s", tokenHasher.value)
	}

	if !txManager.called {
		t.Fatal("transaction not called")
	}

	if !uuidGenerator.called {
		t.Fatal("uuid generator not called")
	}

	if !repo.createCalled {
		t.Fatal("Create not called")
	}

	if repo.updateCalled {
		t.Fatal("UpdateToken should not be called")
	}

	if repo.createData == nil {
		t.Fatal("create data is nil")
	}

	if repo.createData.ID != "uuid-v7" {
		t.Fatalf("unexpected id: %s", repo.createData.ID)
	}

	if repo.createData.Email != "test@example.com" {
		t.Fatalf("unexpected email: %s", repo.createData.Email)
	}

	if repo.createData.TokenHash != "hashed-token" {
		t.Fatalf("unexpected token hash: %s", repo.createData.TokenHash)
	}

	expectedExpiresAt := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	if !repo.createData.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: %v", repo.createData.ExpiresAt)
	}

	expectedCreatedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !repo.createData.CreatedAt.Equal(expectedCreatedAt) {
		t.Fatalf("unexpected created at: %v", repo.createData.CreatedAt)
	}

	if repo.createData.VerifiedAt != nil {
		t.Fatal("VerifiedAt should be nil")
	}

	if !urlBuilder.called {
		t.Fatal("url builder not called")
	}

	if urlBuilder.token != "plain-token" {
		t.Fatalf("unexpected url token: %s", urlBuilder.token)
	}

	if !mailer.called {
		t.Fatal("mail not sent")
	}

	if mailer.data.To != "test@example.com" {
		t.Fatalf("unexpected mail to: %s", mailer.data.To)
	}

	if mailer.data.URL != "http://example.com/user-registration/verify?token=plain-token" {
		t.Fatalf("unexpected mail url: %s", mailer.data.URL)
	}

	if mailer.data.Lang != "ja" {
		t.Fatalf("unexpected mail lang: %s", mailer.data.Lang)
	}
}

func TestUserRegistrationServiceCreateTrimEmail(t *testing.T) {
	repo := &mockUserRegistrationRequestRepository{}
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             " test@example.com ",
		EmailConfirmation: " test@example.com ",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if repo.findEmail != "test@example.com" {
		t.Fatalf("unexpected find email: %s", repo.findEmail)
	}

	if repo.createData.Email != "test@example.com" {
		t.Fatalf("unexpected create email: %s", repo.createData.Email)
	}

	if mailer.data.To != "test@example.com" {
		t.Fatalf("unexpected mail to: %s", mailer.data.To)
	}
}

func TestUserRegistrationServiceCreateExistingRequest(t *testing.T) {
	repo := &mockUserRegistrationRequestRepository{
		findResult: &model.UserRegistrationRequest{
			ID:        "existing-id",
			Email:     "test@example.com",
			TokenHash: "old-token-hash",
		},
	}
	uuidGenerator := &mockUUIDGenerator{}
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		uuidGenerator,
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if repo.createCalled {
		t.Fatal("Create should not be called")
	}

	if !repo.updateCalled {
		t.Fatal("UpdateToken not called")
	}

	if uuidGenerator.called {
		t.Fatal("uuid generator should not be called")
	}

	if repo.updateData == nil {
		t.Fatal("update data is nil")
	}

	if repo.updateData.ID != "existing-id" {
		t.Fatalf("unexpected id: %s", repo.updateData.ID)
	}

	if repo.updateData.TokenHash != "hashed-token" {
		t.Fatalf("unexpected token hash: %s", repo.updateData.TokenHash)
	}

	expectedExpiresAt := time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
	if !repo.updateData.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: %v", repo.updateData.ExpiresAt)
	}

	if !mailer.called {
		t.Fatal("mail not sent")
	}
}

func TestUserRegistrationServiceCreateAlreadyRegistered(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockUserRegistrationRequestRepository{
		findResult: &model.UserRegistrationRequest{
			ID:         "existing-id",
			Email:      "test@example.com",
			VerifiedAt: &now,
		},
	}
	txManager := &mockTxManager{}
	tokenGenerator := &mockTokenGenerator{}
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		repo,
		txManager,
		tokenGenerator,
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	appErr, ok := err.(*app_error.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}

	if appErr.Code != i18n.CodeUserAlreadyRegistered {
		t.Fatalf("unexpected code: %s", appErr.Code)
	}

	if appErr.Status != http.StatusConflict {
		t.Fatalf("unexpected status: %d", appErr.Status)
	}

	if tokenGenerator.called {
		t.Fatal("token generator should not be called")
	}

	if txManager.called {
		t.Fatal("transaction should not be called")
	}

	if mailer.called {
		t.Fatal("mail should not be sent")
	}
}

func TestUserRegistrationServiceCreateEmailRequired(t *testing.T) {
	repo := &mockUserRegistrationRequestRepository{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "",
		EmailConfirmation: "test@example.com",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	appErr, ok := err.(*app_error.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}

	if appErr.Code != i18n.CodeValidationError {
		t.Fatalf("unexpected code: %s", appErr.Code)
	}

	if appErr.Status != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", appErr.Status)
	}

	if appErr.FieldErrors["email"][0].Code != i18n.CodeEmailRequired {
		t.Fatalf("unexpected field error: %s", appErr.FieldErrors["email"][0].Code)
	}

	if repo.findCalled {
		t.Fatal("FindByEmail should not be called")
	}
}

func TestUserRegistrationServiceCreateEmailFormatInvalid(t *testing.T) {
	svc := newTestUserRegistrationService(
		&mockUserRegistrationRequestRepository{},
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "invalid",
		EmailConfirmation: "invalid",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	appErr, ok := err.(*app_error.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}

	if appErr.FieldErrors["email"][0].Code != i18n.CodeEmailFormatInvalid {
		t.Fatalf("unexpected field error: %s", appErr.FieldErrors["email"][0].Code)
	}
}

func TestUserRegistrationServiceCreateEmailConfirmationRequired(t *testing.T) {
	svc := newTestUserRegistrationService(
		&mockUserRegistrationRequestRepository{},
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	appErr, ok := err.(*app_error.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}

	if appErr.FieldErrors["email_confirmation"][0].Code != i18n.CodeEmailConfirmationRequired {
		t.Fatalf("unexpected field error: %s", appErr.FieldErrors["email_confirmation"][0].Code)
	}
}

func TestUserRegistrationServiceCreateEmailConfirmationNotMatch(t *testing.T) {
	svc := newTestUserRegistrationService(
		&mockUserRegistrationRequestRepository{},
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "other@example.com",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	appErr, ok := err.(*app_error.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}

	if appErr.FieldErrors["email_confirmation"][0].Code != i18n.CodeEmailConfirmationNotMatch {
		t.Fatalf("unexpected field error: %s", appErr.FieldErrors["email_confirmation"][0].Code)
	}
}

func TestUserRegistrationServiceCreateFindByEmailError(t *testing.T) {
	expectedErr := errors.New("find failed")
	repo := &mockUserRegistrationRequestRepository{
		findErr: expectedErr,
	}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserRegistrationServiceCreateTokenGenerateError(t *testing.T) {
	expectedErr := errors.New("generate failed")
	txManager := &mockTxManager{}

	svc := newTestUserRegistrationService(
		&mockUserRegistrationRequestRepository{},
		txManager,
		&mockTokenGenerator{err: expectedErr},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if txManager.called {
		t.Fatal("transaction should not be called")
	}
}

func TestUserRegistrationServiceCreateHashError(t *testing.T) {
	expectedErr := errors.New("hash failed")
	txManager := &mockTxManager{}

	svc := newTestUserRegistrationService(
		&mockUserRegistrationRequestRepository{},
		txManager,
		&mockTokenGenerator{},
		&mockTokenHasher{err: expectedErr},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if txManager.called {
		t.Fatal("transaction should not be called")
	}
}

func TestUserRegistrationServiceCreateUUIDError(t *testing.T) {
	expectedErr := errors.New("uuid failed")
	repo := &mockUserRegistrationRequestRepository{}
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{err: expectedErr},
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.createCalled {
		t.Fatal("Create should not be called")
	}

	if mailer.called {
		t.Fatal("mail should not be sent")
	}
}

func TestUserRegistrationServiceCreateCreateError(t *testing.T) {
	expectedErr := errors.New("create failed")
	repo := &mockUserRegistrationRequestRepository{
		createErr: expectedErr,
	}
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.createCalled {
		t.Fatal("Create not called")
	}

	if mailer.called {
		t.Fatal("mail should not be sent")
	}
}

func TestUserRegistrationServiceCreateUpdateError(t *testing.T) {
	expectedErr := errors.New("update failed")
	repo := &mockUserRegistrationRequestRepository{
		findResult: &model.UserRegistrationRequest{
			ID: "existing-id",
		},
		updateErr: expectedErr,
	}
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("UpdateToken not called")
	}

	if mailer.called {
		t.Fatal("mail should not be sent")
	}
}

func TestUserRegistrationServiceCreateMailError(t *testing.T) {
	expectedErr := errors.New("mail failed")
	repo := &mockUserRegistrationRequestRepository{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{err: expectedErr},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.createCalled {
		t.Fatal("Create not called")
	}
}

func TestUserRegistrationServiceCreateTransactionError(t *testing.T) {
	expectedErr := errors.New("tx failed")
	repo := &mockUserRegistrationRequestRepository{}
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{err: expectedErr},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.createCalled {
		t.Fatal("Create should not be called")
	}

	if repo.updateCalled {
		t.Fatal("UpdateToken should not be called")
	}

	if mailer.called {
		t.Fatal("mail should not be sent")
	}
}

func TestUserRegistrationServiceCreateEmptyLanguage(t *testing.T) {
	mailer := &mockMailer{}

	svc := newTestUserRegistrationService(
		&mockUserRegistrationRequestRepository{},
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		mailer,
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "",
	})
	if err != nil {
		t.Fatal(err)
	}

	if mailer.data.Lang != "" {
		t.Fatalf("unexpected mail lang: %s", mailer.data.Lang)
	}
}

type mockConfig30Minutes struct{}

func (m *mockConfig30Minutes) RegistrationTokenExpiresMinutes() int {
	return 30
}

func TestUserRegistrationServiceCreateExpiresAtByConfig(t *testing.T) {
	repo := &mockUserRegistrationRequestRepository{}

	svc := NewUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockClock{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
		&mockConfig30Minutes{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedExpiresAt := time.Date(2026, 1, 1, 0, 30, 0, 0, time.UTC)
	if !repo.createData.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("unexpected expires at: %v", repo.createData.ExpiresAt)
	}
}

func TestUserRegistrationServiceCreateMultipleValidationErrors(t *testing.T) {
	repo := &mockUserRegistrationRequestRepository{}

	svc := newTestUserRegistrationService(
		repo,
		&mockTxManager{},
		&mockTokenGenerator{},
		&mockTokenHasher{},
		&mockUUIDGenerator{},
		&mockMailer{},
		&mockRegistrationURLBuilder{},
	)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "",
		EmailConfirmation: "",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	appErr, ok := err.(*app_error.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}

	if appErr.Code != i18n.CodeValidationError {
		t.Fatalf("unexpected code: %s", appErr.Code)
	}

	if appErr.Status != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", appErr.Status)
	}

	if appErr.FieldErrors["email"][0].Code != i18n.CodeEmailRequired {
		t.Fatalf("unexpected email error: %s", appErr.FieldErrors["email"][0].Code)
	}

	if appErr.FieldErrors["email_confirmation"][0].Code != i18n.CodeEmailConfirmationRequired {
		t.Fatalf("unexpected email confirmation error: %s", appErr.FieldErrors["email_confirmation"][0].Code)
	}

	if repo.findCalled {
		t.Fatal("FindByEmail should not be called")
	}
}
