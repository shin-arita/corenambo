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

	regRepo := repository.NewUserRegistrationRequestRepository(db.SQLDb())
	userRepo := repository.NewUserRepository(db.SQLDb())
	userEmailRepo := repository.NewUserEmailRepository(db.SQLDb())
	userCredentialRepo := repository.NewUserCredentialRepository(db.SQLDb())
	outboxRepo := repository.NewMailOutboxRepository(db.SQLDb())

	return service.NewUserRegistrationService(
		regRepo,
		userRepo,
		userEmailRepo,
		userCredentialRepo,
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
