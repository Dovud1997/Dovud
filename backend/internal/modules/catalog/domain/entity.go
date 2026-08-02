package domain

import (
	"time"

	"github.com/google/uuid"
)

type Manufacturer struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Name      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Category struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	ParentID  *uuid.UUID
	Code      string
	Name      string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Product struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	SKU            string
	Barcode        *string
	Name           string
	Description    *string
	CategoryID     *uuid.UUID
	ManufacturerID *uuid.UUID
	Unit           string
	VATRate        float64
	IsActive       bool
	Version        int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PriceList struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Name      string
	Currency  string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductPrice struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PriceListID uuid.UUID
	ProductID   uuid.UUID
	Amount      float64
	Currency    string
	ValidFrom   *time.Time
	ValidTo     *time.Time
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Promotion struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Code        string
	Name        string
	Description *string
	StartsAt    time.Time
	EndsAt      time.Time
	DiscountPct float64
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PromotionItem struct {
	ID          uuid.UUID
	PromotionID uuid.UUID
	ProductID   *uuid.UUID
	CategoryID  *uuid.UUID
}
