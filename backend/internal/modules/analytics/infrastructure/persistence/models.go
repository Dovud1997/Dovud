package persistence

import (
	"time"

	"github.com/google/uuid"
)

type KpiDefinitionModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID    *uuid.UUID `gorm:"type:uuid;index"`
	Code        string     `gorm:"size:64;not null;index"`
	Name        string     `gorm:"size:255;not null"`
	Description string     `gorm:"type:text"`
	Unit        string     `gorm:"size:32;not null"`
}

func (KpiDefinitionModel) TableName() string { return "kpi_definitions" }

type KpiSnapshotModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	KpiCode     string     `gorm:"size:64;not null;index"`
	Period      string     `gorm:"size:32;not null;index"`
	PeriodStart time.Time  `gorm:"not null;index"`
	ScopeType   string     `gorm:"size:32;not null"`
	ScopeID     *uuid.UUID `gorm:"type:uuid;index"`
	Value       float64    `gorm:"not null;default:0"`
	CreatedAt   time.Time
}

func (KpiSnapshotModel) TableName() string { return "kpi_snapshots" }
