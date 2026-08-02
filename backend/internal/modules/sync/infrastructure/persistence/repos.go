package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/sync/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeviceRepo struct{ db *gorm.DB }

func NewDeviceRepo(db *gorm.DB) *DeviceRepo { return &DeviceRepo{db: db} }

func toDevice(m SyncDeviceModel) domain.SyncDevice {
	return domain.SyncDevice{
		ID: m.ID, TenantID: m.TenantID, UserID: m.UserID, DeviceID: m.DeviceID,
		Platform: m.Platform, AppVersion: m.AppVersion, LastPullCursor: m.LastPullCursor,
		LastPushAt: m.LastPushAt, LastPullAt: m.LastPullAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *DeviceRepo) Upsert(ctx context.Context, device *domain.SyncDevice) error {
	now := time.Now().UTC()
	existing, err := r.Find(ctx, device.TenantID, device.UserID, device.DeviceID)
	if err == nil {
		device.ID = existing.ID
		device.LastPullCursor = existing.LastPullCursor
		device.LastPullAt = existing.LastPullAt
		device.LastPushAt = existing.LastPushAt
		device.CreatedAt = existing.CreatedAt
		device.UpdatedAt = now
		return r.db.WithContext(ctx).Model(&SyncDeviceModel{}).
			Where("id = ?", device.ID).
			Updates(map[string]any{
				"platform":    device.Platform,
				"app_version": device.AppVersion,
				"updated_at":  device.UpdatedAt,
			}).Error
	}
	if err != apperrors.ErrNotFound {
		return err
	}
	if device.ID == uuid.Nil {
		device.ID = uuid.New()
	}
	device.CreatedAt = now
	device.UpdatedAt = now
	return r.db.WithContext(ctx).Create(&SyncDeviceModel{
		ID: device.ID, TenantID: device.TenantID, UserID: device.UserID, DeviceID: device.DeviceID,
		Platform: device.Platform, AppVersion: device.AppVersion, LastPullCursor: device.LastPullCursor,
		LastPushAt: device.LastPushAt, LastPullAt: device.LastPullAt, CreatedAt: device.CreatedAt, UpdatedAt: device.UpdatedAt,
	}).Error
}

func (r *DeviceRepo) Find(ctx context.Context, tenantID, userID uuid.UUID, deviceID string) (*domain.SyncDevice, error) {
	var m SyncDeviceModel
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND device_id = ?", tenantID, userID, deviceID).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	d := toDevice(m)
	return &d, nil
}

func (r *DeviceRepo) UpdateCursors(ctx context.Context, tenantID, userID uuid.UUID, deviceID string, pullCursor string, pullAt, pushAt *time.Time) error {
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if pullCursor != "" {
		updates["last_pull_cursor"] = pullCursor
	}
	if pullAt != nil {
		updates["last_pull_at"] = *pullAt
	}
	if pushAt != nil {
		updates["last_push_at"] = *pushAt
	}
	res := r.db.WithContext(ctx).Model(&SyncDeviceModel{}).
		Where("tenant_id = ? AND user_id = ? AND device_id = ?", tenantID, userID, deviceID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

type ChangeLogRepo struct{ db *gorm.DB }

func NewChangeLogRepo(db *gorm.DB) *ChangeLogRepo { return &ChangeLogRepo{db: db} }

func toChange(m SyncChangeLogModel) domain.SyncChange {
	return domain.SyncChange{
		ID: m.ID, TenantID: m.TenantID, EntityType: m.EntityType, EntityID: m.EntityID,
		Version: m.Version, Deleted: m.Deleted, UpdatedAt: m.UpdatedAt, PayloadJSON: m.PayloadJSON,
	}
}

func EncodeCursor(t time.Time, id uuid.UUID) string {
	return t.UTC().Format(time.RFC3339Nano) + "|" + id.String()
}

func DecodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	if cursor == "" {
		return time.Time{}, uuid.Nil, nil
	}
	parts := strings.SplitN(cursor, "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor time: %w", err)
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor id: %w", err)
	}
	return ts, id, nil
}

func (r *ChangeLogRepo) Append(ctx context.Context, change *domain.SyncChange) error {
	if change.ID == uuid.Nil {
		change.ID = uuid.New()
	}
	if change.UpdatedAt.IsZero() {
		change.UpdatedAt = time.Now().UTC()
	}
	if change.PayloadJSON == "" {
		change.PayloadJSON = "{}"
	}
	return r.db.WithContext(ctx).Create(&SyncChangeLogModel{
		ID: change.ID, TenantID: change.TenantID, EntityType: change.EntityType, EntityID: change.EntityID,
		Version: change.Version, Deleted: change.Deleted, UpdatedAt: change.UpdatedAt, PayloadJSON: change.PayloadJSON,
	}).Error
}

func (r *ChangeLogRepo) ListSince(ctx context.Context, tenantID uuid.UUID, cursor string, types []string, limit int) ([]domain.SyncChange, string, bool, error) {
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	ts, id, err := DecodeCursor(cursor)
	if err != nil {
		return nil, "", false, apperrors.Wrap(err, "VALIDATION_FAILED", "Invalid cursor", 422)
	}

	q := r.db.WithContext(ctx).Model(&SyncChangeLogModel{}).Where("tenant_id = ?", tenantID)
	if !ts.IsZero() {
		q = q.Where("(updated_at > ?) OR (updated_at = ? AND id > ?)", ts, ts, id)
	}
	if len(types) > 0 {
		q = q.Where("entity_type IN ?", types)
	}

	var rows []SyncChangeLogModel
	if err := q.Order("updated_at ASC, id ASC").Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, "", false, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	out := make([]domain.SyncChange, 0, len(rows))
	for _, m := range rows {
		out = append(out, toChange(m))
	}

	nextCursor := cursor
	if len(out) > 0 {
		last := out[len(out)-1]
		nextCursor = EncodeCursor(last.UpdatedAt, last.ID)
	}
	return out, nextCursor, hasMore, nil
}

func (r *ChangeLogRepo) FindLatest(ctx context.Context, tenantID uuid.UUID, entityType, entityID string) (*domain.SyncChange, error) {
	var m SyncChangeLogModel
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND entity_type = ? AND entity_id = ?", tenantID, entityType, entityID).
		Order("version DESC, updated_at DESC").
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	c := toChange(m)
	return &c, nil
}

func (r *ChangeLogRepo) HasAppliedOp(ctx context.Context, tenantID uuid.UUID, opID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&SyncAppliedOpModel{}).
		Where("tenant_id = ? AND op_id = ?", tenantID, opID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ChangeLogRepo) MarkAppliedOp(ctx context.Context, tenantID uuid.UUID, opID string) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&SyncAppliedOpModel{
		ID: uuid.New(), TenantID: tenantID, OpID: opID, CreatedAt: time.Now().UTC(),
	}).Error
}

type ConflictRepo struct{ db *gorm.DB }

func NewConflictRepo(db *gorm.DB) *ConflictRepo { return &ConflictRepo{db: db} }

func toConflict(m SyncConflictModel) domain.SyncConflict {
	return domain.SyncConflict{
		ID: m.ID, TenantID: m.TenantID, DeviceID: m.DeviceID, UserID: m.UserID,
		EntityType: m.EntityType, EntityID: m.EntityID, ClientOpID: m.ClientOpID,
		BaseVersion: m.BaseVersion, ServerVersion: m.ServerVersion,
		ClientPayload: m.ClientPayload, ServerPayload: m.ServerPayload,
		Status: m.Status, Resolution: m.Resolution, CreatedAt: m.CreatedAt, ResolvedAt: m.ResolvedAt,
	}
}

func (r *ConflictRepo) Create(ctx context.Context, conflict *domain.SyncConflict) error {
	if conflict.ID == uuid.Nil {
		conflict.ID = uuid.New()
	}
	if conflict.CreatedAt.IsZero() {
		conflict.CreatedAt = time.Now().UTC()
	}
	if conflict.Status == "" {
		conflict.Status = domain.ConflictStatusOpen
	}
	return r.db.WithContext(ctx).Create(&SyncConflictModel{
		ID: conflict.ID, TenantID: conflict.TenantID, DeviceID: conflict.DeviceID, UserID: conflict.UserID,
		EntityType: conflict.EntityType, EntityID: conflict.EntityID, ClientOpID: conflict.ClientOpID,
		BaseVersion: conflict.BaseVersion, ServerVersion: conflict.ServerVersion,
		ClientPayload: conflict.ClientPayload, ServerPayload: conflict.ServerPayload,
		Status: conflict.Status, Resolution: conflict.Resolution, CreatedAt: conflict.CreatedAt, ResolvedAt: conflict.ResolvedAt,
	}).Error
}

func (r *ConflictRepo) ListOpen(ctx context.Context, tenantID uuid.UUID, deviceID string) ([]domain.SyncConflict, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND status = ?", tenantID, domain.ConflictStatusOpen)
	if deviceID != "" {
		q = q.Where("device_id = ?", deviceID)
	}
	var rows []SyncConflictModel
	if err := q.Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.SyncConflict, 0, len(rows))
	for _, m := range rows {
		out = append(out, toConflict(m))
	}
	return out, nil
}

func (r *ConflictRepo) Resolve(ctx context.Context, tenantID, id uuid.UUID, resolution string) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&SyncConflictModel{}).
		Where("id = ? AND tenant_id = ? AND status = ?", id, tenantID, domain.ConflictStatusOpen).
		Updates(map[string]any{
			"status":      domain.ConflictStatusResolved,
			"resolution":  resolution,
			"resolved_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *ConflictRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.SyncConflict, error) {
	var m SyncConflictModel
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	c := toConflict(m)
	return &c, nil
}

func (r *ConflictRepo) CountOpen(ctx context.Context, tenantID uuid.UUID, deviceID string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&SyncConflictModel{}).
		Where("tenant_id = ? AND status = ?", tenantID, domain.ConflictStatusOpen)
	if deviceID != "" {
		q = q.Where("device_id = ?", deviceID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
