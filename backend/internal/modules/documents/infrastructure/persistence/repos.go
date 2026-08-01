package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/documents/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileRepo struct{ db *gorm.DB }

func NewFileRepo(db *gorm.DB) *FileRepo { return &FileRepo{db: db} }

func toFile(m FileModel) domain.File {
	return domain.File{
		ID: m.ID, TenantID: m.TenantID, Bucket: m.Bucket, ObjectKey: m.ObjectKey,
		FileName: m.FileName, Mime: m.Mime, Size: m.Size, Checksum: m.Checksum,
		Status: m.Status, UploadedBy: m.UploadedBy, CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt, CompletedAt: m.CompletedAt,
	}
}

func (r *FileRepo) Create(ctx context.Context, f *domain.File) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	now := time.Now().UTC()
	f.CreatedAt, f.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&FileModel{
		ID: f.ID, TenantID: f.TenantID, Bucket: f.Bucket, ObjectKey: f.ObjectKey,
		FileName: f.FileName, Mime: f.Mime, Size: f.Size, Checksum: f.Checksum,
		Status: f.Status, UploadedBy: f.UploadedBy, CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt, CompletedAt: f.CompletedAt,
	}).Error
}

func (r *FileRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.File, error) {
	var m FileModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	f := toFile(m)
	return &f, nil
}

func (r *FileRepo) Update(ctx context.Context, f *domain.File) error {
	f.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&FileModel{}).Where("id = ? AND tenant_id = ?", f.ID, f.TenantID).Updates(map[string]any{
		"file_name": f.FileName, "mime": f.Mime, "size": f.Size, "checksum": f.Checksum,
		"status": f.Status, "completed_at": f.CompletedAt, "updated_at": f.UpdatedAt,
	}).Error
}

func (r *FileRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&FileModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *FileRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.File, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&FileModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []FileModel
	if err := q.Order("created_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.File, 0, len(rows))
	for _, m := range rows {
		out = append(out, toFile(m))
	}
	return out, total, nil
}

type DocumentRepo struct{ db *gorm.DB }

func NewDocumentRepo(db *gorm.DB) *DocumentRepo { return &DocumentRepo{db: db} }

func toDoc(m DocumentModel) domain.Document {
	return domain.Document{
		ID: m.ID, TenantID: m.TenantID, Title: m.Title, Description: m.Description,
		DocType: m.DocType, Status: m.Status, CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *DocumentRepo) Create(ctx context.Context, d *domain.Document) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	now := time.Now().UTC()
	d.CreatedAt, d.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&DocumentModel{
		ID: d.ID, TenantID: d.TenantID, Title: d.Title, Description: d.Description,
		DocType: d.DocType, Status: d.Status, CreatedBy: d.CreatedBy,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}).Error
}

func (r *DocumentRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Document, error) {
	var m DocumentModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	d := toDoc(m)
	return &d, nil
}

func (r *DocumentRepo) Update(ctx context.Context, d *domain.Document) error {
	d.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&DocumentModel{}).Where("id = ? AND tenant_id = ?", d.ID, d.TenantID).Updates(map[string]any{
		"title": d.Title, "description": d.Description, "doc_type": d.DocType,
		"status": d.Status, "updated_at": d.UpdatedAt,
	}).Error
}

func (r *DocumentRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&DocumentModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *DocumentRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.Document, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&DocumentModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []DocumentModel
	if err := q.Order("created_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Document, 0, len(rows))
	for _, m := range rows {
		out = append(out, toDoc(m))
	}
	return out, total, nil
}

func (r *DocumentRepo) AttachFile(ctx context.Context, documentID, fileID uuid.UUID, role string) error {
	if role == "" {
		role = "attachment"
	}
	return r.db.WithContext(ctx).Create(&DocumentFileModel{
		DocumentID: documentID, FileID: fileID, Role: role, CreatedAt: time.Now().UTC(),
	}).Error
}

func (r *DocumentRepo) ListFiles(ctx context.Context, documentID uuid.UUID) ([]domain.File, error) {
	var links []DocumentFileModel
	if err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []domain.File{}, nil
	}
	ids := make([]uuid.UUID, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.FileID)
	}
	var rows []FileModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.File, 0, len(rows))
	for _, m := range rows {
		out = append(out, toFile(m))
	}
	return out, nil
}
