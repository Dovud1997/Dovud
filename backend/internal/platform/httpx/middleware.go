package httpx

import (
	"strings"

	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID         uuid.UUID
	TenantID       uuid.UUID
	Roles          []string
	Permissions    []string
	DeviceID       string
	JTI            string
	IsPlatformAdmin bool
}

type JWTValidator interface {
	ParseAccessToken(token string) (*TokenClaims, error)
}

func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Locals("request_id", id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}

func AuthMiddleware(validator JWTValidator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return Fail(c, apperrors.ErrUnauthorized)
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := validator.ParseAccessToken(token)
		if err != nil {
			return Fail(c, apperrors.ErrTokenInvalid)
		}
		c.Locals("claims", claims)
		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		return c.Next()
	}
}

func RequirePermissions(perms ...string) fiber.Handler {
	needed := map[string]struct{}{}
	for _, p := range perms {
		needed[p] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*TokenClaims)
		if !ok || claims == nil {
			return Fail(c, apperrors.ErrUnauthorized)
		}
		if claims.IsPlatformAdmin {
			return c.Next()
		}
		have := map[string]struct{}{}
		for _, p := range claims.Permissions {
			have[p] = struct{}{}
		}
		for p := range needed {
			if _, ok := have[p]; !ok {
				return Fail(c, apperrors.ErrForbidden)
			}
		}
		return c.Next()
	}
}

func ClaimsFromCtx(c *fiber.Ctx) (*TokenClaims, error) {
	claims, ok := c.Locals("claims").(*TokenClaims)
	if !ok || claims == nil {
		return nil, apperrors.ErrUnauthorized
	}
	return claims, nil
}
