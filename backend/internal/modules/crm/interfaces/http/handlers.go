package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/crm/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/customers", httpx.RequirePermissions("customers:read"), h.ListCustomers)
	r.Post("/customers", httpx.RequirePermissions("customers:write"), h.CreateCustomer)
	r.Get("/customers/:id", httpx.RequirePermissions("customers:read"), h.GetCustomer)
	r.Put("/customers/:id", httpx.RequirePermissions("customers:write"), h.UpdateCustomer)
	r.Delete("/customers/:id", httpx.RequirePermissions("customers:write"), h.DeleteCustomer)

	r.Get("/customers/:id/contacts", httpx.RequirePermissions("customers:read"), h.ListContacts)
	r.Post("/customers/:id/contacts", httpx.RequirePermissions("customers:write"), h.CreateContact)
	r.Put("/customers/:id/contacts/:contactId", httpx.RequirePermissions("customers:write"), h.UpdateContact)
	r.Delete("/customers/:id/contacts/:contactId", httpx.RequirePermissions("customers:write"), h.DeleteContact)

	r.Get("/customers/:id/addresses", httpx.RequirePermissions("customers:read"), h.ListAddresses)
	r.Post("/customers/:id/addresses", httpx.RequirePermissions("customers:write"), h.CreateAddress)
	r.Put("/customers/:id/addresses/:addressId", httpx.RequirePermissions("customers:write"), h.UpdateAddress)
	r.Delete("/customers/:id/addresses/:addressId", httpx.RequirePermissions("customers:write"), h.DeleteAddress)

	r.Get("/customer-categories", httpx.RequirePermissions("customers:read"), h.ListCustomerCategories)
	r.Post("/customer-categories", httpx.RequirePermissions("customers:write"), h.CreateCustomerCategory)
	r.Put("/customer-categories/:id", httpx.RequirePermissions("customers:write"), h.UpdateCustomerCategory)
	r.Delete("/customer-categories/:id", httpx.RequirePermissions("customers:write"), h.DeleteCustomerCategory)
}

func (h *Handler) ListCustomers(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.ListCustomers(c.Context(), cl.TenantID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) GetCustomer(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetCustomer(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateCustomer(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CustomerInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateCustomer(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateCustomer(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CustomerInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateCustomer(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteCustomer(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteCustomer(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListContacts(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListContacts(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateContact(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.ContactInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateContact(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateContact(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	contactID, err := uuid.Parse(c.Params("contactId"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.ContactInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateContact(c.Context(), cl.TenantID, id, contactID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteContact(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	contactID, err := uuid.Parse(c.Params("contactId"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteContact(c.Context(), cl.TenantID, id, contactID); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListAddresses(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListAddresses(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateAddress(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.AddressInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateAddress(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateAddress(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	addressID, err := uuid.Parse(c.Params("addressId"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.AddressInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateAddress(c.Context(), cl.TenantID, id, addressID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteAddress(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	addressID, err := uuid.Parse(c.Params("addressId"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteAddress(c.Context(), cl.TenantID, id, addressID); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListCustomerCategories(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListCustomerCategories(c.Context(), cl.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateCustomerCategory(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CustomerCategoryInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateCustomerCategory(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateCustomerCategory(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CustomerCategoryInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateCustomerCategory(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteCustomerCategory(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteCustomerCategory(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}
