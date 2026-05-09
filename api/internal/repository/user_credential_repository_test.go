package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"app-api/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewUserCredentialRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserCredentialRepository(db)

	if repo == nil {
		t.Fatal("repo is nil")
	}
}

func TestUserCredentialRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectExec("INSERT INTO user_credentials").
		WithArgs("user-id", "$2a$10$hash", now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewUserCredentialRepository(db)

	err = repo.Create(context.Background(), &model.UserCredential{
		UserID:            "user-id",
		PasswordHash:      "$2a$10$hash",
		PasswordChangedAt: now,
		CreatedAt:         now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserCredentialRepositoryCreateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedErr := errors.New("insert failed")

	mock.ExpectExec("INSERT INTO user_credentials").
		WithArgs("user-id", "$2a$10$hash", now, now).
		WillReturnError(expectedErr)

	repo := NewUserCredentialRepository(db)

	err = repo.Create(context.Background(), &model.UserCredential{
		UserID:            "user-id",
		PasswordHash:      "$2a$10$hash",
		PasswordChangedAt: now,
		CreatedAt:         now,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
