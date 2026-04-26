package app

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"app-api/internal/config"
	"app-api/internal/repository"

	"github.com/gin-gonic/gin"
)

func TestDummyBuilderBuild(t *testing.T) {
	builder := dummyBuilder{}

	got := builder.Build("token")

	if got != "http://example.com/token" {
		t.Fatalf("unexpected url: %s", got)
	}
}

func TestNewUserRegistrationService(t *testing.T) {
	db := &sql.DB{}
	txManager := &repository.PostgresTxManager{DB: db}

	svc := NewUserRegistrationService(txManager, config.RegistrationConfig{})

	if svc == nil {
		t.Fatal("service is nil")
	}
}

func TestNewUserRegistrationHandler(t *testing.T) {
	handler := NewUserRegistrationHandler(&sql.DB{}, config.RegistrationConfig{})

	if handler == nil {
		t.Fatal("handler is nil")
	}
}

func TestUserRegistrationHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewUserRegistrationHandler(&sql.DB{}, config.RegistrationConfig{})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	handler.Create(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}
