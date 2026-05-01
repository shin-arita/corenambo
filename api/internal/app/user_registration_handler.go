package app

import (
	"database/sql"

	"app-api/internal/config"
	"app-api/internal/handler"
	"app-api/internal/i18n"
	"app-api/internal/mail"
	"app-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type UserRegistrationHandler struct {
	inner *handler.UserRegistrationHandler
}

func NewUserRegistrationHandler(
	db *sql.DB,
	cfg config.RegistrationConfig,
) *UserRegistrationHandler {
	txManager := &repository.PostgresTxManager{DB: db}
	svc := NewUserRegistrationService(txManager, cfg)
	translator := i18n.NewTranslator()

	// SMTP設定をcfgから取得
	_ = mail.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPUseTLS)

	inner := handler.NewUserRegistrationHandler(svc, translator)
	return &UserRegistrationHandler{inner: inner}
}

func (h *UserRegistrationHandler) Create(c *gin.Context) {
	h.inner.Create(c)
}
