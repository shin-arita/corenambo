package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"app-api/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewUserEmailRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserEmailRepository(db)

	if repo == nil {
		t.Fatal("repo is nil")
	}
}

func userEmailRows(now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "user_id", "email", "is_primary", "verified_at", "created_at", "updated_at",
	}).AddRow("email-id", "user-id", "test@example.com", true, now, now, now)
}

func TestUserEmailRepositoryFindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT").
		WithArgs("test@example.com").
		WillReturnRows(userEmailRows(now))

	repo := NewUserEmailRepository(db)

	entity, err := repo.FindByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if entity == nil {
		t.Fatal("entity is nil")
	}

	if entity.Email != "test@example.com" {
		t.Fatalf("unexpected email: %s", entity.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserEmailRepositoryFindByEmailNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT").
		WithArgs("none@example.com").
		WillReturnError(sql.ErrNoRows)

	repo := NewUserEmailRepository(db)

	entity, err := repo.FindByEmail(context.Background(), "none@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if entity != nil {
		t.Fatal("entity should be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserEmailRepositoryFindByEmailError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	expectedErr := errors.New("select failed")

	mock.ExpectQuery("SELECT").
		WithArgs("test@example.com").
		WillReturnError(expectedErr)

	repo := NewUserEmailRepository(db)

	_, err = repo.FindByEmail(context.Background(), "test@example.com")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserEmailRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectExec("INSERT INTO user_emails").
		WithArgs("email-id", "user-id", "test@example.com", true, now, now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewUserEmailRepository(db)

	err = repo.Create(context.Background(), &model.UserEmail{
		ID:         "email-id",
		UserID:     "user-id",
		Email:      "test@example.com",
		IsPrimary:  true,
		VerifiedAt: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserEmailRepositoryCreateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedErr := errors.New("insert failed")

	mock.ExpectExec("INSERT INTO user_emails").
		WithArgs("email-id", "user-id", "test@example.com", true, now, now, now).
		WillReturnError(expectedErr)

	repo := NewUserEmailRepository(db)

	err = repo.Create(context.Background(), &model.UserEmail{
		ID:         "email-id",
		UserID:     "user-id",
		Email:      "test@example.com",
		IsPrimary:  true,
		VerifiedAt: &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserEmailRepositoryInterface(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserEmailRepository(db)
	if repo == nil {
		t.Fatal("repo should not be nil")
	}
}
