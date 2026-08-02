package http

import (
	"github.com/Dovud1997/Dovud/backend/internal/modules/tenant/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc *application.TenantService
}

func NewHandler(svc *application.TenantService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterPublic(r fiber.Router) {
	r.Get("/public/health", func(c *fiber.Ctx) error {
		return httpx.OK(c, fiber.Map{"status": "ok"})
	})
	r.Get("/public/ready", func(c *fiber.Ctx) error {
		return httpx.OK(c, fiber.Map{"status": "ready"})
	})
	r.Get("/public/locales", func(c *fiber.Ctx) error {
		return httpx.OK(c, []string{"ru", "uz", "en"})
	})
	r.Get("/public/branding", h.PublicBranding)
}

func (h *Handler) RegisterProtected(r fiber.Router) {
	r.Get("/tenant", httpx.RequirePermissions("tenant:read"), h.GetTenant)
	r.Put("/tenant", httpx.RequirePermissions("tenant:write"), h.UpdateTenant)
	r.Get("/tenant/branding", httpx.RequirePermissions("tenant:read"), h.GetBranding)
	r.Put("/tenant/branding", httpx.RequirePermissions("tenant:write"), h.UpdateBranding)
	r.Post("/tenant/branding/assets", httpx.RequirePermissions("tenant:write"), h.AttachBrandingAsset)
	r.Get("/tenant/domains", httpx.RequirePermissions("tenant:read"), h.ListDomains)
	r.Post("/tenant/domains", httpx.RequirePermissions("tenant:write"), h.AddDomain)
	r.Delete("/tenant/domains/:id", httpx.RequirePermissions("tenant:write"), h.DeleteDomain)
	r.Get("/tenant/providers", httpx.RequirePermissions("tenant:read"), h.ListProviders)
	r.Put("/tenant/providers/:type", httpx.RequirePermissions("tenant:write"), h.UpsertProvider)
	r.Post("/tenant/providers/:type/test", httpx.RequirePermissions("tenant:write"), h.TestProvider)
}

func (h *Handler) PublicBranding(c *fiber.Ctx) error {
	code := c.Query("tenant")
	host := c.Query("host")
	if host == "" {
		host = c.Hostname()
	}
	res, err := h.svc.PublicBranding(c.Context(), code, host)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) GetTenant(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.Get(c.Context(), claims.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) UpdateTenant(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.UpdateTenantInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.Update(c.Context(), claims.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) GetBranding(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetBranding(c.Context(), claims.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) UpdateBranding(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.UpdateBrandingInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateBranding(c.Context(), claims.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) AttachBrandingAsset(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.AttachBrandingAssetInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.AttachBrandingAsset(c.Context(), claims.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListDomains(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListDomains(c.Context(), claims.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) AddDomain(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CreateDomainInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.AddDomain(c.Context(), claims.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) DeleteDomain(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteDomain(c.Context(), claims.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListProviders(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListProviders(c.Context(), claims.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) UpsertProvider(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.UpsertProviderInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpsertProvider(c.Context(), claims.TenantID, c.Params("type"), in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) TestProvider(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.TestProviderInput
	_ = c.BodyParser(&in)
	if err := h.svc.TestProvider(c.Context(), claims.TenantID, c.Params("type"), in); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}
