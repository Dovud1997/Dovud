package persistence

import (
	"context"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepo struct{ db *gorm.DB }

func NewRefreshTokenRepo(db *gorm.DB) *RefreshTokenRepo { return &RefreshTokenRepo{db: db} }

func (r *RefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	token.CreatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Create(&RefreshTokenModel{
		ID: token.ID, UserID: token.UserID, TenantID: token.TenantID,
		TokenHash: token.TokenHash, DeviceID: token.DeviceID,
		ExpiresAt: token.ExpiresAt, CreatedAt: token.CreatedAt,
	}).Error
}

func (r *RefreshTokenRepo) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var m RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.RefreshToken{
		ID: m.ID, UserID: m.UserID, TenantID: m.TenantID, TokenHash: m.TokenHash,
		DeviceID: m.DeviceID, ExpiresAt: m.ExpiresAt, RevokedAt: m.RevokedAt, CreatedAt: m.CreatedAt,
	}, nil
}

func (r *RefreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("id = ?", id).Update("revoked_at", now).Error
}

func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

type DeviceRepo struct{ db *gorm.DB }

func NewDeviceRepo(db *gorm.DB) *DeviceRepo { return &DeviceRepo{db: db} }

func (r *DeviceRepo) Upsert(ctx context.Context, device *domain.UserDevice) error {
	var existing UserDeviceModel
	err := r.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", device.UserID, device.DeviceID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if device.ID == uuid.Nil {
			device.ID = uuid.New()
		}
		return r.db.WithContext(ctx).Create(&UserDeviceModel{
			ID: device.ID, TenantID: device.TenantID, UserID: device.UserID,
			DeviceID: device.DeviceID, Platform: device.Platform,
			PushToken: device.PushToken, AppVersion: device.AppVersion,
		}).Error
	}
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"platform": device.Platform, "push_token": device.PushToken, "app_version": device.AppVersion,
	}).Error
}

func (r *DeviceRepo) Delete(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&UserDeviceModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
