package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"app-api/internal/config"
	"app-api/internal/mail"
	"app-api/internal/model"
)

// ──── mock repo ────

type mockOutboxRepo struct {
	pending    []*model.MailOutbox
	lastOp     string
	lastReason string
}

func (r *mockOutboxRepo) FetchPending(_ context.Context, _ int) ([]*model.MailOutbox, error) {
	return r.pending, nil
}
func (r *mockOutboxRepo) MarkProcessing(_ context.Context, _ string) error { return nil }
func (r *mockOutboxRepo) MarkSent(_ context.Context, _ string, _ time.Time) error {
	r.lastOp = "sent"
	return nil
}
func (r *mockOutboxRepo) MarkRetry(_ context.Context, _ string, reason string, _ time.Time) error {
	r.lastOp = "retry"
	r.lastReason = reason
	return nil
}
func (r *mockOutboxRepo) MarkFailed(_ context.Context, _ string, reason string, _ time.Time) error {
	r.lastOp = "failed"
	r.lastReason = reason
	return nil
}
func (r *mockOutboxRepo) Create(_ context.Context, _ *model.MailOutbox) error       { return nil }
func (r *mockOutboxRepo) ResetStuckProcessing(_ context.Context, _ time.Time) error { return nil }

// ──── mock mailers ────

type nonRetryableMailer struct{}

func (m *nonRetryableMailer) SendUserRegistrationMail(_ context.Context, _ mail.UserRegistrationMail) error {
	return &mail.NonRetryableMailError{Msg: "registration URL is empty"}
}
func (m *nonRetryableMailer) SendUserAlreadyRegisteredMail(_ context.Context, _ string, _ string) error {
	return nil
}

type retryableMailer struct{}

func (m *retryableMailer) SendUserRegistrationMail(_ context.Context, _ mail.UserRegistrationMail) error {
	return errors.New("connection refused")
}
func (m *retryableMailer) SendUserAlreadyRegisteredMail(_ context.Context, _ string, _ string) error {
	return nil
}

// urlCheckMailer は URL が空のとき NonRetryableMailError を返す。
// payload["url"] = "" となるケースの SMTPMailer 挙動を再現する。
type urlCheckMailer struct{}

func (m *urlCheckMailer) SendUserRegistrationMail(_ context.Context, data mail.UserRegistrationMail) error {
	if data.URL == "" {
		return &mail.NonRetryableMailError{Msg: "registration URL is empty"}
	}
	return nil
}
func (m *urlCheckMailer) SendUserAlreadyRegisteredMail(_ context.Context, _ string, _ string) error {
	return nil
}

// ──── helpers ────

func defaultCfg() config.WorkerConfig {
	return config.WorkerConfig{MaxRetryCount: 3, RegistrationExpiresMinutes: 60, StuckTimeoutMinutes: 15}
}

func pendingOutbox(id, payload string) *model.MailOutbox {
	return &model.MailOutbox{
		ID:            id,
		MailType:      "user_registration",
		ToEmail:       "a@b.com",
		Payload:       payload,
		RetryCount:    0,
		NextAttemptAt: time.Now(),
	}
}

// ──── tests ────

func TestRunNonRetryableMailError_MarkFailed(t *testing.T) {
	repo := &mockOutboxRepo{
		pending: []*model.MailOutbox{
			pendingOutbox("id1", `{"url":"http://example.com/verify?token=x","lang":"ja","email":"a@b.com"}`),
		},
	}

	run(context.Background(), repo, &nonRetryableMailer{}, defaultCfg(), time.Minute)

	if repo.lastOp != "failed" {
		t.Fatalf("expected MarkFailed, got %s", repo.lastOp)
	}
}

func TestRunRetryableSMTPError_MarkRetry(t *testing.T) {
	repo := &mockOutboxRepo{
		pending: []*model.MailOutbox{
			pendingOutbox("id2", `{"url":"http://example.com/verify?token=x","lang":"ja","email":"a@b.com"}`),
		},
	}

	run(context.Background(), repo, &retryableMailer{}, defaultCfg(), time.Minute)

	if repo.lastOp != "retry" {
		t.Fatalf("expected MarkRetry, got %s", repo.lastOp)
	}
}

// payload = '{}' のとき url フィールドが欠落 → url="" → NonRetryableMailError → MarkFailed
func TestRunEmptyURLPayload_MarkFailedWithCorrectMessage(t *testing.T) {
	repo := &mockOutboxRepo{
		pending: []*model.MailOutbox{
			pendingOutbox("id3", `{}`),
		},
	}

	run(context.Background(), repo, &urlCheckMailer{}, defaultCfg(), time.Minute)

	if repo.lastOp != "failed" {
		t.Fatalf("expected MarkFailed, got op=%s", repo.lastOp)
	}
	if repo.lastReason != "registration URL is empty" {
		t.Fatalf("unexpected last_error: %q", repo.lastReason)
	}
}
