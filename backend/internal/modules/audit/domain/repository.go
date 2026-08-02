package domain

import (
	"context"

	"github.com/google/uuid"
)

type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*AuditLog, error)
	List(ctx context.Context, tenantID uuid.UUID, filters ListFilters, page, perPage int) ([]AuditLog, int64, error)
}
