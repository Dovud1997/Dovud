package domain

import (
	"context"

	"github.com/google/uuid"
)

type CompanyRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]Company, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Company, error)
	Create(ctx context.Context, c *Company) error
	Update(ctx context.Context, c *Company) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}

type BranchRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]Branch, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Branch, error)
	Create(ctx context.Context, b *Branch) error
	Update(ctx context.Context, b *Branch) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}

type WarehouseRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, branchID *uuid.UUID, page, perPage int) ([]Warehouse, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Warehouse, error)
	Create(ctx context.Context, w *Warehouse) error
	Update(ctx context.Context, w *Warehouse) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
	ListStocks(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]WarehouseStock, error)
	UpsertStock(ctx context.Context, s *WarehouseStock) error
}
