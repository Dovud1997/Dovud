package application

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/modules/orders/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/syncport"
	"github.com/google/uuid"
)

type Service struct {
	orders domain.OrderRepository
	sync   syncport.ChangeRecorder
}

func NewService(orders domain.OrderRepository) *Service {
	return &Service{orders: orders}
}

func (s *Service) WithSync(rec syncport.ChangeRecorder) *Service {
	s.sync = rec
	return s
}

func (s *Service) record(ctx context.Context, tenantID uuid.UUID, dto *OrderDTO) {
	if s.sync == nil || dto == nil || !syncport.ShouldFanout(ctx) {
		return
	}
	_ = s.sync.RecordChange(ctx, tenantID, "order", dto.ID.String(), dto.Version, false, dto)
}

type OrderLineDTO struct {
	ID              uuid.UUID  `json:"id"`
	ProductID       uuid.UUID  `json:"product_id"`
	Qty             float64    `json:"qty"`
	UnitPrice       float64    `json:"unit_price"`
	Discount        float64    `json:"discount"`
	Tax             float64    `json:"tax"`
	LineTotal       float64    `json:"line_total"`
	PromotionItemID *uuid.UUID `json:"promotion_item_id,omitempty"`
}

type OrderDTO struct {
	ID              uuid.UUID      `json:"id"`
	Number          string         `json:"number"`
	CustomerID      uuid.UUID      `json:"customer_id"`
	AgentID         *uuid.UUID     `json:"agent_id,omitempty"`
	BranchID        *uuid.UUID     `json:"branch_id,omitempty"`
	WarehouseID     *uuid.UUID     `json:"warehouse_id,omitempty"`
	VisitID         *uuid.UUID     `json:"visit_id,omitempty"`
	Status          string         `json:"status"`
	Currency        string         `json:"currency"`
	Subtotal        float64        `json:"subtotal"`
	DiscountTotal   float64        `json:"discount_total"`
	TaxTotal        float64        `json:"tax_total"`
	GrandTotal      float64        `json:"grand_total"`
	OrderedAt       time.Time      `json:"ordered_at"`
	DeliveryDate    *time.Time     `json:"delivery_date,omitempty"`
	PriceListID     *uuid.UUID     `json:"price_list_id,omitempty"`
	PromotionID     *uuid.UUID     `json:"promotion_id,omitempty"`
	Comment         *string        `json:"comment,omitempty"`
	ClientRequestID *string        `json:"client_request_id,omitempty"`
	Version         int64          `json:"version"`
	Lines           []OrderLineDTO `json:"lines,omitempty"`
}

type OrderStatusHistoryDTO struct {
	ID         uuid.UUID  `json:"id"`
	OrderID    uuid.UUID  `json:"order_id"`
	FromStatus string     `json:"from_status"`
	ToStatus   string     `json:"to_status"`
	ChangedBy  *uuid.UUID `json:"changed_by,omitempty"`
	Comment    *string    `json:"comment,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type OrderLineInput struct {
	ProductID uuid.UUID `json:"product_id"`
	Qty       float64   `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Discount  float64   `json:"discount"`
	Tax       float64   `json:"tax"`
}

type CreateOrderInput struct {
	CustomerID      uuid.UUID        `json:"customer_id"`
	AgentID         *uuid.UUID       `json:"agent_id"`
	BranchID        *uuid.UUID       `json:"branch_id"`
	WarehouseID     *uuid.UUID       `json:"warehouse_id"`
	VisitID         *uuid.UUID       `json:"visit_id"`
	Currency        string           `json:"currency"`
	DeliveryDate    *time.Time       `json:"delivery_date"`
	PriceListID     *uuid.UUID       `json:"price_list_id"`
	PromotionID     *uuid.UUID       `json:"promotion_id"`
	Comment         *string          `json:"comment"`
	ClientRequestID *string          `json:"client_request_id"`
	Status          string           `json:"status"`
	Lines           []OrderLineInput `json:"lines"`
}

type UpdateDraftInput struct {
	CustomerID   *uuid.UUID       `json:"customer_id"`
	AgentID      *uuid.UUID       `json:"agent_id"`
	BranchID     *uuid.UUID       `json:"branch_id"`
	WarehouseID  *uuid.UUID       `json:"warehouse_id"`
	VisitID      *uuid.UUID       `json:"visit_id"`
	Currency     *string          `json:"currency"`
	DeliveryDate *time.Time       `json:"delivery_date"`
	PriceListID  *uuid.UUID       `json:"price_list_id"`
	PromotionID  *uuid.UUID       `json:"promotion_id"`
	Comment      *string          `json:"comment"`
	Lines        []OrderLineInput `json:"lines"`
}

type TransitionStatusInput struct {
	Status  string  `json:"status"`
	Comment *string `json:"comment"`
}

func toLineDTO(l domain.OrderLine) OrderLineDTO {
	return OrderLineDTO{
		ID: l.ID, ProductID: l.ProductID, Qty: l.Qty, UnitPrice: l.UnitPrice,
		Discount: l.Discount, Tax: l.Tax, LineTotal: l.LineTotal, PromotionItemID: l.PromotionItemID,
	}
}

func toOrderDTO(o domain.Order, lines []domain.OrderLine) OrderDTO {
	dto := OrderDTO{
		ID: o.ID, Number: o.Number, CustomerID: o.CustomerID, AgentID: o.AgentID,
		BranchID: o.BranchID, WarehouseID: o.WarehouseID, VisitID: o.VisitID, Status: o.Status,
		Currency: o.Currency, Subtotal: o.Subtotal, DiscountTotal: o.DiscountTotal, TaxTotal: o.TaxTotal,
		GrandTotal: o.GrandTotal, OrderedAt: o.OrderedAt, DeliveryDate: o.DeliveryDate,
		PriceListID: o.PriceListID, PromotionID: o.PromotionID, Comment: o.Comment,
		ClientRequestID: o.ClientRequestID, Version: o.Version,
	}
	if lines != nil {
		dto.Lines = make([]OrderLineDTO, 0, len(lines))
		for _, l := range lines {
			dto.Lines = append(dto.Lines, toLineDTO(l))
		}
	}
	return dto
}

func toHistoryDTO(h domain.OrderStatusHistory) OrderStatusHistoryDTO {
	return OrderStatusHistoryDTO{
		ID: h.ID, OrderID: h.OrderID, FromStatus: h.FromStatus, ToStatus: h.ToStatus,
		ChangedBy: h.ChangedBy, Comment: h.Comment, CreatedAt: h.CreatedAt,
	}
}

func buildLines(inputs []OrderLineInput) ([]domain.OrderLine, float64, float64, float64, float64, error) {
	if len(inputs) == 0 {
		return nil, 0, 0, 0, 0, apperrors.ErrValidation
	}
	lines := make([]domain.OrderLine, 0, len(inputs))
	var subtotal, discountTotal, taxTotal, grandTotal float64
	for _, in := range inputs {
		if in.ProductID == uuid.Nil || in.Qty <= 0 {
			return nil, 0, 0, 0, 0, apperrors.ErrValidation
		}
		lineSub := in.Qty * in.UnitPrice
		lineTotal := lineSub - in.Discount + in.Tax
		lines = append(lines, domain.OrderLine{
			ProductID: in.ProductID, Qty: in.Qty, UnitPrice: in.UnitPrice,
			Discount: in.Discount, Tax: in.Tax, LineTotal: lineTotal,
		})
		subtotal += lineSub
		discountTotal += in.Discount
		taxTotal += in.Tax
		grandTotal += lineTotal
	}
	return lines, subtotal, discountTotal, taxTotal, grandTotal, nil
}

func generateOrderNumber() string {
	var n uint32
	_ = binary.Read(rand.Reader, binary.BigEndian, &n)
	return fmt.Sprintf("ORD-%s-%04d", time.Now().UTC().Format("20060102"), n%10000)
}

func (s *Service) ListOrders(ctx context.Context, tenantID uuid.UUID, filters domain.OrderListFilters, page, perPage int) ([]OrderDTO, int64, error) {
	rows, total, err := s.orders.List(ctx, tenantID, filters, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]OrderDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toOrderDTO(r, nil))
	}
	return out, total, nil
}

func (s *Service) GetOrder(ctx context.Context, tenantID, id uuid.UUID) (*OrderDTO, []OrderStatusHistoryDTO, error) {
	o, lines, err := s.orders.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, nil, err
	}
	hist, err := s.orders.ListHistory(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	dto := toOrderDTO(*o, lines)
	hOut := make([]OrderStatusHistoryDTO, 0, len(hist))
	for _, h := range hist {
		hOut = append(hOut, toHistoryDTO(h))
	}
	return &dto, hOut, nil
}

func (s *Service) CreateOrder(ctx context.Context, tenantID uuid.UUID, in CreateOrderInput) (*OrderDTO, error) {
	if in.CustomerID == uuid.Nil {
		return nil, apperrors.ErrValidation
	}

	if in.ClientRequestID != nil {
		key := strings.TrimSpace(*in.ClientRequestID)
		if key != "" {
			existing, lines, err := s.orders.FindByClientRequestID(ctx, tenantID, key)
			if err == nil {
				dto := toOrderDTO(*existing, lines)
				return &dto, nil
			}
			if !errors.Is(err, apperrors.ErrNotFound) {
				return nil, err
			}
			in.ClientRequestID = &key
		} else {
			in.ClientRequestID = nil
		}
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

	lines, subtotal, discountTotal, taxTotal, grandTotal, err := buildLines(in.Lines)
	if err != nil {
		return nil, err
	}

	order := &domain.Order{
		TenantID: tenantID, Number: generateOrderNumber(), CustomerID: in.CustomerID,
		AgentID: in.AgentID, BranchID: in.BranchID, WarehouseID: in.WarehouseID, VisitID: in.VisitID,
		Status: status, Currency: currency, Subtotal: subtotal, DiscountTotal: discountTotal,
		TaxTotal: taxTotal, GrandTotal: grandTotal, OrderedAt: time.Now().UTC(),
		DeliveryDate: in.DeliveryDate, PriceListID: in.PriceListID, PromotionID: in.PromotionID,
		Comment: in.Comment, ClientRequestID: in.ClientRequestID,
	}

	if err := s.orders.Create(ctx, order, lines); err != nil {
		return nil, err
	}
	dto := toOrderDTO(*order, lines)
	s.record(ctx, tenantID, &dto)
	return &dto, nil
}

func (s *Service) UpdateDraft(ctx context.Context, tenantID, id uuid.UUID, in UpdateDraftInput) (*OrderDTO, error) {
	o, _, err := s.orders.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if o.Status != domain.StatusDraft {
		return nil, apperrors.ErrConflict
	}

	if in.CustomerID != nil {
		if *in.CustomerID == uuid.Nil {
			return nil, apperrors.ErrValidation
		}
		o.CustomerID = *in.CustomerID
	}
	if in.AgentID != nil {
		o.AgentID = in.AgentID
	}
	if in.BranchID != nil {
		o.BranchID = in.BranchID
	}
	if in.WarehouseID != nil {
		o.WarehouseID = in.WarehouseID
	}
	if in.VisitID != nil {
		o.VisitID = in.VisitID
	}
	if in.Currency != nil {
		cur := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if cur == "" {
			return nil, apperrors.ErrValidation
		}
		o.Currency = cur
	}
	if in.DeliveryDate != nil {
		o.DeliveryDate = in.DeliveryDate
	}
	if in.PriceListID != nil {
		o.PriceListID = in.PriceListID
	}
	if in.PromotionID != nil {
		o.PromotionID = in.PromotionID
	}
	if in.Comment != nil {
		o.Comment = in.Comment
	}

	var lines []domain.OrderLine
	if in.Lines != nil {
		built, subtotal, discountTotal, taxTotal, grandTotal, err := buildLines(in.Lines)
		if err != nil {
			return nil, err
		}
		lines = built
		o.Subtotal = subtotal
		o.DiscountTotal = discountTotal
		o.TaxTotal = taxTotal
		o.GrandTotal = grandTotal
		if err := s.orders.Update(ctx, o, lines); err != nil {
			return nil, err
		}
	} else {
		if err := s.orders.Update(ctx, o, nil); err != nil {
			return nil, err
		}
		_, lines, err = s.orders.FindByID(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
	}

	dto := toOrderDTO(*o, lines)
	s.record(ctx, tenantID, &dto)
	return &dto, nil
}

func (s *Service) transition(ctx context.Context, tenantID, id uuid.UUID, toStatus string, comment *string, changedBy *uuid.UUID) (*OrderDTO, error) {
	if !domain.ValidStatus(toStatus) {
		return nil, apperrors.ErrValidation
	}
	o, lines, err := s.orders.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if !o.CanTransition(toStatus) {
		return nil, apperrors.ErrConflict
	}
	from := o.Status
	o.Status = toStatus
	if err := s.orders.Update(ctx, o, nil); err != nil {
		return nil, err
	}
	h := &domain.OrderStatusHistory{
		OrderID: o.ID, FromStatus: from, ToStatus: toStatus, ChangedBy: changedBy, Comment: comment,
	}
	if err := s.orders.AddHistory(ctx, h); err != nil {
		return nil, err
	}
	dto := toOrderDTO(*o, lines)
	s.record(ctx, tenantID, &dto)
	return &dto, nil
}

func (s *Service) Submit(ctx context.Context, tenantID, id uuid.UUID, changedBy *uuid.UUID) (*OrderDTO, error) {
	return s.transition(ctx, tenantID, id, domain.StatusSubmitted, nil, changedBy)
}

func (s *Service) Confirm(ctx context.Context, tenantID, id uuid.UUID, changedBy *uuid.UUID) (*OrderDTO, error) {
	return s.transition(ctx, tenantID, id, domain.StatusConfirmed, nil, changedBy)
}

func (s *Service) Cancel(ctx context.Context, tenantID, id uuid.UUID, comment *string, changedBy *uuid.UUID) (*OrderDTO, error) {
	return s.transition(ctx, tenantID, id, domain.StatusCancelled, comment, changedBy)
}

func (s *Service) TransitionStatus(ctx context.Context, tenantID, id uuid.UUID, in TransitionStatusInput, changedBy *uuid.UUID) (*OrderDTO, error) {
	status := strings.TrimSpace(in.Status)
	if status == "" {
		return nil, apperrors.ErrValidation
	}
	return s.transition(ctx, tenantID, id, status, in.Comment, changedBy)
}

func (s *Service) ListHistory(ctx context.Context, tenantID, id uuid.UUID) ([]OrderStatusHistoryDTO, error) {
	if _, _, err := s.orders.FindByID(ctx, tenantID, id); err != nil {
		return nil, err
	}
	hist, err := s.orders.ListHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]OrderStatusHistoryDTO, 0, len(hist))
	for _, h := range hist {
		out = append(out, toHistoryDTO(h))
	}
	return out, nil
}
