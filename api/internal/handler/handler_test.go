package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func (m *mockService) Verify(ctx context.Context, in service.VerifyUserRegistrationInput) (*service.VerifyUserRegistrationOutput, error) {
	return &service.VerifyUserRegistrationOutput{
		Code: i18n.CodeUserRegistrationTokenVerified,
	}, nil
}

type keyedRateLimitStore struct {
	counts map[string]int64
	err    error
}

func (m *keyedRateLimitStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}

	if m.counts == nil {
		m.counts = map[string]int64{}
	}

	m.counts[key]++

	return m.counts[key], nil
}

func newRouter(h *UserRegistrationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", h.Create)
	return r
}

func newTestHandler(svc *mockService) *UserRegistrationHandler {
	return NewUserRegistrationHandlerWithLimiter(svc, i18n.NewTranslator(), nil)
}

func TestNewUserRegistrationHandler(t *testing.T) {
	h := NewUserRegistrationHandler(&mockService{}, i18n.NewTranslator())

	if h.service == nil {
		t.Fatal("service is nil")
	}

	if h.translator == nil {
		t.Fatal("translator is nil")
	}

	if h.rateLimiter == nil {
		t.Fatal("rate limiter is nil")
	}

	if h.rateLimitConfig.RedisAddr() == "" {
		t.Fatal("redis addr is empty")
	}
}

func TestUserRegistrationHandlerCreateSuccess(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationRequestCreated,
		},
	}

	h := newTestHandler(svc)
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

	h := newTestHandler(svc)
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
	h := newTestHandler(&mockService{})
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
	h := newTestHandler(
		&mockService{
			err: app_error.NewValidation(map[string][]app_error.FieldError{
				"email": {
					{Code: i18n.CodeEmailRequired},
				},
			}),
		},
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
	h := newTestHandler(
		&mockService{
			err: app_error.NewConflict(i18n.CodeUserAlreadyRegistered),
		},
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
	h := newTestHandler(
		&mockService{
			err: app_error.NewInternal(nil),
		},
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

func TestUserRegistrationHandlerCreateRateLimitIP(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationRequestCreated,
		},
	}

	h := NewUserRegistrationHandlerWithLimiter(
		svc,
		i18n.NewTranslator(),
		newRateLimiter(&keyedRateLimitStore{}),
	)
	r := newRouter(h)

	for i := 0; i < 5; i++ {
		body := `{"email":"test` + string(rune('a'+i)) + `@example.com","email_confirmation":"test` + string(rune('a'+i)) + `@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatal(w.Code)
		}
	}

	body := `{"email":"testz@example.com","email_confirmation":"testz@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatal(w.Code)
	}
}

func TestUserRegistrationHandlerCreateRateLimitEmail(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationRequestCreated,
		},
	}

	h := NewUserRegistrationHandlerWithLimiter(
		svc,
		i18n.NewTranslator(),
		newRateLimiter(&keyedRateLimitStore{}),
	)
	r := newRouter(h)

	body := `{"email":"test@example.com","email_confirmation":"test@example.com"}`

	req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatal(w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Fatal(w2.Code)
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

func TestUserRegistrationHandlerCreateEnglishLanguage(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationRequestCreated,
		},
	}

	h := newTestHandler(svc)
	r := newRouter(h)

	body := `{"email":"test@example.com","email_confirmation":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en-US")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatal(w.Code)
	}

	if svc.input.Language != "en" {
		t.Fatalf("unexpected language: %s", svc.input.Language)
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ja", "ja"},
		{"en", "en"},
		{"en-US", "en"},
		{"en-GB", "en"},
		{"fr", "ja"},
		{"zh", "ja"},
		{"", "ja"},
		{"<script>", "ja"},
	}

	for _, tt := range tests {
		got := normalizeLanguage(tt.input)
		if got != tt.expected {
			t.Fatalf("normalizeLanguage(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
