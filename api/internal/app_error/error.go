package app_error

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"app-api/internal/i18n"
)

type FieldError struct {
	Code string
}

type AppError struct {
	Code        string
	Status      int
	FieldErrors map[string][]FieldError
	Cause       error
}

func (e *AppError) Error() string {
	return e.Code
}

func (e *AppError) StatusCode() int {
	return e.Status
}

func NewBadRequest(code string) *AppError {
	return &AppError{
		Code:   code,
		Status: http.StatusBadRequest,
	}
}

func NewConflict(code string) *AppError {
	return &AppError{
		Code:   code,
		Status: http.StatusConflict,
	}
}

func NewValidation(fieldErrors map[string][]FieldError) *AppError {
	return &AppError{
		Code:        i18n.CodeValidationError,
		Status:      http.StatusUnprocessableEntity,
		FieldErrors: fieldErrors,
	}
}

func NewInternal(cause error) *AppError {
	return &AppError{
		Code:   i18n.CodeInternalServerError,
		Status: http.StatusInternalServerError,
		Cause:  cause,
	}
}

type stackError struct {
	err   error
	stack []byte
}

func (e *stackError) Error() string {
	return fmt.Sprintf("%v\n%s", e.err, e.stack)
}

func (e *stackError) Unwrap() error {
	return e.err
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}

	var se *stackError
	if errors.As(err, &se) {
		return err
	}

	return &stackError{
		err:   err,
		stack: debug.Stack(),
	}
}

func WrapInternal(cause error) *AppError {
	return NewInternal(WithStack(cause))
}

func Normalize(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Status >= 500 && appErr.Cause != nil {
			appErr.Cause = WithStack(appErr.Cause)
		}
		return appErr
	}

	return NewInternal(WithStack(err))
}
