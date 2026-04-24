package repository

import (
	"context"
	"database/sql"
)

type PostgresTxManager struct {
	DB *sql.DB
}

func (m *PostgresTxManager) WithinTransaction(
	ctx context.Context,
	fn func(txCtx context.Context) error,
) error {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := WithTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
