package config

import (
	"os"
	"strconv"
)

type WorkerConfig struct {
	DatabaseURL                string
	StuckTimeoutMinutes        int
	RegistrationExpiresMinutes int
	MaxRetryCount              int
	SMTPHost                   string
	SMTPPort                   string
	SMTPFrom                   string
	SMTPUser                   string
	SMTPPass                   string
	SMTPUseTLS                 bool
}

func NewWorkerConfig() WorkerConfig {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://app_user:password@db:5432/app_db?sslmode=disable"
	}

	stuckTimeoutMinutes := 15
	if v := os.Getenv("WORKER_STUCK_TIMEOUT_MINUTES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			stuckTimeoutMinutes = parsed
		}
	}

	registrationExpiresMinutes := 60
	if v := os.Getenv("REGISTRATION_TOKEN_EXPIRES_MINUTES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			registrationExpiresMinutes = parsed
		}
	}

	maxRetryCount := 3
	if v := os.Getenv("WORKER_MAX_RETRY_COUNT"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			maxRetryCount = parsed
		}
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

	return WorkerConfig{
		DatabaseURL:                databaseURL,
		StuckTimeoutMinutes:        stuckTimeoutMinutes,
		RegistrationExpiresMinutes: registrationExpiresMinutes,
		MaxRetryCount:              maxRetryCount,
		SMTPHost:                   smtpHost,
		SMTPPort:                   smtpPort,
		SMTPFrom:                   smtpFrom,
		SMTPUser:                   smtpUser,
		SMTPPass:                   smtpPass,
		SMTPUseTLS:                 smtpUseTLS,
	}
}

func (c WorkerConfig) WorkerStuckTimeoutMinutes() int {
	return c.StuckTimeoutMinutes
}

func (c WorkerConfig) WorkerRegistrationExpiresMinutes() int {
	return c.RegistrationExpiresMinutes
}

func (c WorkerConfig) WorkerMaxRetryCount() int {
	return c.MaxRetryCount
}
