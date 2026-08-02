package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusConfirmed = "confirmed"
	StatusPicking   = "picking"
	StatusShipped   = "shipped"
	StatusDelivered = "delivered"
	StatusCancelled = "cancelled"
)

type Order struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	Number          string
	CustomerID      uuid.UUID
	AgentID         *uuid.UUID
	BranchID        *uuid.UUID
	WarehouseID     *uuid.UUID
	VisitID         *uuid.UUID
	Status          string
	Currency        string
	Subtotal        float64
	DiscountTotal   float64
	TaxTotal        float64
	GrandTotal      float64
	OrderedAt       time.Time
	DeliveryDate    *time.Time
	PriceListID     *uuid.UUID
	PromotionID     *uuid.UUID
	Comment         *string
	ClientRequestID *string
	Version         int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderLine struct {
	ID              uuid.UUID
	OrderID         uuid.UUID
	ProductID       uuid.UUID
	Qty             float64
	UnitPrice       float64
	Discount        float64
	Tax             float64
	LineTotal       float64
	PromotionItemID *uuid.UUID
}

type OrderStatusHistory struct {
	ID         uuid.UUID
	OrderID    uuid.UUID
	FromStatus string
	ToStatus   string
	ChangedBy  *uuid.UUID
	Comment    *string
	CreatedAt  time.Time
}

var allowedTransitions = map[string][]string{
	StatusDraft:     {StatusSubmitted, StatusCancelled},
	StatusSubmitted: {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusPicking, StatusCancelled},
	StatusPicking:   {StatusShipped, StatusCancelled},
	StatusShipped:   {StatusDelivered},
	StatusDelivered: {},
	StatusCancelled: {},
}

func (o Order) CanTransition(to string) bool {
	for _, s := range allowedTransitions[o.Status] {
		if s == to {
			return true
		}
	}
	return false
}

func ValidStatus(s string) bool {
	switch s {
	case StatusDraft, StatusSubmitted, StatusConfirmed, StatusPicking, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	default:
		return false
	}
}
