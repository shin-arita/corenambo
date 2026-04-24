package app_error

import (
	"errors"
	"net/http"
	"testing"

	"app-api/internal/i18n"
)

func TestAppErrorError(t *testing.T) {
	err := NewBadRequest("TEST_CODE")

	if err.Error() != "TEST_CODE" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestAppErrorStatusCode(t *testing.T) {
	err := NewBadRequest("TEST_CODE")

	if err.StatusCode() != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", err.StatusCode())
	}
}

func TestNewBadRequest(t *testing.T) {
	err := NewBadRequest("TEST_CODE")

	if err.Code != "TEST_CODE" {
		t.Fatalf("unexpected code: %s", err.Code)
	}

	if err.Status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", err.Status)
	}
}

func TestNewConflict(t *testing.T) {
	err := NewConflict("CONFLICT")

	if err.Code != "CONFLICT" {
		t.Fatalf("unexpected code: %s", err.Code)
	}

	if err.Status != http.StatusConflict {
		t.Fatalf("unexpected status: %d", err.Status)
	}
}

func TestNewValidation(t *testing.T) {
	fieldErrors := map[string][]FieldError{
		"email": {
			{Code: "REQUIRED"},
		},
	}

	err := NewValidation(fieldErrors)

	if err.Code != i18n.CodeValidationError {
		t.Fatalf("unexpected code: %s", err.Code)
	}

	if err.Status != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", err.Status)
	}

	if err.FieldErrors["email"][0].Code != "REQUIRED" {
		t.Fatal("field error not set")
	}
}

func TestNewInternal(t *testing.T) {
	cause := errors.New("boom")

	err := NewInternal(cause)

	if err.Code != i18n.CodeInternalServerError {
		t.Fatalf("unexpected code: %s", err.Code)
	}

	if err.Status != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", err.Status)
	}

	if !errors.Is(err.Cause, cause) {
		t.Fatal("cause mismatch")
	}
}

func TestWithStackNil(t *testing.T) {
	err := WithStack(nil)

	if err != nil {
		t.Fatal("expected nil")
	}
}

func TestWithStack(t *testing.T) {
	cause := errors.New("boom")

	err := WithStack(cause)

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, cause) {
		t.Fatal("cause mismatch")
	}

	if err.Error() == "boom" {
		t.Fatal("stack is not included")
	}
}

func TestWithStackAlreadyStackError(t *testing.T) {
	cause := errors.New("boom")

	err1 := WithStack(cause)
	err2 := WithStack(err1)

	if err1 != err2 {
		t.Fatal("expected same error")
	}
}

func TestStackErrorUnwrap(t *testing.T) {
	cause := errors.New("boom")

	err := WithStack(cause)

	if !errors.Is(err, cause) {
		t.Fatal("unwrap failed")
	}
}

func TestWrapInternal(t *testing.T) {
	cause := errors.New("boom")

	err := WrapInternal(cause)

	if err.Code != i18n.CodeInternalServerError {
		t.Fatalf("unexpected code: %s", err.Code)
	}

	if err.Status != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", err.Status)
	}

	if err.Cause == nil {
		t.Fatal("cause should not be nil")
	}
}

func TestNormalizeNil(t *testing.T) {
	err := Normalize(nil)

	if err != nil {
		t.Fatal("expected nil")
	}
}

func TestNormalize(t *testing.T) {
	cause := errors.New("boom")

	appErr := Normalize(cause)

	if appErr.Code != i18n.CodeInternalServerError {
		t.Fatalf("unexpected code: %s", appErr.Code)
	}

	if appErr.Status != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", appErr.Status)
	}

	if appErr.Cause == nil {
		t.Fatal("cause should not be nil")
	}
}

func TestNormalizeAppError(t *testing.T) {
	err := NewBadRequest("BAD_REQUEST")

	appErr := Normalize(err)

	if appErr != err {
		t.Fatal("expected same AppError")
	}
}

func TestNormalizeInternalAppErrorWithCause(t *testing.T) {
	cause := errors.New("boom")
	err := NewInternal(cause)

	appErr := Normalize(err)

	if appErr != err {
		t.Fatal("expected same AppError")
	}

	if appErr.Cause == nil {
		t.Fatal("cause should not be nil")
	}
}
