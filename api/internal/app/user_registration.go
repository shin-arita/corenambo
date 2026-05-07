package app

import (
	"app-api/internal/clock"
	"app-api/internal/config"
	"app-api/internal/registrationurl"
	"app-api/internal/repository"
	"app-api/internal/service"
	"app-api/internal/token"
	"app-api/internal/uuid"
)

func NewUserRegistrationService(
	db *repository.PostgresTxManager,
	cfg config.RegistrationConfig,
) service.UserRegistrationService {

	userRepo := repository.NewUserRegistrationRequestRepository(db.SQLDb())
	outboxRepo := repository.NewMailOutboxRepository(db.SQLDb())

	return service.NewUserRegistrationService(
		userRepo,
		outboxRepo,
		db,
		token.DefaultGenerator{},
		token.SHA256Hasher{},
		uuid.UUIDv7Generator{},
		clock.SystemClock{},
		registrationurl.NewStaticBuilder(cfg.FrontendBaseURL),
		cfg,
	)
}
