package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/finance/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/finance/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/receivables", httpx.RequirePermissions("finance:read"), h.ListReceivables)
	r.Get("/receivables/:id", httpx.RequirePermissions("finance:read"), h.GetReceivable)
	r.Post("/receivables", httpx.RequirePermissions("finance:write"), h.CreateReceivable)
	r.Post("/receivables/:id/payments", httpx.RequirePermissions("finance:write"), h.RecordPayment)

	r.Get("/finance/customers/:id/balance", httpx.RequirePermissions("finance:read"), h.GetCustomerBalance)
	r.Get("/finance/customers/:id/credit-limit", httpx.RequirePermissions("finance:read"), h.GetCreditLimit)
	r.Put("/finance/customers/:id/credit-limit", httpx.RequirePermissions("finance:write"), h.SetCreditLimit)
	r.Get("/finance/aging", httpx.RequirePermissions("finance:read"), h.AgingReport)

	// Aliases under /customers/... are fine when this handler is mounted at API root;
	// prefer /finance/... paths above to avoid CRM conflicts.
	r.Get("/customers/:id/balance", httpx.RequirePermissions("finance:read"), h.GetCustomerBalance)
	r.Get("/customers/:id/credit-limit", httpx.RequirePermissions("finance:read"), h.GetCreditLimit)
	r.Put("/customers/:id/credit-limit", httpx.RequirePermissions("finance:write"), h.SetCreditLimit)
}

func parseOptionalUUID(q string) (*uuid.UUID, error) {
	if q == "" {
		return nil, nil
	}
	id, err := uuid.Parse(q)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (h *Handler) ListReceivables(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	customerID, err := parseOptionalUUID(c.Query("customer_id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	filters := domain.ReceivableListFilters{
		CustomerID: customerID,
		Status:     c.Query("status"),
	}
	res, total, err := h.svc.ListReceivables(c.Context(), cl.TenantID, filters, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) GetReceivable(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetReceivable(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateReceivable(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CreateReceivableInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateReceivable(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) RecordPayment(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.RecordPaymentInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	createdBy := cl.UserID
	res, err := h.svc.RecordPayment(c.Context(), cl.TenantID, id, in, &createdBy)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) GetCustomerBalance(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetCustomerBalance(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) GetCreditLimit(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetCreditLimit(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) SetCreditLimit(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.SetCreditLimitInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.SetCreditLimit(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) AgingReport(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.AgingReport(c.Context(), cl.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
