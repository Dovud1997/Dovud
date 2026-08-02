package domain

import (
	"time"

	"github.com/google/uuid"
)

type CustomerUser struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	CreatedAt  time.Time
}

type Summary struct {
	CustomerID      uuid.UUID `json:"customer_id"`
	CustomerCode    string    `json:"customer_code"`
	CustomerName    string    `json:"customer_name"`
	OpenOrders      int64     `json:"open_orders"`
	OpenBalance     float64   `json:"open_balance"`
	CreditLimit     float64   `json:"credit_limit"`
	DocumentsCount  int64     `json:"documents_count"`
}
