package httpx

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuditWriter records mutating API calls. Kept as an interface to avoid
// importing the audit module into httpx.
type AuditWriter interface {
	WriteHTTP(c *fiber.Ctx, action string, entityType string, entityID *uuid.UUID, after any)
}

func AuditMiddleware(writer AuditWriter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if writer == nil || err != nil {
			return err
		}
		method := c.Method()
		if method != fiber.MethodPost && method != fiber.MethodPut && method != fiber.MethodPatch && method != fiber.MethodDelete {
			return nil
		}
		if c.Response().StatusCode() >= 400 {
			return nil
		}
		path := c.Path()
		if strings.Contains(path, "/auth/") || strings.Contains(path, "/sync/") || strings.Contains(path, "/files/local/") {
			return nil
		}
		action := method + " " + path
		entityType, entityID := inferEntity(path)
		writer.WriteHTTP(c, action, entityType, entityID, nil)
		return nil
	}
}

func inferEntity(path string) (string, *uuid.UUID) {
	// /api/v1/<resource>/:id...
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return "", nil
	}
	// parts[0]=api parts[1]=v1 parts[2]=resource
	resource := parts[2]
	var id *uuid.UUID
	if len(parts) >= 4 {
		if parsed, err := uuid.Parse(parts[3]); err == nil {
			id = &parsed
		}
	}
	return resource, id
}
