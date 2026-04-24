package app

import (
	"database/sql"
	"testing"

	"app-api/internal/config"
)

func TestNewUserRegistrationHandler(t *testing.T) {
	db := &sql.DB{}

	cfg := config.NewRegistrationConfig()

	h := NewUserRegistrationHandler(db, cfg)

	if h == nil {
		t.Fatal("handler is nil")
	}
}
