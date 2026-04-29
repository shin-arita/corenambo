package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMailOutboxRepositoryFetchPending(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{
		"id", "mail_type", "to_email", "payload",
		"status", "retry_count", "next_attempt_at",
		"sent_at", "last_error", "created_at", "updated_at",
	}).AddRow(
		"id", "user_registration", "a@b.com", "{}",
		"pending", 0, time.Now(),
		nil, nil, time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT .* FROM mail_outboxes").
		WithArgs(10).
		WillReturnRows(rows)

	repo := NewMailOutboxRepository(db)

	list, err := repo.FetchPending(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("unexpected length: %d", len(list))
	}
}

func TestMailOutboxRepositoryMarkSent(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	mock.ExpectExec("UPDATE mail_outboxes").
		WithArgs("id", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewMailOutboxRepository(db)

	err := repo.MarkSent(context.Background(), "id", time.Now())
	if err != nil {
		t.Fatal(err)
	}
}

func TestMailOutboxRepositoryMarkFailed(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	mock.ExpectExec("UPDATE mail_outboxes").
		WithArgs("id", "err", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewMailOutboxRepository(db)

	err := repo.MarkFailed(context.Background(), "id", "err", time.Now())
	if err != nil {
		t.Fatal(err)
	}
}

func TestMailOutboxRepositoryResetStuckProcessing(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	mock.ExpectExec("UPDATE mail_outboxes").
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewMailOutboxRepository(db)

	err := repo.ResetStuckProcessing(context.Background(), time.Now().Add(-15*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
