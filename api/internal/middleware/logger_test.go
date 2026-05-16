package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMaskTokenInPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/verify?token=abc123", "/verify?token=***"},
		{"/verify?token=", "/verify?token=***"},
		{"/verify?other=val&token=abc123", "/verify?other=val&token=***"},
		{"/verify?token=abc123&other=val", "/verify?token=***&other=val"},
		{"/verify", "/verify"},
		{"/verify?other=val", "/verify?other=val"},
		{"/api/v1/user-registration-requests", "/api/v1/user-registration-requests"},
	}
	for _, tt := range tests {
		got := maskTokenInPath(tt.input)
		if got != tt.want {
			t.Errorf("maskTokenInPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLoggerWithTokenMaskTo_MasksToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	router := gin.New()
	router.Use(loggerWithTokenMaskTo(&buf))
	router.POST("/verify", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/verify?token=secrettoken123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := buf.String()
	if strings.Contains(logOutput, "secrettoken123") {
		t.Errorf("token appeared in log output: %s", logOutput)
	}
	if !strings.Contains(logOutput, "token=***") {
		t.Errorf("expected masked token=*** in log output: %s", logOutput)
	}
}

func TestLoggerWithTokenMaskTo_NonTokenQueryUnchanged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	router := gin.New()
	router.Use(loggerWithTokenMaskTo(&buf))
	router.GET("/search", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/search?q=hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !strings.Contains(buf.String(), "q=hello") {
		t.Errorf("non-token query param was unexpectedly removed: %s", buf.String())
	}
}

func TestLoggerWithTokenMask_PublicFunc(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origWriter := gin.DefaultWriter
	var buf bytes.Buffer
	gin.DefaultWriter = &buf
	t.Cleanup(func() { gin.DefaultWriter = origWriter })

	router := gin.New()
	router.Use(LoggerWithTokenMask())
	router.POST("/verify", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/verify?token=mysecret", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if strings.Contains(buf.String(), "mysecret") {
		t.Errorf("token appeared in log: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "token=***") {
		t.Errorf("expected token=*** in log: %s", buf.String())
	}
}

func TestTrustedProxiesNil_IgnoresSpoofedXForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedIP string
	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		t.Fatal(err)
	}
	router.GET("/test", func(c *gin.Context) {
		capturedIP = c.ClientIP()
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "9.9.9.9")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedIP == "9.9.9.9" {
		t.Errorf("X-Forwarded-For spoofing succeeded: ClientIP=%s", capturedIP)
	}
}
