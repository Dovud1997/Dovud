package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/notifications/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/shared/paging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepo struct{ db *gorm.DB }

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func toNotification(m NotificationModel) domain.Notification {
	return domain.Notification{
		ID: m.ID, TenantID: m.TenantID, UserID: m.UserID, Type: m.Type,
		Title: m.Title, Body: m.Body, PayloadJSON: m.PayloadJSON, ReadAt: m.ReadAt,
		Channel: m.Channel, CreatedAt: m.CreatedAt,
	}
}

func toDelivery(m NotificationDeliveryModel) domain.NotificationDelivery {
	return domain.NotificationDelivery{
		ID: m.ID, NotificationID: m.NotificationID, Channel: m.Channel,
		Status: m.Status, Error: m.Error, DeviceID: m.DeviceID, Platform: m.Platform,
		TokenSuffix: m.TokenSuffix, AttemptedAt: m.AttemptedAt,
	}
}

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification, delivery *domain.NotificationDelivery) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	now := time.Now().UTC()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&NotificationModel{
			ID: n.ID, TenantID: n.TenantID, UserID: n.UserID, Type: n.Type,
			Title: n.Title, Body: n.Body, PayloadJSON: n.PayloadJSON, ReadAt: n.ReadAt,
			Channel: n.Channel, CreatedAt: n.CreatedAt,
		}).Error; err != nil {
			return err
		}
		if delivery == nil {
			return nil
		}
		if delivery.ID == uuid.Nil {
			delivery.ID = uuid.New()
		}
		delivery.NotificationID = n.ID
		if delivery.AttemptedAt.IsZero() {
			delivery.AttemptedAt = now
		}
		return tx.Create(&NotificationDeliveryModel{
			ID: delivery.ID, NotificationID: delivery.NotificationID, Channel: delivery.Channel,
			Status: delivery.Status, Error: delivery.Error, DeviceID: delivery.DeviceID,
			Platform: delivery.Platform, TokenSuffix: delivery.TokenSuffix, AttemptedAt: delivery.AttemptedAt,
		}).Error
	})
}

func (r *NotificationRepo) ListByUser(ctx context.Context, tenantID, userID uuid.UUID, filters domain.ListFilters, page, perPage int) ([]domain.Notification, int64, error) {
	page, perPage = paging.Normalize(page, perPage)
	var total int64
	q := r.db.WithContext(ctx).Model(&NotificationModel{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID)
	if filters.UnreadOnly {
		q = q.Where("read_at IS NULL")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []NotificationModel
	if err := q.Order("created_at DESC").Offset(paging.Offset(page, perPage)).Limit(perPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Notification, 0, len(rows))
	for _, m := range rows {
		out = append(out, toNotification(m))
	}
	return out, total, nil
}

func (r *NotificationRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Notification, error) {
	var m NotificationModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	n := toNotification(m)
	return &n, nil
}

func (r *NotificationRepo) MarkRead(ctx context.Context, tenantID, userID, id uuid.UUID) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&NotificationModel{}).
		Where("tenant_id = ? AND user_id = ? AND id = ? AND read_at IS NULL", tenantID, userID, id).
		Update("read_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&NotificationModel{}).
			Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, id).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return apperrors.ErrNotFound
		}
	}
	return nil
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&NotificationModel{}).
		Where("tenant_id = ? AND user_id = ? AND read_at IS NULL", tenantID, userID).
		Update("read_at", now)
	return res.RowsAffected, res.Error
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&NotificationModel{}).
		Where("tenant_id = ? AND user_id = ? AND read_at IS NULL", tenantID, userID).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepo) ListDeliveries(ctx context.Context, notificationID uuid.UUID) ([]domain.NotificationDelivery, error) {
	var rows []NotificationDeliveryModel
	if err := r.db.WithContext(ctx).Where("notification_id = ?", notificationID).
		Order("attempted_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.NotificationDelivery, 0, len(rows))
	for _, m := range rows {
		out = append(out, toDelivery(m))
	}
	return out, nil
}

func (r *NotificationRepo) UpdateDeliveryStatus(ctx context.Context, notificationID uuid.UUID, channel, status string, errMsg *string) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&NotificationDeliveryModel{}).
		Where("notification_id = ? AND channel = ? AND device_id IS NULL", notificationID, channel).
		Updates(map[string]any{
			"status": status, "error": errMsg, "attempted_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return r.db.WithContext(ctx).Create(&NotificationDeliveryModel{
			ID: uuid.New(), NotificationID: notificationID, Channel: channel,
			Status: status, Error: errMsg, AttemptedAt: now,
		}).Error
	}
	return nil
}

func (r *NotificationRepo) UpsertDeviceDelivery(ctx context.Context, delivery *domain.NotificationDelivery) error {
	if delivery == nil {
		return apperrors.ErrValidation
	}
	now := time.Now().UTC()
	if delivery.AttemptedAt.IsZero() {
		delivery.AttemptedAt = now
	}
	q := r.db.WithContext(ctx).Model(&NotificationDeliveryModel{}).
		Where("notification_id = ? AND channel = ?", delivery.NotificationID, delivery.Channel)
	if delivery.DeviceID == nil || *delivery.DeviceID == "" {
		q = q.Where("device_id IS NULL")
	} else {
		q = q.Where("device_id = ?", *delivery.DeviceID)
	}
	res := q.Updates(map[string]any{
		"status": delivery.Status, "error": delivery.Error, "attempted_at": delivery.AttemptedAt,
		"platform": delivery.Platform, "token_suffix": delivery.TokenSuffix,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}
	if delivery.ID == uuid.Nil {
		delivery.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(&NotificationDeliveryModel{
		ID: delivery.ID, NotificationID: delivery.NotificationID, Channel: delivery.Channel,
		Status: delivery.Status, Error: delivery.Error, DeviceID: delivery.DeviceID,
		Platform: delivery.Platform, TokenSuffix: delivery.TokenSuffix, AttemptedAt: delivery.AttemptedAt,
	}).Error
}
