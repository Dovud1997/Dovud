package http

import (
	"strconv"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/application"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct{ svc *application.Service }

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r fiber.Router) {
	r.Get("/agents", httpx.RequirePermissions("agents:read"), h.ListAgents)
	r.Post("/agents", httpx.RequirePermissions("agents:write"), h.CreateAgent)
	r.Get("/agents/:id", httpx.RequirePermissions("agents:read"), h.GetAgent)
	r.Put("/agents/:id", httpx.RequirePermissions("agents:write"), h.UpdateAgent)
	r.Delete("/agents/:id", httpx.RequirePermissions("agents:write"), h.DeleteAgent)

	r.Get("/routes", httpx.RequirePermissions("routes:read"), h.ListRoutes)
	r.Post("/routes", httpx.RequirePermissions("routes:write"), h.CreateRoute)
	r.Get("/routes/:id", httpx.RequirePermissions("routes:read"), h.GetRoute)
	r.Put("/routes/:id", httpx.RequirePermissions("routes:write"), h.UpdateRoute)
	r.Delete("/routes/:id", httpx.RequirePermissions("routes:write"), h.DeleteRoute)
	r.Put("/routes/:id/stops", httpx.RequirePermissions("routes:write"), h.SetStops)

	r.Get("/visits", httpx.RequirePermissions("visits:read"), h.ListVisits)
	r.Get("/visits/:id", httpx.RequirePermissions("visits:read"), h.GetVisit)
	r.Post("/visits/check-in", httpx.RequirePermissions("visits:write"), h.CheckIn)
	r.Post("/visits/:id/check-out", httpx.RequirePermissions("visits:write"), h.CheckOut)
	r.Post("/visits/:id/photos", httpx.RequirePermissions("visits:write"), h.AddPhoto)
	r.Get("/visits/:id/comments", httpx.RequirePermissions("visits:read"), h.ListComments)
	r.Post("/visits/:id/comments", httpx.RequirePermissions("visits:write"), h.AddComment)

	r.Post("/gps/points", httpx.RequirePermissions("visits:write"), h.UploadPoints)
	r.Get("/gps/agents/:id/live", httpx.RequirePermissions("visits:write"), h.LivePosition)
	r.Get("/gps/agents/:id/track", httpx.RequirePermissions("visits:write"), h.TrackHistory)
}

func (h *Handler) ListAgents(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.ListAgents(c.Context(), cl.TenantID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) GetAgent(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetAgent(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateAgent(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.AgentInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateAgent(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateAgent(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.AgentInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateAgent(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteAgent(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteAgent(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) ListRoutes(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	var agentID *uuid.UUID
	if q := c.Query("agent_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			return httpx.Fail(c, err)
		}
		agentID = &id
	}
	var date *time.Time
	if q := c.Query("date"); q != "" {
		t, err := time.Parse("2006-01-02", q)
		if err != nil {
			return httpx.Fail(c, err)
		}
		date = &t
	}
	res, total, err := h.svc.ListRoutes(c.Context(), cl.TenantID, agentID, date, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) GetRoute(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetRoute(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CreateRoute(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.RouteInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CreateRoute(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UpdateRoute(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.RouteInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UpdateRoute(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteRoute(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.svc.DeleteRoute(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"ok": true})
}

func (h *Handler) SetStops(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in []application.RouteStopInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.SetStops(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListVisits(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	var agentID, customerID *uuid.UUID
	if q := c.Query("agent_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			return httpx.Fail(c, err)
		}
		agentID = &id
	}
	if q := c.Query("customer_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			return httpx.Fail(c, err)
		}
		customerID = &id
	}
	res, total, err := h.svc.ListVisits(c.Context(), cl.TenantID, agentID, customerID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) GetVisit(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.GetVisit(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) CheckIn(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CheckInInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CheckIn(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) CheckOut(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CheckOutInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.CheckOut(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) AddPhoto(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.PhotoInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.AddPhoto(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) ListComments(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.ListComments(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) AddComment(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CommentInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.AddComment(c.Context(), cl.TenantID, id, cl.UserID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) UploadPoints(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in []application.GpsPointInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.UploadPoints(c.Context(), cl.TenantID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) LivePosition(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, err := h.svc.LivePosition(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) TrackHistory(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, err)
	}
	var from, to time.Time
	if q := c.Query("from"); q != "" {
		from, err = time.Parse(time.RFC3339, q)
		if err != nil {
			return httpx.Fail(c, err)
		}
	}
	if q := c.Query("to"); q != "" {
		to, err = time.Parse(time.RFC3339, q)
		if err != nil {
			return httpx.Fail(c, err)
		}
	}
	res, err := h.svc.TrackHistory(c.Context(), cl.TenantID, id, from, to)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
