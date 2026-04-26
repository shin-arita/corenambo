package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFetchPending_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	r := NewMailOutboxRepository(db)

	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrConnDone)

	_, err = r.FetchPending(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
