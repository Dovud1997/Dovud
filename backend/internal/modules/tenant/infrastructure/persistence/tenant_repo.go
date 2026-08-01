package persistence

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/tenant/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantRepo struct{ db *gorm.DB }

func NewTenantRepo(db *gorm.DB) *TenantRepo { return &TenantRepo{db: db} }

func toTenant(m TenantModel) *domain.Tenant {
	return &domain.Tenant{
		ID: m.ID, Code: m.Code, Name: m.Name, Status: m.Status,
		DefaultLocale: m.DefaultLocale, DefaultCurrency: m.DefaultCurrency,
		Timezone: m.Timezone, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *TenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var m TenantModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}
	return toTenant(m), nil
}

func (r *TenantRepo) FindByCode(ctx context.Context, code string) (*domain.Tenant, error) {
	var m TenantModel
	err := r.db.WithContext(ctx).Where("lower(code) = ?", strings.ToLower(code)).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}
	return toTenant(m), nil
}

func (r *TenantRepo) FindByHost(ctx context.Context, host string) (*domain.Tenant, error) {
	var domainRow DomainModel
	err := r.db.WithContext(ctx).Where("lower(host) = ?", strings.ToLower(host)).First(&domainRow).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, domainRow.TenantID)
}

func (r *TenantRepo) Update(ctx context.Context, tenant *domain.Tenant) error {
	tenant.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Model(&TenantModel{}).Where("id = ?", tenant.ID).Updates(map[string]any{
		"name": tenant.Name, "status": tenant.Status, "default_locale": tenant.DefaultLocale,
		"default_currency": tenant.DefaultCurrency, "timezone": tenant.Timezone, "updated_at": tenant.UpdatedAt,
	}).Error
}

func (r *TenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	now := time.Now().UTC()
	tenant.CreatedAt, tenant.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&TenantModel{
		ID: tenant.ID, Code: tenant.Code, Name: tenant.Name, Status: tenant.Status,
		DefaultLocale: tenant.DefaultLocale, DefaultCurrency: tenant.DefaultCurrency,
		Timezone: tenant.Timezone, CreatedAt: tenant.CreatedAt, UpdatedAt: tenant.UpdatedAt,
	}).Error
}

type BrandingRepo struct{ db *gorm.DB }

func NewBrandingRepo(db *gorm.DB) *BrandingRepo { return &BrandingRepo{db: db} }

func toBranding(m BrandingModel) *domain.Branding {
	return &domain.Branding{
		ID: m.ID, TenantID: m.TenantID, AppName: m.AppName,
		LogoURL: m.LogoURL, FaviconURL: m.FaviconURL, IconURL: m.IconURL,
		PrimaryColor: m.PrimaryColor, SecondaryColor: m.SecondaryColor, AccentColor: m.AccentColor,
		ThemeModeDefault: m.ThemeModeDefault, BrandingVersion: m.BrandingVersion,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *BrandingRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*domain.Branding, error) {
	var m BrandingModel
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toBranding(m), nil
}

func (r *BrandingRepo) Upsert(ctx context.Context, branding *domain.Branding) error {
	var existing BrandingModel
	err := r.db.WithContext(ctx).Where("tenant_id = ?", branding.TenantID).First(&existing).Error
	now := time.Now().UTC()
	if err == gorm.ErrRecordNotFound {
		if branding.ID == uuid.Nil {
			branding.ID = uuid.New()
		}
		branding.CreatedAt, branding.UpdatedAt = now, now
		branding.BrandingVersion = 1
		return r.db.WithContext(ctx).Create(&BrandingModel{
			ID: branding.ID, TenantID: branding.TenantID, AppName: branding.AppName,
			LogoURL: branding.LogoURL, FaviconURL: branding.FaviconURL, IconURL: branding.IconURL,
			PrimaryColor: branding.PrimaryColor, SecondaryColor: branding.SecondaryColor,
			AccentColor: branding.AccentColor, ThemeModeDefault: branding.ThemeModeDefault,
			BrandingVersion: branding.BrandingVersion, CreatedAt: branding.CreatedAt, UpdatedAt: branding.UpdatedAt,
		}).Error
	}
	if err != nil {
		return err
	}
	branding.ID = existing.ID
	branding.BrandingVersion = existing.BrandingVersion + 1
	branding.UpdatedAt = now
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"app_name": branding.AppName, "logo_url": branding.LogoURL, "favicon_url": branding.FaviconURL,
		"icon_url": branding.IconURL, "primary_color": branding.PrimaryColor,
		"secondary_color": branding.SecondaryColor, "accent_color": branding.AccentColor,
		"theme_mode_default": branding.ThemeModeDefault, "branding_version": branding.BrandingVersion,
		"updated_at": branding.UpdatedAt,
	}).Error
}

func (r *BrandingRepo) GetPublicByCode(ctx context.Context, code string) (*domain.PublicBranding, error) {
	var row domain.PublicBranding
	err := r.db.WithContext(ctx).Table("tenants t").
		Select(`t.code as tenant_code, t.name as tenant_name, t.default_locale,
			b.app_name, b.logo_url, b.favicon_url, b.icon_url,
			b.primary_color, b.secondary_color, b.accent_color,
			b.theme_mode_default, b.branding_version`).
		Joins("JOIN tenant_branding b ON b.tenant_id = t.id").
		Where("lower(t.code) = ? AND t.deleted_at IS NULL", strings.ToLower(code)).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.TenantCode == "" {
		return nil, apperrors.ErrTenantNotFound
	}
	return &row, nil
}

func (r *BrandingRepo) GetPublicByHost(ctx context.Context, host string) (*domain.PublicBranding, error) {
	var row domain.PublicBranding
	err := r.db.WithContext(ctx).Table("tenant_domains d").
		Select(`t.code as tenant_code, t.name as tenant_name, t.default_locale,
			b.app_name, b.logo_url, b.favicon_url, b.icon_url,
			b.primary_color, b.secondary_color, b.accent_color,
			b.theme_mode_default, b.branding_version`).
		Joins("JOIN tenants t ON t.id = d.tenant_id").
		Joins("JOIN tenant_branding b ON b.tenant_id = t.id").
		Where("lower(d.host) = ? AND d.deleted_at IS NULL AND t.deleted_at IS NULL", strings.ToLower(host)).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.TenantCode == "" {
		return nil, apperrors.ErrTenantNotFound
	}
	return &row, nil
}

type DomainRepo struct{ db *gorm.DB }

func NewDomainRepo(db *gorm.DB) *DomainRepo { return &DomainRepo{db: db} }

func (r *DomainRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Domain, error) {
	var rows []DomainModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Domain, 0, len(rows))
	for _, m := range rows {
		out = append(out, domain.Domain{
			ID: m.ID, TenantID: m.TenantID, Host: m.Host, IsPrimary: m.IsPrimary,
			CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		})
	}
	return out, nil
}

func (r *DomainRepo) Create(ctx context.Context, d *domain.Domain) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	now := time.Now().UTC()
	d.CreatedAt, d.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(&DomainModel{
		ID: d.ID, TenantID: d.TenantID, Host: strings.ToLower(d.Host), IsPrimary: d.IsPrimary,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}).Error
}

func (r *DomainRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&DomainModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
