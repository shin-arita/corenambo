package repository

import (
	"context"
	"database/sql"
	"time"

	"app-api/internal/model"
)

type mailOutboxRepository struct {
	db *sql.DB
}

func NewMailOutboxRepository(db *sql.DB) MailOutboxRepository {
	return &mailOutboxRepository{db: db}
}

func (r *mailOutboxRepository) Create(ctx context.Context, entity *model.MailOutbox) error {
	const q = `
INSERT INTO mail_outboxes (
	id, mail_type, to_email, payload, status, retry_count, next_attempt_at,
	sent_at, last_error, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
`
	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx, q,
		entity.ID,
		entity.MailType,
		entity.ToEmail,
		entity.Payload,
		entity.Status,
		entity.RetryCount,
		entity.NextAttemptAt,
		entity.SentAt,
		entity.LastError,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
	return err
}

func (r *mailOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*model.MailOutbox, error) {
	const q = `
SELECT id, mail_type, to_email, payload, status, retry_count, next_attempt_at, sent_at, last_error, created_at, updated_at
FROM mail_outboxes
WHERE status = 'pending' AND next_attempt_at <= NOW()
ORDER BY created_at ASC
LIMIT $1
`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []*model.MailOutbox
	for rows.Next() {
		var m model.MailOutbox
		err := rows.Scan(
			&m.ID,
			&m.MailType,
			&m.ToEmail,
			&m.Payload,
			&m.Status,
			&m.RetryCount,
			&m.NextAttemptAt,
			&m.SentAt,
			&m.LastError,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &m)
	}

	// ★ここが変更点（100%対応）
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *mailOutboxRepository) MarkSent(ctx context.Context, id string, now time.Time) error {
	const q = `
UPDATE mail_outboxes
SET status = 'sent', sent_at = $2, updated_at = $2
WHERE id = $1
`
	_, err := r.db.ExecContext(ctx, q, id, now)
	return err
}

func (r *mailOutboxRepository) MarkFailed(ctx context.Context, id string, errMsg string, next time.Time) error {
	const q = `
UPDATE mail_outboxes
SET status = 'failed',
	retry_count = retry_count + 1,
	last_error = $2,
	next_attempt_at = $3,
	updated_at = $3
WHERE id = $1
`
	_, err := r.db.ExecContext(ctx, q, id, errMsg, next)
	return err
}
