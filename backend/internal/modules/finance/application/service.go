package application

import (
	"context"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/finance/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type Service struct {
	receivables  domain.ReceivableRepository
	creditLimits domain.CreditLimitRepository
}

func NewService(receivables domain.ReceivableRepository, creditLimits domain.CreditLimitRepository) *Service {
	return &Service{receivables: receivables, creditLimits: creditLimits}
}

type ReceivableDTO struct {
	ID           uuid.UUID              `json:"id"`
	CustomerID   uuid.UUID              `json:"customer_id"`
	DocumentType string                 `json:"document_type"`
	DocumentID   *uuid.UUID             `json:"document_id,omitempty"`
	Amount       float64                `json:"amount"`
	PaidAmount   float64                `json:"paid_amount"`
	Balance      float64                `json:"balance"`
	DueDate      *time.Time             `json:"due_date,omitempty"`
	Status       string                 `json:"status"`
	Currency     string                 `json:"currency"`
	Version      int64                  `json:"version"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Payments     []ReceivablePaymentDTO `json:"payments,omitempty"`
}

type ReceivablePaymentDTO struct {
	ID           uuid.UUID  `json:"id"`
	ReceivableID uuid.UUID  `json:"receivable_id"`
	Amount       float64    `json:"amount"`
	PaidAt       time.Time  `json:"paid_at"`
	Method       string     `json:"method"`
	Reference    *string    `json:"reference,omitempty"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreditLimitDTO struct {
	ID         uuid.UUID `json:"id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AgingReportDTO struct {
	Bucket0To30  float64 `json:"bucket_0_30"`
	Bucket31To60 float64 `json:"bucket_31_60"`
	Bucket61To90 float64 `json:"bucket_61_90"`
	Bucket90Plus float64 `json:"bucket_90_plus"`
}

type CustomerBalanceDTO struct {
	CustomerID uuid.UUID `json:"customer_id"`
	Balance    float64   `json:"balance"`
}

type CreateReceivableInput struct {
	CustomerID uuid.UUID  `json:"customer_id"`
	Amount     float64    `json:"amount"`
	DueDate    *time.Time `json:"due_date"`
	Currency   string     `json:"currency"`
	DocumentID *uuid.UUID `json:"document_id"`
}

type RecordPaymentInput struct {
	Amount    float64    `json:"amount"`
	PaidAt    *time.Time `json:"paid_at"`
	Method    string     `json:"method"`
	Reference *string    `json:"reference"`
}

type SetCreditLimitInput struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func toReceivableDTO(r domain.Receivable, now time.Time) ReceivableDTO {
	return ReceivableDTO{
		ID: r.ID, CustomerID: r.CustomerID, DocumentType: r.DocumentType, DocumentID: r.DocumentID,
		Amount: r.Amount, PaidAmount: r.PaidAmount, Balance: r.Balance, DueDate: r.DueDate,
		Status: r.EffectiveStatus(now), Currency: r.Currency, Version: r.Version,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func toPaymentDTO(p domain.ReceivablePayment) ReceivablePaymentDTO {
	return ReceivablePaymentDTO{
		ID: p.ID, ReceivableID: p.ReceivableID, Amount: p.Amount, PaidAt: p.PaidAt,
		Method: p.Method, Reference: p.Reference, CreatedBy: p.CreatedBy, CreatedAt: p.CreatedAt,
	}
}

func toCreditLimitDTO(cl domain.CreditLimit) CreditLimitDTO {
	return CreditLimitDTO{
		ID: cl.ID, CustomerID: cl.CustomerID, Amount: cl.Amount, Currency: cl.Currency, UpdatedAt: cl.UpdatedAt,
	}
}

func (s *Service) ListReceivables(ctx context.Context, tenantID uuid.UUID, filters domain.ReceivableListFilters, page, perPage int) ([]ReceivableDTO, int64, error) {
	if filters.Status != "" && !domain.ValidStatus(filters.Status) {
		return nil, 0, apperrors.ErrValidation
	}
	rows, total, err := s.receivables.List(ctx, tenantID, filters, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	now := time.Now().UTC()
	out := make([]ReceivableDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toReceivableDTO(r, now))
	}
	return out, total, nil
}

func (s *Service) GetReceivable(ctx context.Context, tenantID, id uuid.UUID) (*ReceivableDTO, error) {
	rec, err := s.receivables.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	payments, err := s.receivables.ListPayments(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toReceivableDTO(*rec, time.Now().UTC())
	dto.Payments = make([]ReceivablePaymentDTO, 0, len(payments))
	for _, p := range payments {
		dto.Payments = append(dto.Payments, toPaymentDTO(p))
	}
	return &dto, nil
}

func (s *Service) CreateReceivable(ctx context.Context, tenantID uuid.UUID, in CreateReceivableInput) (*ReceivableDTO, error) {
	if in.CustomerID == uuid.Nil || in.Amount <= 0 {
		return nil, apperrors.ErrValidation
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "UZS"
	}

	now := time.Now().UTC()
	status := domain.StatusOpen
	if in.DueDate != nil && in.DueDate.Before(now) {
		status = domain.StatusOverdue
	}

	rec := &domain.Receivable{
		TenantID: tenantID, CustomerID: in.CustomerID, DocumentType: domain.DocumentTypeManual,
		DocumentID: in.DocumentID, Amount: in.Amount, PaidAmount: 0, Balance: in.Amount,
		DueDate: in.DueDate, Status: status, Currency: currency,
	}
	if err := s.receivables.Create(ctx, rec); err != nil {
		return nil, err
	}
	dto := toReceivableDTO(*rec, now)
	return &dto, nil
}

func (s *Service) RecordPayment(ctx context.Context, tenantID, receivableID uuid.UUID, in RecordPaymentInput, createdBy *uuid.UUID) (*ReceivableDTO, error) {
	method := strings.TrimSpace(in.Method)
	if in.Amount <= 0 || !domain.ValidPaymentMethod(method) {
		return nil, apperrors.ErrValidation
	}

	rec, err := s.receivables.FindByID(ctx, tenantID, receivableID)
	if err != nil {
		return nil, err
	}
	if rec.Status == domain.StatusClosed || rec.Balance <= 0 {
		return nil, apperrors.ErrConflict
	}
	if in.Amount > rec.Balance {
		return nil, apperrors.ErrValidation
	}

	now := time.Now().UTC()
	paidAt := now
	if in.PaidAt != nil {
		paidAt = in.PaidAt.UTC()
	}

	payment := &domain.ReceivablePayment{
		ReceivableID: rec.ID, Amount: in.Amount, PaidAt: paidAt,
		Method: method, Reference: in.Reference, CreatedBy: createdBy,
	}
	if err := s.receivables.AddPayment(ctx, payment); err != nil {
		return nil, err
	}

	rec.PaidAmount += in.Amount
	rec.Balance = rec.Amount - rec.PaidAmount
	if rec.Balance <= 0 {
		rec.Balance = 0
		rec.Status = domain.StatusClosed
	} else {
		rec.Status = domain.StatusPartial
		if rec.DueDate != nil && rec.DueDate.Before(now) {
			rec.Status = domain.StatusOverdue
		}
	}
	if err := s.receivables.Update(ctx, rec); err != nil {
		return nil, err
	}

	dto := toReceivableDTO(*rec, now)
	payments, err := s.receivables.ListPayments(ctx, rec.ID)
	if err != nil {
		return nil, err
	}
	dto.Payments = make([]ReceivablePaymentDTO, 0, len(payments))
	for _, p := range payments {
		dto.Payments = append(dto.Payments, toPaymentDTO(p))
	}
	return &dto, nil
}

func (s *Service) GetCustomerBalance(ctx context.Context, tenantID, customerID uuid.UUID) (*CustomerBalanceDTO, error) {
	if customerID == uuid.Nil {
		return nil, apperrors.ErrValidation
	}
	balance, err := s.receivables.SumBalanceByCustomer(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}
	return &CustomerBalanceDTO{CustomerID: customerID, Balance: balance}, nil
}

func (s *Service) GetCreditLimit(ctx context.Context, tenantID, customerID uuid.UUID) (*CreditLimitDTO, error) {
	if customerID == uuid.Nil {
		return nil, apperrors.ErrValidation
	}
	cl, err := s.creditLimits.GetByCustomer(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}
	dto := toCreditLimitDTO(*cl)
	return &dto, nil
}

func (s *Service) SetCreditLimit(ctx context.Context, tenantID, customerID uuid.UUID, in SetCreditLimitInput) (*CreditLimitDTO, error) {
	if customerID == uuid.Nil || in.Amount < 0 {
		return nil, apperrors.ErrValidation
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "UZS"
	}
	cl := &domain.CreditLimit{
		TenantID: tenantID, CustomerID: customerID, Amount: in.Amount, Currency: currency,
	}
	if err := s.creditLimits.Upsert(ctx, cl); err != nil {
		return nil, err
	}
	dto := toCreditLimitDTO(*cl)
	return &dto, nil
}

func (s *Service) AgingReport(ctx context.Context, tenantID uuid.UUID) (*AgingReportDTO, error) {
	report, err := s.receivables.AgingReport(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &AgingReportDTO{
		Bucket0To30: report.Bucket0To30, Bucket31To60: report.Bucket31To60,
		Bucket61To90: report.Bucket61To90, Bucket90Plus: report.Bucket90Plus,
	}, nil
}
