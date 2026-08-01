package domain

import (
	"context"

	"github.com/google/uuid"
)

type TenantRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	FindByCode(ctx context.Context, code string) (*Tenant, error)
	FindByHost(ctx context.Context, host string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Create(ctx context.Context, tenant *Tenant) error
}

type BrandingRepository interface {
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*Branding, error)
	Upsert(ctx context.Context, branding *Branding) error
	GetPublicByCode(ctx context.Context, code string) (*PublicBranding, error)
	GetPublicByHost(ctx context.Context, host string) (*PublicBranding, error)
}

type DomainRepository interface {
	List(ctx context.Context, tenantID uuid.UUID) ([]Domain, error)
	Create(ctx context.Context, domain *Domain) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}
