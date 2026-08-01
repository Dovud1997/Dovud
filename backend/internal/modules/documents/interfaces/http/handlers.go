package http

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/documents/application"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/httpx"
	"github.com/Dovud1997/Dovud/backend/internal/platform/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc   *application.Service
	store storage.ObjectStore
}

func NewHandler(svc *application.Service, store storage.ObjectStore) *Handler {
	return &Handler{svc: svc, store: store}
}

func (h *Handler) RegisterPublic(r fiber.Router) {
	// Signed local upload/download (no JWT; HMAC in query).
	r.Put("/files/local/put", h.LocalPut)
	r.Get("/files/local/get", h.LocalGet)
}

func (h *Handler) Register(r fiber.Router) {
	// Static paths before /files/:id
	r.Post("/files/presign", httpx.RequirePermissions("documents:write"), h.Presign)
	r.Get("/files", httpx.RequirePermissions("documents:read"), h.ListFiles)
	r.Post("/files/:id/complete", httpx.RequirePermissions("documents:write"), h.Complete)
	r.Get("/files/:id", httpx.RequirePermissions("documents:read"), h.GetFile)
	r.Delete("/files/:id", httpx.RequirePermissions("documents:write"), h.DeleteFile)

	r.Get("/documents", httpx.RequirePermissions("documents:read"), h.ListDocuments)
	r.Post("/documents", httpx.RequirePermissions("documents:write"), h.CreateDocument)
	r.Get("/documents/:id", httpx.RequirePermissions("documents:read"), h.GetDocument)
	r.Put("/documents/:id", httpx.RequirePermissions("documents:write"), h.UpdateDocument)
	r.Delete("/documents/:id", httpx.RequirePermissions("documents:write"), h.DeleteDocument)
	r.Post("/documents/:id/files", httpx.RequirePermissions("documents:write"), h.AttachFile)
}

func (h *Handler) Presign(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.PresignInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	res, err := h.svc.PresignUpload(c.Context(), cl.TenantID, cl.UserID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) Complete(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	var in application.CompleteInput
	_ = c.BodyParser(&in)
	res, err := h.svc.CompleteUpload(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) GetFile(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	res, err := h.svc.GetFile(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) ListFiles(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.ListFiles(c.Context(), cl.TenantID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) DeleteFile(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	if err := h.svc.DeleteFile(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"deleted": true})
}

func (h *Handler) verifyLocal(c *fiber.Ctx, op string) (string, error) {
	local, ok := h.store.(*storage.LocalStore)
	if !ok {
		return "", apperrors.ErrUnavailable
	}
	key := strings.TrimSpace(c.Query("key"))
	sig := strings.TrimSpace(c.Query("sig"))
	exp, err := storage.ParseExp(c.Query("exp"))
	if err != nil || key == "" || sig == "" {
		return "", apperrors.ErrUnauthorized
	}
	if !local.Verify(op, key, sig, exp) {
		return "", apperrors.ErrUnauthorized
	}
	return key, nil
}

func (h *Handler) LocalPut(c *fiber.Ctx) error {
	key, err := h.verifyLocal(c, "put")
	if err != nil {
		return httpx.Fail(c, err)
	}
	local := h.store.(*storage.LocalStore)
	body := c.Body()
	ct := c.Get("Content-Type")
	if err := local.Put(c.Context(), key, ct, bytes.NewReader(body), int64(len(body))); err != nil {
		return httpx.Fail(c, apperrors.Wrap(err, "STORAGE_ERROR", "Local upload failed", 500))
	}
	return httpx.OK(c, fiber.Map{"uploaded": true, "key": key})
}

func (h *Handler) LocalGet(c *fiber.Ctx) error {
	key, err := h.verifyLocal(c, "get")
	if err != nil {
		return httpx.Fail(c, err)
	}
	local := h.store.(*storage.LocalStore)
	f, err := local.Open(key)
	if err != nil {
		return httpx.Fail(c, apperrors.ErrNotFound)
	}
	defer f.Close()
	c.Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(c.Response().BodyWriter(), f)
	return err
}

func (h *Handler) ListDocuments(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	res, total, err := h.svc.ListDocuments(c.Context(), cl.TenantID, page, perPage)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OKMeta(c, res, httpx.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *Handler) CreateDocument(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var in application.CreateDocumentInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	res, err := h.svc.CreateDocument(c.Context(), cl.TenantID, cl.UserID, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.Created(c, res)
}

func (h *Handler) GetDocument(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	res, err := h.svc.GetDocument(c.Context(), cl.TenantID, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) UpdateDocument(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	var in application.UpdateDocumentInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	res, err := h.svc.UpdateDocument(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}

func (h *Handler) DeleteDocument(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	if err := h.svc.DeleteDocument(c.Context(), cl.TenantID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"deleted": true})
}

func (h *Handler) AttachFile(c *fiber.Ctx) error {
	cl, err := httpx.ClaimsFromCtx(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	var in application.AttachFileInput
	if err := c.BodyParser(&in); err != nil {
		return httpx.Fail(c, apperrors.ErrValidation)
	}
	res, err := h.svc.AttachFile(c.Context(), cl.TenantID, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res)
}
