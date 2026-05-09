package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"app-api/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewUserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserRepository(db)

	if repo == nil {
		t.Fatal("repo is nil")
	}
}

func TestUserRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectExec("INSERT INTO users").
		WithArgs("user-id", "testuser", "active", now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewUserRepository(db)

	err = repo.Create(context.Background(), &model.User{
		ID:          "user-id",
		DisplayName: "testuser",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepositoryCreateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedErr := errors.New("insert failed")

	mock.ExpectExec("INSERT INTO users").
		WithArgs("user-id", "testuser", "active", now, now).
		WillReturnError(expectedErr)

	repo := NewUserRepository(db)

	err = repo.Create(context.Background(), &model.User{
		ID:          "user-id",
		DisplayName: "testuser",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepositoryInterface(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserRepository(db)
	if repo == nil {
		t.Fatal("repo should not be nil")
	}
}
