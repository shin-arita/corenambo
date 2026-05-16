package app

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app-api/internal/config"
	"app-api/internal/handler"
	"app-api/internal/i18n"
	"app-api/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestNewUserRegistrationService(t *testing.T) {
	db := &sql.DB{}
	txManager := repository.NewPostgresTxManager(db)

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

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM user_registration_requests").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "token_hash", "expires_at", "verified_at", "last_sent_at", "created_at"}))
	mock.ExpectExec("INSERT INTO user_registration_requests").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO mail_outboxes").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	txManager := repository.NewPostgresTxManager(db)
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

func TestUserRegistrationHandlerCheckToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	futureTime := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	pastTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT .* FROM user_registration_requests").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "token_hash", "expires_at", "verified_at", "last_sent_at", "created_at"}).
			AddRow("reg-id", "test@example.com", "hash", futureTime, nil, nil, pastTime))
	mock.ExpectQuery("SELECT .* FROM user_emails").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "is_primary", "verified_at", "created_at", "updated_at"}))

	txManager := repository.NewPostgresTxManager(db)
	svc := NewUserRegistrationService(txManager, config.RegistrationConfig{})
	translator := i18n.NewTranslator()
	h := handler.NewUserRegistrationHandlerWithLimiter(svc, translator, nil)
	w := &UserRegistrationHandler{inner: h}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request, _ = http.NewRequest(http.MethodGet, "/?token=raw-token", nil)

	w.CheckToken(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body: %s", recorder.Code, recorder.Body.String())
	}
}

func TestUserRegistrationHandlerVerify(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	futureTime := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	pastTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM user_registration_requests").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "token_hash", "expires_at", "verified_at", "last_sent_at", "created_at"}).
			AddRow("reg-id", "test@example.com", "hash", futureTime, nil, nil, pastTime))
	mock.ExpectQuery("SELECT .* FROM user_emails").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "is_primary", "verified_at", "created_at", "updated_at"}))
	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO user_emails").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO user_credentials").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE user_registration_requests").
		WithArgs("reg-id", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	txManager := repository.NewPostgresTxManager(db)
	svc := NewUserRegistrationService(txManager, config.RegistrationConfig{})
	translator := i18n.NewTranslator()
	h := handler.NewUserRegistrationHandlerWithLimiter(svc, translator, nil)
	w := &UserRegistrationHandler{inner: h}

	body := strings.NewReader(`{"display_name":"testuser","password":"password123","password_confirmation":"password123","agreed_to_terms":true}`)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request, _ = http.NewRequest(http.MethodPost, "/?token=raw-token", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	w.Verify(ctx)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d body: %s", recorder.Code, recorder.Body.String())
	}
}
