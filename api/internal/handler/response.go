package handler

import (
	"app-api/internal/app_error"
	"app-api/internal/i18n"
)

type SuccessResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code    string                          `json:"code"`
	Message string                          `json:"message"`
	Errors  map[string][]FieldErrorResponse `json:"errors,omitempty"`
}

type FieldErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToResponse(err *app_error.AppError, lang string, t i18n.Translator) ErrorResponse {
	res := ErrorResponse{
		Code:    err.Code,
		Message: t.Translate(lang, err.Code),
	}

	if len(err.FieldErrors) == 0 {
		return res
	}

	res.Errors = map[string][]FieldErrorResponse{}
	for field, errs := range err.FieldErrors {
		for _, fe := range errs {
			res.Errors[field] = append(res.Errors[field], FieldErrorResponse{
				Code:    fe.Code,
				Message: t.Translate(lang, fe.Code),
			})
		}
	}

	return res
}
