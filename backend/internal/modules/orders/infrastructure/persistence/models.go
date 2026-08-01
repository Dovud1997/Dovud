package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	Number          string     `gorm:"size:64;not null"`
	CustomerID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	AgentID         *uuid.UUID `gorm:"type:uuid;index"`
	BranchID        *uuid.UUID `gorm:"type:uuid;index"`
	WarehouseID     *uuid.UUID `gorm:"type:uuid;index"`
	VisitID         *uuid.UUID `gorm:"type:uuid;index"`
	Status          string     `gorm:"size:32;not null;index"`
	Currency        string     `gorm:"size:8;not null"`
	Subtotal        float64    `gorm:"not null;default:0"`
	DiscountTotal   float64    `gorm:"not null;default:0"`
	TaxTotal        float64    `gorm:"not null;default:0"`
	GrandTotal      float64    `gorm:"not null;default:0"`
	OrderedAt       time.Time  `gorm:"not null"`
	DeliveryDate    *time.Time
	PriceListID     *uuid.UUID `gorm:"type:uuid"`
	PromotionID     *uuid.UUID `gorm:"type:uuid"`
	Comment         *string
	ClientRequestID *string `gorm:"size:128;index"`
	Version         int64   `gorm:"not null;default:1"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (OrderModel) TableName() string { return "orders" }

type OrderLineModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OrderID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	ProductID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Qty             float64    `gorm:"not null"`
	UnitPrice       float64    `gorm:"not null;default:0"`
	Discount        float64    `gorm:"not null;default:0"`
	Tax             float64    `gorm:"not null;default:0"`
	LineTotal       float64    `gorm:"not null;default:0"`
	PromotionItemID *uuid.UUID `gorm:"type:uuid"`
}

func (OrderLineModel) TableName() string { return "order_lines" }

type OrderStatusHistoryModel struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OrderID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	FromStatus string     `gorm:"size:32;not null"`
	ToStatus   string     `gorm:"size:32;not null"`
	ChangedBy  *uuid.UUID `gorm:"type:uuid"`
	Comment    *string
	CreatedAt  time.Time
}

func (OrderStatusHistoryModel) TableName() string { return "order_status_history" }
