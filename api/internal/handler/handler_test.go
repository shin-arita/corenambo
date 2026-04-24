package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"app-api/internal/app_error"
	"app-api/internal/i18n"
	"app-api/internal/service"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	out   *service.CreateUserRegistrationOutput
	err   error
	input service.CreateUserRegistrationInput
}

func (m *mockService) Create(ctx context.Context, in service.CreateUserRegistrationInput) (*service.CreateUserRegistrationOutput, error) {
	m.input = in
	return m.out, m.err
}

func newRouter(h *UserRegistrationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", h.Create)
	return r
}

func TestUserRegistrationHandlerCreateSuccess(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationRequestCreated,
		},
	}

	h := NewUserRegistrationHandler(svc, i18n.NewTranslator())
	r := newRouter(h)

	body := `{"email":"test@example.com","email_confirmation":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "ja")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatal(w.Code)
	}

	if svc.input.Language != "ja" {
		t.Fatalf("unexpected language: %s", svc.input.Language)
	}
}

func TestUserRegistrationHandlerCreateDefaultLanguage(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationRequestCreated,
		},
	}

	h := NewUserRegistrationHandler(svc, i18n.NewTranslator())
	r := newRouter(h)

	body := `{"email":"test@example.com","email_confirmation":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatal(w.Code)
	}

	if svc.input.Language != "ja" {
		t.Fatalf("unexpected language: %s", svc.input.Language)
	}
}

func TestUserRegistrationHandlerCreateBadRequest(t *testing.T) {
	h := NewUserRegistrationHandler(&mockService{}, i18n.NewTranslator())
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
}

func TestUserRegistrationHandlerCreateValidationError(t *testing.T) {
	h := NewUserRegistrationHandler(
		&mockService{
			err: app_error.NewValidation(map[string][]app_error.FieldError{
				"email": {
					{Code: i18n.CodeEmailRequired},
				},
			}),
		},
		i18n.NewTranslator(),
	)

	r := newRouter(h)

	body := `{"email":"","email_confirmation":""}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatal(w.Code)
	}
}

func TestUserRegistrationHandlerCreateConflict(t *testing.T) {
	h := NewUserRegistrationHandler(
		&mockService{
			err: app_error.NewConflict(i18n.CodeUserAlreadyRegistered),
		},
		i18n.NewTranslator(),
	)

	r := newRouter(h)

	body := `{"email":"test@example.com","email_confirmation":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatal(w.Code)
	}
}

func TestUserRegistrationHandlerCreateInternalError(t *testing.T) {
	h := NewUserRegistrationHandler(
		&mockService{
			err: app_error.NewInternal(nil),
		},
		i18n.NewTranslator(),
	)

	r := newRouter(h)

	body := `{"email":"test@example.com","email_confirmation":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatal(w.Code)
	}
}

func TestToResponse(t *testing.T) {
	tr := i18n.NewTranslator()

	err := app_error.NewValidation(map[string][]app_error.FieldError{
		"email": {
			{Code: i18n.CodeEmailRequired},
		},
	})

	res := ToResponse(err, "ja", tr)

	if res.Code == "" {
		t.Fatal("code empty")
	}

	if len(res.Errors) == 0 {
		t.Fatal("errors empty")
	}
}
