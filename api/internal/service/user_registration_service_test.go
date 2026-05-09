package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

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

	// dummyTokenGen returns "token", dummyTokenHasher returns "hash"
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

	// dummyTokenGen returns "token", dummyTokenHasher returns "hash"
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

	// dummyTokenGen returns "token", so URL must contain "token=token"
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

func TestVerify(t *testing.T) {
	svc := newTestUserRegistrationService()
	_, err := svc.Verify(context.Background(), VerifyUserRegistrationInput{
		Token: "token",
	})
	if err == nil {
		t.Fatal("expected error: Verify is not implemented")
	}
}
