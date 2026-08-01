package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:255;not null"`
	Inn       *string        `gorm:"size:64"`
	Status    string         `gorm:"size:32;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CompanyModel) TableName() string { return "companies" }

type BranchModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	CompanyID *uuid.UUID     `gorm:"type:uuid;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:255;not null"`
	Address   *string
	Lat       *float64
	Lng       *float64
	Status    string `gorm:"size:32;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (BranchModel) TableName() string { return "branches" }

type WarehouseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	BranchID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:255;not null"`
	Type      string         `gorm:"size:32;not null;default:main"`
	Status    string         `gorm:"size:32;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (WarehouseModel) TableName() string { return "warehouses" }

type WarehouseStockModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	WarehouseID uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null;index"`
	QtyOnHand   float64   `gorm:"not null;default:0"`
	QtyReserved float64   `gorm:"not null;default:0"`
	UpdatedAt   time.Time
}

func (WarehouseStockModel) TableName() string { return "warehouse_stocks" }
