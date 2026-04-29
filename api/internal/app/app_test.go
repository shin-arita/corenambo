package app

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"app-api/internal/config"
	"app-api/internal/handler"
	"app-api/internal/i18n"
	"app-api/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

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

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT .* FROM user_registration_requests").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "token_hash", "expires_at", "verified_at", "created_at", "last_sent_at"}))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO user_registration_requests").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO mail_outboxes").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	txManager := &repository.PostgresTxManager{DB: db}
	svc := NewUserRegistrationService(txManager, config.RegistrationConfig{})
	translator := i18n.NewTranslator()
	h := handler.NewUserRegistrationHandlerWithLimiter(svc, translator, nil)
	w := &UserRegistrationHandler{inner: h}

	body := strings.NewReader(`{"email":"test@example.com","email_confirmation":"test@example.com"}`)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request, _ = http.NewRequest(http.MethodPost, "/", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	w.Create(ctx)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d body: %s", recorder.Code, recorder.Body.String())
	}
}
