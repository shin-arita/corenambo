package repository

import (
	"context"
	"database/sql"
)

type Executor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type txKeyType struct{}

var txKey = txKeyType{}

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func getExecutor(ctx context.Context, db *sql.DB) Executor {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return db
}
