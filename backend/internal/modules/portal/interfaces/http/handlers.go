package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/portal/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/portal/summary", httpx.RequirePermissions("portal:read"), h.Summary)
	r.Get("/portal/orders", httpx.RequirePermissions("portal:read"), h.Orders)
	r.Get("/portal/receivables", httpx.RequirePermissions("portal:read"), h.Receivables)
	r.Get("/portal/documents", httpx.RequirePermissions("portal:read"), h.Documents)
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
