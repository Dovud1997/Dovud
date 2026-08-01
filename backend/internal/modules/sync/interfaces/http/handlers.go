package http

import (
	"strconv"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Post("/sync/bootstrap", httpx.RequirePermissions("sync:use"), h.Bootstrap)
	r.Get("/sync/pull", httpx.RequirePermissions("sync:use"), h.Pull)
	r.Post("/sync/push", httpx.RequirePermissions("sync:use"), h.Push)
	r.Get("/sync/conflicts", httpx.RequirePermissions("sync:use"), h.ListConflicts)
	r.Post("/sync/conflicts/:id/resolve", httpx.RequirePermissions("sync:use"), h.ResolveConflict)
	r.Get("/sync/status", httpx.RequirePermissions("sync:use"), h.Status)
}

func deviceIDFrom(c *fiber.Ctx, bodyDeviceID string) string {
	if id := strings.TrimSpace(bodyDeviceID); id != "" {
		return id
	}
	if q := strings.TrimSpace(c.Query("device_id")); q != "" {
		return q
	}
	if cl, err := httpx.ClaimsFromCtx(c); err == nil && cl.DeviceID != "" {
		return cl.DeviceID
	}
	return ""
}

func (h *Handler) Bootstrap(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.BootstrapInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	in.DeviceID = deviceIDFrom(c, in.DeviceID)
	res, err := h.svc.Bootstrap(c.Context(), cl.TenantID, cl.UserID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Pull(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	deviceID := deviceIDFrom(c, "")
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.Query("limit", "100"))

	var types []string
	if raw := strings.TrimSpace(c.Query("types")); raw != "" {
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				types = append(types, t)
			}
		}
	}

	res, err := h.svc.Pull(c.Context(), cl.TenantID, cl.UserID, deviceID, cursor, types, limit)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Push(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.PushInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	in.DeviceID = deviceIDFrom(c, in.DeviceID)
	res, err := h.svc.Push(c.Context(), cl.TenantID, cl.UserID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListConflicts(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	deviceID := deviceIDFrom(c, "")
	res, err := h.svc.ListConflicts(c.Context(), cl.TenantID, deviceID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ResolveConflict(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.ResolveConflictInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ResolveConflict(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Status(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	deviceID := deviceIDFrom(c, "")
	res, err := h.svc.Status(c.Context(), cl.TenantID, cl.UserID, deviceID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
