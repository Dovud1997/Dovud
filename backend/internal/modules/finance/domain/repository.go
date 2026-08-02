package domain

import (
	"context"

	"github.com/google/uuid"
)

type ReceivableListFilters struct {
	CustomerID *uuid.UUID
	Status     string
}

type ReceivableRepository interface {
	List(ctx context.Context, tenantID uuid.UUID, filters ReceivableListFilters, page, perPage int) ([]Receivable, int64, error)
	FindByID(ctx context.Context, tenantID, id uuid.UUID) (*Receivable, error)
	Create(ctx context.Context, r *Receivable) error
	Update(ctx context.Context, r *Receivable) error
	SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
	ListPayments(ctx context.Context, receivableID uuid.UUID) ([]ReceivablePayment, error)
	AddPayment(ctx context.Context, p *ReceivablePayment) error
	AgingReport(ctx context.Context, tenantID uuid.UUID) (*AgingReport, error)
	SumBalanceByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (float64, error)
}

type CreditLimitRepository interface {
	GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*CreditLimit, error)
	Upsert(ctx context.Context, cl *CreditLimit) error
}
