package persistence

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/organization/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyRepo struct{ db *gorm.DB }

func NewCompanyRepo(db *gorm.DB) *CompanyRepo { return &CompanyRepo{db: db} }

func (r *CompanyRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.Company, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&CompanyModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []CompanyModel
	if err := q.Order("code").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Company, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Company{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Inn: m.Inn, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, total, nil
}

func (r *CompanyRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Company, error) {
	var m CompanyModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.Company{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Inn: m.Inn, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *CompanyRepo) Create(ctx context.Context, c *domain.Company) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt, c.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&CompanyModel{ID: c.ID, TenantID: c.TenantID, Code: c.Code, Name: c.Name, Inn: c.Inn, Status: c.Status, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}).Error
}

func (r *CompanyRepo) Update(ctx context.Context, c *domain.Company) error {
	c.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&CompanyModel{}).Where("id = ? AND tenant_id = ?", c.ID, c.TenantID).Updates(map[string]any{
		"code": c.Code, "name": c.Name, "inn": c.Inn, "status": c.Status, "updated_at": c.UpdatedAt,
	}).Error
}

func (r *CompanyRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&CompanyModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type BranchRepo struct{ db *gorm.DB }

func NewBranchRepo(db *gorm.DB) *BranchRepo { return &BranchRepo{db: db} }

func (r *BranchRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.Branch, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&BranchModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []BranchModel
	if err := q.Order("code").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Branch, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Branch{ID: m.ID, TenantID: m.TenantID, CompanyID: m.CompanyID, Code: m.Code, Name: m.Name, Address: m.Address, Lat: m.Lat, Lng: m.Lng, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, total, nil
}

func (r *BranchRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Branch, error) {
	var m BranchModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.Branch{ID: m.ID, TenantID: m.TenantID, CompanyID: m.CompanyID, Code: m.Code, Name: m.Name, Address: m.Address, Lat: m.Lat, Lng: m.Lng, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *BranchRepo) Create(ctx context.Context, b *domain.Branch) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	now := time.Now().UTC()
	b.CreatedAt, b.UpdatedAt = now, now
	b.Code = strings.ToUpper(strings.TrimSpace(b.Code))
	return r.db.WithContext(ctx).Create(&BranchModel{ID: b.ID, TenantID: b.TenantID, CompanyID: b.CompanyID, Code: b.Code, Name: b.Name, Address: b.Address, Lat: b.Lat, Lng: b.Lng, Status: b.Status, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt}).Error
}

func (r *BranchRepo) Update(ctx context.Context, b *domain.Branch) error {
	b.UpdatedAt = time.Now().UTC()
	b.Code = strings.ToUpper(strings.TrimSpace(b.Code))
	return r.db.WithContext(ctx).Model(&BranchModel{}).Where("id = ? AND tenant_id = ?", b.ID, b.TenantID).Updates(map[string]any{
		"company_id": b.CompanyID, "code": b.Code, "name": b.Name, "address": b.Address, "lat": b.Lat, "lng": b.Lng, "status": b.Status, "updated_at": b.UpdatedAt,
	}).Error
}

func (r *BranchRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&BranchModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type WarehouseRepo struct{ db *gorm.DB }

func NewWarehouseRepo(db *gorm.DB) *WarehouseRepo { return &WarehouseRepo{db: db} }

func (r *WarehouseRepo) List(ctx context.Context, tenantID uuid.UUID, branchID *uuid.UUID, page, perPage int) ([]domain.Warehouse, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&WarehouseModel{}).Where("tenant_id = ?", tenantID)
	if branchID != nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []WarehouseModel
	if err := q.Order("code").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Warehouse, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Warehouse{ID: m.ID, TenantID: m.TenantID, BranchID: m.BranchID, Code: m.Code, Name: m.Name, Type: m.Type, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, total, nil
}

func (r *WarehouseRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Warehouse, error) {
	var m WarehouseModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.Warehouse{ID: m.ID, TenantID: m.TenantID, BranchID: m.BranchID, Code: m.Code, Name: m.Name, Type: m.Type, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *WarehouseRepo) Create(ctx context.Context, w *domain.Warehouse) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	now := time.Now().UTC()
	w.CreatedAt, w.UpdatedAt = now, now
	w.Code = strings.ToUpper(strings.TrimSpace(w.Code))
	if w.Type == "" {
		w.Type = "main"
	}
	return r.db.WithContext(ctx).Create(&WarehouseModel{ID: w.ID, TenantID: w.TenantID, BranchID: w.BranchID, Code: w.Code, Name: w.Name, Type: w.Type, Status: w.Status, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt}).Error
}

func (r *WarehouseRepo) Update(ctx context.Context, w *domain.Warehouse) error {
	w.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&WarehouseModel{}).Where("id = ? AND tenant_id = ?", w.ID, w.TenantID).Updates(map[string]any{
		"branch_id": w.BranchID, "code": w.Code, "name": w.Name, "type": w.Type, "status": w.Status, "updated_at": w.UpdatedAt,
	}).Error
}

func (r *WarehouseRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&WarehouseModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *WarehouseRepo) ListStocks(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.WarehouseStock, error) {
	var rows []WarehouseStockModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND warehouse_id = ?", tenantID, warehouseID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.WarehouseStock, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.WarehouseStock{ID: m.ID, TenantID: m.TenantID, WarehouseID: m.WarehouseID, ProductID: m.ProductID, QtyOnHand: m.QtyOnHand, QtyReserved: m.QtyReserved, UpdatedAt: m.UpdatedAt})
	}
	return out, nil
}

func (r *WarehouseRepo) UpsertStock(ctx context.Context, s *domain.WarehouseStock) error {
	var existing WarehouseStockModel
	err := r.db.WithContext(ctx).Where("warehouse_id = ? AND product_id = ?", s.WarehouseID, s.ProductID).First(&existing).Error
	now := time.Now().UTC()
	if err == gorm.ErrRecordNotFound {
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		s.UpdatedAt = now
		return r.db.WithContext(ctx).Create(&WarehouseStockModel{ID: s.ID, TenantID: s.TenantID, WarehouseID: s.WarehouseID, ProductID: s.ProductID, QtyOnHand: s.QtyOnHand, QtyReserved: s.QtyReserved, UpdatedAt: s.UpdatedAt}).Error
	}
	if err != nil {
		return err
	}
	s.ID = existing.ID
	s.UpdatedAt = now
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"qty_on_hand": s.QtyOnHand, "qty_reserved": s.QtyReserved, "updated_at": s.UpdatedAt,
	}).Error
}
