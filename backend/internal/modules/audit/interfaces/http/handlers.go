package http

import (
	"strconv"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/audit/application"
	"github.com/Dovud1997/Dovud/backend/internal/modules/audit/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/audit-logs", httpx.RequirePermissions("audit:read"), h.List)
	r.Get("/audit-logs/:id", httpx.RequirePermissions("audit:read"), h.Get)
}

func (h *Handler) List(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	filters := domain.ListFilters{
		EntityType: strings.TrimSpace(c.Query("entity_type")),
		Action:     strings.TrimSpace(c.Query("action")),
	}
	if raw := c.Query("actor_user_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return httpx.Fail(c, apperrors.ErrValidation)
		}
		filters.ActorUserID = &id
	}
	if raw := c.Query("entity_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return httpx.Fail(c, apperrors.ErrValidation)
		}
		filters.EntityID = &id
	}
	if raw := c.Query("from"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return httpx.Fail(c, apperrors.ErrValidation)
		}
		filters.From = &t
	}
	if raw := c.Query("to"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return httpx.Fail(c, apperrors.ErrValidation)
		}
		filters.To = &t
	}
	res, total, err := h.svc.List(c.Context(), cl.TenantID, filters, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	res, err := h.svc.Get(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
