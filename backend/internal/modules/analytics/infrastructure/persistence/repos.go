package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/analytics/domain"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KpiRepo struct{ db *gorm.DB }

func NewKpiRepo(db *gorm.DB) *KpiRepo { return &KpiRepo{db: db} }

func toDefinition(m KpiDefinitionModel) domain.KpiDefinition {
	return domain.KpiDefinition{
		ID: m.ID, TenantID: m.TenantID, Code: m.Code, Name: m.Name,
		Description: m.Description, Unit: m.Unit,
	}
}

func toSnapshot(m KpiSnapshotModel) domain.KpiSnapshot {
	return domain.KpiSnapshot{
		ID: m.ID, TenantID: m.TenantID, KpiCode: m.KpiCode, Period: m.Period,
		PeriodStart: m.PeriodStart, ScopeType: m.ScopeType, ScopeID: m.ScopeID,
		Value: m.Value, CreatedAt: m.CreatedAt,
	}
}

func (r *KpiRepo) ListDefinitions(ctx context.Context, tenantID uuid.UUID) ([]domain.KpiDefinition, error) {
	var rows []KpiDefinitionModel
	err := r.db.WithContext(ctx).
		Where("tenant_id IS NULL OR tenant_id = ?", tenantID).
		Order("code ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.KpiDefinition, 0, len(rows))
	for _, m := range rows {
		out = append(out, toDefinition(m))
	}
	return out, nil
}

func (r *KpiRepo) CreateDefinition(ctx context.Context, d *domain.KpiDefinition) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(&KpiDefinitionModel{
		ID: d.ID, TenantID: d.TenantID, Code: d.Code, Name: d.Name,
		Description: d.Description, Unit: d.Unit,
	}).Error
}

func (r *KpiRepo) ListSnapshots(ctx context.Context, tenantID uuid.UUID, filters domain.SnapshotFilters, page, perPage int) ([]domain.KpiSnapshot, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&KpiSnapshotModel{}).Where("tenant_id = ?", tenantID)
	if filters.Code != "" {
		q = q.Where("kpi_code = ?", filters.Code)
	}
	if filters.Period != "" {
		q = q.Where("period = ?", filters.Period)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []KpiSnapshotModel
	if err := q.Order("period_start DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.KpiSnapshot, 0, len(rows))
	for _, m := range rows {
		out = append(out, toSnapshot(m))
	}
	return out, total, nil
}

func (r *KpiRepo) CreateSnapshot(ctx context.Context, s *domain.KpiSnapshot) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(&KpiSnapshotModel{
		ID: s.ID, TenantID: s.TenantID, KpiCode: s.KpiCode, Period: s.Period,
		PeriodStart: s.PeriodStart, ScopeType: s.ScopeType, ScopeID: s.ScopeID,
		Value: s.Value, CreatedAt: s.CreatedAt,
	}).Error
}
