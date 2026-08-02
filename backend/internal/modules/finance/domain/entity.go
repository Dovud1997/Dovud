package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	DocumentTypeOrder  = "order"
	DocumentTypeManual = "manual"
	DocumentTypeReturn = "return"

	StatusOpen    = "open"
	StatusPartial = "partial"
	StatusClosed  = "closed"
	StatusOverdue = "overdue"

	MethodCash     = "cash"
	MethodCard     = "card"
	MethodTransfer = "transfer"
)

type Receivable struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CustomerID   uuid.UUID
	DocumentType string
	DocumentID   *uuid.UUID
	Amount       float64
	PaidAmount   float64
	Balance      float64
	DueDate      *time.Time
	Status       string
	Currency     string
	Version      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ReceivablePayment struct {
	ID           uuid.UUID
	ReceivableID uuid.UUID
	Amount       float64
	PaidAt       time.Time
	Method       string
	Reference    *string
	CreatedBy    *uuid.UUID
	CreatedAt    time.Time
}

type CreditLimit struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	Amount     float64
	Currency   string
	UpdatedAt  time.Time
}

type AgingReport struct {
	Bucket0To30  float64
	Bucket31To60 float64
	Bucket61To90 float64
	Bucket90Plus float64
}

func ValidDocumentType(s string) bool {
	switch s {
	case DocumentTypeOrder, DocumentTypeManual, DocumentTypeReturn:
		return true
	default:
		return false
	}
}

func ValidStatus(s string) bool {
	switch s {
	case StatusOpen, StatusPartial, StatusClosed, StatusOverdue:
		return true
	default:
		return false
	}
}

func ValidPaymentMethod(s string) bool {
	switch s {
	case MethodCash, MethodCard, MethodTransfer:
		return true
	default:
		return false
	}
}

// EffectiveStatus returns overdue when due date has passed and the receivable is not closed.
func (r Receivable) EffectiveStatus(now time.Time) string {
	if r.Status == StatusClosed {
		return StatusClosed
	}
	if r.DueDate != nil && r.DueDate.Before(now) {
		return StatusOverdue
	}
	return r.Status
}
