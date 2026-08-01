package httpx

import (
	"errors"

	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Meta struct {
	Page    int   `json:"page,omitempty"`
	PerPage int   `json:"per_page,omitempty"`
	Total   int64 `json:"total,omitempty"`
}

type Envelope struct {
	Success   bool           `json:"success"`
	Data      any            `json:"data,omitempty"`
	Meta      *Meta          `json:"meta,omitempty"`
	Error     *ErrorBody     `json:"error"`
	RequestID string         `json:"request_id"`
}

type ErrorBody struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	MessageKey string         `json:"message_key,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

func RequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("request_id").(string); ok && id != "" {
		return id
	}
	id := c.Get("X-Request-ID")
	if id == "" {
		id = uuid.NewString()
	}
	c.Locals("request_id", id)
	return id
}

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{
		Success:   true,
		Data:      data,
		Error:     nil,
		RequestID: RequestID(c),
	})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{
		Success:   true,
		Data:      data,
		Error:     nil,
		RequestID: RequestID(c),
	})
}

func OKMeta(c *fiber.Ctx, data any, meta Meta) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{
		Success:   true,
		Data:      data,
		Meta:      &meta,
		Error:     nil,
		RequestID: RequestID(c),
	})
}

func Fail(c *fiber.Ctx, err error) error {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		status := appErr.HTTPStatus
		if status == 0 {
			status = fiber.StatusInternalServerError
		}
		return c.Status(status).JSON(Envelope{
			Success: false,
			Error: &ErrorBody{
				Code:       appErr.Code,
				Message:    appErr.Message,
				MessageKey: appErr.MessageKey,
				Details:    appErr.Details,
			},
			RequestID: RequestID(c),
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:       apperrors.ErrInternal.Code,
			Message:    apperrors.ErrInternal.Message,
			MessageKey: apperrors.ErrInternal.MessageKey,
		},
		RequestID: RequestID(c),
	})
}
