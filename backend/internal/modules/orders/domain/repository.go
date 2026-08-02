package domain

import (
	"context"

	"github.com/google/uuid"
)

type OrderListFilters struct {
	Status     string
	CustomerID *uuid.UUID
	AgentID    *uuid.UUID
}

type OrderRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, filters OrderListFilters, page, perPage int) ([]Order, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Order, []OrderLine, error)
	FindByClientRequestID(ctx context.Context, tenantID uuid.UUID, clientRequestID string) (*Order, []OrderLine, error)
	Create(ctx context.Context, order *Order, lines []OrderLine) error
	Update(ctx context.Context, order *Order, lines []OrderLine) error
	AddHistory(ctx context.Context, h *OrderStatusHistory) error
	ListHistory(ctx context.Context, orderID uuid.UUID) ([]OrderStatusHistory, error)
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}
