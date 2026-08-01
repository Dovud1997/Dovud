package http

import (
	"strconv"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	auth *application.AuthService
	rbac *application.RBACService
}

func NewHandler(auth *application.AuthService, rbac *application.RBACService) *Handler {
	return &Handler{auth: auth, rbac: rbac}
}

func (h *Handler) RegisterPublic(r fiber.Router) {
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/logout", h.Logout)
}

func (h *Handler) RegisterProtected(r fiber.Router) {
	r.Get("/auth/me", h.Me)
	r.Patch("/auth/me", h.UpdateMe)
	r.Post("/auth/change-password", h.ChangePassword)

	r.Get("/permissions", httpx.RequirePermissions("roles:read"), h.ListPermissions)
	r.Get("/roles", httpx.RequirePermissions("roles:read"), h.ListRoles)
	r.Post("/roles", httpx.RequirePermissions("roles:write"), h.CreateRole)
	r.Put("/roles/:id/permissions", httpx.RequirePermissions("roles:write"), h.SetRolePermissions)

	r.Get("/users", httpx.RequirePermissions("users:read"), h.ListUsers)
	r.Post("/users", httpx.RequirePermissions("users:write"), h.CreateUser)
	r.Put("/users/:id/roles", httpx.RequirePermissions("users:write"), h.AssignRoles)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var in application.LoginInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.auth.Login(c.Context(), in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
		DeviceID     string `json:"device_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.auth.Refresh(c.Context(), body.RefreshToken, body.DeviceID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.BodyParser(&body)
	_ = h.auth.Logout(c.Context(), body.RefreshToken)
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) Me(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.auth.Me(c.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) UpdateMe(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.UpdateMeInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.auth.UpdateMe(c.Context(), claims.TenantID, claims.UserID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.auth.ChangePassword(c.Context(), claims.TenantID, claims.UserID, body.CurrentPassword, body.NewPassword); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListPermissions(c *fiber.Ctx) error {
	res, err := h.rbac.ListPermissions(c.Context())
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListRoles(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.rbac.ListRoles(c.Context(), claims.TenantID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateRole(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CreateRoleInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.rbac.CreateRole(c.Context(), claims.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) SetRolePermissions(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	roleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var body struct {
		PermissionCodes []string `json:"permission_codes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.rbac.SetRolePermissions(c.Context(), claims.TenantID, roleID, body.PermissionCodes); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.rbac.ListUsers(c.Context(), claims.TenantID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CreateUserInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.rbac.CreateUser(c.Context(), claims.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) AssignRoles(c *fiber.Ctx) error {
	claims, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var body struct {
		RoleIDs []uuid.UUID `json:"role_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.rbac.AssignRoles(c.Context(), claims.TenantID, userID, body.RoleIDs); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}
