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

func (r *CustomerUserRepo) FindByUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.CustomerUser, error) {
	var m CustomerUserModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &domain.CustomerUser{
		ID: m.ID, TenantID: m.TenantID, UserID: m.UserID, CustomerID: m.CustomerID, CreatedAt: m.CreatedAt,
	}, nil
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
	return r.db.WithContext(ctx).Model(&CustomerUserModel{}).Where("id = ?", existing.ID).
		Update("customer_id", link.CustomerID).Error
}
