package domain

import (
	"context"

	"github.com/google/uuid"
)

type ReturnListFilters struct {
	Status string
}

type ReturnRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, filters ReturnListFilters, page, perPage int) ([]Return, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Return, []ReturnLine, error)
	Create(ctx context.Context, ret *Return, lines []ReturnLine) error
	Update(ctx context.Context, ret *Return, lines []ReturnLine) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}
