package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/portal/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerUserRepo struct{ db *gorm.DB }

func NewCustomerUserRepo(db *gorm.DB) *CustomerUserRepo { return &CustomerUserRepo{db: db} }

func toLink(m CustomerUserModel) domain.CustomerUser {
	return domain.CustomerUser{
		ID: m.ID, TenantID: m.TenantID, UserID: m.UserID, CustomerID: m.CustomerID, CreatedAt: m.CreatedAt,
	}
}

func (r *CustomerUserRepo) FindByUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.CustomerUser, error) {
	var m CustomerUserModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	link := toLink(m)
	return &link, nil
}

func (r *CustomerUserRepo) Upsert(ctx context.Context, link *domain.CustomerUser) error {
	var existing CustomerUserModel
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", link.TenantID, link.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if link.ID == uuid.Nil {
			link.ID = uuid.New()
		}
		if link.CreatedAt.IsZero() {
			link.CreatedAt = time.Now().UTC()
		}
		return r.db.WithContext(ctx).Create(&CustomerUserModel{
			ID: link.ID, TenantID: link.TenantID, UserID: link.UserID, CustomerID: link.CustomerID, CreatedAt: link.CreatedAt,
		}).Error
	}
	if err != nil {
		return err
	}
	link.ID = existing.ID
	link.CreatedAt = existing.CreatedAt
	return r.db.WithContext(ctx).Model(&CustomerUserModel{}).Where("id = ?", existing.ID).
		Update("customer_id", link.CustomerID).Error
}

func (r *CustomerUserRepo) DeleteByUser(ctx context.Context, tenantID, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Delete(&CustomerUserModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *CustomerUserRepo) List(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]domain.CustomerUser, error) {
	q := r.db.WithContext(ctx).Model(&CustomerUserModel{}).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	var rows []CustomerUserModel
	if err := q.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.CustomerUser, 0, len(rows))
	for _, m := range rows {
		out = append(out, toLink(m))
	}
	return out, nil
}
