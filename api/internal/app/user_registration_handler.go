package app

import (
	"database/sql"

	"app-api/internal/config"
	"app-api/internal/handler"
	"app-api/internal/i18n"
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
	inner := handler.NewUserRegistrationHandler(svc, translator)
	return &UserRegistrationHandler{inner: inner}
}

func (h *UserRegistrationHandler) Create(c *gin.Context) {
	h.inner.Create(c)
}
