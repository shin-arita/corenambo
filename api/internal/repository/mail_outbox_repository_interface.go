package repository

import (
	"context"
	"time"

	"app-api/internal/model"
)

type MailOutboxRepository interface {
	Create(ctx context.Context, entity *model.MailOutbox) error
	FetchPending(ctx context.Context, limit int) ([]*model.MailOutbox, error)
	MarkSent(ctx context.Context, id string, now time.Time) error
	MarkFailed(ctx context.Context, id string, errMsg string, next time.Time) error
}
