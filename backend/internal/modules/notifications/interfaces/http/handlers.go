package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/notifications/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/notifications", httpx.RequirePermissions("notifications:read"), h.List)
	r.Get("/notifications/unread-count", httpx.RequirePermissions("notifications:read"), h.UnreadCount)
	r.Post("/notifications/read-all", httpx.RequirePermissions("notifications:write"), h.MarkAllRead)
	r.Post("/notifications/test", httpx.RequirePermissions("notifications:write"), h.CreateTest)
	r.Post("/notifications", httpx.RequirePermissions("notifications:write"), h.Create)
	r.Post("/notifications/:id/read", httpx.RequirePermissions("notifications:write"), h.MarkRead)
}

func (h *Handler) List(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	unreadOnly := c.Query("unread") == "true" || c.Query("unread") == "1"
	res, total, err := h.svc.ListByUser(c.Context(), cl.TenantID, cl.UserID, unreadOnly, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) UnreadCount(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	count, err := h.svc.UnreadCount(c.Context(), cl.TenantID, cl.UserID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"count": count})
}

func (h *Handler) MarkRead(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.MarkRead(c.Context(), cl.TenantID, cl.UserID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) MarkAllRead(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	n, err := h.svc.MarkAllRead(c.Context(), cl.TenantID, cl.UserID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"updated": n})
}

func (h *Handler) Create(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CreateInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.Create(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) CreateTest(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateTest(c.Context(), cl.TenantID, cl.UserID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}
