package app

import (
	"database/sql"

	"app-api/internal/clock"
	"app-api/internal/config"
	"app-api/internal/handler"
	"app-api/internal/i18n"
	"app-api/internal/mail"
	"app-api/internal/registrationurl"
	"app-api/internal/repository"
	"app-api/internal/service"
	"app-api/internal/token"
	"app-api/internal/uuid"
)

func NewUserRegistrationHandler(db *sql.DB, cfg config.RegistrationConfig) *handler.UserRegistrationHandler {
	userRegistrationRequestRepository := repository.NewUserRegistrationRequestRepository(db)
	txManager := &repository.PostgresTxManager{DB: db}

	userRegistrationService := service.NewUserRegistrationService(
		userRegistrationRequestRepository,
		txManager,
		token.DefaultGenerator{},
		token.SHA256Hasher{},
		uuid.UUIDv7Generator{},
		clock.SystemClock{},
		mail.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom),
		registrationurl.NewStaticBuilder(cfg.FrontendBaseURL),
		cfg,
	)

	return handler.NewUserRegistrationHandler(
		userRegistrationService,
		i18n.NewTranslator(),
	)
}
