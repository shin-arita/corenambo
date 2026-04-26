package repository

import (
	"context"
	"testing"
	"time"

	"app-api/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMailOutboxRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectExec("INSERT INTO mail_outboxes").
		WithArgs(
			"id",
			"user_registration",
			"test@example.com",
			"{}",
			"pending",
			0,
			now,
			nil,
			nil,
			now,
			now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewMailOutboxRepository(db)

	err = repo.Create(context.Background(), &model.MailOutbox{
		ID:            "id",
		MailType:      "user_registration",
		ToEmail:       "test@example.com",
		Payload:       "{}",
		Status:        "pending",
		RetryCount:    0,
		NextAttemptAt: now,
		SentAt:        nil,
		LastError:     nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
