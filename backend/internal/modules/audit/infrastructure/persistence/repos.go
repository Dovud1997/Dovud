package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/audit/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepo struct{ db *gorm.DB }

func NewAuditRepo(db *gorm.DB) *AuditRepo { return &AuditRepo{db: db} }

func toDomain(m AuditLogModel) domain.AuditLog {
	return domain.AuditLog{
		ID: m.ID, TenantID: m.TenantID, ActorUserID: m.ActorUserID, Action: m.Action,
		EntityType: m.EntityType, EntityID: m.EntityID, BeforeJSON: m.BeforeJSON, AfterJSON: m.AfterJSON,
		IP: m.IP, UserAgent: m.UserAgent, RequestID: m.RequestID, CreatedAt: m.CreatedAt,
	}
}

func (r *AuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(&AuditLogModel{
		ID: log.ID, TenantID: log.TenantID, ActorUserID: log.ActorUserID, Action: log.Action,
		EntityType: log.EntityType, EntityID: log.EntityID, BeforeJSON: log.BeforeJSON, AfterJSON: log.AfterJSON,
		IP: log.IP, UserAgent: log.UserAgent, RequestID: log.RequestID, CreatedAt: log.CreatedAt,
	}).Error
}

func (r *AuditRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.AuditLog, error) {
	var m AuditLogModel
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	row := toDomain(m)
	return &row, nil
}

func (r *AuditRepo) List(ctx context.Context, tenantID uuid.UUID, filters domain.ListFilters, page, perPage int) ([]domain.AuditLog, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	q := r.db.WithContext(ctx).Model(&AuditLogModel{}).Where("tenant_id = ?", tenantID)
	if filters.ActorUserID != nil {
		q = q.Where("actor_user_id = ?", *filters.ActorUserID)
	}
	if filters.EntityType != "" {
		q = q.Where("entity_type = ?", filters.EntityType)
	}
	if filters.EntityID != nil {
		q = q.Where("entity_id = ?", *filters.EntityID)
	}
	if filters.Action != "" {
		q = q.Where("action = ?", filters.Action)
	}
	if filters.From != nil {
		q = q.Where("created_at >= ?", *filters.From)
	}
	if filters.To != nil {
		q = q.Where("created_at <= ?", *filters.To)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []AuditLogModel
	if err := q.Order("created_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.AuditLog, 0, len(rows))
	for _, m := range rows {
		out = append(out, toDomain(m))
	}
	return out, total, nil
}
