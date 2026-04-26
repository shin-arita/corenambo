package config

import (
	"os"
	"strconv"
)

type RegistrationConfig struct {
	TokenExpiresMinutes   int
	ResendIntervalMinutes int
	FrontendBaseURL       string
	SMTPHost              string
	SMTPPort              string
	SMTPFrom              string
}

func NewRegistrationConfig() RegistrationConfig {
	expiresMinutes := 60
	if v := os.Getenv("REGISTRATION_TOKEN_EXPIRES_MINUTES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			expiresMinutes = parsed
		}
	}

	resendIntervalMinutes := 5
	if v := os.Getenv("REGISTRATION_RESEND_INTERVAL_MINUTES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			resendIntervalMinutes = parsed
		}
	}

	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendBaseURL == "" {
		frontendBaseURL = "http://localhost:5173"
	}

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "mail"
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "1025"
	}

	smtpFrom := os.Getenv("SMTP_FROM")
	if smtpFrom == "" {
		smtpFrom = "noreply@example.com"
	}

	return RegistrationConfig{
		TokenExpiresMinutes:   expiresMinutes,
		ResendIntervalMinutes: resendIntervalMinutes,
		FrontendBaseURL:       frontendBaseURL,
		SMTPHost:              smtpHost,
		SMTPPort:              smtpPort,
		SMTPFrom:              smtpFrom,
	}
}

func (c RegistrationConfig) RegistrationTokenExpiresMinutes() int {
	return c.TokenExpiresMinutes
}

func (c RegistrationConfig) RegistrationResendIntervalMinutes() int {
	return c.ResendIntervalMinutes
}
