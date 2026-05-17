package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"app-api/internal/token"
)

// ──── Helpers ────

var validVerifyInput = VerifyUserRegistrationInput{
	Token:                "token",
	DisplayName:          "testuser",
	Password:             "password123",
	PasswordConfirmation: "password123",
	AgreedToTerms:        true,
}

func newVerifyServiceDefault() UserRegistrationService {
	future := time.Now().Add(60 * time.Minute)
	return newVerifyService(
		&tokenFoundUserRegRepo{expiresAt: future},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
}

// ──── Create tests ────

func TestCreate(t *testing.T) {
	svc := newTestUserRegistrationService()
	out, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ExpiresMinutes != 60 {
		t.Fatalf("expected ExpiresMinutes=60, got %d", out.ExpiresMinutes)
	}
}

func TestCreate_EmailEmpty(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "",
		EmailConfirmation: "",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_EmailFormatInvalid(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "not-an-email",
		EmailConfirmation: "not-an-email",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_EmailWithDisplayNameRejected(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "Name <user@example.com>",
		EmailConfirmation: "Name <user@example.com>",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error for display-name format email")
	}
}

func TestCreate_EmailWithAngleBracketsRejected(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "<user@example.com>",
		EmailConfirmation: "<user@example.com>",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error for angle-bracket format email")
	}
}

func TestCreate_EmailConfirmationNotMatch(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "other@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_EmailConfirmationEmpty(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_AlreadyVerified(t *testing.T) {
	svc := newTestUserRegistrationServiceWithVerifiedUser()
	out, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ExpiresMinutes != 60 {
		t.Fatalf("expected ExpiresMinutes=60, got %d", out.ExpiresMinutes)
	}
}

func TestCreate_ResendNotAvailable(t *testing.T) {
	svc := newTestUserRegistrationServiceWithRecentlySentUser()
	out, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.ExpiresMinutes != 60 {
		t.Fatalf("expected ExpiresMinutes=60, got %d", out.ExpiresMinutes)
	}
}

func TestCreate_UpdateToken(t *testing.T) {
	svc := newTestUserRegistrationServiceWithExpiredUser()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate_DBError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithDBError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_TokenGenError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithTokenGenError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_TokenHashError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithTokenHashError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_FirstUUIDError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithFirstUUIDError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_SecondUUIDError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithSecondUUIDError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_CreateUserError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithCreateUserError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_UpdateTokenError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithUpdateTokenError()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_MarshalError(t *testing.T) {
	orig := jsonMarshal
	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("mock marshal error")
	}
	t.Cleanup(func() { jsonMarshal = orig })

	svc := newTestUserRegistrationService()
	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreate_EmailNormalized(t *testing.T) {
	repo := &captureCreateRepo{}
	svc := newTestUserRegistrationServiceWithCaptureRepo(repo)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "Test@Example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.capturedEmail != "test@example.com" {
		t.Fatalf("expected test@example.com, got %s", repo.capturedEmail)
	}
}

func TestCreate_RawTokenPassedToURLBuilder(t *testing.T) {
	builder := &captureURLBuilder{}
	svc := newTestUserRegistrationServiceWithCaptureURLBuilder(builder)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if builder.capturedToken != "token" {
		t.Fatalf("expected plainToken %q to be passed to URL builder, got %q", "token", builder.capturedToken)
	}
	if builder.capturedToken == "hash" {
		t.Fatal("token hash was passed to URL builder instead of plain token")
	}
}

func TestCreate_DBStoresHashNotRawToken(t *testing.T) {
	repo := &captureCreateUserRegistrationRepo{}
	svc := newTestUserRegistrationServiceWithCaptureUserRegistrationRepo(repo)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if repo.capturedTokenHash == "token" {
		t.Fatal("raw token was stored in DB instead of hash")
	}
	if repo.capturedTokenHash != "hash" {
		t.Fatalf("expected hash %q stored in DB, got %q", "hash", repo.capturedTokenHash)
	}
}

func TestCreate_OutboxPayloadContainsTokenURL(t *testing.T) {
	outboxRepo := &captureOutboxRepo{}
	svc := newTestUserRegistrationServiceWithCaptureOutbox(outboxRepo)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(outboxRepo.capturedPayload), &payload); err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	if !strings.Contains(payload["url"], "token=token") {
		t.Fatalf("outbox payload URL does not contain raw token: %s", payload["url"])
	}
}

func TestCreate_OutboxPayloadURLNotEmptyToken(t *testing.T) {
	outboxRepo := &captureOutboxRepo{}
	svc := newTestUserRegistrationServiceWithCaptureOutbox(outboxRepo)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(outboxRepo.capturedPayload), &payload); err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}

	if strings.HasSuffix(payload["url"], "token=") {
		t.Fatalf("outbox payload URL ends with empty token: %s", payload["url"])
	}
}

func TestCreate_EmptyTokenFails(t *testing.T) {
	svc := newTestUserRegistrationServiceWithEmptyTokenGen()

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err == nil {
		t.Fatal("expected error when token generator returns empty token")
	}
}

func TestCreate_OutboxMailTypeIsUserRegistration(t *testing.T) {
	outbox := &captureFullOutboxRepo{}
	svc := newTestUserRegistrationServiceWithCaptureFullOutbox(outbox, fixedClock{t: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)})

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if outbox.capturedOutbox.MailType != "user_registration" {
		t.Fatalf("unexpected mail_type: got=%q want=%q", outbox.capturedOutbox.MailType, "user_registration")
	}
}

func TestCreate_OutboxStatusIsPending(t *testing.T) {
	outbox := &captureFullOutboxRepo{}
	svc := newTestUserRegistrationServiceWithCaptureFullOutbox(outbox, fixedClock{t: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)})

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if outbox.capturedOutbox.Status != "pending" {
		t.Fatalf("unexpected status: got=%q want=%q", outbox.capturedOutbox.Status, "pending")
	}
}

func TestCreate_OutboxNextAttemptAtIsNow(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	outbox := &captureFullOutboxRepo{}
	svc := newTestUserRegistrationServiceWithCaptureFullOutbox(outbox, fixedClock{t: fixedTime})

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !outbox.capturedOutbox.NextAttemptAt.Equal(fixedTime) {
		t.Fatalf("unexpected next_attempt_at: got=%v want=%v", outbox.capturedOutbox.NextAttemptAt, fixedTime)
	}
}

func TestCreate_TokenHashRoundTrip(t *testing.T) {
	repo := &captureCreateUserRegistrationRepo{}
	builder := &captureURLBuilder{}
	svc := newTestUserRegistrationServiceWithRealHasherAndCapture(repo, builder)

	_, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatal(err)
	}

	// builder.capturedToken はメールURLに埋め込まれた平文token
	// repo.capturedTokenHash はDBに保存されたhash
	// SHA256(url_token) == stored_token_hash であることを検証
	hasher := token.SHA256Hasher{}
	wantHash, err := hasher.Hash(builder.capturedToken)
	if err != nil {
		t.Fatalf("failed to hash URL token: %v", err)
	}
	if wantHash != repo.capturedTokenHash {
		t.Fatalf("token hash mismatch: SHA256(url_token)=%q stored_hash=%q", wantHash, repo.capturedTokenHash)
	}
}

// ──── Verify validation tests ────

func TestVerify_Validation_DisplayNameRequired(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = ""
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestVerify_Validation_DisplayNameWhitespace(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "   "
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for whitespace display_name")
	}
}

func TestVerify_Validation_PasswordRequired(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Password = ""
	in.PasswordConfirmation = ""
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestVerify_Validation_PasswordConfirmationRequired(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.PasswordConfirmation = ""
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestVerify_Validation_PasswordConfirmationNotMatch(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.PasswordConfirmation = "different"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestVerify_Validation_PasswordTooShort(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Password = "pass123"
	in.PasswordConfirmation = "pass123"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for short password")
	}
}

func TestVerify_Validation_PasswordNoLetter(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Password = "12345678"
	in.PasswordConfirmation = "12345678"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for password with no letter")
	}
}

func TestVerify_Validation_PasswordNoDigit(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Password = "password"
	in.PasswordConfirmation = "password"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for password with no digit")
	}
}

func TestVerify_Validation_PasswordMinLength(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Password = "passw0rd"
	in.PasswordConfirmation = "passw0rd"
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for 8-char valid password: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_PasswordTooLong(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Password = strings.Repeat("a", 72) + "1" // 73バイト: bcrypt上限超え
	in.PasswordConfirmation = in.Password
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for 73-byte password")
	}
}

func TestVerify_Validation_PasswordMaxLength(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Password = strings.Repeat("a", 71) + "1" // 72バイト: bcrypt上限内（英字+数字）
	in.PasswordConfirmation = in.Password
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for 72-byte password: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_AgreedToTermsRequired(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.AgreedToTerms = false
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

// ──── Verify success and token error tests ────

func TestVerify(t *testing.T) {
	svc := newVerifyServiceDefault()
	out, err := svc.Verify(context.Background(), validVerifyInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_EmptyToken(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.Token = ""
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestVerify_TokenHashError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := NewUserRegistrationService(
		&tokenFoundUserRegRepo{expiresAt: future},
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
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from token hasher")
	}
}

func TestVerify_PasswordHashError(t *testing.T) {
	orig := hashPassword
	hashPassword = func(p string) (string, error) { return "", errors.New("bcrypt error") }
	t.Cleanup(func() { hashPassword = orig })

	svc := newVerifyServiceDefault()
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from password hasher")
	}
}

func TestVerify_TokenNotFound(t *testing.T) {
	svc := newVerifyService(
		&tokenNotFoundUserRegRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error for token not found")
	}
}

func TestVerify_AlreadyVerified(t *testing.T) {
	svc := newVerifyService(
		&dummyVerifiedUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error for already verified")
	}
}

func TestVerify_TokenExpired(t *testing.T) {
	svc := newVerifyService(
		&tokenExpiredUserRegRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerify_EmailAlreadyExists(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newVerifyService(
		&tokenFoundUserRegRepo{expiresAt: future},
		&dummyUserModelRepo{},
		&dummyExistingEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error for existing email")
	}
}

func TestVerify_FindEmailError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newVerifyService(
		&tokenFoundUserRegRepo{expiresAt: future},
		&dummyUserModelRepo{},
		&dummyErrorFindEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from FindByEmail")
	}
}

func TestVerify_UserCreateError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newVerifyService(
		&tokenFoundUserRegRepo{expiresAt: future},
		&dummyErrorUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from user create")
	}
}

func TestVerify_UserEmailCreateError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newVerifyService(
		&tokenFoundUserRegRepo{expiresAt: future},
		&dummyUserModelRepo{},
		&dummyErrorUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from user email create")
	}
}

func TestVerify_UserCredentialCreateError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newVerifyService(
		&tokenFoundUserRegRepo{expiresAt: future},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyErrorUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from credential create")
	}
}

func TestVerify_UpdateVerifiedAtError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newVerifyService(
		&errorUpdateVerifiedAtRepo{expiresAt: future},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from update verified at")
	}
}

func TestVerify_DBError(t *testing.T) {
	svc := newVerifyService(
		&dummyErrorUserRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected db error")
	}
}

func TestVerify_FirstUUIDError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := NewUserRegistrationService(
		&tokenFoundUserRegRepo{expiresAt: future},
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
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from first UUID in Verify")
	}
}

func TestVerify_SecondUUIDError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := NewUserRegistrationService(
		&tokenFoundUserRegRepo{expiresAt: future},
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
	_, err := svc.Verify(context.Background(), validVerifyInput)
	if err == nil {
		t.Fatal("expected error from second UUID in Verify")
	}
}

func TestVerify_DisplayNameTrimmed(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "  testuser  "
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameTooShort(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "ab"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for too short display_name")
	}
}

func TestVerify_Validation_DisplayNameMinLength(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "abc"
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for 3-char display_name: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameMaxLength(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = strings.Repeat("a", 30)
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for 30-char display_name: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameTooLong(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = strings.Repeat("a", 31)
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for too long display_name")
	}
}

func TestVerify_Validation_DisplayNameNewline(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "user\nname"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for newline in display_name")
	}
}

func TestVerify_Validation_DisplayNameTab(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "user\tname"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for tab in display_name")
	}
}

func TestVerify_Validation_DisplayNameControlChar(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "user\x01name"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for control char in display_name")
	}
}

func TestVerify_Validation_DisplayNameLineSeparator(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "user name"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for LINE SEPARATOR in display_name")
	}
}

func TestVerify_Validation_DisplayNameZeroWidthSpace(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "user\u200bname"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for ZERO WIDTH SPACE in display_name")
	}
}

func TestVerify_Validation_DisplayNameSoftHyphen(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "user\u00adname"
	_, err := svc.Verify(context.Background(), in)
	if err == nil {
		t.Fatal("expected validation error for SOFT HYPHEN in display_name")
	}
}

func TestVerify_Validation_DisplayNameZWJAllowed(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "usr\u200dnam"
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for ZWJ in display_name: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameEmojiAllowed(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "abc😀def"
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for emoji in display_name: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameNFCNormalized(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	// NFD: "か" (U+304B) + combining dakuten (U+3099) = "が" in NFD form
	// 8 NFD code points but represents valid NFC after normalization
	in.DisplayName = "abcがabc"
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for NFD display_name: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameJapaneseAllowed(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "コレナンボユーザ"
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for Japanese display_name: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameInternalSpaceAllowed(t *testing.T) {
	svc := newVerifyServiceDefault()
	in := validVerifyInput
	in.DisplayName = "user name"
	out, err := svc.Verify(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error for display_name with internal space: %v", err)
	}
	if out.Code == "" {
		t.Fatal("expected non-empty code")
	}
}

func TestVerify_Validation_DisplayNameReservedJA(t *testing.T) {
	reserved := []string{"管理者", "運営", "公式", "サポート", "システム"}
	for _, name := range reserved {
		svc := newVerifyServiceDefault()
		in := validVerifyInput
		in.DisplayName = name
		_, err := svc.Verify(context.Background(), in)
		if err == nil {
			t.Fatalf("expected validation error for reserved display_name %q", name)
		}
	}
}

func TestVerify_Validation_DisplayNameReservedEN(t *testing.T) {
	reserved := []string{"admin", "administrator", "official", "support", "system", "root"}
	for _, name := range reserved {
		svc := newVerifyServiceDefault()
		in := validVerifyInput
		in.DisplayName = name
		_, err := svc.Verify(context.Background(), in)
		if err == nil {
			t.Fatalf("expected validation error for reserved display_name %q", name)
		}
	}
}

func TestVerify_Validation_DisplayNameReservedCaseInsensitive(t *testing.T) {
	cases := []string{"Admin", "ADMIN", "AdMiN", "System", "SYSTEM"}
	for _, name := range cases {
		svc := newVerifyServiceDefault()
		in := validVerifyInput
		in.DisplayName = name
		_, err := svc.Verify(context.Background(), in)
		if err == nil {
			t.Fatalf("expected validation error for reserved display_name (case variant) %q", name)
		}
	}
}

func TestVerify_Validation_DisplayNameReservedZH(t *testing.T) {
	reserved := []string{"管理员", "官方", "客服", "系统"}
	for _, name := range reserved {
		svc := newVerifyServiceDefault()
		in := validVerifyInput
		in.DisplayName = name
		_, err := svc.Verify(context.Background(), in)
		if err == nil {
			t.Fatalf("expected validation error for reserved display_name %q", name)
		}
	}
}

func TestVerify_Validation_DisplayNamePartialReservedAllowed(t *testing.T) {
	// 部分一致では禁止しない
	allowed := []string{"adminuser", "myadmin", "system1", "支持admin"}
	for _, name := range allowed {
		svc := newVerifyServiceDefault()
		in := validVerifyInput
		in.DisplayName = name
		out, err := svc.Verify(context.Background(), in)
		if err != nil {
			t.Fatalf("unexpected error for partial-match display_name %q: %v", name, err)
		}
		if out.Code == "" {
			t.Fatalf("expected non-empty code for %q", name)
		}
	}
}

func TestCreate_AlreadyVerified_IgnoresTokenGenError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithVerifiedUserAndTokenGenError()
	out, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatalf("already-verified user must succeed even if token gen would fail, got: %v", err)
	}
	if out.ExpiresMinutes != 60 {
		t.Fatalf("expected ExpiresMinutes=60, got %d", out.ExpiresMinutes)
	}
}

func TestCreate_ResendNotAvailable_IgnoresTokenGenError(t *testing.T) {
	svc := newTestUserRegistrationServiceWithRecentlySentUserAndTokenGenError()
	out, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatalf("recently-sent user must succeed even if token gen would fail, got: %v", err)
	}
	if out.ExpiresMinutes != 60 {
		t.Fatalf("expected ExpiresMinutes=60, got %d", out.ExpiresMinutes)
	}
}

func TestCreate_ConcurrentDuplicateEmail(t *testing.T) {
	svc := newTestUserRegistrationServiceWithDuplicateEmailRepo()
	out, err := svc.Create(context.Background(), CreateUserRegistrationInput{
		Email:             "test@example.com",
		EmailConfirmation: "test@example.com",
		Language:          "ja",
	})
	if err != nil {
		t.Fatalf("expected success on concurrent duplicate email, got: %v", err)
	}
	if out.ExpiresMinutes != 60 {
		t.Fatalf("expected ExpiresMinutes=60, got %d", out.ExpiresMinutes)
	}
}

func TestVerify_BcryptSkippedForInvalidToken(t *testing.T) {
	called := false
	orig := bcryptGenerate
	bcryptGenerate = func(_ []byte, _ int) ([]byte, error) {
		called = true
		return []byte("hash"), nil
	}
	t.Cleanup(func() { bcryptGenerate = orig })

	svc := newVerifyService(
		&tokenNotFoundUserRegRepo{},
		&dummyUserModelRepo{},
		&dummyUserEmailRepo{},
		&dummyUserCredentialRepo{},
	)
	_, _ = svc.Verify(context.Background(), validVerifyInput)
	if called {
		t.Fatal("bcrypt was called even though token was not found — CPU DoS vector open")
	}
}

func TestHashPasswordBcryptError(t *testing.T) {
	orig := bcryptGenerate
	bcryptGenerate = func(_ []byte, _ int) ([]byte, error) {
		return nil, errors.New("bcrypt error")
	}
	t.Cleanup(func() { bcryptGenerate = orig })

	_, err := hashPassword("test")
	if err == nil {
		t.Fatal("expected error from bcrypt")
	}
}

// ──── CheckToken tests ────

func TestCheckToken_Valid(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newCheckTokenService(
		&checkTokenFoundRepo{expiresAt: future},
		&dummyUserEmailRepo{},
	)

	out, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Code != "REGISTRATION_TOKEN_VALID" {
		t.Fatalf("expected REGISTRATION_TOKEN_VALID, got %s", out.Code)
	}
}

func TestCheckToken_EmptyToken(t *testing.T) {
	svc := newCheckTokenService(&checkTokenFoundRepo{}, &dummyUserEmailRepo{})

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: ""})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckToken_TokenNotFound(t *testing.T) {
	svc := newCheckTokenService(&checkTokenNotFoundRepo{}, &dummyUserEmailRepo{})

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckToken_TokenExpired(t *testing.T) {
	past := time.Now().Add(-2 * time.Hour)
	svc := newCheckTokenService(
		&checkTokenFoundRepo{expiresAt: past},
		&dummyUserEmailRepo{},
	)

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckToken_TokenUsed(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	now := time.Now()
	svc := newCheckTokenService(
		&checkTokenFoundRepo{expiresAt: future, verifiedAt: &now},
		&dummyUserEmailRepo{},
	)

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckToken_UserAlreadyRegistered(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newCheckTokenService(
		&checkTokenFoundRepo{expiresAt: future},
		&dummyExistingEmailRepo{},
	)

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckToken_DBError(t *testing.T) {
	svc := newCheckTokenService(&checkTokenFindErrorRepo{}, &dummyUserEmailRepo{})

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckToken_HashError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := NewUserRegistrationService(
		&checkTokenFoundRepo{expiresAt: future},
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

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error from hash")
	}
}

func TestCheckToken_EmailFindError(t *testing.T) {
	future := time.Now().Add(60 * time.Minute)
	svc := newCheckTokenService(
		&checkTokenFoundRepo{expiresAt: future},
		&dummyErrorFindEmailRepo{},
	)

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckToken_VerifiedBeforeExpired(t *testing.T) {
	past := time.Now().Add(-2 * time.Hour)
	now := time.Now()
	// verified_at あり + 期限切れ → verified_at チェックが先に検出される
	svc := newCheckTokenService(
		&checkTokenFoundRepo{expiresAt: past, verifiedAt: &now},
		&dummyUserEmailRepo{},
	)

	_, err := svc.CheckToken(context.Background(), CheckTokenInput{Token: "token"})
	if err == nil {
		t.Fatal("expected error")
	}
}
