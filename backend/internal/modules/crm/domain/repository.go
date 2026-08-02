package domain

import (
	"context"

	"github.com/google/uuid"
)

type CustomerRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]Customer, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Customer, error)
	Create(ctx context.Context, c *Customer) error
	Update(ctx context.Context, c *Customer) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}

type CustomerContactRepository interface {
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]CustomerContact, error)
	FindByID(ctx context.Context, customerID, id uuid.UUID) (*CustomerContact, error)
	Create(ctx context.Context, c *CustomerContact) error
	Update(ctx context.Context, c *CustomerContact) error
	SoftDelete(ctx context.Context, customerID, id uuid.UUID) error
}

type CustomerAddressRepository interface {
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]CustomerAddress, error)
	FindByID(ctx context.Context, customerID, id uuid.UUID) (*CustomerAddress, error)
	Create(ctx context.Context, a *CustomerAddress) error
	Update(ctx context.Context, a *CustomerAddress) error
	SoftDelete(ctx context.Context, customerID, id uuid.UUID) error
}

type CustomerCategoryRepository interface {
	List(ctx context.Context, tenantID uuid.UUID) ([]CustomerCategory, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*CustomerCategory, error)
	Create(ctx context.Context, c *CustomerCategory) error
	Update(ctx context.Context, c *CustomerCategory) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}
