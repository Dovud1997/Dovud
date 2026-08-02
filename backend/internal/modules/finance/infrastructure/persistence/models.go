package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReceivableModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	CustomerID   uuid.UUID      `gorm:"type:uuid;not null;index"`
	DocumentType string         `gorm:"size:32;not null"`
	DocumentID   *uuid.UUID     `gorm:"type:uuid;index"`
	Amount       float64        `gorm:"not null;default:0"`
	PaidAmount   float64        `gorm:"not null;default:0"`
	Balance      float64        `gorm:"not null;default:0"`
	DueDate      *time.Time
	Status       string         `gorm:"size:32;not null;index"`
	Currency     string         `gorm:"size:8;not null"`
	Version      int64          `gorm:"not null;default:1"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (ReceivableModel) TableName() string { return "receivables" }

type ReceivablePaymentModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ReceivableID uuid.UUID  `gorm:"type:uuid;not null;index"`
	Amount       float64    `gorm:"not null"`
	PaidAt       time.Time  `gorm:"not null"`
	Method       string     `gorm:"size:32;not null"`
	Reference    *string    `gorm:"size:255"`
	CreatedBy    *uuid.UUID `gorm:"type:uuid"`
	CreatedAt    time.Time
}

func (ReceivablePaymentModel) TableName() string { return "receivable_payments" }

type CreditLimitModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Amount     float64   `gorm:"not null;default:0"`
	Currency   string    `gorm:"size:8;not null"`
	UpdatedAt  time.Time
}

func (CreditLimitModel) TableName() string { return "credit_limits" }
