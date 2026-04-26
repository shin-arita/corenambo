package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMailOutboxRepositoryFetchPending_ScanError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	// 型不一致でScanエラーを発生させる
	rows := sqlmock.NewRows([]string{
		"id", "mail_type", "to_email", "payload",
		"status", "retry_count", "next_attempt_at",
		"sent_at", "last_error", "created_at", "updated_at",
	}).AddRow(
		"id",
		"user_registration",
		"a@b.com",
		"{}",
		"pending",
		"INVALID_INT", // ← ここでScanエラー
		time.Now(),
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery("SELECT .* FROM mail_outboxes").
		WithArgs(10).
		WillReturnRows(rows)

	repo := NewMailOutboxRepository(db)

	_, err := repo.FetchPending(context.Background(), 10)
	if err == nil {
		t.Fatal("expected scan error")
	}
}
