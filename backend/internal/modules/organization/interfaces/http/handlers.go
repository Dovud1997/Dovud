package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/organization/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/companies", httpx.RequirePermissions("branches:read"), h.ListCompanies)
	r.Post("/companies", httpx.RequirePermissions("branches:write"), h.CreateCompany)
	r.Put("/companies/:id", httpx.RequirePermissions("branches:write"), h.UpdateCompany)
	r.Delete("/companies/:id", httpx.RequirePermissions("branches:write"), h.DeleteCompany)

	r.Get("/branches", httpx.RequirePermissions("branches:read"), h.ListBranches)
	r.Post("/branches", httpx.RequirePermissions("branches:write"), h.CreateBranch)
	r.Put("/branches/:id", httpx.RequirePermissions("branches:write"), h.UpdateBranch)
	r.Delete("/branches/:id", httpx.RequirePermissions("branches:write"), h.DeleteBranch)

	r.Get("/warehouses", httpx.RequirePermissions("warehouses:read"), h.ListWarehouses)
	r.Post("/warehouses", httpx.RequirePermissions("warehouses:write"), h.CreateWarehouse)
	r.Put("/warehouses/:id", httpx.RequirePermissions("warehouses:write"), h.UpdateWarehouse)
	r.Delete("/warehouses/:id", httpx.RequirePermissions("warehouses:write"), h.DeleteWarehouse)
	r.Get("/warehouses/:id/stocks", httpx.RequirePermissions("warehouses:read"), h.ListStocks)
	r.Put("/warehouses/:id/stocks/:productId", httpx.RequirePermissions("warehouses:write"), h.UpsertStock)
}

func claims(c *fiber.Ctx) (*httpx.TokenClaims, error) { return httpx.ClaimsFromCtx(c) }

func (h *Handler) ListCompanies(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.ListCompanies(c.Context(), cl.TenantID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) CreateCompany(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CompanyInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateCompany(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateCompany(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CompanyInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateCompany(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteCompany(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteCompany(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListBranches(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.ListBranches(c.Context(), cl.TenantID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) CreateBranch(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.BranchInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateBranch(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateBranch(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.BranchInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateBranch(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteBranch(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteBranch(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListWarehouses(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	var branchID *uuid.UUID
	if raw := c.Query("branch_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return httpx.Fail(c, err)
		}
		branchID = &id
	}
	res, total, err := h.svc.ListWarehouses(c.Context(), cl.TenantID, branchID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) CreateWarehouse(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.WarehouseInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateWarehouse(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateWarehouse(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.WarehouseInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateWarehouse(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteWarehouse(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteWarehouse(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListStocks(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListStocks(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) UpsertStock(c *fiber.Ctx) error {
	cl, err := claims(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	wid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	pid, err := uuid.Parse(c.Params("productId"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.StockInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpsertStock(c.Context(), cl.TenantID, wid, pid, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
