package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/orders/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/orders/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/orders", httpx.RequirePermissions("orders:read"), h.ListOrders)
	r.Post("/orders", httpx.RequirePermissions("orders:write"), h.CreateOrder)
	r.Get("/orders/:id", httpx.RequirePermissions("orders:read"), h.GetOrder)
	r.Put("/orders/:id", httpx.RequirePermissions("orders:write"), h.UpdateDraft)
	r.Post("/orders/:id/submit", httpx.RequirePermissions("orders:write"), h.Submit)
	r.Post("/orders/:id/confirm", httpx.RequirePermissions("orders:approve"), h.Confirm)
	r.Post("/orders/:id/cancel", httpx.RequirePermissions("orders:write"), h.Cancel)
	r.Post("/orders/:id/status", httpx.RequirePermissions("orders:write"), h.TransitionStatus)
	r.Get("/orders/:id/history", httpx.RequirePermissions("orders:read"), h.ListHistory)
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

func (h *Handler) ListOrders(c *fiber.Ctx) error {
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
	agentID, err := parseOptionalUUID(c.Query("agent_id"))
	if err != nil {
		return httpx.Fail(c, err)
	}

	filters := domain.OrderListFilters{
		Status:     c.Query("status"),
		CustomerID: customerID,
		AgentID:    agentID,
	}
	res, total, err := h.svc.ListOrders(c.Context(), cl.TenantID, filters, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) GetOrder(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	order, history, err := h.svc.GetOrder(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"order": order, "history": history})
}

func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CreateOrderInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateOrder(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateDraft(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.UpdateDraftInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateDraft(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Submit(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	userID := cl.UserID
	res, err := h.svc.Submit(c.Context(), cl.TenantID, id, &userID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Confirm(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	userID := cl.UserID
	res, err := h.svc.Confirm(c.Context(), cl.TenantID, id, &userID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Cancel(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var body struct {
		Comment *string `json:"comment"`
	}
	_ = c.BodyParser(&body)
	userID := cl.UserID
	res, err := h.svc.Cancel(c.Context(), cl.TenantID, id, body.Comment, &userID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) TransitionStatus(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.TransitionStatusInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	userID := cl.UserID
	res, err := h.svc.TransitionStatus(c.Context(), cl.TenantID, id, in, &userID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListHistory(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListHistory(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
