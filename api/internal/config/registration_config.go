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
	SMTPUser              string
	SMTPPass              string
	SMTPUseTLS            bool
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

	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	smtpUseTLS := false
	if v := os.Getenv("SMTP_USE_TLS"); v == "true" {
		smtpUseTLS = true
	}

	return RegistrationConfig{
		TokenExpiresMinutes:   expiresMinutes,
		ResendIntervalMinutes: resendIntervalMinutes,
		FrontendBaseURL:       frontendBaseURL,
		SMTPHost:              smtpHost,
		SMTPPort:              smtpPort,
		SMTPFrom:              smtpFrom,
		SMTPUser:              smtpUser,
		SMTPPass:              smtpPass,
		SMTPUseTLS:            smtpUseTLS,
	}
}

func (c RegistrationConfig) RegistrationTokenExpiresMinutes() int {
	return c.TokenExpiresMinutes
}

func (c RegistrationConfig) RegistrationResendIntervalMinutes() int {
	return c.ResendIntervalMinutes
}
