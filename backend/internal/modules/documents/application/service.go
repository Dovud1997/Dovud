package application

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/documents/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/outbox"
	"github.com/Dovud1997/Dovud/backend/internal/platform/storage"
	"github.com/google/uuid"
)

type Service struct {
	files     domain.FileRepository
	docs      domain.DocumentRepository
	store     storage.ObjectStore
	outbox    *outbox.Store
	presignTTL time.Duration
}

func NewService(files domain.FileRepository, docs domain.DocumentRepository, store storage.ObjectStore, outboxStore *outbox.Store) *Service {
	return &Service{
		files: files, docs: docs, store: store, outbox: outboxStore,
		presignTTL: 15 * time.Minute,
	}
}

type FileDTO struct {
	ID          uuid.UUID  `json:"id"`
	FileName    string     `json:"file_name"`
	Mime        string     `json:"mime"`
	Size        int64      `json:"size"`
	Status      string     `json:"status"`
	Bucket      string     `json:"bucket"`
	ObjectKey   string     `json:"object_key"`
	Checksum    *string    `json:"checksum,omitempty"`
	UploadedBy  *uuid.UUID `json:"uploaded_by,omitempty"`
	DownloadURL string     `json:"download_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type PresignResult struct {
	FileID    uuid.UUID `json:"file_id"`
	UploadURL string    `json:"upload_url"`
	Method    string    `json:"method"`
	ObjectKey string    `json:"object_key"`
	ExpiresIn int       `json:"expires_in_seconds"`
	Driver    string    `json:"driver"`
}

type DocumentDTO struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	DocType     string     `json:"doc_type"`
	Status      string     `json:"status"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Files       []FileDTO  `json:"files,omitempty"`
}

type PresignInput struct {
	FileName    string `json:"file_name"`
	Mime        string `json:"mime"`
	Size        int64  `json:"size"`
	Checksum    string `json:"checksum"`
}

type CompleteInput struct {
	Size     *int64  `json:"size"`
	Checksum *string `json:"checksum"`
}

type CreateDocumentInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	DocType     string  `json:"doc_type"`
}

type UpdateDocumentInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	DocType     *string `json:"doc_type"`
	Status      *string `json:"status"`
}

type AttachFileInput struct {
	FileID uuid.UUID `json:"file_id"`
	Role   string    `json:"role"`
}

func toFileDTO(f domain.File, downloadURL string) FileDTO {
	return FileDTO{
		ID: f.ID, FileName: f.FileName, Mime: f.Mime, Size: f.Size, Status: f.Status,
		Bucket: f.Bucket, ObjectKey: f.ObjectKey, Checksum: f.Checksum, UploadedBy: f.UploadedBy,
		DownloadURL: downloadURL, CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt, CompletedAt: f.CompletedAt,
	}
}

func toDocDTO(d domain.Document, files []FileDTO) DocumentDTO {
	return DocumentDTO{
		ID: d.ID, Title: d.Title, Description: d.Description, DocType: d.DocType, Status: d.Status,
		CreatedBy: d.CreatedBy, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, Files: files,
	}
}

func (s *Service) PresignUpload(ctx context.Context, tenantID, userID uuid.UUID, in PresignInput) (*PresignResult, error) {
	if s.store == nil {
		return nil, apperrors.ErrUnavailable
	}
	name := strings.TrimSpace(in.FileName)
	mime := strings.TrimSpace(in.Mime)
	if name == "" || mime == "" {
		return nil, apperrors.ErrValidation
	}
	if in.Size < 0 {
		return nil, apperrors.ErrValidation
	}

	fileID := uuid.New()
	ext := path.Ext(name)
	objectKey := fmt.Sprintf("tenants/%s/files/%s%s", tenantID.String(), fileID.String(), ext)

	var checksum *string
	if c := strings.TrimSpace(in.Checksum); c != "" {
		checksum = &c
	}

	f := &domain.File{
		ID: fileID, TenantID: tenantID, Bucket: s.store.Bucket(), ObjectKey: objectKey,
		FileName: name, Mime: mime, Size: in.Size, Checksum: checksum,
		Status: domain.FileStatusPending, UploadedBy: &userID,
	}
	if err := s.files.Create(ctx, f); err != nil {
		return nil, err
	}

	url, err := s.store.PresignPut(ctx, objectKey, mime, s.presignTTL)
	if err != nil {
		return nil, apperrors.Wrap(err, "STORAGE_ERROR", "Failed to create upload URL", 502)
	}

	return &PresignResult{
		FileID: fileID, UploadURL: url, Method: "PUT", ObjectKey: objectKey,
		ExpiresIn: int(s.presignTTL.Seconds()), Driver: s.store.Driver(),
	}, nil
}

func (s *Service) CompleteUpload(ctx context.Context, tenantID, fileID uuid.UUID, in CompleteInput) (*FileDTO, error) {
	f, err := s.files.FindByID(ctx, tenantID, fileID)
	if err != nil {
		return nil, err
	}
	if f.Status == domain.FileStatusReady {
		dto := toFileDTO(*f, "")
		return &dto, nil
	}
	if in.Size != nil {
		f.Size = *in.Size
	}
	if in.Checksum != nil {
		f.Checksum = in.Checksum
	}
	now := time.Now().UTC()
	f.Status = domain.FileStatusReady
	f.CompletedAt = &now
	if err := s.files.Update(ctx, f); err != nil {
		return nil, err
	}

	if s.outbox != nil {
		_ = s.outbox.Append(ctx, tenantID, "file", &f.ID, "document.uploaded", map[string]any{
			"file_id": f.ID.String(), "object_key": f.ObjectKey, "mime": f.Mime, "size": f.Size,
		})
		_ = s.outbox.Append(ctx, tenantID, "file", &f.ID, "media.uploaded", map[string]any{
			"file_id": f.ID.String(), "object_key": f.ObjectKey, "mime": f.Mime,
		})
	}

	downloadURL, _ := s.store.PresignGet(ctx, f.ObjectKey, s.presignTTL)
	dto := toFileDTO(*f, downloadURL)
	return &dto, nil
}

func (s *Service) GetFile(ctx context.Context, tenantID, fileID uuid.UUID) (*FileDTO, error) {
	f, err := s.files.FindByID(ctx, tenantID, fileID)
	if err != nil {
		return nil, err
	}
	var downloadURL string
	if f.Status == domain.FileStatusReady && s.store != nil {
		downloadURL, _ = s.store.PresignGet(ctx, f.ObjectKey, s.presignTTL)
	}
	dto := toFileDTO(*f, downloadURL)
	return &dto, nil
}

func (s *Service) DeleteFile(ctx context.Context, tenantID, fileID uuid.UUID) error {
	f, err := s.files.FindByID(ctx, tenantID, fileID)
	if err != nil {
		return err
	}
	if s.store != nil {
		_ = s.store.Delete(ctx, f.ObjectKey)
	}
	f.Status = domain.FileStatusDeleted
	_ = s.files.Update(ctx, f)
	return s.files.SoftDelete(ctx, tenantID, fileID)
}

func (s *Service) ListFiles(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]FileDTO, int64, error) {
	rows, total, err := s.files.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]FileDTO, 0, len(rows))
	for _, f := range rows {
		out = append(out, toFileDTO(f, ""))
	}
	return out, total, nil
}

func (s *Service) CreateDocument(ctx context.Context, tenantID, userID uuid.UUID, in CreateDocumentInput) (*DocumentDTO, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, apperrors.ErrValidation
	}
	docType := strings.TrimSpace(in.DocType)
	if docType == "" {
		docType = "general"
	}
	d := &domain.Document{
		TenantID: tenantID, Title: title, Description: in.Description,
		DocType: docType, Status: domain.DocStatusActive, CreatedBy: &userID,
	}
	if err := s.docs.Create(ctx, d); err != nil {
		return nil, err
	}
	dto := toDocDTO(*d, nil)
	return &dto, nil
}

func (s *Service) GetDocument(ctx context.Context, tenantID, id uuid.UUID) (*DocumentDTO, error) {
	d, err := s.docs.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	files, err := s.docs.ListFiles(ctx, id)
	if err != nil {
		return nil, err
	}
	fileDTOs := make([]FileDTO, 0, len(files))
	for _, f := range files {
		fileDTOs = append(fileDTOs, toFileDTO(f, ""))
	}
	dto := toDocDTO(*d, fileDTOs)
	return &dto, nil
}

func (s *Service) UpdateDocument(ctx context.Context, tenantID, id uuid.UUID, in UpdateDocumentInput) (*DocumentDTO, error) {
	d, err := s.docs.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return nil, apperrors.ErrValidation
		}
		d.Title = t
	}
	if in.Description != nil {
		d.Description = in.Description
	}
	if in.DocType != nil {
		d.DocType = strings.TrimSpace(*in.DocType)
	}
	if in.Status != nil {
		d.Status = strings.TrimSpace(*in.Status)
	}
	if err := s.docs.Update(ctx, d); err != nil {
		return nil, err
	}
	return s.GetDocument(ctx, tenantID, id)
}

func (s *Service) DeleteDocument(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.docs.SoftDelete(ctx, tenantID, id)
}

func (s *Service) ListDocuments(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]DocumentDTO, int64, error) {
	rows, total, err := s.docs.List(ctx, tenantID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]DocumentDTO, 0, len(rows))
	for _, d := range rows {
		out = append(out, toDocDTO(d, nil))
	}
	return out, total, nil
}

func (s *Service) AttachFile(ctx context.Context, tenantID, documentID uuid.UUID, in AttachFileInput) (*DocumentDTO, error) {
	if _, err := s.docs.FindByID(ctx, tenantID, documentID); err != nil {
		return nil, err
	}
	f, err := s.files.FindByID(ctx, tenantID, in.FileID)
	if err != nil {
		return nil, err
	}
	if f.Status != domain.FileStatusReady {
		return nil, apperrors.New("FILE_NOT_READY", "File upload is not completed", 422)
	}
	if err := s.docs.AttachFile(ctx, documentID, in.FileID, in.Role); err != nil {
		return nil, err
	}
	return s.GetDocument(ctx, tenantID, documentID)
}
