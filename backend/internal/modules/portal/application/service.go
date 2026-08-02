package application

import (
	"context"
	"time"

	crmdomain "github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	docsdomain "github.com/Dovud1997/Dovud/backend/internal/modules/documents/domain"
	financedomain "github.com/Dovud1997/Dovud/backend/internal/modules/finance/domain"
	identitydomain "github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	ordersdomain "github.com/Dovud1997/Dovud/backend/internal/modules/orders/domain"
	"github.com/Dovud1997/Dovud/backend/internal/modules/portal/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/google/uuid"
)

type Service struct {
	links       domain.CustomerUserRepository
	customers   crmdomain.CustomerRepository
	orders      ordersdomain.OrderRepository
	receivables financedomain.ReceivableRepository
	documents   docsdomain.DocumentRepository
	users       identitydomain.UserRepository
}

func NewService(
	links domain.CustomerUserRepository,
	customers crmdomain.CustomerRepository,
	orders ordersdomain.OrderRepository,
	receivables financedomain.ReceivableRepository,
	documents docsdomain.DocumentRepository,
	users identitydomain.UserRepository,
) *Service {
	return &Service{
		links: links, customers: customers, orders: orders,
		receivables: receivables, documents: documents, users: users,
	}
}

type OrderDTO struct {
	ID         uuid.UUID `json:"id"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Currency   string    `json:"currency"`
	GrandTotal float64   `json:"grand_total"`
	OrderedAt  time.Time `json:"ordered_at"`
}

type ReceivableDTO struct {
	ID       uuid.UUID  `json:"id"`
	Amount   float64    `json:"amount"`
	Balance  float64    `json:"balance"`
	Status   string     `json:"status"`
	Currency string     `json:"currency"`
	DueDate  *time.Time `json:"due_date,omitempty"`
}

type DocumentDTO struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	DocType   string    `json:"doc_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type LinkDTO struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type LinkInput struct {
	UserID     uuid.UUID `json:"user_id"`
	CustomerID uuid.UUID `json:"customer_id"`
}

func (s *Service) resolveCustomer(ctx context.Context, tenantID, userID uuid.UUID) (*crmdomain.Customer, error) {
	link, err := s.links.FindByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, apperrors.New("PORTAL_NOT_LINKED", "User is not linked to a customer", 403)
	}
	return s.customers.FindByID(ctx, tenantID, link.CustomerID)
}

func (s *Service) Summary(ctx context.Context, tenantID, userID uuid.UUID) (*domain.Summary, error) {
	cust, err := s.resolveCustomer(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	orders, _, err := s.orders.List(ctx, tenantID, ordersdomain.OrderListFilters{CustomerID: &cust.ID}, 1, 100)
	if err != nil {
		return nil, err
	}
	var openOrders int64
	for _, o := range orders {
		switch o.Status {
		case ordersdomain.StatusDelivered, ordersdomain.StatusCancelled:
		default:
			openOrders++
		}
	}
	recs, _, err := s.receivables.List(ctx, tenantID, financedomain.ReceivableListFilters{CustomerID: &cust.ID}, 1, 200)
	if err != nil {
		return nil, err
	}
	var balance float64
	for _, r := range recs {
		if r.Status != financedomain.StatusClosed {
			balance += r.Balance
		}
	}
	_, totalDocs, err := s.documents.ListByCustomer(ctx, tenantID, cust.ID, 1, 1)
	if err != nil {
		totalDocs = 0
	}
	return &domain.Summary{
		CustomerID: cust.ID, CustomerCode: cust.Code, CustomerName: cust.Name,
		OpenOrders: openOrders, OpenBalance: balance, CreditLimit: cust.CreditLimit,
		DocumentsCount: totalDocs,
	}, nil
}

func (s *Service) Orders(ctx context.Context, tenantID, userID uuid.UUID, page, perPage int) ([]OrderDTO, int64, error) {
	cust, err := s.resolveCustomer(ctx, tenantID, userID)
	if err != nil {
		return nil, 0, err
	}
	rows, total, err := s.orders.List(ctx, tenantID, ordersdomain.OrderListFilters{CustomerID: &cust.ID}, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]OrderDTO, 0, len(rows))
	for _, o := range rows {
		out = append(out, OrderDTO{
			ID: o.ID, Number: o.Number, Status: o.Status, Currency: o.Currency,
			GrandTotal: o.GrandTotal, OrderedAt: o.OrderedAt,
		})
	}
	return out, total, nil
}

func (s *Service) Receivables(ctx context.Context, tenantID, userID uuid.UUID, page, perPage int) ([]ReceivableDTO, int64, error) {
	cust, err := s.resolveCustomer(ctx, tenantID, userID)
	if err != nil {
		return nil, 0, err
	}
	rows, total, err := s.receivables.List(ctx, tenantID, financedomain.ReceivableListFilters{CustomerID: &cust.ID}, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ReceivableDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, ReceivableDTO{
			ID: r.ID, Amount: r.Amount, Balance: r.Balance, Status: r.Status,
			Currency: r.Currency, DueDate: r.DueDate,
		})
	}
	return out, total, nil
}

func (s *Service) Documents(ctx context.Context, tenantID, userID uuid.UUID, page, perPage int) ([]DocumentDTO, int64, error) {
	cust, err := s.resolveCustomer(ctx, tenantID, userID)
	if err != nil {
		return nil, 0, err
	}
	rows, total, err := s.documents.ListByCustomer(ctx, tenantID, cust.ID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]DocumentDTO, 0, len(rows))
	for _, d := range rows {
		out = append(out, DocumentDTO{
			ID: d.ID, Title: d.Title, DocType: d.DocType, Status: d.Status, CreatedAt: d.CreatedAt,
		})
	}
	return out, total, nil
}

func (s *Service) LinkUser(ctx context.Context, tenantID uuid.UUID, in LinkInput) (*LinkDTO, error) {
	if in.UserID == uuid.Nil || in.CustomerID == uuid.Nil {
		return nil, apperrors.ErrValidation
	}
	if _, err := s.users.FindByID(ctx, tenantID, in.UserID); err != nil {
		return nil, apperrors.ErrNotFound
	}
	if _, err := s.customers.FindByID(ctx, tenantID, in.CustomerID); err != nil {
		return nil, apperrors.ErrNotFound
	}
	link := &domain.CustomerUser{
		TenantID: tenantID, UserID: in.UserID, CustomerID: in.CustomerID,
	}
	if err := s.links.Upsert(ctx, link); err != nil {
		return nil, err
	}
	stored, err := s.links.FindByUser(ctx, tenantID, in.UserID)
	if err != nil {
		return nil, err
	}
	return &LinkDTO{
		ID: stored.ID, UserID: stored.UserID, CustomerID: stored.CustomerID, CreatedAt: stored.CreatedAt,
	}, nil
}

func (s *Service) UnlinkUser(ctx context.Context, tenantID, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return apperrors.ErrValidation
	}
	return s.links.DeleteByUser(ctx, tenantID, userID)
}

func (s *Service) ListLinks(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]LinkDTO, error) {
	rows, err := s.links.List(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}
	out := make([]LinkDTO, 0, len(rows))
	for _, l := range rows {
		out = append(out, LinkDTO{
			ID: l.ID, UserID: l.UserID, CustomerID: l.CustomerID, CreatedAt: l.CreatedAt,
		})
	}
	return out, nil
}
