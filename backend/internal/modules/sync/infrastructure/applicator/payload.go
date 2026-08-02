package applicator

import (
	"fmt"
	"strings"

	crmdomain "github.com/Dovud1997/Dovud/backend/internal/modules/crm/domain"
	ffdomain "github.com/Dovud1997/Dovud/backend/internal/modules/fieldforce/domain"
	ordersdomain "github.com/Dovud1997/Dovud/backend/internal/modules/orders/domain"
	returnsdomain "github.com/Dovud1997/Dovud/backend/internal/modules/returns/domain"
	"github.com/google/uuid"
)

func strOr(m map[string]any, key, fallback string) string {
	if m == nil {
		return fallback
	}
	v, ok := m[key]
	if !ok || v == nil {
		return fallback
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "" || s == "<nil>" {
		return fallback
	}
	return s
}

func strPtrOr(m map[string]any, key string) *string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "" || s == "<nil>" {
		empty := ""
		return &empty
	}
	return &s
}

func floatOr(m map[string]any, key string, fallback float64) float64 {
	if m == nil {
		return fallback
	}
	v, ok := m[key]
	if !ok || v == nil {
		return fallback
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		var f float64
		_, err := fmt.Sscan(fmt.Sprint(v), &f)
		if err != nil {
			return fallback
		}
		return f
	}
}

func floatPtrOr(m map[string]any, key string) *float64 {
	if m == nil {
		return nil
	}
	if _, ok := m[key]; !ok || m[key] == nil {
		return nil
	}
	f := floatOr(m, key, 0)
	return &f
}

func uuidPtrOr(m map[string]any, key string) *uuid.UUID {
	s := strOr(m, key, "")
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func parseOrderLines(payload map[string]any) []ordersdomain.OrderLine {
	raw, ok := payload["lines"].([]any)
	if !ok {
		return nil
	}
	out := make([]ordersdomain.OrderLine, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		pid, err := uuid.Parse(strOr(m, "product_id", ""))
		if err != nil {
			continue
		}
		qty := floatOr(m, "qty", 0)
		unit := floatOr(m, "unit_price", 0)
		disc := floatOr(m, "discount", 0)
		tax := floatOr(m, "tax", 0)
		total := qty*unit - disc + tax
		if m["line_total"] != nil {
			total = floatOr(m, "line_total", total)
		}
		out = append(out, ordersdomain.OrderLine{
			ProductID: pid, Qty: qty, UnitPrice: unit, Discount: disc, Tax: tax, LineTotal: total,
		})
	}
	return out
}

func sumLines(lines []ordersdomain.OrderLine) (sub, disc, tax, grand float64) {
	for _, l := range lines {
		sub += l.Qty * l.UnitPrice
		disc += l.Discount
		tax += l.Tax
		grand += l.LineTotal
	}
	return
}

func customerPayload(c *crmdomain.Customer) map[string]any {
	return map[string]any{
		"id": c.ID.String(), "code": c.Code, "name": c.Name, "type": c.Type, "status": c.Status,
		"credit_limit": c.CreditLimit, "inn": c.Inn, "address": c.Address, "lat": c.Lat, "lng": c.Lng,
		"branch_id": c.BranchID, "version": c.Version,
	}
}

func orderPayload(o *ordersdomain.Order, lines []ordersdomain.OrderLine) map[string]any {
	lineMaps := make([]map[string]any, 0, len(lines))
	for _, l := range lines {
		lineMaps = append(lineMaps, map[string]any{
			"id": l.ID.String(), "product_id": l.ProductID.String(), "qty": l.Qty,
			"unit_price": l.UnitPrice, "discount": l.Discount, "tax": l.Tax, "line_total": l.LineTotal,
		})
	}
	return map[string]any{
		"id": o.ID.String(), "number": o.Number, "customer_id": o.CustomerID.String(),
		"status": o.Status, "currency": o.Currency, "subtotal": o.Subtotal,
		"discount_total": o.DiscountTotal, "tax_total": o.TaxTotal, "grand_total": o.GrandTotal,
		"comment": o.Comment, "client_request_id": o.ClientRequestID, "version": o.Version,
		"lines": lineMaps,
	}
}

func visitPayload(v *ffdomain.Visit) map[string]any {
	return map[string]any{
		"id": v.ID.String(), "agent_id": v.AgentID.String(), "customer_id": v.CustomerID.String(),
		"route_stop_id": v.RouteStopID, "started_at": v.StartedAt, "ended_at": v.EndedAt,
		"checkin_lat": v.CheckinLat, "checkin_lng": v.CheckinLng,
		"checkout_lat": v.CheckoutLat, "checkout_lng": v.CheckoutLng,
		"result": v.Result, "notes": v.Notes, "version": v.Version,
	}
}

func parseReturnLines(payload map[string]any) []returnsdomain.ReturnLine {
	raw, ok := payload["lines"].([]any)
	if !ok {
		return nil
	}
	out := make([]returnsdomain.ReturnLine, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		pid, err := uuid.Parse(strOr(m, "product_id", ""))
		if err != nil {
			continue
		}
		qty := floatOr(m, "qty", 0)
		unit := floatOr(m, "unit_price", 0)
		total := qty * unit
		if m["line_total"] != nil {
			total = floatOr(m, "line_total", total)
		}
		out = append(out, returnsdomain.ReturnLine{
			ProductID: pid, Qty: qty, UnitPrice: unit, LineTotal: total, Reason: strPtrOr(m, "reason"),
		})
	}
	return out
}

func returnPayload(r *returnsdomain.Return, lines []returnsdomain.ReturnLine) map[string]any {
	lineMaps := make([]map[string]any, 0, len(lines))
	for _, l := range lines {
		lineMaps = append(lineMaps, map[string]any{
			"id": l.ID.String(), "product_id": l.ProductID.String(), "qty": l.Qty,
			"unit_price": l.UnitPrice, "line_total": l.LineTotal, "reason": l.Reason,
		})
	}
	return map[string]any{
		"id": r.ID.String(), "number": r.Number, "customer_id": r.CustomerID.String(),
		"order_id": r.OrderID, "agent_id": r.AgentID, "status": r.Status, "reason": r.Reason,
		"currency": r.Currency, "subtotal": r.Subtotal, "tax_total": r.TaxTotal,
		"grand_total": r.GrandTotal, "version": r.Version, "lines": lineMaps,
	}
}
