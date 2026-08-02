package persistence

import (
	"time"

	"github.com/google/uuid"
)

type AuditLogModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID    *uuid.UUID `gorm:"type:uuid;index"`
	ActorUserID *uuid.UUID `gorm:"type:uuid;index"`
	Action      string     `gorm:"size:128;not null;index"`
	EntityType  *string    `gorm:"size:64;index"`
	EntityID    *uuid.UUID `gorm:"type:uuid;index"`
	BeforeJSON  *string    `gorm:"type:text"`
	AfterJSON   *string    `gorm:"type:text"`
	IP          *string    `gorm:"size:64"`
	UserAgent   *string    `gorm:"type:text"`
	RequestID   *string    `gorm:"size:64;index"`
	CreatedAt   time.Time  `gorm:"index"`
}

func (AuditLogModel) TableName() string { return "audit_logs" }
