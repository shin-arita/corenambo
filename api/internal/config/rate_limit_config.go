package config

import (
	"os"
	"strconv"
)

type RateLimitConfig struct {
	RedisURLStr        string
	IPPerMinute        int
	EmailPerFiveMinute int
}

func NewRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RedisURLStr:        buildRedisURL(),
		IPPerMinute:        getEnvInt("RATE_LIMIT_IP_PER_MINUTE", 5),
		EmailPerFiveMinute: getEnvInt("RATE_LIMIT_EMAIL_PER_5MIN", 1),
	}
}

// buildRedisURL returns the Redis URL from REDIS_URL env var (default: redis://redis:6379/0).
func buildRedisURL() string {
	return getEnvString("REDIS_URL", "redis://redis:6379/0")
}

func (c RateLimitConfig) RedisURL() string {
	return c.RedisURLStr
}

func (c RateLimitConfig) RateLimitIPPerMinute() int {
	return c.IPPerMinute
}

func (c RateLimitConfig) RateLimitEmailPer5Min() int {
	return c.EmailPerFiveMinute
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}

	return parsed
}
