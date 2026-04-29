package config

import "os"

type ServerConfig struct {
	CORSAllowOrigin string
	Port            string
}

func NewServerConfig() ServerConfig {
	corsAllowOrigin := os.Getenv("CORS_ALLOW_ORIGIN")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return ServerConfig{
		CORSAllowOrigin: corsAllowOrigin,
		Port:            port,
	}
}

func (c ServerConfig) GetCORSAllowOrigin() string {
	return c.CORSAllowOrigin
}

func (c ServerConfig) GetPort() string {
	return c.Port
}
