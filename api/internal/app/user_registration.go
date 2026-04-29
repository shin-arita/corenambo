package app

import (
	"os"

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
	cfg config.Config,
) service.UserRegistrationService {

	userRepo := repository.NewUserRegistrationRequestRepository(db.DB)
	outboxRepo := repository.NewMailOutboxRepository(db.DB)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return service.NewUserRegistrationService(
		userRepo,
		outboxRepo,
		db,
		token.DefaultGenerator{},
		token.SHA256Hasher{},
		uuid.UUIDv7Generator{},
		clock.SystemClock{},
		registrationurl.NewStaticBuilder(frontendURL),
		cfg,
	)
}
