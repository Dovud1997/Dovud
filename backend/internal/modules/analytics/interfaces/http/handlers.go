package http

import (
	"strconv"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/analytics/application"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/dashboard/summary", httpx.RequirePermissions("analytics:read"), h.DashboardSummary)
	r.Get("/kpi", httpx.RequirePermissions("analytics:read"), h.ListKPI)
	r.Get("/kpi/snapshots", httpx.RequirePermissions("analytics:read"), h.ListSnapshots)
	r.Get("/analytics/sales", httpx.RequirePermissions("analytics:read"), h.SalesAnalytics)
	r.Get("/analytics/visits", httpx.RequirePermissions("analytics:read"), h.VisitsAnalytics)
}

func (h *Handler) DashboardSummary(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.DashboardSummary(c.Context(), cl.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListKPI(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListKPIDefinitions(c.Context(), cl.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListSnapshots(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.ListKPISnapshots(c.Context(), cl.TenantID, c.Query("code"), c.Query("period"), page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func parseDateRange(c *fiber.Ctx) (time.Time, time.Time, error) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, apperrors.ErrValidation
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.ErrValidation
		}
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.ErrValidation
		}
	} else {
		// inclusive end date → exclusive next day
		to = to.Add(24 * time.Hour)
	}
	return from.UTC(), to.UTC(), nil
}

func (h *Handler) SalesAnalytics(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	from, to, err := parseDateRange(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.SalesAnalytics(c.Context(), cl.TenantID, from, to)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) VisitsAnalytics(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	from, to, err := parseDateRange(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.VisitsAnalytics(c.Context(), cl.TenantID, from, to)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
