package http

import (
	"context"
	"log/slog"

	"github.com/Dovud1997/Dovud/backend/internal/modules/audit/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// HTTPWriter adapts audit.Service to httpx.AuditWriter.
type HTTPWriter struct {
	svc *application.Service
	log *slog.Logger
}

func NewHTTPWriter(svc *application.Service, log *slog.Logger) *HTTPWriter {
	if log == nil {
		log = slog.Default()
	}
	return &HTTPWriter{svc: svc, log: log}
}

func (w *HTTPWriter) WriteHTTP(c *fiber.Ctx, action string, entityType string, entityID *uuid.UUID, after any) {
	if w == nil || w.svc == nil {
		return
	}
	var tenantID *uuid.UUID
	var actor *uuid.UUID
	if cl, err := httpx.ClaimsFromCtx(c); err == nil {
		tenantID = &cl.TenantID
		actor = &cl.UserID
	}
	reqID := httpx.RequestID(c)
	ip := c.IP()
	ua := c.Get("User-Agent")
	go func() {
		_, err := w.svc.Write(context.Background(), application.WriteInput{
			TenantID: tenantID, ActorUserID: actor, Action: action,
			EntityType: entityType, EntityID: entityID, After: after,
			IP: ip, UserAgent: ua, RequestID: reqID,
		})
		if err != nil {
			w.log.Warn("audit write failed", "error", err, "action", action)
		}
	}()
}
