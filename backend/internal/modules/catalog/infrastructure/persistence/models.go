package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ManufacturerModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:255;not null"`
	Status    string         `gorm:"size:32;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (ManufacturerModel) TableName() string { return "manufacturers" }

type CategoryModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	ParentID  *uuid.UUID     `gorm:"type:uuid;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:255;not null"`
	SortOrder int            `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CategoryModel) TableName() string { return "categories" }

type ProductModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index"`
	SKU            string         `gorm:"size:64;not null"`
	Barcode        *string        `gorm:"size:64"`
	Name           string         `gorm:"size:255;not null"`
	Description    *string
	CategoryID     *uuid.UUID `gorm:"type:uuid;index"`
	ManufacturerID *uuid.UUID `gorm:"type:uuid;index"`
	Unit           string     `gorm:"size:32;not null;default:pcs"`
	VATRate        float64    `gorm:"not null;default:0"`
	IsActive       bool       `gorm:"not null;default:true"`
	Version        int64      `gorm:"not null;default:1"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (ProductModel) TableName() string { return "products" }

type PriceListModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	Code      string         `gorm:"size:64;not null"`
	Name      string         `gorm:"size:255;not null"`
	Currency  string         `gorm:"size:3;not null"`
	IsDefault bool           `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (PriceListModel) TableName() string { return "price_lists" }

type ProductPriceModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	PriceListID uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Amount      float64   `gorm:"not null"`
	Currency    string    `gorm:"size:3;not null"`
	ValidFrom   *time.Time
	ValidTo     *time.Time
	Version     int64 `gorm:"not null;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (ProductPriceModel) TableName() string { return "product_prices" }

type PromotionModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Code        string    `gorm:"size:64;not null"`
	Name        string    `gorm:"size:255;not null"`
	Description *string
	StartsAt    time.Time
	EndsAt      time.Time
	DiscountPct float64 `gorm:"not null;default:0"`
	IsActive    bool    `gorm:"not null;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (PromotionModel) TableName() string { return "promotions" }

type PromotionItemModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PromotionID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ProductID   *uuid.UUID `gorm:"type:uuid;index"`
	CategoryID  *uuid.UUID `gorm:"type:uuid;index"`
}

func (PromotionItemModel) TableName() string { return "promotion_items" }
