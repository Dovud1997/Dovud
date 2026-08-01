package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerRepo struct{ db *gorm.DB }

func NewCustomerRepo(db *gorm.DB) *CustomerRepo { return &CustomerRepo{db: db} }

func toCustomer(m CustomerModel) domain.Customer {
	return domain.Customer{
		ID: m.ID, TenantID: m.TenantID, BranchID: m.BranchID, Code: m.Code, Name: m.Name,
		Type: m.Type, Inn: m.Inn, Status: m.Status, CreditLimit: m.CreditLimit, BalanceCached: m.BalanceCached,
		Lat: m.Lat, Lng: m.Lng, Address: m.Address, Version: m.Version, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *CustomerRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]domain.Customer, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&CustomerModel{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []CustomerModel
	if err := q.Order("name").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Customer, 0, len(rows))
	for _, m := range rows {
		out = append(out, toCustomer(m))
	}
	return out, total, nil
}

func (r *CustomerRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Customer, error) {
	var m CustomerModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	c := toCustomer(m)
	return &c, nil
}

func (r *CustomerRepo) Create(ctx context.Context, c *domain.Customer) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt, c.UpdatedAt = now, now
	if c.Version == 0 {
		c.Version = 1
	}
	return r.db.WithContext(ctx).Create(&CustomerModel{
		ID: c.ID, TenantID: c.TenantID, BranchID: c.BranchID, Code: c.Code, Name: c.Name,
		Type: c.Type, Inn: c.Inn, Status: c.Status, CreditLimit: c.CreditLimit, BalanceCached: c.BalanceCached,
		Lat: c.Lat, Lng: c.Lng, Address: c.Address, Version: c.Version, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}).Error
}

func (r *CustomerRepo) Update(ctx context.Context, c *domain.Customer) error {
	c.UpdatedAt = time.Now().UTC()
	c.Version++
	return r.db.WithContext(ctx).Model(&CustomerModel{}).Where("id = ? AND tenant_id = ?", c.ID, c.TenantID).Updates(map[string]any{
		"branch_id": c.BranchID, "code": c.Code, "name": c.Name, "type": c.Type, "inn": c.Inn,
		"status": c.Status, "credit_limit": c.CreditLimit, "balance_cached": c.BalanceCached,
		"lat": c.Lat, "lng": c.Lng, "address": c.Address, "version": c.Version, "updated_at": c.UpdatedAt,
	}).Error
}

func (r *CustomerRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&CustomerModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type CustomerContactRepo struct{ db *gorm.DB }

func NewCustomerContactRepo(db *gorm.DB) *CustomerContactRepo { return &CustomerContactRepo{db: db} }

func toContact(m CustomerContactModel) domain.CustomerContact {
	return domain.CustomerContact{
		ID: m.ID, CustomerID: m.CustomerID, FullName: m.FullName, Phone: m.Phone,
		Email: m.Email, Position: m.Position, IsPrimary: m.IsPrimary, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *CustomerContactRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.CustomerContact, error) {
	var rows []CustomerContactModel
	if err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("is_primary DESC, full_name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.CustomerContact, 0, len(rows))
	for _, m := range rows {
		out = append(out, toContact(m))
	}
	return out, nil
}

func (r *CustomerContactRepo) FindByID(ctx context.Context, customerID, id uuid.UUID) (*domain.CustomerContact, error) {
	var m CustomerContactModel
	if err := r.db.WithContext(ctx).Where("customer_id = ? AND id = ?", customerID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	c := toContact(m)
	return &c, nil
}

func (r *CustomerContactRepo) Create(ctx context.Context, c *domain.CustomerContact) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt, c.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&CustomerContactModel{
		ID: c.ID, CustomerID: c.CustomerID, FullName: c.FullName, Phone: c.Phone,
		Email: c.Email, Position: c.Position, IsPrimary: c.IsPrimary, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}).Error
}

func (r *CustomerContactRepo) Update(ctx context.Context, c *domain.CustomerContact) error {
	c.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&CustomerContactModel{}).Where("id = ? AND customer_id = ?", c.ID, c.CustomerID).Updates(map[string]any{
		"full_name": c.FullName, "phone": c.Phone, "email": c.Email, "position": c.Position,
		"is_primary": c.IsPrimary, "updated_at": c.UpdatedAt,
	}).Error
}

func (r *CustomerContactRepo) SoftDelete(ctx context.Context, customerID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND customer_id = ?", id, customerID).Delete(&CustomerContactModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type CustomerAddressRepo struct{ db *gorm.DB }

func NewCustomerAddressRepo(db *gorm.DB) *CustomerAddressRepo { return &CustomerAddressRepo{db: db} }

func toAddress(m CustomerAddressModel) domain.CustomerAddress {
	return domain.CustomerAddress{
		ID: m.ID, CustomerID: m.CustomerID, Label: m.Label, Address: m.Address,
		Lat: m.Lat, Lng: m.Lng, IsPrimary: m.IsPrimary, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *CustomerAddressRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.CustomerAddress, error) {
	var rows []CustomerAddressModel
	if err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("is_primary DESC, label").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.CustomerAddress, 0, len(rows))
	for _, m := range rows {
		out = append(out, toAddress(m))
	}
	return out, nil
}

func (r *CustomerAddressRepo) FindByID(ctx context.Context, customerID, id uuid.UUID) (*domain.CustomerAddress, error) {
	var m CustomerAddressModel
	if err := r.db.WithContext(ctx).Where("customer_id = ? AND id = ?", customerID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	a := toAddress(m)
	return &a, nil
}

func (r *CustomerAddressRepo) Create(ctx context.Context, a *domain.CustomerAddress) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	now := time.Now().UTC()
	a.CreatedAt, a.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&CustomerAddressModel{
		ID: a.ID, CustomerID: a.CustomerID, Label: a.Label, Address: a.Address,
		Lat: a.Lat, Lng: a.Lng, IsPrimary: a.IsPrimary, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}).Error
}

func (r *CustomerAddressRepo) Update(ctx context.Context, a *domain.CustomerAddress) error {
	a.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&CustomerAddressModel{}).Where("id = ? AND customer_id = ?", a.ID, a.CustomerID).Updates(map[string]any{
		"label": a.Label, "address": a.Address, "lat": a.Lat, "lng": a.Lng,
		"is_primary": a.IsPrimary, "updated_at": a.UpdatedAt,
	}).Error
}

func (r *CustomerAddressRepo) SoftDelete(ctx context.Context, customerID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND customer_id = ?", id, customerID).Delete(&CustomerAddressModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type CustomerCategoryRepo struct{ db *gorm.DB }

func NewCustomerCategoryRepo(db *gorm.DB) *CustomerCategoryRepo { return &CustomerCategoryRepo{db: db} }

func (r *CustomerCategoryRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.CustomerCategory, error) {
	var rows []CustomerCategoryModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.CustomerCategory, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.CustomerCategory{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt})
	}
	return out, nil
}

func (r *CustomerCategoryRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.CustomerCategory, error) {
	var m CustomerCategoryModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.CustomerCategory{ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}, nil
}

func (r *CustomerCategoryRepo) Create(ctx context.Context, c *domain.CustomerCategory) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt, c.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&CustomerCategoryModel{
		ID: c.ID, TenantID: c.TenantID, Code: c.Code, Name: c.Name, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}).Error
}

func (r *CustomerCategoryRepo) Update(ctx context.Context, c *domain.CustomerCategory) error {
	c.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&CustomerCategoryModel{}).Where("id = ? AND tenant_id = ?", c.ID, c.TenantID).Updates(map[string]any{
		"code": c.Code, "name": c.Name, "updated_at": c.UpdatedAt,
	}).Error
}

func (r *CustomerCategoryRepo) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&CustomerCategoryModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
