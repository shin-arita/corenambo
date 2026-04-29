package main

import (
	"database/sql"
	"net/http"
	"os"
	"runtime/debug"

	"app-api/internal/app"
	"app-api/internal/config"
	"app-api/internal/logger"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close db: %v", err)
		}
	}()

	serverCfg := config.NewServerConfig()
	if serverCfg.GetCORSAllowOrigin() == "" {
		panic("CORS_ALLOW_ORIGIN is required")
	}

	cfg := config.NewRegistrationConfig()
	userRegistrationHandler := app.NewUserRegistrationHandler(db, cfg)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(corsMiddleware(serverCfg.GetCORSAllowOrigin()))
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error(
			"method=%s path=%s panic=%v stack=%s",
			c.Request.Method,
			c.FullPath(),
			recovered,
			string(debug.Stack()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_SERVER_ERROR",
			"message": "internal server error",
		})
	}))

	router.POST("/api/v1/user-registration-requests", userRegistrationHandler.Create)

	logger.Info("msg=server started port=%s", serverCfg.GetPort())
	if err := router.Run(":" + serverCfg.GetPort()); err != nil {
		panic(err)
	}
}

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept-Language")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
