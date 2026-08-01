package domain

import (
	"context"

	"github.com/google/uuid"
)

type ManufacturerRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]Manufacturer, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Manufacturer, error)
	Create(ctx context.Context, m *Manufacturer) error
	Update(ctx context.Context, m *Manufacturer) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}

type CategoryRepository interface {
	List(ctx context.Context, tenantID uuid.UUID) ([]Category, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Category, error)
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}

type ProductRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, q string, page, perPage int) ([]Product, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Product, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}

type PriceRepository interface {
	ListPriceLists(ctx context.Context, tenantID uuid.UUID) ([]PriceList, error)
	FindPriceList(ctx context.Context, tenantID, id uuid.UUID) (*PriceList, error)
	CreatePriceList(ctx context.Context, pl *PriceList) error
	UpdatePriceList(ctx context.Context, pl *PriceList) error
	SoftDeletePriceList(ctx context.Context, tenantID, id uuid.UUID) error
	ListPrices(ctx context.Context, tenantID, priceListID uuid.UUID) ([]ProductPrice, error)
	UpsertPrice(ctx context.Context, price *ProductPrice) error
}

type PromotionRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]Promotion, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Promotion, error)
	Create(ctx context.Context, p *Promotion) error
	Update(ctx context.Context, p *Promotion) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
	ReplaceItems(ctx context.Context, promotionID uuid.UUID, items []PromotionItem) error
	ListItems(ctx context.Context, promotionID uuid.UUID) ([]PromotionItem, error)
}
