package config

import "testing"

func TestNewRegistrationConfigDefault(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "")
	t.Setenv("FRONTEND_BASE_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")

	c := NewRegistrationConfig()

	if c.TokenExpiresMinutes != 60 {
		t.Fatalf("unexpected expires minutes: %d", c.TokenExpiresMinutes)
	}

	if c.FrontendBaseURL != "http://localhost:5173" {
		t.Fatalf("unexpected frontend base url: %s", c.FrontendBaseURL)
	}

	if c.SMTPHost != "mail" {
		t.Fatalf("unexpected smtp host: %s", c.SMTPHost)
	}

	if c.SMTPPort != "1025" {
		t.Fatalf("unexpected smtp port: %s", c.SMTPPort)
	}

	if c.SMTPFrom != "noreply@example.com" {
		t.Fatalf("unexpected smtp from: %s", c.SMTPFrom)
	}
}

func TestNewRegistrationConfigFromEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "30")
	t.Setenv("FRONTEND_BASE_URL", "http://example.com")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_FROM", "system@example.com")

	c := NewRegistrationConfig()

	if c.TokenExpiresMinutes != 30 {
		t.Fatalf("unexpected expires minutes: %d", c.TokenExpiresMinutes)
	}

	if c.FrontendBaseURL != "http://example.com" {
		t.Fatalf("unexpected frontend base url: %s", c.FrontendBaseURL)
	}

	if c.SMTPHost != "smtp.example.com" {
		t.Fatalf("unexpected smtp host: %s", c.SMTPHost)
	}

	if c.SMTPPort != "2525" {
		t.Fatalf("unexpected smtp port: %s", c.SMTPPort)
	}

	if c.SMTPFrom != "system@example.com" {
		t.Fatalf("unexpected smtp from: %s", c.SMTPFrom)
	}
}

func TestNewRegistrationConfigInvalidEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "abc")
	t.Setenv("FRONTEND_BASE_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")

	c := NewRegistrationConfig()

	if c.TokenExpiresMinutes != 60 {
		t.Fatalf("unexpected expires minutes: %d", c.TokenExpiresMinutes)
	}
}

func TestNewRegistrationConfigZeroEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "0")
	t.Setenv("FRONTEND_BASE_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")

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

func TestNewRateLimitConfigDefault(t *testing.T) {
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("RATE_LIMIT_IP_PER_MINUTE", "")
	t.Setenv("RATE_LIMIT_EMAIL_PER_5MIN", "")

	c := NewRateLimitConfig()

	if c.RedisAddr() != "redis:6379" {
		t.Fatalf("unexpected redis addr: %s", c.RedisAddr())
	}

	if c.RateLimitIPPerMinute() != 5 {
		t.Fatalf("unexpected ip limit: %d", c.RateLimitIPPerMinute())
	}

	if c.RateLimitEmailPer5Min() != 1 {
		t.Fatalf("unexpected email limit: %d", c.RateLimitEmailPer5Min())
	}
}

func TestNewRateLimitConfigFromEnv(t *testing.T) {
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("RATE_LIMIT_IP_PER_MINUTE", "10")
	t.Setenv("RATE_LIMIT_EMAIL_PER_5MIN", "2")

	c := NewRateLimitConfig()

	if c.RedisAddr() != "localhost:6379" {
		t.Fatalf("unexpected redis addr: %s", c.RedisAddr())
	}

	if c.RateLimitIPPerMinute() != 10 {
		t.Fatalf("unexpected ip limit: %d", c.RateLimitIPPerMinute())
	}

	if c.RateLimitEmailPer5Min() != 2 {
		t.Fatalf("unexpected email limit: %d", c.RateLimitEmailPer5Min())
	}
}

func TestNewRateLimitConfigInvalidEnv(t *testing.T) {
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("RATE_LIMIT_IP_PER_MINUTE", "abc")
	t.Setenv("RATE_LIMIT_EMAIL_PER_5MIN", "0")

	c := NewRateLimitConfig()

	if c.RedisAddr() != "redis:6379" {
		t.Fatalf("unexpected redis addr: %s", c.RedisAddr())
	}

	if c.RateLimitIPPerMinute() != 5 {
		t.Fatalf("unexpected ip limit: %d", c.RateLimitIPPerMinute())
	}

	if c.RateLimitEmailPer5Min() != 1 {
		t.Fatalf("unexpected email limit: %d", c.RateLimitEmailPer5Min())
	}
}

func TestRegistrationConfigRegistrationResendIntervalMinutes(t *testing.T) {
	c := RegistrationConfig{
		ResendIntervalMinutes: 15,
	}

	if c.RegistrationResendIntervalMinutes() != 15 {
		t.Fatalf("unexpected resend interval minutes: %d", c.RegistrationResendIntervalMinutes())
	}
}

func TestNewRegistrationConfigInvalidResendIntervalEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "")
	t.Setenv("REGISTRATION_RESEND_INTERVAL_MINUTES", "abc")
	t.Setenv("FRONTEND_BASE_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")

	c := NewRegistrationConfig()

	if c.ResendIntervalMinutes != 5 {
		t.Fatalf("unexpected resend interval minutes: %d", c.ResendIntervalMinutes)
	}
}

func TestNewRegistrationConfigZeroResendIntervalEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "")
	t.Setenv("REGISTRATION_RESEND_INTERVAL_MINUTES", "0")
	t.Setenv("FRONTEND_BASE_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")

	c := NewRegistrationConfig()

	if c.ResendIntervalMinutes != 5 {
		t.Fatalf("unexpected resend interval minutes: %d", c.ResendIntervalMinutes)
	}
}

func TestNewRegistrationConfigResendIntervalFromEnv(t *testing.T) {
	t.Setenv("REGISTRATION_TOKEN_EXPIRES_MINUTES", "")
	t.Setenv("REGISTRATION_RESEND_INTERVAL_MINUTES", "15")
	t.Setenv("FRONTEND_BASE_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM", "")

	c := NewRegistrationConfig()

	if c.ResendIntervalMinutes != 15 {
		t.Fatalf("unexpected resend interval minutes: %d", c.ResendIntervalMinutes)
	}
}
