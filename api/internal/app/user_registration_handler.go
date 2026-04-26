package app

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"app-api/internal/config"
)

type UserRegistrationHandler struct{}

func NewUserRegistrationHandler(
	db *sql.DB,
	cfg config.RegistrationConfig,
) *UserRegistrationHandler {
	return &UserRegistrationHandler{}
}

func (h *UserRegistrationHandler) Create(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": "OK",
	})
}
