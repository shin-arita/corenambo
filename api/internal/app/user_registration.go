package app

import (
	"app-api/internal/clock"
	"app-api/internal/config"
	"app-api/internal/mail"
	"app-api/internal/repository"
	"app-api/internal/service"
	"app-api/internal/token"
	"app-api/internal/uuid"
)

type dummyBuilder struct{}

func (d dummyBuilder) Build(token string) string {
	return "http://example.com/" + token
}

func NewUserRegistrationService(
	db *repository.PostgresTxManager,
	cfg config.Config,
) service.UserRegistrationService {

	userRepo := repository.NewUserRegistrationRequestRepository(db.DB)
	outboxRepo := repository.NewMailOutboxRepository(db.DB)

	return service.NewUserRegistrationService(
		userRepo,
		outboxRepo,
		db,
		token.DefaultGenerator{},
		token.SHA256Hasher{},
		uuid.UUIDv7Generator{},
		clock.SystemClock{},
		&mail.NoopMailer{},
		dummyBuilder{},
		cfg,
	)
}
