package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/portal/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/portal/summary", httpx.RequirePermissions("portal:read"), h.Summary)
	r.Get("/portal/orders", httpx.RequirePermissions("portal:read"), h.Orders)
	r.Get("/portal/receivables", httpx.RequirePermissions("portal:read"), h.Receivables)
	r.Get("/portal/documents", httpx.RequirePermissions("portal:read"), h.Documents)

	r.Get("/portal/links", httpx.RequirePermissions("portal:write"), h.ListLinks)
	r.Post("/portal/links", httpx.RequirePermissions("portal:write"), h.LinkUser)
	r.Delete("/portal/links/:user_id", httpx.RequirePermissions("portal:write"), h.UnlinkUser)
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.Summary(c.Context(), cl.TenantID, cl.UserID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Orders(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.Orders(c.Context(), cl.TenantID, cl.UserID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) Receivables(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.Receivables(c.Context(), cl.TenantID, cl.UserID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) Documents(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.Documents(c.Context(), cl.TenantID, cl.UserID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) LinkUser(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.LinkInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.LinkUser(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UnlinkUser(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.UnlinkUser(c.Context(), cl.TenantID, userID); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListLinks(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var customerID *uuid.UUID
	if raw := c.Query("customer_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return httpx.Fail(c, err)
		}
		customerID = &id
	}
	res, err := h.svc.ListLinks(c.Context(), cl.TenantID, customerID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
