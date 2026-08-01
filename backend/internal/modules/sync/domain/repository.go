package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceRepository interface {
	Upsert(ctx context.Context, device *SyncDevice) error
	Find(ctx context.Context, tenantID, userID uuid.UUID, deviceID string) (*SyncDevice, error)
	UpdateCursors(ctx context.Context, tenantID, userID uuid.UUID, deviceID string, pullCursor string, pullAt, pushAt *time.Time) error
}

type ChangeLogRepository interface {
	Append(ctx context.Context, change *SyncChange) error
	ListSince(ctx context.Context, tenantID uuid.UUID, cursor string, types []string, limit int) (changes []SyncChange, nextCursor string, hasMore bool, err error)
	FindLatest(ctx context.Context, tenantID uuid.UUID, entityType, entityID string) (*SyncChange, error)
	HasAppliedOp(ctx context.Context, tenantID uuid.UUID, opID string) (bool, error)
	MarkAppliedOp(ctx context.Context, tenantID uuid.UUID, opID string) error
}

type ConflictRepository interface {
	Create(ctx context.Context, conflict *SyncConflict) error
	ListOpen(ctx context.Context, tenantID uuid.UUID, deviceID string) ([]SyncConflict, error)
	Resolve(ctx context.Context, tenantID, id uuid.UUID, resolution string) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*SyncConflict, error)
	CountOpen(ctx context.Context, tenantID uuid.UUID, deviceID string) (int64, error)
}
