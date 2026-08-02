package applicator

import (
	"context"
	"fmt"
	"strings"
	"time"

	crmdomain "github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	ffdomain "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/domain"
	ordersdomain "github.com/Dovud1997/Dovud/backend/internal/modules/orders/domain"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/Dovud1997/Dovud/backend/internal/platform/syncport"
	"github.com/google/uuid"
)

// DomainApplicator mutates CRM / orders / visits from sync push ops.
type DomainApplicator struct {
	customers crmdomain.CustomerRepository
	orders    ordersdomain.OrderRepository
	visits    ffdomain.VisitRepository
}

func New(
	customers crmdomain.CustomerRepository,
	orders ordersdomain.OrderRepository,
	visits ffdomain.VisitRepository,
) *DomainApplicator {
	return &DomainApplicator{customers: customers, orders: orders, visits: visits}
}

func (a *DomainApplicator) Supports(entityType string) bool {
	switch strings.TrimSpace(strings.ToLower(entityType)) {
	case "customer", "order", "visit":
		return true
	default:
		return false
	}
}

func (a *DomainApplicator) Apply(ctx context.Context, req syncport.ApplyRequest) (*syncport.ApplyResult, error) {
	typ := strings.TrimSpace(strings.ToLower(req.EntityType))
	op := strings.TrimSpace(strings.ToLower(req.Op))
	id, err := uuid.Parse(strings.TrimSpace(req.EntityID))
	if err != nil {
		return nil, apperrors.ErrValidation
	}
	switch typ {
	case "customer":
		return a.applyCustomer(ctx, req.TenantID, id, op, req.Payload)
	case "order":
		return a.applyOrder(ctx, req.TenantID, id, op, req.Payload)
	case "visit":
		return a.applyVisit(ctx, req.TenantID, id, op, req.Payload)
	default:
		return nil, apperrors.ErrValidation
	}
}

func (a *DomainApplicator) applyCustomer(ctx context.Context, tenantID, id uuid.UUID, op string, payload map[string]any) (*syncport.ApplyResult, error) {
	if a.customers == nil {
		return nil, apperrors.ErrValidation
	}
	switch op {
	case "create":
		c := &crmdomain.Customer{
			ID: id, TenantID: tenantID,
			Code: strings.ToUpper(strOr(payload, "code", id.String()[:8])),
			Name: strOr(payload, "name", "Synced customer"),
			Type: strOr(payload, "type", "outlet"),
			Status: strOr(payload, "status", "active"),
			CreditLimit: floatOr(payload, "credit_limit", 0),
			Inn: strPtrOr(payload, "inn"),
			Address: strPtrOr(payload, "address"),
			Lat: floatPtrOr(payload, "lat"),
			Lng: floatPtrOr(payload, "lng"),
			BranchID: uuidPtrOr(payload, "branch_id"),
		}
		if err := a.customers.Create(ctx, c); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: customerPayload(c), Version: c.Version}, nil
	case "update":
		c, err := a.customers.FindByID(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
		if v := strOr(payload, "code", ""); v != "" {
			c.Code = strings.ToUpper(v)
		}
		if v := strOr(payload, "name", ""); v != "" {
			c.Name = v
		}
		if v := strOr(payload, "type", ""); v != "" {
			c.Type = v
		}
		if v := strOr(payload, "status", ""); v != "" {
			c.Status = v
		}
		if payload["credit_limit"] != nil {
			c.CreditLimit = floatOr(payload, "credit_limit", c.CreditLimit)
		}
		if payload["inn"] != nil {
			c.Inn = strPtrOr(payload, "inn")
		}
		if payload["address"] != nil {
			c.Address = strPtrOr(payload, "address")
		}
		if payload["lat"] != nil {
			c.Lat = floatPtrOr(payload, "lat")
		}
		if payload["lng"] != nil {
			c.Lng = floatPtrOr(payload, "lng")
		}
		if payload["branch_id"] != nil {
			c.BranchID = uuidPtrOr(payload, "branch_id")
		}
		if err := a.customers.Update(ctx, c); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: customerPayload(c), Version: c.Version}, nil
	case "delete":
		if err := a.customers.SoftDelete(ctx, tenantID, id); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: map[string]any{"id": id.String()}, Version: 0, Deleted: true}, nil
	default:
		return nil, apperrors.ErrValidation
	}
}

func (a *DomainApplicator) applyOrder(ctx context.Context, tenantID, id uuid.UUID, op string, payload map[string]any) (*syncport.ApplyResult, error) {
	if a.orders == nil {
		return nil, apperrors.ErrValidation
	}
	switch op {
	case "create":
		customerID, err := uuid.Parse(strOr(payload, "customer_id", ""))
		if err != nil {
			return nil, apperrors.ErrValidation
		}
		lines := parseOrderLines(payload)
		sub, disc, tax, grand := sumLines(lines)
		o := &ordersdomain.Order{
			ID: id, TenantID: tenantID, Number: strOr(payload, "number", fmt.Sprintf("SYNC-%s", id.String()[:8])),
			CustomerID: customerID, Status: strOr(payload, "status", ordersdomain.StatusDraft),
			Currency: strOr(payload, "currency", "UZS"),
			Subtotal: sub, DiscountTotal: disc, TaxTotal: tax, GrandTotal: grand,
			Comment: strPtrOr(payload, "comment"),
			ClientRequestID: strPtrOr(payload, "client_request_id"),
			AgentID: uuidPtrOr(payload, "agent_id"),
			BranchID: uuidPtrOr(payload, "branch_id"),
			WarehouseID: uuidPtrOr(payload, "warehouse_id"),
			VisitID: uuidPtrOr(payload, "visit_id"),
		}
		if err := a.orders.Create(ctx, o, lines); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: orderPayload(o, lines), Version: o.Version}, nil
	case "update":
		o, lines, err := a.orders.FindByID(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
		if o.Status != ordersdomain.StatusDraft {
			return nil, apperrors.ErrConflict
		}
		if v := strOr(payload, "currency", ""); v != "" {
			o.Currency = strings.ToUpper(v)
		}
		if payload["comment"] != nil {
			o.Comment = strPtrOr(payload, "comment")
		}
		if payload["customer_id"] != nil {
			if cid, err := uuid.Parse(strOr(payload, "customer_id", "")); err == nil {
				o.CustomerID = cid
			}
		}
		newLines := lines
		if payload["lines"] != nil {
			newLines = parseOrderLines(payload)
			o.Subtotal, o.DiscountTotal, o.TaxTotal, o.GrandTotal = sumLines(newLines)
		}
		if err := a.orders.Update(ctx, o, newLines); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: orderPayload(o, newLines), Version: o.Version}, nil
	case "delete":
		if err := a.orders.SoftDelete(ctx, tenantID, id); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: map[string]any{"id": id.String()}, Version: 0, Deleted: true}, nil
	default:
		return nil, apperrors.ErrValidation
	}
}

func (a *DomainApplicator) applyVisit(ctx context.Context, tenantID, id uuid.UUID, op string, payload map[string]any) (*syncport.ApplyResult, error) {
	if a.visits == nil {
		return nil, apperrors.ErrValidation
	}
	switch op {
	case "create":
		agentID, err := uuid.Parse(strOr(payload, "agent_id", ""))
		if err != nil {
			return nil, apperrors.ErrValidation
		}
		customerID, err := uuid.Parse(strOr(payload, "customer_id", ""))
		if err != nil {
			return nil, apperrors.ErrValidation
		}
		v := &ffdomain.Visit{
			ID: id, TenantID: tenantID, AgentID: agentID, CustomerID: customerID,
			RouteStopID: uuidPtrOr(payload, "route_stop_id"),
			StartedAt: time.Now().UTC(),
			CheckinLat: floatPtrOr(payload, "checkin_lat"),
			CheckinLng: floatPtrOr(payload, "checkin_lng"),
			Notes: strPtrOr(payload, "notes"),
			Result: strOr(payload, "result", ""),
		}
		if err := a.visits.Create(ctx, v); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: visitPayload(v), Version: v.Version}, nil
	case "update":
		v, err := a.visits.FindByID(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
		if payload["notes"] != nil {
			v.Notes = strPtrOr(payload, "notes")
		}
		if payload["result"] != nil {
			v.Result = strOr(payload, "result", v.Result)
		}
		if payload["checkout_lat"] != nil {
			v.CheckoutLat = floatPtrOr(payload, "checkout_lat")
		}
		if payload["checkout_lng"] != nil {
			v.CheckoutLng = floatPtrOr(payload, "checkout_lng")
		}
		if payload["ended_at"] != nil || v.Result != "" {
			if v.EndedAt == nil {
				now := time.Now().UTC()
				v.EndedAt = &now
			}
		}
		if err := a.visits.Update(ctx, v); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: visitPayload(v), Version: v.Version}, nil
	case "delete":
		// Visits are not soft-deleted; mark ended cancelled-style in payload only.
		v, err := a.visits.FindByID(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
		now := time.Now().UTC()
		v.EndedAt = &now
		v.Result = "cancelled"
		if err := a.visits.Update(ctx, v); err != nil {
			return nil, err
		}
		return &syncport.ApplyResult{Payload: visitPayload(v), Version: v.Version, Deleted: true}, nil
	default:
		return nil, apperrors.ErrValidation
	}
}
