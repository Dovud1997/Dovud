package persistence

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/catalog/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ManufacturerRepo struct{ db *gorm.DB }

func NewManufacturerRepo(db *gorm.DB) *ManufacturerRepo { return &ManufacturerRepo{db: db} }

func (r *ManufacturerRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.Manufacturer, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&ManufacturerModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ManufacturerModel
	if err := q.Order("name").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Manufacturer, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Manufacturer{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, total, nil
}

func (r *ManufacturerRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Manufacturer, error) {
	var m ManufacturerModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.Manufacturer{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *ManufacturerRepo) Create(ctx context.Context, m *domain.Manufacturer) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now().UTC()
	m.CreatedAt, m.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&ManufacturerModel{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Status: m.Status, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}).Error
}

func (r *ManufacturerRepo) Update(ctx context.Context, m *domain.Manufacturer) error {
	m.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&ManufacturerModel{}).Where("id = ? AND tenant_id = ?", m.ID, m.TenantID).Updates(map[string]any{
		"code": m.Code, "name": m.Name, "status": m.Status, "updated_at": m.UpdatedAt,
	}).Error
}

func (r *ManufacturerRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&ManufacturerModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Category, error) {
	var rows []CategoryModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("sort_order, name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Category, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Category{ID: m.ID, TenantID: m.TenantID, ParentID: m.ParentID, Code: m.Code, Name: m.Name, SortOrder: m.SortOrder, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, nil
}

func (r *CategoryRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Category, error) {
	var m CategoryModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.Category{ID: m.ID, TenantID: m.TenantID, ParentID: m.ParentID, Code: m.Code, Name: m.Name, SortOrder: m.SortOrder, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt, c.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&CategoryModel{ID: c.ID, TenantID: c.TenantID, ParentID: c.ParentID, Code: c.Code, Name: c.Name, SortOrder: c.SortOrder, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}).Error
}

func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	c.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&CategoryModel{}).Where("id = ? AND tenant_id = ?", c.ID, c.TenantID).Updates(map[string]any{
		"parent_id": c.ParentID, "code": c.Code, "name": c.Name, "sort_order": c.SortOrder, "updated_at": c.UpdatedAt,
	}).Error
}

func (r *CategoryRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&CategoryModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type ProductRepo struct{ db *gorm.DB }

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{db: db} }

func (r *ProductRepo) List(ctx context.Context, tenantID uuid.UUID, qstr string, page, perPage int) ([]domain.Product, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&ProductModel{}).Where("tenant_id = ?", tenantID)
	if qstr = strings.TrimSpace(qstr); qstr != "" {
		like := "%" + strings.ToLower(qstr) + "%"
		q = q.Where("lower(name) LIKE ? OR lower(sku) LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ProductModel
	if err := q.Order("name").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Product, 0, len(rows))
	for _, m := range rows {
		out = append(out, toProduct(m))
	}
	return out, total, nil
}

func toProduct(m ProductModel) domain.Product {
	return domain.Product{ID: m.ID, TenantID: m.TenantID, SKU: m.SKU, Barcode: m.Barcode, Name: m.Name, Description: m.Description, CategoryID: m.CategoryID, ManufacturerID: m.ManufacturerID, Unit: m.Unit, VATRate: m.VATRate, IsActive: m.IsActive, Version: m.Version, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func (r *ProductRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Product, error) {
	var m ProductModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	p := toProduct(m)
	return &p, nil
}

func (r *ProductRepo) Create(ctx context.Context, p *domain.Product) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.CreatedAt, p.UpdatedAt = now, now
	if p.Version == 0 {
		p.Version = 1
	}
	if p.Unit == "" {
		p.Unit = "pcs"
	}
	return r.db.WithContext(ctx).Create(&ProductModel{ID: p.ID, TenantID: p.TenantID, SKU: p.SKU, Barcode: p.Barcode, Name: p.Name, Description: p.Description, CategoryID: p.CategoryID, ManufacturerID: p.ManufacturerID, Unit: p.Unit, VATRate: p.VATRate, IsActive: p.IsActive, Version: p.Version, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}).Error
}

func (r *ProductRepo) Update(ctx context.Context, p *domain.Product) error {
	p.UpdatedAt = time.Now().UTC()
	p.Version++
	return r.db.WithContext(ctx).Model(&ProductModel{}).Where("id = ? AND tenant_id = ?", p.ID, p.TenantID).Updates(map[string]any{
		"sku": p.SKU, "barcode": p.Barcode, "name": p.Name, "description": p.Description, "category_id": p.CategoryID,
		"manufacturer_id": p.ManufacturerID, "unit": p.Unit, "vat_rate": p.VATRate, "is_active": p.IsActive,
		"version": p.Version, "updated_at": p.UpdatedAt,
	}).Error
}

func (r *ProductRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&ProductModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type PriceRepo struct{ db *gorm.DB }

func NewPriceRepo(db *gorm.DB) *PriceRepo { return &PriceRepo{db: db} }

func (r *PriceRepo) ListPriceLists(ctx context.Context, tenantID uuid.UUID) ([]domain.PriceList, error) {
	var rows []PriceListModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.PriceList, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.PriceList{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Currency: m.Currency, IsDefault: m.IsDefault, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, nil
}

func (r *PriceRepo) FindPriceList(ctx context.Context, tenantID, id uuid.UUID) (*domain.PriceList, error) {
	var m PriceListModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.PriceList{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Currency: m.Currency, IsDefault: m.IsDefault, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *PriceRepo) CreatePriceList(ctx context.Context, pl *domain.PriceList) error {
	if pl.ID == uuid.Nil {
		pl.ID = uuid.New()
	}
	now := time.Now().UTC()
	pl.CreatedAt, pl.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&PriceListModel{ID: pl.ID, TenantID: pl.TenantID, Code: pl.Code, Name: pl.Name, Currency: pl.Currency, IsDefault: pl.IsDefault, CreatedAt: pl.CreatedAt, UpdatedAt: pl.UpdatedAt}).Error
}

func (r *PriceRepo) UpdatePriceList(ctx context.Context, pl *domain.PriceList) error {
	pl.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&PriceListModel{}).Where("id = ? AND tenant_id = ?", pl.ID, pl.TenantID).Updates(map[string]any{
		"code": pl.Code, "name": pl.Name, "currency": pl.Currency, "is_default": pl.IsDefault, "updated_at": pl.UpdatedAt,
	}).Error
}

func (r *PriceRepo) SoftDeletePriceList(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&PriceListModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PriceRepo) ListPrices(ctx context.Context, tenantID, priceListID uuid.UUID) ([]domain.ProductPrice, error) {
	var rows []ProductPriceModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND price_list_id = ?", tenantID, priceListID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.ProductPrice, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.ProductPrice{ID: m.ID, TenantID: m.TenantID, PriceListID: m.PriceListID, ProductID: m.ProductID, Amount: m.Amount, Currency: m.Currency, ValidFrom: m.ValidFrom, ValidTo: m.ValidTo, Version: m.Version, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, nil
}

func (r *PriceRepo) UpsertPrice(ctx context.Context, price *domain.ProductPrice) error {
	var existing ProductPriceModel
	err := r.db.WithContext(ctx).Where("price_list_id = ? AND product_id = ?", price.PriceListID, price.ProductID).First(&existing).Error
	now := time.Now().UTC()
	if err == gorm.ErrRecordNotFound {
		if price.ID == uuid.Nil {
			price.ID = uuid.New()
		}
		price.CreatedAt, price.UpdatedAt = now, now
		if price.Version == 0 {
			price.Version = 1
		}
		return r.db.WithContext(ctx).Create(&ProductPriceModel{ID: price.ID, TenantID: price.TenantID, PriceListID: price.PriceListID, ProductID: price.ProductID, Amount: price.Amount, Currency: price.Currency, ValidFrom: price.ValidFrom, ValidTo: price.ValidTo, Version: price.Version, CreatedAt: price.CreatedAt, UpdatedAt: price.UpdatedAt}).Error
	}
	if err != nil {
		return err
	}
	price.ID = existing.ID
	price.Version = existing.Version + 1
	price.UpdatedAt = now
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"amount": price.Amount, "currency": price.Currency, "valid_from": price.ValidFrom, "valid_to": price.ValidTo,
		"version": price.Version, "updated_at": price.UpdatedAt,
	}).Error
}

type PromotionRepo struct{ db *gorm.DB }

func NewPromotionRepo(db *gorm.DB) *PromotionRepo { return &PromotionRepo{db: db} }

func (r *PromotionRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.Promotion, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&PromotionModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []PromotionModel
	if err := q.Order("starts_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Promotion, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Promotion{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Description: m.Description, StartsAt: m.StartsAt, EndsAt: m.EndsAt, DiscountPct: m.DiscountPct, IsActive: m.IsActive, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, total, nil
}

func (r *PromotionRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Promotion, error) {
	var m PromotionModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.Promotion{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, Description: m.Description, StartsAt: m.StartsAt, EndsAt: m.EndsAt, DiscountPct: m.DiscountPct, IsActive: m.IsActive, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *PromotionRepo) Create(ctx context.Context, p *domain.Promotion) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.CreatedAt, p.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&PromotionModel{ID: p.ID, TenantID: p.TenantID, Code: p.Code, Name: p.Name, Description: p.Description, StartsAt: p.StartsAt, EndsAt: p.EndsAt, DiscountPct: p.DiscountPct, IsActive: p.IsActive, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}).Error
}

func (r *PromotionRepo) Update(ctx context.Context, p *domain.Promotion) error {
	p.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&PromotionModel{}).Where("id = ? AND tenant_id = ?", p.ID, p.TenantID).Updates(map[string]any{
		"code": p.Code, "name": p.Name, "description": p.Description, "starts_at": p.StartsAt, "ends_at": p.EndsAt,
		"discount_pct": p.DiscountPct, "is_active": p.IsActive, "updated_at": p.UpdatedAt,
	}).Error
}

func (r *PromotionRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&PromotionModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PromotionRepo) ReplaceItems(ctx context.Context, promotionID uuid.UUID, items []domain.PromotionItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("promotion_id = ?", promotionID).Delete(&PromotionItemModel{}).Error; err != nil {
			return err
		}
		for _, item := range items {
			if item.ID == uuid.Nil {
				item.ID = uuid.New()
			}
			if err := tx.Create(&PromotionItemModel{ID: item.ID, PromotionID: promotionID, ProductID: item.ProductID, CategoryID: item.CategoryID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PromotionRepo) ListItems(ctx context.Context, promotionID uuid.UUID) ([]domain.PromotionItem, error) {
	var rows []PromotionItemModel
	if err := r.db.WithContext(ctx).Where("promotion_id = ?", promotionID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.PromotionItem, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.PromotionItem{ID: m.ID, PromotionID: m.PromotionID, ProductID: m.ProductID, CategoryID: m.CategoryID})
	}
	return out, nil
}
