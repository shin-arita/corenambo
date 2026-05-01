package handler

import (
	"net/http"
	"strings"

	"app-api/internal/app_error"
	"app-api/internal/config"
	"app-api/internal/i18n"
	"app-api/internal/logger"
	"app-api/internal/service"

	"github.com/gin-gonic/gin"
)

const codeTooManyRequests = "TOO_MANY_REQUESTS"

type UserRegistrationHandler struct {
	service         service.UserRegistrationService
	translator      i18n.Translator
	rateLimiter     *rateLimiter
	rateLimitConfig config.RateLimitConfig
}

func NewUserRegistrationHandler(
	service service.UserRegistrationService,
	translator i18n.Translator,
) *UserRegistrationHandler {
	rateLimitConfig := config.NewRateLimitConfig()

	return &UserRegistrationHandler{
		service:         service,
		translator:      translator,
		rateLimiter:     newRateLimiter(newRedisRateLimitStore(rateLimitConfig.RedisAddr())),
		rateLimitConfig: rateLimitConfig,
	}
}

func NewUserRegistrationHandlerWithLimiter(
	service service.UserRegistrationService,
	translator i18n.Translator,
	limiter *rateLimiter,
) *UserRegistrationHandler {
	rateLimitConfig := config.NewRateLimitConfig()

	return &UserRegistrationHandler{
		service:         service,
		translator:      translator,
		rateLimiter:     limiter,
		rateLimitConfig: rateLimitConfig,
	}
}

type CreateUserRegistrationRequest struct {
	Email             string `json:"email"`
	EmailConfirmation string `json:"email_confirmation"`
}

// normalizeLanguage はAccept-Languageヘッダーを許可リスト方式で正規化する
// ja/en のみ受け付け、それ以外はjaにフォールバックする
func normalizeLanguage(lang string) string {
	switch {
	case strings.HasPrefix(lang, "en"):
		return "en"
	default:
		return "ja"
	}
}

func (h *UserRegistrationHandler) Create(c *gin.Context) {
	lang := normalizeLanguage(c.GetHeader("Accept-Language"))

	var req CreateUserRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := app_error.NewBadRequest(i18n.CodeBadRequest)
		c.JSON(appErr.StatusCode(), ToResponse(appErr, lang, h.translator))
		return
	}

	if h.rateLimiter != nil {
		if !h.rateLimiter.AllowIP(
			c.Request.Context(),
			c.ClientIP(),
			h.rateLimitConfig.RateLimitIPPerMinute(),
		) {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Code:    codeTooManyRequests,
				Message: "リクエストが多すぎます。しばらく待ってから再度お試しください。",
			})
			return
		}

		if !h.rateLimiter.AllowEmail(
			c.Request.Context(),
			req.Email,
			h.rateLimitConfig.RateLimitEmailPer5Min(),
		) {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Code:    codeTooManyRequests,
				Message: "リクエストが多すぎます。しばらく待ってから再度お試しください。",
			})
			return
		}
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
