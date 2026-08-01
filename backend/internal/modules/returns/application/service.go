package application

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/returns/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type Service struct {
	returns domain.ReturnRepository
}

func NewService(returns domain.ReturnRepository) *Service {
	return &Service{returns: returns}
}

type ReturnLineDTO struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Qty       float64   `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	LineTotal float64   `json:"line_total"`
	Reason    *string   `json:"reason,omitempty"`
}

type ReturnDTO struct {
	ID         uuid.UUID       `json:"id"`
	Number     string          `json:"number"`
	OrderID    *uuid.UUID      `json:"order_id,omitempty"`
	CustomerID uuid.UUID       `json:"customer_id"`
	AgentID    *uuid.UUID      `json:"agent_id,omitempty"`
	Status     string          `json:"status"`
	Reason     *string         `json:"reason,omitempty"`
	Currency   string          `json:"currency"`
	Subtotal   float64         `json:"subtotal"`
	TaxTotal   float64         `json:"tax_total"`
	GrandTotal float64         `json:"grand_total"`
	Version    int64           `json:"version"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Lines      []ReturnLineDTO `json:"lines,omitempty"`
}

type ReturnLineInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Qty       float64   `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Reason    *string   `json:"reason"`
}

type CreateReturnInput struct {
	OrderID    *uuid.UUID        `json:"order_id"`
	CustomerID uuid.UUID         `json:"customer_id"`
	AgentID    *uuid.UUID        `json:"agent_id"`
	Reason     *string           `json:"reason"`
	Currency   string            `json:"currency"`
	TaxTotal   float64           `json:"tax_total"`
	Status     string            `json:"status"`
	Lines      []ReturnLineInput `json:"lines"`
}

type UpdateDraftInput struct {
	OrderID    *uuid.UUID        `json:"order_id"`
	CustomerID *uuid.UUID        `json:"customer_id"`
	AgentID    *uuid.UUID        `json:"agent_id"`
	Reason     *string           `json:"reason"`
	Currency   *string           `json:"currency"`
	TaxTotal   *float64          `json:"tax_total"`
	Lines      []ReturnLineInput `json:"lines"`
}

func toLineDTO(l domain.ReturnLine) ReturnLineDTO {
	return ReturnLineDTO{
		ID: l.ID, ProductID: l.ProductID, Qty: l.Qty, UnitPrice: l.UnitPrice,
		LineTotal: l.LineTotal, Reason: l.Reason,
	}
}

func toReturnDTO(r domain.Return, lines []domain.ReturnLine) ReturnDTO {
	dto := ReturnDTO{
		ID: r.ID, Number: r.Number, OrderID: r.OrderID, CustomerID: r.CustomerID,
		AgentID: r.AgentID, Status: r.Status, Reason: r.Reason, Currency: r.Currency,
		Subtotal: r.Subtotal, TaxTotal: r.TaxTotal, GrandTotal: r.GrandTotal,
		Version: r.Version, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if lines != nil {
		dto.Lines = make([]ReturnLineDTO, 0, len(lines))
		for _, l := range lines {
			dto.Lines = append(dto.Lines, toLineDTO(l))
		}
	}
	return dto
}

func buildLines(inputs []ReturnLineInput) ([]domain.ReturnLine, float64, error) {
	if len(inputs) == 0 {
		return nil, 0, apperrors.ErrValidation
	}
	lines := make([]domain.ReturnLine, 0, len(inputs))
	var subtotal float64
	for _, in := range inputs {
		if in.ProductID == uuid.Nil || in.Qty <= 0 {
			return nil, 0, apperrors.ErrValidation
		}
		lineTotal := in.Qty * in.UnitPrice
		lines = append(lines, domain.ReturnLine{
			ProductID: in.ProductID, Qty: in.Qty, UnitPrice: in.UnitPrice,
			LineTotal: lineTotal, Reason: in.Reason,
		})
		subtotal += lineTotal
	}
	return lines, subtotal, nil
}

func generateReturnNumber() string {
	var n uint32
	_ = binary.Read(rand.Reader, binary.BigEndian, &n)
	return fmt.Sprintf("RET-%s-%04d", time.Now().UTC().Format("20060102"), n%10000)
}

func (s *Service) ListReturns(ctx context.Context, tenantID uuid.UUID, filters domain.ReturnListFilters, page, perPage int) ([]ReturnDTO, int64, error) {
	rows, total, err := s.returns.List(ctx, tenantID, filters, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ReturnDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toReturnDTO(r, nil))
	}
	return out, total, nil
}

func (s *Service) GetReturn(ctx context.Context, tenantID, id uuid.UUID) (*ReturnDTO, error) {
	r, lines, err := s.returns.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	dto := toReturnDTO(*r, lines)
	return &dto, nil
}

func (s *Service) CreateReturn(ctx context.Context, tenantID uuid.UUID, in CreateReturnInput) (*ReturnDTO, error) {
	if in.CustomerID == uuid.Nil {
		return nil, apperrors.ErrValidation
	}

	status := strings.TrimSpace(in.Status)
	if status == "" {
		status = domain.StatusDraft
	}
	if status != domain.StatusDraft && status != domain.StatusSubmitted {
		return nil, apperrors.ErrValidation
	}

	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "UZS"
	}

	lines, subtotal, err := buildLines(in.Lines)
	if err != nil {
		return nil, err
	}
	taxTotal := in.TaxTotal
	if taxTotal < 0 {
		return nil, apperrors.ErrValidation
	}

	ret := &domain.Return{
		TenantID: tenantID, Number: generateReturnNumber(), OrderID: in.OrderID,
		CustomerID: in.CustomerID, AgentID: in.AgentID, Status: status, Reason: in.Reason,
		Currency: currency, Subtotal: subtotal, TaxTotal: taxTotal, GrandTotal: subtotal + taxTotal,
	}

	if err := s.returns.Create(ctx, ret, lines); err != nil {
		return nil, err
	}
	dto := toReturnDTO(*ret, lines)
	return &dto, nil
}

func (s *Service) UpdateDraft(ctx context.Context, tenantID, id uuid.UUID, in UpdateDraftInput) (*ReturnDTO, error) {
	r, _, err := s.returns.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if r.Status != domain.StatusDraft {
		return nil, apperrors.ErrConflict
	}

	if in.OrderID != nil {
		r.OrderID = in.OrderID
	}
	if in.CustomerID != nil {
		if *in.CustomerID == uuid.Nil {
			return nil, apperrors.ErrValidation
		}
		r.CustomerID = *in.CustomerID
	}
	if in.AgentID != nil {
		r.AgentID = in.AgentID
	}
	if in.Reason != nil {
		r.Reason = in.Reason
	}
	if in.Currency != nil {
		cur := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if cur == "" {
			return nil, apperrors.ErrValidation
		}
		r.Currency = cur
	}
	if in.TaxTotal != nil {
		if *in.TaxTotal < 0 {
			return nil, apperrors.ErrValidation
		}
		r.TaxTotal = *in.TaxTotal
	}

	var lines []domain.ReturnLine
	if in.Lines != nil {
		built, subtotal, err := buildLines(in.Lines)
		if err != nil {
			return nil, err
		}
		lines = built
		r.Subtotal = subtotal
		r.GrandTotal = r.Subtotal + r.TaxTotal
		if err := s.returns.Update(ctx, r, lines); err != nil {
			return nil, err
		}
	} else {
		r.GrandTotal = r.Subtotal + r.TaxTotal
		if err := s.returns.Update(ctx, r, nil); err != nil {
			return nil, err
		}
		_, lines, err = s.returns.FindByID(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
	}

	dto := toReturnDTO(*r, lines)
	return &dto, nil
}

func (s *Service) transition(ctx context.Context, tenantID, id uuid.UUID, toStatus string) (*ReturnDTO, error) {
	if !domain.ValidStatus(toStatus) {
		return nil, apperrors.ErrValidation
	}
	r, lines, err := s.returns.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if !r.CanTransition(toStatus) {
		return nil, apperrors.ErrConflict
	}
	r.Status = toStatus
	if err := s.returns.Update(ctx, r, nil); err != nil {
		return nil, err
	}
	dto := toReturnDTO(*r, lines)
	return &dto, nil
}

func (s *Service) Submit(ctx context.Context, tenantID, id uuid.UUID) (*ReturnDTO, error) {
	return s.transition(ctx, tenantID, id, domain.StatusSubmitted)
}

func (s *Service) Approve(ctx context.Context, tenantID, id uuid.UUID) (*ReturnDTO, error) {
	return s.transition(ctx, tenantID, id, domain.StatusApproved)
}

func (s *Service) Reject(ctx context.Context, tenantID, id uuid.UUID, reason *string) (*ReturnDTO, error) {
	r, lines, err := s.returns.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if !r.CanTransition(domain.StatusRejected) {
		return nil, apperrors.ErrConflict
	}
	r.Status = domain.StatusRejected
	if reason != nil {
		r.Reason = reason
	}
	if err := s.returns.Update(ctx, r, nil); err != nil {
		return nil, err
	}
	dto := toReturnDTO(*r, lines)
	return &dto, nil
}

func (s *Service) Complete(ctx context.Context, tenantID, id uuid.UUID) (*ReturnDTO, error) {
	return s.transition(ctx, tenantID, id, domain.StatusCompleted)
}
