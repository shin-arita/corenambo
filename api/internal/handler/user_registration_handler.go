package handler

import (
	"net/http"

	"app-api/internal/app_error"
	"app-api/internal/i18n"
	"app-api/internal/logger"
	"app-api/internal/service"

	"github.com/gin-gonic/gin"
)

type UserRegistrationHandler struct {
	service    service.UserRegistrationService
	translator i18n.Translator
}

func NewUserRegistrationHandler(
	service service.UserRegistrationService,
	translator i18n.Translator,
) *UserRegistrationHandler {
	return &UserRegistrationHandler{
		service:    service,
		translator: translator,
	}
}

type CreateUserRegistrationRequest struct {
	Email             string `json:"email"`
	EmailConfirmation string `json:"email_confirmation"`
}

func (h *UserRegistrationHandler) Create(c *gin.Context) {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		lang = "ja"
	}

	var req CreateUserRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := app_error.NewBadRequest(i18n.CodeBadRequest)
		c.JSON(appErr.StatusCode(), ToResponse(appErr, lang, h.translator))
		return
	}

	output, err := h.service.Create(
		c.Request.Context(),
		service.CreateUserRegistrationInput{
			Email:             req.Email,
			EmailConfirmation: req.EmailConfirmation,
			Language:          lang,
		},
	)
	if err != nil {
		appErr := app_error.Normalize(err)

		if appErr.StatusCode() >= 500 {
			logger.Error(
				"method=%s path=%s code=%s cause=%v",
				c.Request.Method,
				c.FullPath(),
				appErr.Code,
				appErr.Cause,
			)
		} else {
			logger.Warn(
				"method=%s path=%s code=%s",
				c.Request.Method,
				c.FullPath(),
				appErr.Code,
			)
		}

		c.JSON(appErr.StatusCode(), ToResponse(appErr, lang, h.translator))
		return
	}

	logger.Info("method=%s path=%s code=%s", c.Request.Method, c.FullPath(), output.Code)

	c.JSON(http.StatusCreated, SuccessResponse{
		Code:    output.Code,
		Message: h.translator.Translate(lang, output.Code),
	})
}
