package config

import "testing"

func TestNewRegistrationConfigDefault(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "")
	t.Setenv("FRONTEND_BASE_URL", "")

	c := NewRegistrationConfig()

	if c.TokenExpiresMinutes != 60 {
		t.Fatalf("unexpected expires minutes: %d", c.TokenExpiresMinutes)
	}

	if c.FrontendBaseURL != "http://localhost:5173" {
		t.Fatalf("unexpected frontend base url: %s", c.FrontendBaseURL)
	}
}

func TestNewRegistrationConfigFromEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "30")
	t.Setenv("FRONTEND_BASE_URL", "http://example.com")

	c := NewRegistrationConfig()

	if c.TokenExpiresMinutes != 30 {
		t.Fatalf("unexpected expires minutes: %d", c.TokenExpiresMinutes)
	}

	if c.FrontendBaseURL != "http://example.com" {
		t.Fatalf("unexpected frontend base url: %s", c.FrontendBaseURL)
	}
}

func TestNewRegistrationConfigInvalidEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "abc")
	t.Setenv("FRONTEND_BASE_URL", "")

	c := NewRegistrationConfig()

	if c.TokenExpiresMinutes != 60 {
		t.Fatalf("unexpected expires minutes: %d", c.TokenExpiresMinutes)
	}
}

func TestNewRegistrationConfigZeroEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "0")
	t.Setenv("FRONTEND_BASE_URL", "")

	c := NewRegistrationConfig()

	if c.TokenExpiresMinutes != 60 {
		t.Fatalf("unexpected expires minutes: %d", c.TokenExpiresMinutes)
	}
}

func TestRegistrationConfigRegistrationTokenExpiresMinutes(t *testing.T) {
	c := RegistrationConfig{
		TokenExpiresMinutes: 45,
	}

	if c.RegistrationTokenExpiresMinutes() != 45 {
		t.Fatalf("unexpected expires minutes: %d", c.RegistrationTokenExpiresMinutes())
	}
}
