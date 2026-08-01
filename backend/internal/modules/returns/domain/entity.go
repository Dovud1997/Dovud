package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

type Return struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Number     string
	OrderID    *uuid.UUID
	CustomerID uuid.UUID
	AgentID    *uuid.UUID
	Status     string
	Reason     *string
	Currency   string
	Subtotal   float64
	TaxTotal   float64
	GrandTotal float64
	Version    int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ReturnLine struct {
	ID        uuid.UUID
	ReturnID  uuid.UUID
	ProductID uuid.UUID
	Qty       float64
	UnitPrice float64
	LineTotal float64
	Reason    *string
}

var allowedTransitions = map[string][]string{
	StatusDraft:     {StatusSubmitted, StatusCancelled},
	StatusSubmitted: {StatusApproved, StatusRejected},
	StatusApproved:  {StatusCompleted},
	StatusRejected:  {},
	StatusCompleted: {},
	StatusCancelled: {},
}

func (r Return) CanTransition(to string) bool {
	for _, s := range allowedTransitions[r.Status] {
		if s == to {
			return true
		}
	}
	return false
}

func ValidStatus(s string) bool {
	switch s {
	case StatusDraft, StatusSubmitted, StatusApproved, StatusRejected, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}
