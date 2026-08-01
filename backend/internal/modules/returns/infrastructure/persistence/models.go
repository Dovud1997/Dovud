package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReturnModel struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	Number     string         `gorm:"size:64;not null"`
	OrderID    *uuid.UUID     `gorm:"type:uuid;index"`
	CustomerID uuid.UUID      `gorm:"type:uuid;not null;index"`
	AgentID    *uuid.UUID     `gorm:"type:uuid;index"`
	Status     string         `gorm:"size:32;not null;index"`
	Reason     *string
	Currency   string         `gorm:"size:8;not null"`
	Subtotal   float64        `gorm:"not null;default:0"`
	TaxTotal   float64        `gorm:"not null;default:0"`
	GrandTotal float64        `gorm:"not null;default:0"`
	Version    int64          `gorm:"not null;default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (ReturnModel) TableName() string { return "returns" }

type ReturnLineModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ReturnID  uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
	Qty       float64   `gorm:"not null"`
	UnitPrice float64   `gorm:"not null;default:0"`
	LineTotal float64   `gorm:"not null;default:0"`
	Reason    *string
}

func (ReturnLineModel) TableName() string { return "return_lines" }
