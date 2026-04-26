package config

import (
	"os"
	"strconv"
)

type RateLimitConfig struct {
	RedisAddress       string
	IPPerMinute        int
	EmailPerFiveMinute int
}

func NewRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RedisAddress:       getEnvString("REDIS_ADDR", "redis:6379"),
		IPPerMinute:        getEnvInt("RATE_LIMIT_IP_PER_MINUTE", 5),
		EmailPerFiveMinute: getEnvInt("RATE_LIMIT_EMAIL_PER_5MIN", 1),
	}
}

func (c RateLimitConfig) RedisAddr() string {
	return c.RedisAddress
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
