package main

import (
	"database/sql"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"app-api/internal/app"
	"app-api/internal/config"
	"app-api/internal/logger"
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

	cfg := config.NewRegistrationConfig()

	userRegistrationHandler := app.NewUserRegistrationHandler(db, cfg)

	router := gin.New()
	router.Use(gin.Logger())
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("msg=server started port=%s", port)

	if err := router.Run(":" + port); err != nil {
		panic(err)
	}
}
