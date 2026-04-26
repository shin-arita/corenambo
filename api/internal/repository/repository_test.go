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

func TestNewUserRegistrationRequestRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserRegistrationRequestRepository(db)

	if repo == nil {
		t.Fatal("repo is nil")
	}
}

func userRegistrationRequestRows(now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"email",
		"token_hash",
		"expires_at",
		"verified_at",
		"last_sent_at",
		"created_at",
	}).AddRow(
		"id",
		"test@example.com",
		"token-hash",
		now,
		nil,
		nil,
		now,
	)
}

func TestUserRegistrationRequestRepositoryFindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT").
		WithArgs("test@example.com").
		WillReturnRows(userRegistrationRequestRows(now))

	repo := NewUserRegistrationRequestRepository(db)

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

	if entity.LastSentAt != nil {
		t.Fatal("last sent at should be nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRegistrationRequestRepositoryFindByEmailNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT").
		WithArgs("none@example.com").
		WillReturnError(sql.ErrNoRows)

	repo := NewUserRegistrationRequestRepository(db)

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

func TestUserRegistrationRequestRepositoryFindByEmailError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	expectedErr := errors.New("select failed")

	mock.ExpectQuery("SELECT").
		WithArgs("test@example.com").
		WillReturnError(expectedErr)

	repo := NewUserRegistrationRequestRepository(db)

	_, err = repo.FindByEmail(context.Background(), "test@example.com")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRegistrationRequestRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectExec("INSERT INTO user_registration_requests").
		WithArgs("id", "test@example.com", "token-hash", now, nil, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewUserRegistrationRequestRepository(db)

	err = repo.Create(context.Background(), &model.UserRegistrationRequest{
		ID:         "id",
		Email:      "test@example.com",
		TokenHash:  "token-hash",
		ExpiresAt:  now,
		VerifiedAt: nil,
		LastSentAt: &now,
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRegistrationRequestRepositoryCreateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedErr := errors.New("insert failed")

	mock.ExpectExec("INSERT INTO user_registration_requests").
		WithArgs("id", "test@example.com", "token-hash", now, nil, sqlmock.AnyArg(), now).
		WillReturnError(expectedErr)

	repo := NewUserRegistrationRequestRepository(db)

	err = repo.Create(context.Background(), &model.UserRegistrationRequest{
		ID:         "id",
		Email:      "test@example.com",
		TokenHash:  "token-hash",
		ExpiresAt:  now,
		VerifiedAt: nil,
		LastSentAt: &now,
		CreatedAt:  now,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRegistrationRequestRepositoryUpdateToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectExec("UPDATE user_registration_requests").
		WithArgs("id", "token-hash", now, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := NewUserRegistrationRequestRepository(db)

	err = repo.UpdateToken(context.Background(), &model.UserRegistrationRequest{
		ID:         "id",
		TokenHash:  "token-hash",
		ExpiresAt:  now,
		LastSentAt: &now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRegistrationRequestRepositoryUpdateTokenError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedErr := errors.New("update failed")

	mock.ExpectExec("UPDATE user_registration_requests").
		WithArgs("id", "token-hash", now, sqlmock.AnyArg()).
		WillReturnError(expectedErr)

	repo := NewUserRegistrationRequestRepository(db)

	err = repo.UpdateToken(context.Background(), &model.UserRegistrationRequest{
		ID:         "id",
		TokenHash:  "token-hash",
		ExpiresAt:  now,
		LastSentAt: &now,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresTxManagerWithinTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	m := &PostgresTxManager{DB: db}

	err = m.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		if getExecutor(txCtx, db) == db {
			t.Fatal("expected tx executor")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresTxManagerWithinTransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	expectedErr := errors.New("begin failed")

	mock.ExpectBegin().WillReturnError(expectedErr)

	m := &PostgresTxManager{DB: db}

	err = m.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		t.Fatal("should not be called")
		return nil
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresTxManagerWithinTransactionRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	expectedErr := errors.New("fn failed")

	mock.ExpectBegin()
	mock.ExpectRollback()

	m := &PostgresTxManager{DB: db}

	err = m.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresTxManagerWithinTransactionCommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	expectedErr := errors.New("commit failed")

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(expectedErr)

	m := &PostgresTxManager{DB: db}

	err = m.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		return nil
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetExecutorWithoutTx(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	executor := getExecutor(context.Background(), db)

	if executor != db {
		t.Fatal("expected db executor")
	}
}

func TestTxManagerInterface(t *testing.T) {
	var _ TxManager = &PostgresTxManager{}
}

func TestUserRegistrationRequestRepositoryInterface(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	_ = NewUserRegistrationRequestRepository(db)
}

func TestUserRegistrationRequestRepositoryFindByTokenHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT").
		WithArgs("token-hash").
		WillReturnRows(userRegistrationRequestRows(now))

	repo := NewUserRegistrationRequestRepository(db)

	entity, err := repo.FindByTokenHash(context.Background(), "token-hash")
	if err != nil {
		t.Fatal(err)
	}

	if entity == nil {
		t.Fatal("entity is nil")
	}

	if entity.TokenHash != "token-hash" {
		t.Fatalf("unexpected token hash: %s", entity.TokenHash)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRegistrationRequestRepositoryFindByTokenHashNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT").
		WithArgs("none").
		WillReturnError(sql.ErrNoRows)

	repo := NewUserRegistrationRequestRepository(db)

	entity, err := repo.FindByTokenHash(context.Background(), "none")
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

func TestUserRegistrationRequestRepositoryFindByTokenHashError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	expectedErr := errors.New("select failed")

	mock.ExpectQuery("SELECT").
		WithArgs("token-hash").
		WillReturnError(expectedErr)

	repo := NewUserRegistrationRequestRepository(db)

	_, err = repo.FindByTokenHash(context.Background(), "token-hash")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
