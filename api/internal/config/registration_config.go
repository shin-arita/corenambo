package config

import (
	"os"
	"strconv"
)

type RegistrationConfig struct {
	TokenExpiresMinutes int
	FrontendBaseURL     string
}

func NewRegistrationConfig() RegistrationConfig {
	expiresMinutes := 60
	if v := os.Getenv("REGISTRATION_TOKEN_EXPIRES_MINUTES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			expiresMinutes = parsed
		}
	}

	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendBaseURL == "" {
		frontendBaseURL = "http://localhost:5173"
	}

	return RegistrationConfig{
		TokenExpiresMinutes: expiresMinutes,
		FrontendBaseURL:     frontendBaseURL,
	}
}

func (c RegistrationConfig) RegistrationTokenExpiresMinutes() int {
	return c.TokenExpiresMinutes
}
