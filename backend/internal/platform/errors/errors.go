package errors

import "fmt"

type AppError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	MessageKey string         `json:"message_key,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	HTTPStatus int            `json:"-"`
	Err        error          `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return e.Code + ": " + e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, MessageKey: code, HTTPStatus: httpStatus}
}

func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, MessageKey: code, HTTPStatus: httpStatus, Err: err}
}

var (
	ErrUnauthorized     = New("AUTH_INVALID", "Unauthorized", 401)
	ErrForbidden        = New("AUTH_FORBIDDEN", "Forbidden", 403)
	ErrNotFound         = New("NOT_FOUND", "Resource not found", 404)
	ErrValidation       = New("VALIDATION_FAILED", "Validation failed", 422)
	ErrConflict         = New("CONFLICT", "Conflict", 409)
	ErrInternal         = New("INTERNAL_ERROR", "Internal server error", 500)
	ErrTenantNotFound   = New("TENANT_NOT_FOUND", "Tenant not found", 404)
	ErrInvalidCreds     = New("AUTH_INVALID", "Invalid credentials", 401)
	ErrTokenInvalid     = New("AUTH_INVALID", "Invalid or expired token", 401)
)
