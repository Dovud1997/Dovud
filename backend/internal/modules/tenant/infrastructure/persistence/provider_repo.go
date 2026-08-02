package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/tenant/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProviderRepo struct{ db *gorm.DB }

func NewProviderRepo(db *gorm.DB) *ProviderRepo { return &ProviderRepo{db: db} }

func toProvider(m ProviderModel) domain.TenantProvider {
	return domain.TenantProvider{
		ID: m.ID, TenantID: m.TenantID, Type: m.Type, Driver: m.Driver,
		ConfigEnc: m.ConfigEnc, IsEnabled: m.IsEnabled,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *ProviderRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantProvider, error) {
	var rows []ProviderModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("type ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.TenantProvider, 0, len(rows))
	for _, m := range rows {
		out = append(out, toProvider(m))
	}
	return out, nil
}

func (r *ProviderRepo) FindByType(ctx context.Context, tenantID uuid.UUID, providerType string) (*domain.TenantProvider, error) {
	var m ProviderModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND type = ?", tenantID, providerType).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	p := toProvider(m)
	return &p, nil
}

func (r *ProviderRepo) Upsert(ctx context.Context, p *domain.TenantProvider) error {
	var existing ProviderModel
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND type = ?", p.TenantID, p.Type).First(&existing).Error
	now := time.Now().UTC()
	if err == gorm.ErrRecordNotFound {
		if p.ID == uuid.Nil {
			p.ID = uuid.New()
		}
		p.CreatedAt, p.UpdatedAt = now, now
		return r.db.WithContext(ctx).Create(&ProviderModel{
			ID: p.ID, TenantID: p.TenantID, Type: p.Type, Driver: p.Driver,
			ConfigEnc: p.ConfigEnc, IsEnabled: p.IsEnabled,
			CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
		}).Error
	}
	if err != nil {
		return err
	}
	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt
	p.UpdatedAt = now
	return r.db.WithContext(ctx).Model(&ProviderModel{}).Where("id = ?", existing.ID).Updates(map[string]any{
		"driver": p.Driver, "config_enc": p.ConfigEnc, "is_enabled": p.IsEnabled, "updated_at": p.UpdatedAt,
	}).Error
}
