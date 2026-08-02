package persistence

import (
	"time"

	"github.com/google/uuid"
)

type SyncDeviceModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_sync_devices_tenant_user_device"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_sync_devices_tenant_user_device"`
	DeviceID       string    `gorm:"size:128;not null;uniqueIndex:uq_sync_devices_tenant_user_device"`
	Platform       *string   `gorm:"size:64"`
	AppVersion     *string   `gorm:"size:64"`
	LastPullCursor string    `gorm:"size:256;not null;default:''"`
	LastPushAt     *time.Time
	LastPullAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (SyncDeviceModel) TableName() string { return "sync_devices" }

type SyncChangeLogModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	EntityType  string    `gorm:"size:64;not null;index"`
	EntityID    string    `gorm:"size:64;not null;index"`
	Version     int64     `gorm:"not null"`
	Deleted     bool      `gorm:"not null;default:false"`
	UpdatedAt   time.Time `gorm:"not null;index"`
	PayloadJSON string    `gorm:"type:text;not null;default:'{}'"`
}

func (SyncChangeLogModel) TableName() string { return "sync_change_log" }

type SyncConflictModel struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index"`
	DeviceID      string    `gorm:"size:128;not null;index"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	EntityType    string    `gorm:"size:64;not null"`
	EntityID      string    `gorm:"size:64;not null"`
	ClientOpID    string    `gorm:"size:128;not null"`
	BaseVersion   int64     `gorm:"not null"`
	ServerVersion int64     `gorm:"not null"`
	ClientPayload string    `gorm:"type:text;not null;default:''"`
	ServerPayload string    `gorm:"type:text;not null;default:''"`
	Status        string    `gorm:"size:32;not null;index"`
	Resolution    *string   `gorm:"size:64"`
	CreatedAt     time.Time
	ResolvedAt    *time.Time
}

func (SyncConflictModel) TableName() string { return "sync_conflicts" }

type SyncAppliedOpModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_sync_applied_ops_tenant_op"`
	OpID      string    `gorm:"size:128;not null;uniqueIndex:uq_sync_applied_ops_tenant_op"`
	CreatedAt time.Time
}

func (SyncAppliedOpModel) TableName() string { return "sync_applied_ops" }
