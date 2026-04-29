package repository

import (
	"context"
	"database/sql"
	"time"

	"app-api/internal/model"
)

type MailOutboxRepository interface {
	FetchPending(ctx context.Context, limit int) ([]*model.MailOutbox, error)
	MarkProcessing(ctx context.Context, id string) error
	MarkSent(ctx context.Context, id string, sentAt time.Time) error
	MarkFailed(ctx context.Context, id string, reason string, nextRetryAt time.Time) error
	Create(ctx context.Context, m *model.MailOutbox) error
	ResetStuckProcessing(ctx context.Context, stuckBefore time.Time) error
}

type mailOutboxRepository struct {
	db *sql.DB
}

func NewMailOutboxRepository(db *sql.DB) MailOutboxRepository {
	return &mailOutboxRepository{db: db}
}

func (r *mailOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*model.MailOutbox, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, mail_type, to_email, payload, status, retry_count, next_attempt_at,
       sent_at, last_error, created_at, updated_at
FROM mail_outboxes
WHERE status = 'pending'
  AND next_attempt_at <= NOW()
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*model.MailOutbox
	for rows.Next() {
		m := &model.MailOutbox{}
		if err := rows.Scan(
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
		); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, nil
}

func (r *mailOutboxRepository) MarkProcessing(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE mail_outboxes
SET status = 'processing', updated_at = NOW()
WHERE id = $1
`, id)
	return err
}

func (r *mailOutboxRepository) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE mail_outboxes
SET status = 'sent',
    sent_at = $2,
    updated_at = $2
WHERE id = $1
`, id, sentAt)
	return err
}

func (r *mailOutboxRepository) MarkFailed(ctx context.Context, id string, reason string, nextRetryAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE mail_outboxes
SET status = 'failed',
    last_error = $2,
    next_attempt_at = $3,
    retry_count = retry_count + 1,
    updated_at = NOW()
WHERE id = $1
`, id, reason, nextRetryAt)
	return err
}

func (r *mailOutboxRepository) Create(ctx context.Context, m *model.MailOutbox) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO mail_outboxes (
	id, mail_type, to_email, payload, status,
	retry_count, next_attempt_at, sent_at, last_error, created_at, updated_at
) VALUES (
	$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
)
`, m.ID, m.MailType, m.ToEmail, m.Payload, m.Status,
		m.RetryCount, m.NextAttemptAt, m.SentAt, m.LastError, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *mailOutboxRepository) ResetStuckProcessing(ctx context.Context, stuckBefore time.Time) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE mail_outboxes
SET status = 'pending',
    next_attempt_at = NOW(),
    updated_at = NOW()
WHERE status = 'processing'
  AND updated_at < $1
`, stuckBefore)
	return err
}
