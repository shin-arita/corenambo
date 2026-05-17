package handler

import (
	"bytes"
	"context"
	"encoding/json"
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
	out             *service.CreateUserRegistrationOutput
	err             error
	input           service.CreateUserRegistrationInput
	verifyOut       *service.VerifyUserRegistrationOutput
	verifyErr       error
	verifyInput     service.VerifyUserRegistrationInput
	checkTokenOut   *service.CheckTokenOutput
	checkTokenErr   error
	checkTokenInput service.CheckTokenInput
}

func (m *mockService) Create(ctx context.Context, in service.CreateUserRegistrationInput) (*service.CreateUserRegistrationOutput, error) {
	m.input = in
	return m.out, m.err
}

func (m *mockService) Verify(ctx context.Context, in service.VerifyUserRegistrationInput) (*service.VerifyUserRegistrationOutput, error) {
	m.verifyInput = in
	if m.verifyOut != nil || m.verifyErr != nil {
		return m.verifyOut, m.verifyErr
	}
	return &service.VerifyUserRegistrationOutput{
		Code: i18n.CodeUserRegistrationVerified,
	}, nil
}

func (m *mockService) CheckToken(ctx context.Context, in service.CheckTokenInput) (*service.CheckTokenOutput, error) {
	m.checkTokenInput = in
	if m.checkTokenOut != nil || m.checkTokenErr != nil {
		return m.checkTokenOut, m.checkTokenErr
	}
	return &service.CheckTokenOutput{
		Code: i18n.CodeRegistrationTokenValid,
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

func newVerifyRouter(h *UserRegistrationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/verify", h.Verify)
	return r
}

func newVerifyRouterWithSizeLimit(h *UserRegistrationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})
	r.POST("/verify", h.Verify)
	return r
}

func newRouterWithSizeLimit(h *UserRegistrationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})
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

	if h.rateLimitConfig.RedisURL() == "" {
		t.Fatal("redis url is empty")
	}
}

func TestUserRegistrationHandlerCreateSuccess(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: 60,
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
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: 60,
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
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: 60,
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

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse rate limit response body: %v", err)
	}
	if res["code"] != i18n.CodeTooManyRequests {
		t.Fatalf("expected code %q, got %v", i18n.CodeTooManyRequests, res["code"])
	}
}

func TestUserRegistrationHandlerCreateRateLimitEmail(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: 60,
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

	var res map[string]any
	if err := json.Unmarshal(w2.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse rate limit response body: %v", err)
	}
	if res["code"] != i18n.CodeTooManyRequests {
		t.Fatalf("expected code %q, got %v", i18n.CodeTooManyRequests, res["code"])
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
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: 60,
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

func TestUserRegistrationHandlerCreateSuccessResponseBody(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: 60,
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

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}

	if res["expires_minutes"] != float64(60) {
		t.Fatalf("expected expires_minutes=60, got %v", res["expires_minutes"])
	}

	if res["code"] == "" {
		t.Fatal("code is empty")
	}

	if res["message"] == "" {
		t.Fatal("message is empty")
	}
}

func TestUserRegistrationHandlerCreateBodyUnderSizeLimit(t *testing.T) {
	svc := &mockService{
		out: &service.CreateUserRegistrationOutput{
			Code:           i18n.CodeUserRegistrationRequestCreated,
			ExpiresMinutes: 60,
		},
	}

	h := newTestHandler(svc)
	r := newRouterWithSizeLimit(h)

	body := `{"email":"test@example.com","email_confirmation":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerCreateBodyExceedsSizeLimit(t *testing.T) {
	h := newTestHandler(&mockService{})
	r := newRouterWithSizeLimit(h)

	body := make([]byte, 1<<20+1)
	for i := range body {
		body[i] = 'a'
	}
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

const validVerifyBody = `{"display_name":"testuser","password":"password123","password_confirmation":"password123","agreed_to_terms":true}`
const validVerifyURL = "/verify?token=valid-token"

func TestUserRegistrationHandlerVerifySuccess(t *testing.T) {
	svc := &mockService{
		verifyOut: &service.VerifyUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationVerified,
		},
	}

	h := newTestHandler(svc)
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body: %s", w.Code, w.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}

	if res["code"] == "" {
		t.Fatal("code is empty")
	}

	if res["message"] == "" {
		t.Fatal("message is empty")
	}

	if svc.verifyInput.Token != "valid-token" {
		t.Fatalf("expected query token passed to service, got %q", svc.verifyInput.Token)
	}
}

func TestUserRegistrationHandlerVerifyMissingQueryToken(t *testing.T) {
	h := newTestHandler(&mockService{})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/verify", bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["code"] != i18n.CodeInvalidRegistrationToken {
		t.Fatalf("expected code %q, got %v", i18n.CodeInvalidRegistrationToken, res["code"])
	}
}

func TestUserRegistrationHandlerVerifyBadRequest(t *testing.T) {
	h := newTestHandler(&mockService{})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyTokenInvalid(t *testing.T) {
	h := newTestHandler(&mockService{
		verifyErr: app_error.NewBadRequest(i18n.CodeInvalidRegistrationToken),
	})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyTokenExpired(t *testing.T) {
	h := newTestHandler(&mockService{
		verifyErr: app_error.NewBadRequest(i18n.CodeExpiredRegistrationToken),
	})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyAlreadyVerified(t *testing.T) {
	h := newTestHandler(&mockService{
		verifyErr: app_error.NewConflict(i18n.CodeUsedRegistrationToken),
	})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyUserAlreadyRegistered(t *testing.T) {
	h := newTestHandler(&mockService{
		verifyErr: app_error.NewConflict(i18n.CodeUserAlreadyRegistered),
	})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyValidationError(t *testing.T) {
	h := newTestHandler(&mockService{
		verifyErr: app_error.NewValidation(map[string][]app_error.FieldError{
			"display_name": {{Code: i18n.CodeDisplayNameRequired}},
		}),
	})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyInternalError(t *testing.T) {
	h := newTestHandler(&mockService{
		verifyErr: app_error.NewInternal(nil),
	})
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyRateLimitIP(t *testing.T) {
	svc := &mockService{
		verifyOut: &service.VerifyUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationVerified,
		},
	}

	h := NewUserRegistrationHandlerWithLimiter(
		svc,
		i18n.NewTranslator(),
		newRateLimiter(&keyedRateLimitStore{}),
	)
	r := newVerifyRouter(h)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("request %d: expected 201, got %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse rate limit response body: %v", err)
	}
	if res["code"] != i18n.CodeTooManyRequests {
		t.Fatalf("expected code %q, got %v", i18n.CodeTooManyRequests, res["code"])
	}
}

func TestUserRegistrationHandlerVerifyBodyExceedsSizeLimit(t *testing.T) {
	h := newTestHandler(&mockService{})
	r := newVerifyRouterWithSizeLimit(h)

	body := make([]byte, 1<<20+1)
	for i := range body {
		body[i] = 'a'
	}
	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerVerifyDefaultLanguage(t *testing.T) {
	svc := &mockService{
		verifyOut: &service.VerifyUserRegistrationOutput{
			Code: i18n.CodeUserRegistrationVerified,
		},
	}

	h := newTestHandler(svc)
	r := newVerifyRouter(h)

	req := httptest.NewRequest(http.MethodPost, validVerifyURL, bytes.NewBufferString(validVerifyBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func newCheckTokenRouter(h *UserRegistrationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/verify", h.CheckToken)
	return r
}

const validCheckTokenURL = "/verify?token=valid-token"

func TestUserRegistrationHandlerCheckTokenSuccess(t *testing.T) {
	svc := &mockService{
		checkTokenOut: &service.CheckTokenOutput{
			Code: i18n.CodeRegistrationTokenValid,
		},
	}

	h := newTestHandler(svc)
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)
	req.Header.Set("Accept-Language", "ja")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}

	if res["code"] != i18n.CodeRegistrationTokenValid {
		t.Fatalf("expected code %q, got %v", i18n.CodeRegistrationTokenValid, res["code"])
	}

	if res["message"] == "" {
		t.Fatal("message is empty")
	}

	if svc.checkTokenInput.Token != "valid-token" {
		t.Fatalf("expected token passed to service, got %q", svc.checkTokenInput.Token)
	}
}

func TestUserRegistrationHandlerCheckTokenMissingToken(t *testing.T) {
	h := newTestHandler(&mockService{})
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/verify", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["code"] != i18n.CodeInvalidRegistrationToken {
		t.Fatalf("expected code %q, got %v", i18n.CodeInvalidRegistrationToken, res["code"])
	}
}

func TestUserRegistrationHandlerCheckTokenInvalid(t *testing.T) {
	h := newTestHandler(&mockService{
		checkTokenErr: app_error.NewBadRequest(i18n.CodeInvalidRegistrationToken),
	})
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerCheckTokenExpired(t *testing.T) {
	h := newTestHandler(&mockService{
		checkTokenErr: app_error.NewBadRequest(i18n.CodeExpiredRegistrationToken),
	})
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerCheckTokenUsed(t *testing.T) {
	h := newTestHandler(&mockService{
		checkTokenErr: app_error.NewConflict(i18n.CodeUsedRegistrationToken),
	})
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerCheckTokenUserAlreadyRegistered(t *testing.T) {
	h := newTestHandler(&mockService{
		checkTokenErr: app_error.NewConflict(i18n.CodeUserAlreadyRegistered),
	})
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerCheckTokenInternalError(t *testing.T) {
	h := newTestHandler(&mockService{
		checkTokenErr: app_error.NewInternal(nil),
	})
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserRegistrationHandlerCheckTokenRateLimitIP(t *testing.T) {
	svc := &mockService{
		checkTokenOut: &service.CheckTokenOutput{
			Code: i18n.CodeRegistrationTokenValid,
		},
	}

	h := NewUserRegistrationHandlerWithLimiter(
		svc,
		i18n.NewTranslator(),
		newRateLimiter(&keyedRateLimitStore{}),
	)
	r := newCheckTokenRouter(h)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse rate limit response body: %v", err)
	}
	if res["code"] != i18n.CodeTooManyRequests {
		t.Fatalf("expected code %q, got %v", i18n.CodeTooManyRequests, res["code"])
	}
}

func TestUserRegistrationHandlerCheckTokenDefaultLanguage(t *testing.T) {
	h := newTestHandler(&mockService{})
	r := newCheckTokenRouter(h)

	req := httptest.NewRequest(http.MethodGet, validCheckTokenURL, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["message"] == "" {
		t.Fatal("message is empty")
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
